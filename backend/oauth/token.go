package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/utils"
)

func generateRefreshToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// HandleToken exchanges an authorization code or refresh token for tokens
// @Summary OAuth token endpoint
// @Tags OAuth
// @Accept x-www-form-urlencoded
// @Produce json
// @Param grant_type formData string true "Grant type (authorization_code or refresh_token)"
// @Param client_id formData string true "OAuth client ID"
// @Param code formData string false "Authorization code (required for authorization_code grant)"
// @Param code_verifier formData string false "PKCE code verifier (required for authorization_code grant)"
// @Param redirect_uri formData string false "Redirect URI (optional, must match authorize request)"
// @Param refresh_token formData string false "Refresh token (required for refresh_token grant)"
// @Success 200 {object} gin.H "Token response with access_token, refresh_token, expires_in"
// @Failure 400 {object} gin.H "Invalid request"
// @Router /oauth/token [post]
func HandleToken(c *gin.Context) {
	grantType := c.PostForm("grant_type")
	clientID := c.PostForm("client_id")

	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id is required"})
		return
	}

	client := DefaultStore.GetClient(clientID)
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client"})
		return
	}

	switch grantType {
	case "authorization_code":
		handleAuthCodeGrant(c, clientID)
	case "refresh_token":
		handleRefreshTokenGrant(c, clientID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported grant_type"})
	}
}

func handleAuthCodeGrant(c *gin.Context, clientID string) {
	code := c.PostForm("code")
	verifier := c.PostForm("code_verifier")
	redirectURI := c.PostForm("redirect_uri")

	if code == "" || verifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code and code_verifier are required"})
		return
	}

	authCode := DefaultStore.GetCode(code)
	if authCode == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code"})
		return
	}

	if authCode.ClientID != clientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code was not issued for this client"})
		return
	}

	if authCode.Used {
		DefaultStore.DeleteCode(code)
		c.JSON(http.StatusBadRequest, gin.H{"error": "code already used"})
		return
	}

	if time.Now().After(authCode.ExpiresAt) {
		DefaultStore.DeleteCode(code)
		c.JSON(http.StatusBadRequest, gin.H{"error": "code expired"})
		return
	}

	if redirectURI != "" && redirectURI != authCode.RedirectURI {
		c.JSON(http.StatusBadRequest, gin.H{"error": "redirect_uri mismatch"})
		return
	}

	expected := codeVerifierToChallenge(verifier)
	if expected != authCode.Challenge {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code_verifier mismatch"})
		return
	}

	DefaultStore.UseCode(code)

	var user struct {
		ID   uint
		Role string
	}
	if err := config.DB.Table("users").Select("id, role").Where("id = ?", authCode.UserID).Scan(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	refreshToken := generateRefreshToken()
	DefaultStore.AddToken(&Token{
		Refresh:       refreshToken,
		ClientID:      clientID,
		UserID:        user.ID,
		AccessExpiry:  time.Now().Add(45 * time.Minute),
		RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
	})

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"token_type":    "bearer",
		"expires_in":    2700,
		"refresh_token": refreshToken,
	})
}

func handleRefreshTokenGrant(c *gin.Context, clientID string) {
	refreshToken := c.PostForm("refresh_token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token is required"})
		return
	}

	t := DefaultStore.GetToken(refreshToken)
	if t == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refresh_token"})
		return
	}

	if t.Revoked {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token revoked"})
		return
	}

	if t.ClientID != clientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token was not issued for this client"})
		return
	}

	if time.Now().After(t.RefreshExpiry) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token expired"})
		return
	}

	DefaultStore.RevokeToken(refreshToken)

	var user struct {
		ID   uint
		Role string
	}
	if err := config.DB.Table("users").Select("id, role").Where("id = ?", t.UserID).Scan(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	newRefresh := generateRefreshToken()
	DefaultStore.AddToken(&Token{
		Refresh:       newRefresh,
		ClientID:      clientID,
		UserID:        user.ID,
		AccessExpiry:  time.Now().Add(45 * time.Minute),
		RefreshExpiry: time.Now().Add(7 * 24 * time.Hour),
	})

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"token_type":    "bearer",
		"expires_in":    2700,
		"refresh_token": newRefresh,
	})
}
