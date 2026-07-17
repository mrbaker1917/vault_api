package crypto

import "errors"

const (
	EncryptedBlobVersion1  byte = 0x01
	MaxEncryptedBlobSize        = 1 << 20 // 1 MB
	MinEncryptedBlobSize        = 2       // version byte + payload
	MaxEncryptedItemKeySize     = 4096
)

var (
	ErrEmptyEncryptedBlob     = errors.New("encrypted data is required")
	ErrEncryptedBlobTooLarge  = errors.New("encrypted data too large")
	ErrUnsupportedBlobVersion = errors.New("unsupported encrypted blob version")
	ErrEmptyEncryptedItemKey  = errors.New("encrypted item key is required")
	ErrEncryptedItemKeyTooLarge = errors.New("encrypted item key too large")
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

func ValidateEncryptedItemKey(data []byte) error {
	if len(data) == 0 {
		return ErrEmptyEncryptedItemKey
	}
	if len(data) > MaxEncryptedItemKeySize {
		return ErrEncryptedItemKeyTooLarge
	}
	return nil
}
