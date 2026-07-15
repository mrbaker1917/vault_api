package crypto

import "errors"

const (
	EncryptedBlobVersion1  byte = 0x01
	MaxEncryptedBlobSize        = 1 << 20 // 1 MB
	MinEncryptedBlobSize        = 2       // version byte + payload
)

var (
	ErrEmptyEncryptedBlob     = errors.New("encrypted data is required")
	ErrEncryptedBlobTooLarge  = errors.New("encrypted data too large")
	ErrUnsupportedBlobVersion = errors.New("unsupported encrypted blob version")
)

func ValidateEncryptedBlob(data []byte) error {
	if len(data) == 0 {
		return ErrEmptyEncryptedBlob
	}
	if len(data) > MaxEncryptedBlobSize {
		return ErrEncryptedBlobTooLarge
	}
	if len(data) < MinEncryptedBlobSize {
		return ErrUnsupportedBlobVersion
	}
	if data[0] != EncryptedBlobVersion1 {
		return ErrUnsupportedBlobVersion
	}
	return nil
}
