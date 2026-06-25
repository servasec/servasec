package sso

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCProvider struct {
	config  *oauth2.Config
	verifier *oidc.IDTokenVerifier
	provider *oidc.Provider
}

func NewOIDCProvider() *OIDCProvider {
	clientID := os.Getenv("SSO_OIDC_CLIENT_ID")
	clientSecret := os.Getenv("SSO_OIDC_CLIENT_SECRET")
	issuerURL := os.Getenv("SSO_OIDC_ISSUER_URL")
	if clientID == "" || clientSecret == "" || issuerURL == "" {
		return &OIDCProvider{}
	}

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return &OIDCProvider{}
	}

	scopes := os.Getenv("SSO_OIDC_SCOPES")
	if scopes == "" {
		scopes = "openid profile email"
	}

	oauthCfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  PublicURL() + "/api/auth/oidc/callback",
		Scopes:       strings.Split(scopes, " "),
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCProvider{
		config:   oauthCfg,
		verifier: verifier,
		provider: provider,
	}
}

func (p *OIDCProvider) Name() string { return "oidc" }

func (p *OIDCProvider) IsEnabled() bool {
	return p.config != nil
}

func (p *OIDCProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *OIDCProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

func (p *OIDCProvider) UserInfo(ctx context.Context, token *oauth2.Token) (*SSOUser, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in OIDC response")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("oidc verify id_token: %w", err)
	}

	var claims struct {
		Sub       string `json:"sub"`
		Email     string `json:"email"`
		Username  string `json:"preferred_username"`
		Name      string `json:"name"`
		AvatarURL string `json:"picture"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oidc decode claims: %w", err)
	}

	username := claims.Username
	if username == "" {
		username = claims.Name
	}
	if username == "" {
		username = claims.Email
	}

	user := &SSOUser{
		Provider:   "oidc",
		ProviderID: claims.Sub,
		Email:      claims.Email,
		Username:   username,
		AvatarURL:  claims.AvatarURL,
	}

	if user.Email == "" {
		userInfo, err := p.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
		if err == nil {
			var uiClaims struct {
				Email string `json:"email"`
				Sub   string `json:"sub"`
			}
			if err := userInfo.Claims(&uiClaims); err == nil {
				user.Email = uiClaims.Email
				if user.ProviderID == "" {
					user.ProviderID = uiClaims.Sub
				}
			}
		}
	}

	return user, nil
}
