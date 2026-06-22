package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/models"
)

func collectSubjects(uid uint) []string {
	subs := []string{fmt.Sprintf("user:%d", uid)}
	var teamMembers []models.TeamMember
	config.DB.Where("user_id = ?", uid).Find(&teamMembers)
	for _, tm := range teamMembers {
		subs = append(subs, fmt.Sprintf("team:%d", tm.TeamID))
	}
	return subs
}

func extractIDsFromPolicies(policies [][]string, prefix string) []string {
	seen := make(map[string]struct{})
	for _, p := range policies {
		if len(p) >= 2 && strings.HasPrefix(p[1], prefix) {
			parts := strings.Split(p[1], "/")
			if len(parts) == 3 && parts[2] != "" {
				seen[parts[2]] = struct{}{}
			}
		}
	}
	var ids []string
	for id := range seen {
		ids = append(ids, id)
	}
	return ids
}

func GetAccessibleAppIDs(c *gin.Context) []string {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	currentUser := user.(*models.User)
	if currentUser.Role == "admin" {
		return nil
	}

	uid := c.GetUint("user_id")
	seen := make(map[string]struct{})

	subs := collectSubjects(uid)
	for _, sub := range subs {
		policies, _ := config.CEF.GetFilteredPolicy(0, sub)
		for _, id := range extractIDsFromPolicies(policies, "/applications/") {
			seen[id] = struct{}{}
		}
	}

	groupIDs := make(map[string]struct{})
	for _, sub := range subs {
		policies, _ := config.CEF.GetFilteredPolicy(0, sub)
		for _, id := range extractIDsFromPolicies(policies, "/groups/") {
			groupIDs[id] = struct{}{}
		}
	}

	if len(groupIDs) > 0 {
		var groupAppIDs []string
		config.DB.Model(&models.Application{}).
			Where("group_id IN ?", mapKeys(groupIDs)).
			Pluck("CAST(id AS TEXT)", &groupAppIDs)
		for _, id := range groupAppIDs {
			seen[id] = struct{}{}
		}
	}

	var ids []string
	for id := range seen {
		ids = append(ids, id)
	}
	return ids
}

func mapKeys(m map[string]struct{}) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
