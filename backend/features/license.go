package features

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ecdsaPublicKeyHex = "3059301306072a8648ce3d020106082a8648ce3d030107034200047aa16aee6884f5df9ad9ec38b532afd558cd330f4b139739527c8bc4943dc8a8654f5e96d5c9431153cb083d893331830c2ad5ff797e6f6f9a885e99d9decb16"

var cachedPublicKey *ecdsa.PublicKey

// PublicKeyHex returns the license signing public key in DER PKIX hex.
// SSC_LICENSE_PUBLIC_KEY_HEX overrides the compiled-in default so managed
// instances can validate licenses issued by the control plane.
func PublicKeyHex() string {
	if v := os.Getenv("SSC_LICENSE_PUBLIC_KEY_HEX"); v != "" {
		return v
	}
	return ecdsaPublicKeyHex
}

func publicKey() *ecdsa.PublicKey {
	if cachedPublicKey != nil {
		return cachedPublicKey
	}

	der, err := hex.DecodeString(PublicKeyHex())
	if err != nil {
		return nil
	}

	parsed, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil
	}

	key, ok := parsed.(*ecdsa.PublicKey)
	if !ok {
		return nil
	}

	cachedPublicKey = key
	return key
}

type Quota struct {
	Groups       int `json:"groups"`
	Applications int `json:"applications"`
	Versions     int `json:"versions"`
	Users        int `json:"users"`
}

type licenseClaims struct {
	Features []string `json:"features"`
	Quota    *Quota   `json:"quota,omitempty"`
	jwt.RegisteredClaims
}

type parsedLicense struct {
	Features []string
	Quota    *Quota
}

// expectedSubject returns the tenant slug a bound license must match. It reads
// SSC_SITE_NAME (the instance's slug) when set. Tenant-bound licenses carry a
// `sub` claim; an unbound license (no `sub`) is accepted on any instance.
func expectedSubject() string {
	return os.Getenv("SSC_SITE_NAME")
}

func ParseLicenseFull(licenseKey string) *parsedLicense {
	key := strings.TrimSpace(licenseKey)
	if key == "" {
		return nil
	}

	pub := publicKey()
	if pub == nil {
		return nil
	}

	token, err := jwt.ParseWithClaims(key, &licenseClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return pub, nil
	})
	if err != nil {
		return nil
	}

	claims, ok := token.Claims.(*licenseClaims)
	if !ok || !token.Valid {
		return nil
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil
	}

	// Tenant-bound licenses (claim `sub`) must match this instance's slug.
	// Unbound licenses (no `sub`) are accepted on any instance for
	// backward-compatibility with standalone/self-hosted installs.
	if claims.Subject != "" {
		if want := expectedSubject(); want == "" || claims.Subject != want {
			return nil
		}
	}

	return &parsedLicense{
		Features: claims.Features,
		Quota:    claims.Quota,
	}
}

func ParseLicense(licenseKey string) []string {
	pl := ParseLicenseFull(licenseKey)
	if pl == nil {
		return nil
	}
	return pl.Features
}
