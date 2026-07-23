package crypto

import "testing"

func TestEncryptDecryptMFASecretRoundTrip(t *testing.T) {
	const master = "unit-test-master-secret-value"
	const secret = "JBSWY3DPEHPK3PXP"

	encrypted, err := EncryptMFASecret(secret, master)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if encrypted == secret {
		t.Fatal("expected ciphertext to differ from plaintext")
	}

	plain, err := DecryptMFASecret(encrypted, master)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if plain != secret {
		t.Fatalf("got %q, want %q", plain, secret)
	}
}

func TestDecryptMFASecretLegacyPlaintext(t *testing.T) {
	const legacy = "JBSWY3DPEHPK3PXP"
	plain, err := DecryptMFASecret(legacy, "any-key")
	if err != nil {
		t.Fatalf("decrypt legacy: %v", err)
	}
	if plain != legacy {
		t.Fatalf("got %q, want %q", plain, legacy)
	}
}

func TestDecryptMFASecretWrongKeyFails(t *testing.T) {
	encrypted, err := EncryptMFASecret("SECRET", "correct-key")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := DecryptMFASecret(encrypted, "wrong-key"); err == nil {
		t.Fatal("expected decrypt with wrong key to fail")
	}
}
