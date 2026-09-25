package crypto

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"
)

const (
	MinPasswordLength = 12
	hibpRangeURL      = "https://api.pwnedpasswords.com/range/"
)

var (
	ErrWeakPassword             = errors.New("password does not meet requirements")
	ErrPasswordBreached         = errors.New("password has appeared in a data breach")
	ErrPasswordBreachCheckFailed = errors.New("password breach check failed")
)

type PasswordBreachChecker interface {
	IsBreached(ctx context.Context, password string) (bool, error)
}

type HIBPPasswordBreachChecker struct {
	client  *http.Client
	baseURL string
}

func NewHIBPPasswordBreachChecker(client *http.Client) *HIBPPasswordBreachChecker {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HIBPPasswordBreachChecker{
		client:  client,
		baseURL: hibpRangeURL,
	}
}

func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("%w: must be at least %d characters", ErrWeakPassword, MinPasswordLength)
	}

	var hasLower, hasUpper, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasLower || !hasUpper || !hasDigit {
		return fmt.Errorf("%w: must include uppercase, lowercase, and a number", ErrWeakPassword)
	}

	return nil
}

func (c *HIBPPasswordBreachChecker) IsBreached(ctx context.Context, password string) (bool, error) {
	sum := sha1.Sum([]byte(password))
	hash := strings.ToUpper(hex.EncodeToString(sum[:]))
	prefix := hash[:5]
	suffix := hash[5:]

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+prefix, nil)
	if err != nil {
		return false, fmt.Errorf("create hibp request: %w", err)
	}
	req.Header.Set("User-Agent", "vault-api")
	req.Header.Set("Add-Padding", "true")

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("hibp request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("hibp status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.EqualFold(parts[0], suffix) {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("read hibp response: %w", err)
	}

	return false, nil
}

func ValidatePassword(ctx context.Context, password string, checker PasswordBreachChecker) error {
	if err := ValidatePasswordStrength(password); err != nil {
		return err
	}
	if checker == nil {
		return nil
	}

	breached, err := checker.IsBreached(ctx, password)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPasswordBreachCheckFailed, err)
	}
	if breached {
		return ErrPasswordBreached
	}
	return nil
}
