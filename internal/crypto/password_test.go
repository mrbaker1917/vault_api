package crypto

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidatePasswordStrength(t *testing.T) {
	t.Run("accepts strong password", func(t *testing.T) {
		if err := ValidatePasswordStrength("StrongPass123"); err != nil {
			t.Fatalf("expected valid password, got %v", err)
		}
	})

	t.Run("rejects short password", func(t *testing.T) {
		err := ValidatePasswordStrength("Short1A")
		if !errors.Is(err, ErrWeakPassword) {
			t.Fatalf("expected ErrWeakPassword, got %v", err)
		}
	})

	t.Run("rejects missing character classes", func(t *testing.T) {
		err := ValidatePasswordStrength("alllowercase123")
		if !errors.Is(err, ErrWeakPassword) {
			t.Fatalf("expected ErrWeakPassword, got %v", err)
		}
	})
}

func TestHIBPPasswordBreachChecker(t *testing.T) {
	sum := sha1.Sum([]byte("password"))
	hash := strings.ToUpper(hex.EncodeToString(sum[:]))
	prefix := hash[:5]
	suffix := hash[5:]

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/range/") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if strings.TrimPrefix(r.URL.Path, "/range/") == prefix {
			_, _ = w.Write([]byte(suffix + ":12345\r\n000000000000000000000000000000000:1\r\n"))
			return
		}
		_, _ = w.Write([]byte("000000000000000000000000000000000:1\r\n"))
	}))
	defer server.Close()

	checker := &HIBPPasswordBreachChecker{
		client:  server.Client(),
		baseURL: server.URL + "/range/",
	}

	breached, err := checker.IsBreached(context.Background(), "password")
	if err != nil {
		t.Fatalf("IsBreached: %v", err)
	}
	if !breached {
		t.Fatal("expected password to be marked breached")
	}

	breached, err = checker.IsBreached(context.Background(), "UniqueLocalPass987")
	if err != nil {
		t.Fatalf("IsBreached unique: %v", err)
	}
	if breached {
		t.Fatal("expected unique password not to be breached")
	}
}
