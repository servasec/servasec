package features

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
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
	claims := licenseClaims{
		Features: features,
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
