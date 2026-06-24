package features

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ecdsaPublicKeyHex = "3059301306072a8648ce3d020106082a8648ce3d030107034200047aa16aee6884f5df9ad9ec38b532afd558cd330f4b139739527c8bc4943dc8a8654f5e96d5c9431153cb083d893331830c2ad5ff797e6f6f9a885e99d9decb16"

var cachedPublicKey *ecdsa.PublicKey

func publicKey() *ecdsa.PublicKey {
	if cachedPublicKey != nil {
		return cachedPublicKey
	}

	der, err := hex.DecodeString(ecdsaPublicKeyHex)
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

type licenseClaims struct {
	Features []string `json:"features"`
	jwt.RegisteredClaims
}

func ParseLicense(licenseKey string) []string {
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

	return claims.Features
}
