package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/servasec/servasec/backend/debug"
	"golang.org/x/oauth2"
)

type GitHubProvider struct {
	config *oauth2.Config
}

func NewGitHubProvider() *GitHubProvider {
	clientID := os.Getenv("SSO_GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("SSO_GITHUB_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return &GitHubProvider{}
	}
	return &GitHubProvider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
			RedirectURL: PublicURL() + "/api/auth/github/callback",
			Scopes:      []string{"read:user", "user:email"},
		},
	}
}

func (p *GitHubProvider) Name() string { return "github" }

func (p *GitHubProvider) IsEnabled() bool {
	return p.config != nil
}

func (p *GitHubProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *GitHubProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

type githubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (p *GitHubProvider) UserInfo(ctx context.Context, token *oauth2.Token) (*SSOUser, error) {
	client := p.config.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("github user info: %w", err)
	}
	defer resp.Body.Close()

	var gu githubUser
	if err := json.NewDecoder(resp.Body).Decode(&gu); err != nil {
		return nil, fmt.Errorf("github decode user: %w", err)
	}

	email := gu.Email
	if email == "" {
		var err error
		email, err = fetchPrimaryGitHubEmail(client)
		if err != nil {
			debug.Log("SSO github: fetchPrimaryGitHubEmail error: %v", err)
		}
	}

	return &SSOUser{
		Provider:   "github",
		ProviderID: fmt.Sprint(gu.ID),
		Email:      email,
		Username:   gu.Login,
		AvatarURL:  gu.AvatarURL,
	}, nil
}

func fetchPrimaryGitHubEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no primary verified email")
}
