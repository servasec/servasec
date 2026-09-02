package features

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testPrivateKeyPEM = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIOTT8khnIZsi6oJnWGO2ki9SVnrAshIjADDxGib+6KIloAoGCCqGSM49
AwEHoUQDQgAEeqFq7miE9d+a2ew4tTKv1VjNMw9LE5c5UnyLxJQ9yKhlT16W1clD
EVPLCD2JMzGDDCrV/3l+b2+aiF6Z2d7LFg==
-----END EC PRIVATE KEY-----`

func parseTestPrivateKey() *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(testPrivateKeyPEM))
	if block == nil {
		panic("failed to decode test private key PEM")
	}
	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		panic("failed to parse test private key: " + err.Error())
	}
	return key
}

func signTestLicense(features []string, expiresAt *time.Time) string {
	return signTestLicenseWithQuota(features, nil, expiresAt)
}

func signTestLicenseWithQuota(features []string, quota *Quota, expiresAt *time.Time) string {
	claims := licenseClaims{
		Features: features,
		Quota:    quota,
	}
	if expiresAt != nil {
		claims.ExpiresAt = jwt.NewNumericDate(*expiresAt)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	signed, err := token.SignedString(parseTestPrivateKey())
	if err != nil {
		panic("failed to sign test license: " + err.Error())
	}
	return signed
}

func TestParseLicense_Empty(t *testing.T) {
	result := ParseLicense("")
	if result != nil {
		t.Error("expected nil for empty license key")
	}
}

func TestParseLicense_Blank(t *testing.T) {
	result := ParseLicense("  ")
	if result != nil {
		t.Error("expected nil for blank license key")
	}
}

func TestParseLicense_InvalidJWT(t *testing.T) {
	result := ParseLicense("not-a-valid-jwt")
	if result != nil {
		t.Error("expected nil for invalid JWT")
	}

	result = ParseLicense("eyJhbGciOiJFUzI1NiJ9.eyJmZWF0dXJlcyI6W119.tampered")
	if result != nil {
		t.Error("expected nil for tampered JWT")
	}
}

func TestParseLicense_Valid(t *testing.T) {
	token := signTestLicense([]string{FeatureAuditLog}, nil)
	result := ParseLicense(token)
	if result == nil {
		t.Fatal("expected features for valid license")
	}
	if len(result) != 1 || result[0] != FeatureAuditLog {
		t.Errorf("expected [audit_log], got %v", result)
	}
}

func TestParseLicense_MultipleFeatures(t *testing.T) {
	expected := []string{FeatureAuditLog, "sso", "sla"}
	token := signTestLicense(expected, nil)
	result := ParseLicense(token)
	if result == nil {
		t.Fatal("expected features for valid license")
	}
	if len(result) != len(expected) {
		t.Fatalf("expected %d features, got %v", len(expected), result)
	}
}

func TestParseLicense_Expired(t *testing.T) {
	expired := time.Now().Add(-1 * time.Hour)
	token := signTestLicense([]string{FeatureAuditLog}, &expired)
	result := ParseLicense(token)
	if result != nil {
		t.Error("expected nil for expired license")
	}
}

func TestParseLicense_NotExpired(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	token := signTestLicense([]string{FeatureAuditLog}, &future)
	result := ParseLicense(token)
	if result == nil {
		t.Error("expected features for non-expired license")
	}
}

func TestParseLicense_WrongSigningMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, licenseClaims{
		Features: []string{FeatureAuditLog},
	})
	signed, err := token.SignedString([]byte("some-hmac-secret"))
	if err != nil {
		t.Fatal("failed to sign with HMAC:", err)
	}
	result := ParseLicense(signed)
	if result != nil {
		t.Error("expected nil for HMAC-signed token against ECDSA verifier")
	}
}

func TestParseLicense_WrongKey(t *testing.T) {
	otherKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal("failed to generate other key:", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, licenseClaims{
		Features: []string{FeatureAuditLog},
	})
	signed, err := token.SignedString(otherKey)
	if err != nil {
		t.Fatal("failed to sign with other key:", err)
	}

	result := ParseLicense(signed)
	if result != nil {
		t.Error("expected nil for token signed with different key")
	}
}

func TestParseLicense_NoFeatures(t *testing.T) {
	token := signTestLicense([]string{}, nil)
	result := ParseLicense(token)
	if result == nil {
		t.Fatal("expected non-nil for valid license with empty features")
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

func TestParseLicenseFull_Quota(t *testing.T) {
	quota := &Quota{Groups: 3, Applications: 5, Versions: 5, Users: 5}
	token := signTestLicenseWithQuota([]string{FeatureAuditLog}, quota, nil)
	result := ParseLicenseFull(token)
	if result == nil {
		t.Fatal("expected parsed license")
	}
	if result.Quota == nil {
		t.Fatal("expected quota claims")
	}
	if result.Quota.Groups != 3 || result.Quota.Applications != 5 || result.Quota.Versions != 5 || result.Quota.Users != 5 {
		t.Errorf("unexpected quota: %+v", result.Quota)
	}
	if len(result.Features) != 1 || result.Features[0] != FeatureAuditLog {
		t.Errorf("expected [audit_log], got %v", result.Features)
	}
}

func TestParseLicenseFull_NoQuota(t *testing.T) {
	token := signTestLicense([]string{FeatureAuditLog}, nil)
	result := ParseLicenseFull(token)
	if result == nil {
		t.Fatal("expected parsed license")
	}
	if result.Quota != nil {
		t.Errorf("expected nil quota, got %+v", result.Quota)
	}
}

func signTestLicenseSubject(subject string, features []string, expiresAt *time.Time) string {
	claims := licenseClaims{
		Features: features,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: subject,
		},
	}
	if expiresAt != nil {
		claims.ExpiresAt = jwt.NewNumericDate(*expiresAt)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	signed, err := token.SignedString(parseTestPrivateKey())
	if err != nil {
		panic("failed to sign test license: " + err.Error())
	}
	return signed
}

func TestParseLicense_SubjectBinding(t *testing.T) {
	t.Setenv("SSC_SITE_NAME", "atmos-test")

	// Bound to the matching slug => valid.
	token := signTestLicenseSubject("atmos-test", []string{FeatureAuditLog}, nil)
	if result := ParseLicense(token); result == nil {
		t.Fatal("expected license matching SSC_SITE_NAME to be accepted")
	}

	// Bound to a different slug => rejected.
	other := signTestLicenseSubject("other-tenant", []string{FeatureAuditLog}, nil)
	if result := ParseLicense(other); result != nil {
		t.Errorf("expected license bound to another tenant to be rejected, got %v", result)
	}
}

func TestParseLicense_SubjectBindingNoSiteName(t *testing.T) {
	os.Unsetenv("SSC_SITE_NAME")

	// A bound license must be rejected when the instance has no slug set.
	token := signTestLicenseSubject("atmos-test", []string{FeatureAuditLog}, nil)
	if result := ParseLicense(token); result != nil {
		t.Errorf("expected bound license rejected when SSC_SITE_NAME is empty, got %v", result)
	}

	// An unbound license (no sub) is accepted regardless.
	unbound := signTestLicense([]string{FeatureAuditLog}, nil)
	if result := ParseLicense(unbound); result == nil {
		t.Error("expected unbound license accepted when SSC_SITE_NAME is empty")
	}
}

func TestPublicKeyHex_EnvOverride(t *testing.T) {
	cachedPublicKey = nil
	t.Setenv("SSC_LICENSE_PUBLIC_KEY_HEX", "abcd")
	if got := PublicKeyHex(); got != "abcd" {
		t.Errorf("expected env override, got %s", got)
	}
	cachedPublicKey = nil
	os.Unsetenv("SSC_LICENSE_PUBLIC_KEY_HEX")
	if got := PublicKeyHex(); got != ecdsaPublicKeyHex {
		t.Errorf("expected compiled-in default, got %s", got)
	}
}
