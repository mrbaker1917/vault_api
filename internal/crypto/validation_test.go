package crypto

import (
	"errors"
	"testing"
)

func TestValidateEncryptedBlob(t *testing.T) {
	t.Run("accepts valid v1 blob", func(t *testing.T) {
		if err := ValidateEncryptedBlob([]byte{EncryptedBlobVersion1, 0xAA, 0xBB}); err != nil {
			t.Fatalf("expected valid blob, got %v", err)
		}
	})

	t.Run("rejects empty blob", func(t *testing.T) {
		err := ValidateEncryptedBlob(nil)
		if !errors.Is(err, ErrEmptyEncryptedBlob) {
			t.Fatalf("expected ErrEmptyEncryptedBlob, got %v", err)
		}
	})

	t.Run("rejects oversized blob", func(t *testing.T) {
		data := make([]byte, MaxEncryptedBlobSize+1)
		data[0] = EncryptedBlobVersion1
		err := ValidateEncryptedBlob(data)
		if !errors.Is(err, ErrEncryptedBlobTooLarge) {
			t.Fatalf("expected ErrEncryptedBlobTooLarge, got %v", err)
		}
	})

	t.Run("rejects unsupported version", func(t *testing.T) {
		err := ValidateEncryptedBlob([]byte{0x02, 0x01})
		if !errors.Is(err, ErrUnsupportedBlobVersion) {
			t.Fatalf("expected ErrUnsupportedBlobVersion, got %v", err)
		}
	})
}
