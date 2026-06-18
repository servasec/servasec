package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
)

var (
	AccessSecret  = []byte(os.Getenv("JWT_SECRET"))
	RefreshSecret = []byte(os.Getenv("REFRESH_SECRET"))
)

func ValidateSecrets() {
	if len(AccessSecret) == 0 {
		log.Fatal("JWT_SECRET environment variable is required but not set")
	}
	if len(RefreshSecret) == 0 {
		log.Fatal("REFRESH_SECRET environment variable is required but not set")
	}
}

func hashToken(tokenStr string) string {
	hash := sha256.Sum256([]byte(tokenStr))
	return hex.EncodeToString(hash[:])
}

func IsTokenBlacklisted(tokenStr string) bool {
	tokenHash := hashToken(tokenStr)
	var entry models.BlacklistedToken
	result := config.DB.Where("token_hash = ?", tokenHash).First(&entry)
	return result.Error == nil
}

func StartBlacklistCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().Add(-7 * 24 * time.Hour)
			config.DB.Where("created_at < ?", cutoff).Delete(&models.BlacklistedToken{})
		}
	}()
}

type TokenClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID uint, role string) (string, error) {
	claims := TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(45 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(AccessSecret)
}

func GenerateRefreshToken(userID uint) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(int(userID)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(RefreshSecret)
}

func GetClaimsFromCookie(c *gin.Context) (*TokenClaims, error) {
	tokenStr, err := c.Cookie("access_token")
	if err != nil {
		return nil, err
	}

	if IsTokenBlacklisted(tokenStr) {
		return nil, jwt.ErrTokenInvalidClaims
	}

	token, err := jwt.ParseWithClaims(tokenStr, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return AccessSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(*TokenClaims), nil
}

func BlacklistToken(tokenStr string) {
	tokenHash := hashToken(tokenStr)
	entry := models.BlacklistedToken{
		TokenHash: tokenHash,
	}
	if err := config.DB.Create(&entry).Error; err != nil {
		log.Printf("Failed to blacklist token: %v", err)
	}
}
