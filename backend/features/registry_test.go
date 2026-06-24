package features

import (
	"os"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry([]string{"findings", "audit_log"})

	if !r.IsEnabled("findings") {
		t.Error("expected findings to be enabled")
	}
	if !r.IsEnabled("audit_log") {
		t.Error("expected audit_log to be enabled")
	}
	if r.IsEnabled("nonexistent") {
		t.Error("expected nonexistent feature to be disabled")
	}
}

func TestFreeFeatures(t *testing.T) {
	ff := FreeFeatures()
	expected := []string{"findings", "webhooks", "teams", "dashboard"}
	if len(ff) != len(expected) {
		t.Fatalf("expected %d free features, got %d", len(expected), len(ff))
	}
	for _, exp := range expected {
		found := false
		for _, f := range ff {
			if f == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected free feature %q not found", exp)
		}
	}
}

func TestEnabledFeatures(t *testing.T) {
	r := NewRegistry([]string{"a", "b"})
	ef := r.EnabledFeatures()
	if len(ef) != 2 {
		t.Fatalf("expected 2 features, got %d", len(ef))
	}
}

func TestInitFreeOnly(t *testing.T) {
	os.Clearenv()
	F = nil

	r := Init("")

	if !r.IsEnabled(FeatureFindings) {
		t.Error("findings should be enabled in free tier")
	}
	if !r.IsEnabled(FeatureWebhooks) {
		t.Error("webhooks should be enabled in free tier")
	}
	if !r.IsEnabled(FeatureTeams) {
		t.Error("teams should be enabled in free tier")
	}
	if !r.IsEnabled(FeatureDashboard) {
		t.Error("dashboard should be enabled in free tier")
	}
	if r.IsEnabled(FeatureAuditLog) {
		t.Error("audit_log should NOT be enabled in free tier")
	}
}

func TestInitWithEnvIgnored(t *testing.T) {
	os.Clearenv()
	os.Setenv("SSC_PRO_AUDIT_LOG", "true")
	F = nil

	r := Init("")

	if r.IsEnabled(FeatureAuditLog) {
		t.Error("audit_log should NOT be enabled via env var without license")
	}
	if !r.IsEnabled(FeatureFindings) {
		t.Error("findings should still be enabled")
	}
}

func TestInitWithInvalidLicense(t *testing.T) {
	os.Clearenv()
	F = nil

	r := Init("this-is-not-a-valid-jwt")

	if r.IsEnabled(FeatureAuditLog) {
		t.Error("audit_log should NOT be enabled with invalid license")
	}
}
