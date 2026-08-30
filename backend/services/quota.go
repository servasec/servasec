package services

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/features"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/utils"
)

type QuotaResource string

const (
	QuotaGroups       QuotaResource = "groups"
	QuotaApplications QuotaResource = "applications"
	QuotaVersions     QuotaResource = "versions"
	QuotaUsers        QuotaResource = "users"
)

// QuotaExceededError is returned when creating a resource would exceed the
// limits granted by the current license.
type QuotaExceededError struct {
	Resource QuotaResource
	Limit    int
	Usage    int
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf("quota exceeded for %s (usage %d, limit %d)", e.Resource, e.Usage, e.Limit)
}

// RespondQuotaExceeded writes a structured 402 response when err is a
// QuotaExceededError; non-quota errors are re-raised as 500.
func RespondQuotaExceeded(c *gin.Context, err error) {
	var qe *QuotaExceededError
	if errors.As(err, &qe) {
		utils.QuotaExceededResponse(c, string(qe.Resource), qe.Limit, qe.Usage)
		return
	}
	utils.InternalServerError(c, "failed to check quota")
}

// QuotaLimits returns the license quota, or nil when unlimited (community
// build, or license without quota claims).
func QuotaLimits() *features.Quota {
	if features.F == nil {
		return nil
	}
	return features.F.Quota()
}

func quotaLimit(resource QuotaResource) int {
	q := QuotaLimits()
	if q == nil {
		return 0
	}
	switch resource {
	case QuotaGroups:
		return q.Groups
	case QuotaApplications:
		return q.Applications
	case QuotaVersions:
		return q.Versions
	case QuotaUsers:
		return q.Users
	}
	return 0
}

func countGlobal(resource QuotaResource) (int, error) {
	var count int64
	switch resource {
	case QuotaGroups:
		err := config.DB.Model(&models.Group{}).Count(&count).Error
		return int(count), err
	case QuotaApplications:
		err := config.DB.Model(&models.Application{}).Count(&count).Error
		return int(count), err
	case QuotaUsers:
		err := config.DB.Model(&models.User{}).Count(&count).Error
		return int(count), err
	}
	return 0, nil
}

// CheckGroupQuota verifies the instance can create a new group.
func CheckGroupQuota() error {
	return checkQuota(QuotaGroups)
}

// CheckApplicationQuota verifies the instance can create a new application.
func CheckApplicationQuota() error {
	return checkQuota(QuotaApplications)
}

// CheckVersionQuota verifies an application can host a new version.
func CheckVersionQuota(applicationID uint) error {
	limit := quotaLimit(QuotaVersions)
	if limit <= 0 {
		return nil
	}
	var count int64
	if err := config.DB.Model(&models.ApplicationVersion{}).
		Where("application_id = ?", applicationID).
		Count(&count).Error; err != nil {
		return err
	}
	if int(count) >= limit {
		return &QuotaExceededError{Resource: QuotaVersions, Limit: limit, Usage: int(count)}
	}
	return nil
}

// CheckUserQuota verifies the instance can register a new user.
func CheckUserQuota() error {
	return checkQuota(QuotaUsers)
}

func checkQuota(resource QuotaResource) error {
	limit := quotaLimit(resource)
	if limit <= 0 {
		return nil
	}
	usage, err := countGlobal(resource)
	if err != nil {
		return err
	}
	if usage >= limit {
		return &QuotaExceededError{Resource: resource, Limit: limit, Usage: usage}
	}
	return nil
}

// QuotaUsage returns the current global usage for the resources covered by a
// license, used by GET /me/quotas. Values are only present for limited
// resources; unlimited resources are omitted by the caller via limits.
func QuotaUsage() (map[QuotaResource]int, error) {
	usage := make(map[QuotaResource]int, 3)
	for _, r := range []QuotaResource{QuotaGroups, QuotaApplications, QuotaUsers} {
		n, err := countGlobal(r)
		if err != nil {
			return nil, err
		}
		usage[r] = n
	}
	return usage, nil
}

// QuotaStatus describes a single resource's limit and current usage. Limit is
// nil when the instance is unlimited for that resource (no license, or license
// without quota claims).
type QuotaStatus struct {
	Resource QuotaResource `json:"resource"`
	Limit    *int          `json:"limit"`
	Usage    int           `json:"usage"`
}

// versionUsage returns the highest number of versions hosted by a single
// application, which is the binding constraint for the per-application
// versions quota.
func versionUsage() (int, error) {
	var rows []struct {
		Cnt int
	}
	if err := config.DB.Model(&models.ApplicationVersion{}).
		Select("count(*) as cnt").
		Group("application_id").
		Scan(&rows).Error; err != nil {
		return 0, err
	}
	max := 0
	for _, r := range rows {
		if r.Cnt > max {
			max = r.Cnt
		}
	}
	return max, nil
}

// GetQuotaStatuses returns limit + usage for every quota resource.
func GetQuotaStatuses() ([]QuotaStatus, error) {
	q := QuotaLimits()
	limitFor := func(r QuotaResource) *int {
		if q == nil {
			return nil
		}
		var l int
		switch r {
		case QuotaGroups:
			l = q.Groups
		case QuotaApplications:
			l = q.Applications
		case QuotaVersions:
			l = q.Versions
		case QuotaUsers:
			l = q.Users
		}
		if l <= 0 {
			return nil
		}
		return &l
	}

	usage, err := QuotaUsage()
	if err != nil {
		return nil, err
	}
	versions, err := versionUsage()
	if err != nil {
		return nil, err
	}
	usage[QuotaVersions] = versions

	resources := []QuotaResource{QuotaGroups, QuotaApplications, QuotaVersions, QuotaUsers}
	statuses := make([]QuotaStatus, 0, len(resources))
	for _, r := range resources {
		statuses = append(statuses, QuotaStatus{
			Resource: r,
			Limit:    limitFor(r),
			Usage:    usage[r],
		})
	}
	return statuses, nil
}
