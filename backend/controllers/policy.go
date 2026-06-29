package controllers

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

func GetPolicies(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID)

	var policies []models.Policy
	q := config.DB.Model(&models.Policy{})

	if scopeType := c.Query("scopeType"); scopeType != "" {
		q = q.Where("scope_type = ?", scopeType)
	}
	if scopeValue := c.Query("scopeValue"); scopeValue != "" {
		q = q.Where("scope_value = ?", scopeValue)
	}
	if eventType := c.Query("eventType"); eventType != "" {
		q = q.Where("event_types LIKE ?", "%"+eventType+"%")
	}

	q.Order("priority DESC, created_at DESC").Find(&policies)

	if user.Role != "admin" {
		accessibleIDs := utils.GetAccessibleAppIDs(c)
		var filtered []models.Policy
		for _, p := range policies {
			if policyAccessibleToUser(p, accessibleIDs) {
				filtered = append(filtered, p)
			}
		}
		policies = filtered
	}

	utils.OKResponse(c, policies)
}

func GetPolicy(c *gin.Context) {
	var policy models.Policy
	if err := config.DB.First(&policy, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Policy not found")
		return
	}

	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID)

	if user.Role != "admin" && !policyAccessibleToUser(policy, utils.GetAccessibleAppIDs(c)) {
		utils.ForbiddenError(c, "access denied")
		return
	}

	utils.OKResponse(c, policy)
}

func CreatePolicy(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required,min=1,max=200"`
		Description string `json:"description" binding:"max=1000"`
		ScopeType   string `json:"scopeType" binding:"required,oneof=application group global"`
		ScopeValue  string `json:"scopeValue" binding:"required"`
		EventTypes  string `json:"eventTypes" binding:"required"`
		Conditions  string `json:"conditions"`
		Actions     string `json:"actions" binding:"required"`
		IsActive    *bool  `json:"isActive"`
		Priority    int    `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	if input.Conditions != "" && !isValidJSON(input.Conditions) {
		utils.BadRequestError(c, "Conditions must be valid JSON array")
		return
	}
	if !isValidJSON(input.Actions) {
		utils.BadRequestError(c, "Actions must be valid JSON array")
		return
	}

	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID)

	if !userCanManageScope(&user, input.ScopeType, input.ScopeValue, c) {
		utils.ForbiddenError(c, "access denied")
		return
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	policy := models.Policy{
		Name:        input.Name,
		Description: input.Description,
		ScopeType:   input.ScopeType,
		ScopeValue:  input.ScopeValue,
		EventTypes:  input.EventTypes,
		Conditions:  input.Conditions,
		Actions:     input.Actions,
		IsActive:    isActive,
		Priority:    input.Priority,
	}
	if err := config.DB.Create(&policy).Error; err != nil {
		utils.InternalServerError(c, "Failed to create policy")
		return
	}
	utils.CreatedResponse(c, policy)
}

func UpdatePolicy(c *gin.Context) {
	var policy models.Policy
	if err := config.DB.First(&policy, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Policy not found")
		return
	}

	var input struct {
		Name        *string `json:"name" binding:"omitnil,min=1,max=200"`
		Description *string `json:"description" binding:"max=1000"`
		ScopeType   *string `json:"scopeType" binding:"omitnil,oneof=application group global"`
		ScopeValue  *string `json:"scopeValue"`
		EventTypes  *string `json:"eventTypes"`
		Conditions  *string `json:"conditions"`
		Actions     *string `json:"actions"`
		IsActive    *bool   `json:"isActive"`
		Priority    *int    `json:"priority"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequestError(c, "Invalid input")
		return
	}

	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID)

	if user.Role != "admin" && !policyAccessibleToUser(policy, utils.GetAccessibleAppIDs(c)) {
		utils.ForbiddenError(c, "access denied")
		return
	}

	if input.ScopeType != nil || input.ScopeValue != nil {
		st := policy.ScopeType
		sv := policy.ScopeValue
		if input.ScopeType != nil {
			st = *input.ScopeType
		}
		if input.ScopeValue != nil {
			sv = *input.ScopeValue
		}
		if !userCanManageScope(&user, st, sv, c) {
			utils.ForbiddenError(c, "access denied for new scope")
			return
		}
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.ScopeType != nil {
		updates["scope_type"] = *input.ScopeType
	}
	if input.ScopeValue != nil {
		updates["scope_value"] = *input.ScopeValue
	}
	if input.EventTypes != nil {
		updates["event_types"] = *input.EventTypes
	}
	if input.Conditions != nil {
		if *input.Conditions != "" && !isValidJSON(*input.Conditions) {
			utils.BadRequestError(c, "Conditions must be valid JSON array")
			return
		}
		updates["conditions"] = *input.Conditions
	}
	if input.Actions != nil {
		if !isValidJSON(*input.Actions) {
			utils.BadRequestError(c, "Actions must be valid JSON array")
			return
		}
		updates["actions"] = *input.Actions
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.Priority != nil {
		updates["priority"] = *input.Priority
	}

	if len(updates) == 0 {
		utils.OKResponse(c, policy)
		return
	}

	if err := config.DB.Model(&policy).Updates(updates).Error; err != nil {
		utils.InternalServerError(c, "Failed to update policy")
		return
	}
	config.DB.First(&policy, policy.ID)
	utils.OKResponse(c, policy)
}

func DeletePolicy(c *gin.Context) {
	var policy models.Policy
	if err := config.DB.First(&policy, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Policy not found")
		return
	}

	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID)

	if user.Role != "admin" && !policyAccessibleToUser(policy, utils.GetAccessibleAppIDs(c)) {
		utils.ForbiddenError(c, "access denied")
		return
	}

	config.DB.Where("policy_id = ?", policy.ID).Delete(&models.PolicyLog{})
	if err := config.DB.Delete(&policy).Error; err != nil {
		utils.InternalServerError(c, "Failed to delete policy")
		return
	}
	utils.NoContentResponse(c)
}

func TogglePolicy(c *gin.Context) {
	var policy models.Policy
	if err := config.DB.First(&policy, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Policy not found")
		return
	}

	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID)

	if user.Role != "admin" && !policyAccessibleToUser(policy, utils.GetAccessibleAppIDs(c)) {
		utils.ForbiddenError(c, "access denied")
		return
	}

	policy.IsActive = !policy.IsActive
	if err := config.DB.Save(&policy).Error; err != nil {
		utils.InternalServerError(c, "Failed to toggle policy")
		return
	}
	utils.OKResponse(c, policy)
}

func GetPolicyLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID)

	q := config.DB.Model(&models.PolicyLog{})

	if policyID := c.Query("policyId"); policyID != "" {
		q = q.Where("policy_id = ?", policyID)
	}
	if findingID := c.Query("findingId"); findingID != "" {
		q = q.Where("finding_id = ?", findingID)
	}
	if eventType := c.Query("eventType"); eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}

	if user.Role != "admin" {
		accessibleIDs := utils.GetAccessibleAppIDs(c)
		policyIDs := accessiblePolicyIDs(accessibleIDs)
		if len(policyIDs) == 0 {
			utils.OKResponse(c, []models.PolicyLog{})
			return
		}
		q = q.Where("policy_id IN ?", policyIDs)
	}

	var logs []models.PolicyLog
	q.Order("created_at DESC").Limit(200).Find(&logs)
	utils.OKResponse(c, logs)
}

func GetPolicyLogsByPolicy(c *gin.Context) {
	var policy models.Policy
	if err := config.DB.First(&policy, c.Param("id")).Error; err != nil {
		utils.NotFoundError(c, "Policy not found")
		return
	}

	userID, _ := c.Get("user_id")
	var user models.User
	config.DB.First(&user, userID)

	if user.Role != "admin" && !policyAccessibleToUser(policy, utils.GetAccessibleAppIDs(c)) {
		utils.ForbiddenError(c, "access denied")
		return
	}

	var logs []models.PolicyLog
	config.DB.Where("policy_id = ?", policy.ID).Order("created_at DESC").Limit(100).Find(&logs)
	utils.OKResponse(c, logs)
}

func isValidJSON(s string) bool {
	var v interface{}
	return json.Unmarshal([]byte(s), &v) == nil
}

func policyAccessibleToUser(policy models.Policy, accessibleIDs []string) bool {
	switch policy.ScopeType {
	case "global":
		return true
	case "application":
		return stringSliceContains(accessibleIDs, policy.ScopeValue)
	case "group":
		var count int64
		config.DB.Model(&models.Application{}).
			Where("group_id = ? AND id IN ?", policy.ScopeValue, accessibleIDs).
			Count(&count)
		return count > 0
	default:
		return false
	}
}

func userCanManageScope(user *models.User, scopeType, scopeValue string, c *gin.Context) bool {
	if user.Role == "admin" {
		return true
	}
	switch scopeType {
	case "global":
		return false
	case "group":
		var count int64
		config.DB.Model(&models.Application{}).
			Where("group_id = ? AND id IN ?", scopeValue, utils.GetAccessibleAppIDs(c)).
			Count(&count)
		return count > 0
	case "application":
		return stringSliceContains(utils.GetAccessibleAppIDs(c), scopeValue)
	default:
		return false
	}
}

func accessiblePolicyIDs(accessibleIDs []string) []uint {
	var all []models.Policy
	config.DB.Where("is_active = ?", true).Find(&all)

	var ids []uint
	for _, p := range all {
		if policyAccessibleToUser(p, accessibleIDs) {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

func stringSliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func parseUintOrZero(s string) uint {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint(id)
}
