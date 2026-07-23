package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const mfaSecretPrefix = "enc:v1:"

// DeriveMFAKey derives a 32-byte AES key from the server master secret (JWT secret).
func DeriveMFAKey(masterSecret string) []byte {
	sum := sha256.Sum256([]byte("vault-api-mfa-v1:" + masterSecret))
	return sum[:]
}

// EncryptMFASecret encrypts a TOTP secret for storage. Returns a prefixed ciphertext string.
func EncryptMFASecret(plaintext, masterSecret string) (string, error) {
	plaintext = strings.TrimSpace(plaintext)
	if plaintext == "" {
		return "", fmt.Errorf("mfa secret is empty")
	}

	block, err := aes.NewCipher(DeriveMFAKey(masterSecret))
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}

	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return mfaSecretPrefix + base64.RawURLEncoding.EncodeToString(sealed), nil
}

// DecryptMFASecret decrypts a stored TOTP secret. Legacy plaintext values (no prefix) are returned as-is.
func DecryptMFASecret(stored, masterSecret string) (string, error) {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return "", fmt.Errorf("mfa secret is empty")
	}
	if !strings.HasPrefix(stored, mfaSecretPrefix) {
		return stored, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(stored, mfaSecretPrefix))
	if err != nil {
		return "", fmt.Errorf("decode mfa secret: %w", err)
	}

	block, err := aes.NewCipher(DeriveMFAKey(masterSecret))
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create gcm: %w", err)
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("mfa secret ciphertext too short")
	}

	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt mfa secret: %w", err)
	}
	return string(plain), nil
}
