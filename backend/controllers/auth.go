package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/dto"
	"github.com/servasec/servasec/backend/features"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func Register(c *gin.Context) {
	var input dto.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input: username (max 32), email (max 254), password (8-72)")
		return
	}

	if errKey := utils.ValidateUsername(input.Username); errKey != "" {
		utils.BadRequestError(c, errKey)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalServerError(c, "Failed to hash password")
		return
	}

	user := models.User{
		Username: input.Username,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     "member",
	}

	var existing models.User
	if err := config.DB.Where("username = ?", input.Username).First(&existing).Error; err == nil {
		utils.ConflictError(c, "Username already exists")
		return
	}
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		utils.ConflictError(c, "Email already exists")
		return
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utils.InternalServerError(c, "Failed to create user")
		return
	}

	utils.CreatedResponse(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

func Login(c *gin.Context) {
	var input dto.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "invalid_input")
		return
	}

	usernameOrEmail := strings.TrimSpace(input.Username)
	if usernameOrEmail == "" || strings.TrimSpace(input.Password) == "" {
		utils.BadRequestError(c, "please_fill_fields")
		return
	}

	var user models.User
	if err := config.DB.Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).First(&user).Error; err != nil {
		utils.UnauthorizedError(c, "invalid_credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		utils.UnauthorizedError(c, "invalid_credentials")
		return
	}

	if user.Banned {
		utils.ForbiddenError(c, "banned")
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		utils.InternalServerError(c, "could not create access token")
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		utils.InternalServerError(c, "could not create refresh token")
		return
	}

	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("access_token", accessToken, 3600, "/", "", true, true)
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", "", true, true)

	var enabledFeatures []string
	if features.F != nil {
		enabledFeatures = features.F.EnabledFeatures()
	}

	userResp := gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"role":      user.Role,
		"avatarUrl": user.AvatarURL,
		"features":  enabledFeatures,
	}
	if user.OAuthProvider != "" {
		userResp["oauthProvider"] = user.OAuthProvider
	}
	utils.OKResponse(c, gin.H{
		"message": "Login successful",
		"user":    userResp,
	})
}

func Refresh(c *gin.Context) {
	tokenStr, err := c.Cookie("refresh_token")
	if err != nil {
		utils.UnauthorizedError(c, "missing refresh token")
		return
	}

	if utils.IsTokenBlacklisted(tokenStr) {
		utils.UnauthorizedError(c, "token invalidated")
		return
	}

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return utils.RefreshSecret, nil
	})
	if err != nil || !token.Valid {
		utils.UnauthorizedError(c, "invalid refresh token")
		return
	}

	claims := token.Claims.(*jwt.RegisteredClaims)
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		utils.InternalServerError(c, "can't read subject")
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.UnauthorizedError(c, "user not found")
		return
	}

	if user.Banned {
		utils.UnauthorizedError(c, "banned")
		return
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		utils.InternalServerError(c, "failed to generate access token")
		return
	}

	c.SetCookie("access_token", accessToken, 3600, "/", "", true, true)

	utils.OKResponse(c, gin.H{"message": "Token refreshed"})
}

func Logout(c *gin.Context) {
	accessToken, _ := c.Cookie("access_token")
	refreshToken, _ := c.Cookie("refresh_token")

	if accessToken != "" {
		utils.BlacklistToken(accessToken)
	}
	if refreshToken != "" {
		utils.BlacklistToken(refreshToken)
	}

	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	utils.OKResponse(c, gin.H{"message": "logged out"})
}

func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedError(c, "unauthorized")
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.NotFoundError(c, "User not found")
		return
	}

	var enabledFeatures []string
	if features.F != nil {
		enabledFeatures = features.F.EnabledFeatures()
	}

	resp := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"banned":   user.Banned,
		"avatarUrl": user.AvatarURL,
		"features": enabledFeatures,
	}
	if user.OAuthProvider != "" {
		resp["oauthProvider"] = user.OAuthProvider
	}
	utils.OKResponse(c, resp)
}

func UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedError(c, "unauthorized")
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.NotFoundError(c, "User not found")
		return
	}

	var input struct {
		Username string `json:"username" binding:"omitnil,min=1,max=32"`
		Email    string `json:"email" binding:"omitnil,email,max=254"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	if input.Email != "" {
		if user.OAuthProvider != "" {
			utils.ForbiddenError(c, "SSO users cannot change their email")
			return
		}
		var existing models.User
		if err := config.DB.Where("email = ? AND id != ?", input.Email, user.ID).First(&existing).Error; err == nil {
			utils.ConflictError(c, "Email already taken")
			return
		}
		user.Email = input.Email
	}

	if input.Username != "" {
		user.Username = input.Username
	}

	if err := config.DB.Save(&user).Error; err != nil {
		utils.InternalServerError(c, "Failed to update profile")
		return
	}

	utils.OKResponse(c, gin.H{"message": "Profile updated"})
}

func UpdateCurrentUserPassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedError(c, "unauthorized")
		return
	}

	var input struct {
		Current string `json:"current"`
		New     string `json:"new" binding:"min=8,max=72"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Password must be 8-72 characters")
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		utils.NotFoundError(c, "User not found")
		return
	}

	if user.OAuthProvider != "" {
		utils.ForbiddenError(c, "SSO users cannot change their password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Current)); err != nil {
		utils.UnauthorizedError(c, "Current password is incorrect")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.New), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalServerError(c, "Failed to hash new password")
		return
	}

	user.Password = string(hashedPassword)
	if err := config.DB.Save(&user).Error; err != nil {
		utils.InternalServerError(c, "Failed to update password")
		return
	}

	utils.OKResponse(c, gin.H{"message": "Password updated"})
}
