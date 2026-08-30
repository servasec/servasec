package services

import (
	"strconv"
	"testing"

	"github.com/servasec/servasec/backend/config"
	"github.com/servasec/servasec/backend/features"
	"github.com/servasec/servasec/backend/models"
	"github.com/servasec/servasec/backend/testutil"
)

func withQuota(t *testing.T, q *features.Quota) {
	t.Helper()
	testutil.SetupTestDB()
	purgeQuotaTables(t)
	if q == nil {
		features.F = nil
		return
	}
	features.F = features.NewRegistryWithQuota(features.FreeFeatures(), q)
}

func purgeQuotaTables(t *testing.T) {
	t.Helper()
	for _, m := range []interface{}{&models.Group{}, &models.Application{}, &models.ApplicationVersion{}, &models.User{}} {
		if err := config.DB.Unscoped().Where("1 = 1").Delete(m).Error; err != nil {
			t.Fatalf("failed to purge table: %v", err)
		}
	}
}

func resetQuota() {
	features.F = nil
}

func seedGroups(t *testing.T, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		g := models.Group{Name: "g", Path: "quota-group-" + strconv.Itoa(i)}
		if err := config.DB.Create(&g).Error; err != nil {
			t.Fatalf("failed to seed group: %v", err)
		}
	}
}

func TestQuota_UnlimitedWithoutLicense(t *testing.T) {
	withQuota(t, nil)
	defer resetQuota()
	seedGroups(t, 5)
	if err := CheckGroupQuota(); err != nil {
		t.Fatalf("expected no quota error without license, got %v", err)
	}
}

func TestQuota_GroupLimitEnforced(t *testing.T) {
	withQuota(t, &features.Quota{Groups: 2, Applications: 5, Versions: 5, Users: 5})
	defer resetQuota()
	seedGroups(t, 2)
	err := CheckGroupQuota()
	qe, ok := err.(*QuotaExceededError)
	if !ok {
		t.Fatalf("expected QuotaExceededError, got %v", err)
	}
	if qe.Resource != QuotaGroups || qe.Limit != 2 || qe.Usage != 2 {
		t.Fatalf("unexpected quota error details: %+v", qe)
	}
}

func TestQuota_VersionLimitPerApplication(t *testing.T) {
	withQuota(t, &features.Quota{Groups: 5, Applications: 5, Versions: 2, Users: 5})
	defer resetQuota()

	var group models.Group
	if err := config.DB.Create(&models.Group{Name: "g", Path: "quota-app-group"}).Error; err != nil {
		t.Fatalf("failed to seed group: %v", err)
	}
	config.DB.Where("path = ?", "quota-app-group").First(&group)

	var app models.Application
	app = models.Application{Name: "app", Slug: "quota-app-1", GroupID: group.ID, ApiToken: "token-1"}
	if err := config.DB.Create(&app).Error; err != nil {
		t.Fatalf("failed to seed application: %v", err)
	}
	seedVersion(t, app.ID, "v1")
	seedVersion(t, app.ID, "v2")

	err := CheckVersionQuota(app.ID)
	if _, ok := err.(*QuotaExceededError); !ok {
		t.Fatalf("expected QuotaExceededError at limit, got %v", err)
	}

	// A different application still has headroom.
	var app2 models.Application
	app2 = models.Application{Name: "app2", Slug: "quota-app-2", GroupID: group.ID, ApiToken: "token-2"}
	if err := config.DB.Create(&app2).Error; err != nil {
		t.Fatalf("failed to seed application 2: %v", err)
	}
	seedVersion(t, app2.ID, "v1")
	if err := CheckVersionQuota(app2.ID); err != nil {
		t.Fatalf("expected no error for fresh application, got %v", err)
	}
}

func seedVersion(t *testing.T, appID uint, name string) {
	t.Helper()
	v := models.ApplicationVersion{ApplicationID: appID, Name: name}
	if err := config.DB.Create(&v).Error; err != nil {
		t.Fatalf("failed to seed version: %v", err)
	}
}

func TestQuota_StatusesIncludeUnlimitedAsNilLimit(t *testing.T) {
	withQuota(t, &features.Quota{Groups: 3})
	defer resetQuota()

	statuses, err := GetQuotaStatuses()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	byResource := map[QuotaResource]QuotaStatus{}
	for _, s := range statuses {
		byResource[s.Resource] = s
	}
	if g := byResource[QuotaGroups]; g.Limit == nil || *g.Limit != 3 {
		t.Fatalf("expected groups limit 3, got %+v", g)
	}
	if a := byResource[QuotaApplications]; a.Limit != nil {
		t.Fatalf("expected nil limit for applications, got %+v", a)
	}
	if u := byResource[QuotaUsers]; u.Limit != nil {
		t.Fatalf("expected nil limit for users, got %+v", u)
	}
}
