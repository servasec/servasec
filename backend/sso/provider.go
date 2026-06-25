package sso

import (
	"context"
	"os"

	"golang.org/x/oauth2"
	"github.com/servasec/servasec/backend/features"
)

type SSOUser struct {
	Provider   string
	ProviderID string
	Email      string
	Username   string
	AvatarURL  string
}

type Provider interface {
	Name() string
	IsEnabled() bool
	AuthURL(state string) string
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	UserInfo(ctx context.Context, token *oauth2.Token) (*SSOUser, error)
}

type ProviderInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func PublicURL() string {
	u := os.Getenv("SSC_PUBLIC_URL")
	if u == "" {
		u = "https://servasec.local"
	}
	return u
}

func EnabledProviders() []Provider {
	if !features.F.IsEnabled(features.FeatureSSO) {
		return nil
	}
	var providers []Provider
	if p := NewGitHubProvider(); p.IsEnabled() {
		providers = append(providers, p)
	}
	if p := NewGitLabProvider(); p.IsEnabled() {
		providers = append(providers, p)
	}
	if p := NewOIDCProvider(); p.IsEnabled() {
		providers = append(providers, p)
	}
	return providers
}

func EnabledProviderInfos() []ProviderInfo {
	var infos []ProviderInfo
	for _, p := range EnabledProviders() {
		infos = append(infos, ProviderInfo{
			Name: p.Name(),
			URL:  PublicURL() + "/api/auth/" + p.Name(),
		})
	}
	return infos
}

func GetProvider(name string) Provider {
	for _, p := range EnabledProviders() {
		if p.Name() == name {
			return p
		}
	}
	return nil
}
