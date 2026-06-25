package sso

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/debug"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
	"golang.org/x/crypto/bcrypt"
)

func HandleAuthRedirect(c *gin.Context) {
	providerName := c.Param("provider")
	provider := GetProvider(providerName)
	if provider == nil {
		redirectError(c, "sso_provider_not_found")
		return
	}

	state, err := generateState()
	if err != nil {
		redirectError(c, "sso_internal_error")
		return
	}

	setStateCookie(c, state)

	authURL := provider.AuthURL(state)
	c.Redirect(http.StatusFound, authURL)
}

func HandleCallback(c *gin.Context) {
	providerName := c.Param("provider")
	provider := GetProvider(providerName)
	if provider == nil {
		redirectError(c, "sso_provider_not_found")
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		redirectError(c, "sso_missing_params")
		return
	}

	cookieState, err := getStateCookie(c)
	if err != nil || cookieState == "" {
		redirectError(c, "sso_invalid_state")
		return
	}

	if state != cookieState {
		redirectError(c, "sso_state_mismatch")
		return
	}

	clearStateCookie(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	token, err := provider.Exchange(ctx, code)
	if err != nil {
		debug.Log("SSO exchange error [%s]: %v", providerName, err)
		redirectError(c, "sso_exchange_failed")
		return
	}

	ssoUser, err := provider.UserInfo(ctx, token)
	if err != nil {
		debug.Log("SSO userinfo error [%s]: %v", providerName, err)
		redirectError(c, "sso_userinfo_failed")
		return
	}

	user, err := findOrCreateUser(ssoUser)
	if err != nil {
		debug.Log("SSO findOrCreate error [%s]: %v", providerName, err)
		redirectError(c, "sso_auth_failed")
		return
	}

	if user.Banned {
		redirectError(c, "banned")
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		redirectError(c, "sso_internal_error")
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		redirectError(c, "sso_internal_error")
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", accessToken, 3600, "/", "", true, true)
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", true, true)

	redirectSuccess(c)
}

func findOrCreateUser(ssoUser *SSOUser) (*models.User, error) {
	var user models.User

	err := config.DB.
		Where("o_auth_provider = ? AND o_auth_id = ?", ssoUser.Provider, ssoUser.ProviderID).
		First(&user).Error
	if err == nil {
		if ssoUser.AvatarURL != "" {
			user.AvatarURL = ssoUser.AvatarURL
			config.DB.Save(&user)
		}
		return &user, nil
	}

	if ssoUser.Email != "" {
		err = config.DB.Where("email = ?", ssoUser.Email).First(&user).Error
		if err == nil && user.OAuthProvider == "" {
			user.OAuthProvider = ssoUser.Provider
			user.OAuthID = ssoUser.ProviderID
			if ssoUser.AvatarURL != "" {
				user.AvatarURL = ssoUser.AvatarURL
			}
			config.DB.Save(&user)
			return &user, nil
		}
	}

	username := sanitizeUsername(ssoUser.Username)
	if username == "" {
		username = fmt.Sprintf("user-%s", ssoUser.ProviderID)
	}
	username = ensureUniqueUsername(username)

	password, err := generateRandomPassword()
	if err != nil {
		return nil, fmt.Errorf("generate password: %w", err)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user = models.User{
		Username:      username,
		Email:         ssoUser.Email,
		Password:      string(hashedPassword),
		Role:          "member",
		OAuthProvider: ssoUser.Provider,
		OAuthID:       ssoUser.ProviderID,
		AvatarURL:     ssoUser.AvatarURL,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func sanitizeUsername(name string) string {
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}
	s := result.String()
	if len(s) > 32 {
		s = s[:32]
	}
	return s
}

func ensureUniqueUsername(base string) string {
	name := base
	attempt := 0
	for {
		var existing models.User
		if err := config.DB.Where("username = ?", name).First(&existing).Error; err != nil {
			return name
		}
		attempt++
		suffix := fmt.Sprintf("-%d", attempt)
		if len(base)+len(suffix) > 32 {
			base = base[:32-len(suffix)]
		}
		name = base + suffix
	}
}

func generateRandomPassword() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
