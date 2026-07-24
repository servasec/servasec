package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

func GetEncryptionKey() ([]byte, error) {
	keyEnv := os.Getenv("SSC_FEATURES_ENCRYPTION_KEY")
	if keyEnv == "" {
		return nil, fmt.Errorf("SSC_FEATURES_ENCRYPTION_KEY environment variable is not set")
	}
	hash := sha256.Sum256([]byte(keyEnv))
	return hash[:], nil
}

func Encrypt(plaintext []byte, key []byte) (ciphertext string, iv string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("failed to create GCM: %w", err)
	}

	ivBytes := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, ivBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate IV: %w", err)
	}

	encrypted := gcm.Seal(nil, ivBytes, plaintext, nil)
	return base64.StdEncoding.EncodeToString(encrypted), base64.StdEncoding.EncodeToString(ivBytes), nil
}

func Decrypt(ciphertext string, iv string, key []byte) (string, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", fmt.Errorf("failed to decode IV: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, ivBytes, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
