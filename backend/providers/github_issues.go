package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type GitHubIssueTracker struct{}

type githubAppTokenCache struct {
	token     string
	expiresAt time.Time
}

var (
	githubAppCache   = make(map[string]*githubAppTokenCache)
	githubAppCacheMu sync.Mutex
)

func (g *GitHubIssueTracker) Name() string {
	return "github"
}

func (g *GitHubIssueTracker) getAuthToken(cfg IssueTrackerConfig) (string, error) {
	if cfg.AuthType != "github_app" {
		return cfg.Token, nil
	}

	if cfg.GitHubAppID == nil || cfg.GitHubInstallationID == nil || cfg.GitHubPrivateKey == "" {
		return "", fmt.Errorf("GitHub App configuration incomplete")
	}

	cacheKey := fmt.Sprintf("%d:%d", *cfg.GitHubAppID, *cfg.GitHubInstallationID)

	githubAppCacheMu.Lock()
	if cached, ok := githubAppCache[cacheKey]; ok {
		if time.Now().Before(cached.expiresAt.Add(-5 * time.Minute)) {
			token := cached.token
			githubAppCacheMu.Unlock()
			return token, nil
		}
	}
	githubAppCacheMu.Unlock()

	token, expiresAt, err := g.fetchInstallationToken(*cfg.GitHubAppID, cfg.GitHubPrivateKey, *cfg.GitHubInstallationID)
	if err != nil {
		return "", fmt.Errorf("failed to get installation token: %w", err)
	}

	githubAppCacheMu.Lock()
	githubAppCache[cacheKey] = &githubAppTokenCache{token: token, expiresAt: expiresAt}
	githubAppCacheMu.Unlock()

	return token, nil
}

func (g *GitHubIssueTracker) fetchInstallationToken(appID int64, privateKeyPEM string, installationID int64) (string, time.Time, error) {
	jwtToken, err := g.generateJWT(appID, privateKeyPEM)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate JWT: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 201 {
		return "", time.Time{}, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Token, result.ExpiresAt, nil
}

func (g *GitHubIssueTracker) generateJWT(appID int64, privateKeyPEM string) (string, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": fmt.Sprintf("%d", appID),
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}

func (g *GitHubIssueTracker) CreateIssue(cfg IssueTrackerConfig, finding IssueData) (*ExternalIssue, error) {
	token, err := g.getAuthToken(cfg)
	if err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}

	labels := []string{"servasec", "security"}
	if sevLabel, ok := SeverityLabels[finding.Severity]; ok {
		labels = append(labels, sevLabel)
	}

	body := FormatFindingBody(cfg, finding)

	payload := map[string]interface{}{
		"title":  TruncateTitle(finding.Title, 256),
		"body":   body,
		"labels": labels,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	repoPath := strings.TrimPrefix(cfg.RepositoryURL, "https://github.com/")
	repoPath = strings.TrimPrefix(repoPath, "http://github.com/")
	repoPath = strings.TrimSuffix(repoPath, ".git")
	repoPath = strings.TrimSuffix(repoPath, "/")

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues", repoPath)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 401 && cfg.AuthType == "github_app" {
			githubAppCacheMu.Lock()
			cacheKey := fmt.Sprintf("%d:%d", *cfg.GitHubAppID, *cfg.GitHubInstallationID)
			delete(githubAppCache, cacheKey)
			githubAppCacheMu.Unlock()

			token, err = g.getAuthToken(cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to refresh token: %w", err)
			}
			continue
		}

		if resp.StatusCode == 403 || resp.StatusCode == 429 {
			lastErr = fmt.Errorf("rate limited (status %d)", resp.StatusCode)
			continue
		}

		if resp.StatusCode != 201 {
			lastErr = fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(bodyBytes))
			return nil, lastErr
		}

		var result struct {
			Number int    `json:"number"`
			HTMLURL string `json:"html_url"`
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &ExternalIssue{
			IssueID:  fmt.Sprintf("%d", result.Number),
			IssueURL: result.HTMLURL,
		}, nil
	}

	return nil, fmt.Errorf("failed after retries: %w", lastErr)
}

func (g *GitHubIssueTracker) TestConnection(cfg IssueTrackerConfig) error {
	token, err := g.getAuthToken(cfg)
	if err != nil {
		return fmt.Errorf("auth failed: %w", err)
	}

	repoPath := strings.TrimPrefix(cfg.RepositoryURL, "https://github.com/")
	repoPath = strings.TrimPrefix(repoPath, "http://github.com/")
	repoPath = strings.TrimSuffix(repoPath, ".git")
	repoPath = strings.TrimSuffix(repoPath, "/")

	url := fmt.Sprintf("https://api.github.com/repos/%s", repoPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid token")
	}
	if resp.StatusCode == 404 {
		return fmt.Errorf("repository not found")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
