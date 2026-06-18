package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func GetUsers(c *gin.Context) {
	var users []models.User
	result := config.DB.Find(&users)
	if result.Error != nil {
		utils.InternalServerError(c, result.Error.Error())
		return
	}
	utils.OKResponse(c, users)
}

func GetUser(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	result := config.DB.First(&user, id)
	if result.Error != nil {
		utils.NotFoundError(c, "User not found")
		return
	}
	utils.OKResponse(c, user)
}

func UpdateUser(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		utils.NotFoundError(c, "User not found")
		return
	}

	var input struct {
		Email  *string `json:"email" binding:"omitempty,email,max=254"`
		Role   *string `json:"role" binding:"omitempty,oneof=admin member"`
		Banned *bool   `json:"banned"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.Banned != nil {
		user.Banned = *input.Banned
	}

	if err := config.DB.Save(&user).Error; err != nil {
		utils.InternalServerError(c, "Failed to update user")
		return
	}

	utils.OKResponse(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"banned":   user.Banned,
	})
}

func DeleteUser(c *gin.Context) {
	var user models.User
	id := c.Param("id")

	if err := config.DB.First(&user, id).Error; err != nil {
		utils.NotFoundError(c, "User not found")
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		utils.InternalServerError(c, "Failed to delete user")
		return
	}
	utils.OKResponse(c, gin.H{"message": "User deleted"})
}
