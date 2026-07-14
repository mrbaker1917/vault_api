package crypto

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

const recoveryCodeCount = 10
const recoveryCodeBytes = 5 // 8 base32 chars per segment

func GenerateRecoveryCode() (string, error) {
	b := make([]byte, recoveryCodeBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	encoded := strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), "=")
	if len(encoded) < 8 {
		return encoded, nil
	}
	return encoded[:4] + "-" + encoded[4:8], nil
}

func RecoveryCodeCount() int {
	return recoveryCodeCount
}
