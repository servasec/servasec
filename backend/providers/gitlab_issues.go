package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GitLabIssueTracker struct{}

func (g *GitLabIssueTracker) Name() string {
	return "gitlab"
}

func (g *GitLabIssueTracker) baseURL(cfg IssueTrackerConfig) string {
	if strings.Contains(cfg.RepositoryURL, "/-/") {
		parts := strings.SplitN(cfg.RepositoryURL, "/-/commit", 2)
		return parts[0]
	}
	if strings.HasPrefix(cfg.RepositoryURL, "https://") || strings.HasPrefix(cfg.RepositoryURL, "http://") {
		u, err := url.Parse(cfg.RepositoryURL)
		if err == nil {
			return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		}
	}
	return "https://gitlab.com"
}

func (g *GitLabIssueTracker) projectPath(cfg IssueTrackerConfig) string {
	repoURL := cfg.RepositoryURL
	repoURL = strings.TrimSuffix(repoURL, ".git")
	repoURL = strings.TrimSuffix(repoURL, "/")

	if strings.HasPrefix(repoURL, "https://") || strings.HasPrefix(repoURL, "http://") {
		u, err := url.Parse(repoURL)
		if err == nil {
			return strings.TrimPrefix(u.Path, "/")
		}
	}

	return strings.TrimPrefix(repoURL, "/")
}

func (g *GitLabIssueTracker) CreateIssue(cfg IssueTrackerConfig, finding IssueData) (*ExternalIssue, error) {
	labels := "servasec,security"
	if sevLabel, ok := SeverityLabels[finding.Severity]; ok {
		labels += "," + sevLabel
	}

	body := FormatFindingBody(cfg, finding)

	payload := map[string]interface{}{
		"title":       TruncateTitle(finding.Title, 255),
		"description": body,
		"labels":      labels,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	baseURL := g.baseURL(cfg)
	projectPath := url.PathEscape(g.projectPath(cfg))
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/issues", baseURL, projectPath)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}

		req, err := http.NewRequest("POST", apiURL, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("PRIVATE-TOKEN", cfg.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 429 {
			lastErr = fmt.Errorf("rate limited")
			continue
		}

		if resp.StatusCode != 201 {
			lastErr = fmt.Errorf("GitLab API returned status %d: %s", resp.StatusCode, string(bodyBytes))
			return nil, lastErr
		}

		var result struct {
			IID   int    `json:"iid"`
			WebURL string `json:"web_url"`
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &ExternalIssue{
			IssueID:  fmt.Sprintf("%d", result.IID),
			IssueURL: result.WebURL,
		}, nil
	}

	return nil, fmt.Errorf("failed after retries: %w", lastErr)
}

func (g *GitLabIssueTracker) TestConnection(cfg IssueTrackerConfig) error {
	baseURL := g.baseURL(cfg)
	projectPath := url.PathEscape(g.projectPath(cfg))
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s", baseURL, projectPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", cfg.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid token")
	}
	if resp.StatusCode == 404 {
		return fmt.Errorf("project not found")
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
