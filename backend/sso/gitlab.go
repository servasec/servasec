package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

type GitLabProvider struct {
	config *oauth2.Config
	baseURL string
}

func NewGitLabProvider() *GitLabProvider {
	clientID := os.Getenv("SSO_GITLAB_CLIENT_ID")
	clientSecret := os.Getenv("SSO_GITLAB_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return &GitLabProvider{}
	}
	baseURL := os.Getenv("SSO_GITLAB_BASE_URL")
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	return &GitLabProvider{
		baseURL: baseURL,
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  baseURL + "/oauth/authorize",
				TokenURL: baseURL + "/oauth/token",
			},
			RedirectURL: PublicURL() + "/api/auth/gitlab/callback",
			Scopes:      []string{"read_user", "openid"},
		},
	}
}

func (p *GitLabProvider) Name() string { return "gitlab" }

func (p *GitLabProvider) IsEnabled() bool {
	return p.config != nil
}

func (p *GitLabProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *GitLabProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

type gitlabUser struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func (p *GitLabProvider) UserInfo(ctx context.Context, token *oauth2.Token) (*SSOUser, error) {
	client := p.config.Client(ctx, token)

	resp, err := client.Get(p.baseURL + "/api/v4/user")
	if err != nil {
		return nil, fmt.Errorf("gitlab user info: %w", err)
	}
	defer resp.Body.Close()

	var gu gitlabUser
	if err := json.NewDecoder(resp.Body).Decode(&gu); err != nil {
		return nil, fmt.Errorf("gitlab decode user: %w", err)
	}

	return &SSOUser{
		Provider:   "gitlab",
		ProviderID: fmt.Sprint(gu.ID),
		Email:      gu.Email,
		Username:   gu.Username,
		AvatarURL:  gu.AvatarURL,
	}, nil
}
