package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
	"vault_api/internal/config"
	"vault_api/internal/redis"
	"vault_api/internal/repository"
	"vault_api/internal/service"
	"vault_api/internal/sessioncache"
)

func main() {
	email := flag.String("email", "", "account email")
	disableMFA := flag.Bool("disable-mfa", false, "turn off the authenticator so sign-in only needs the new password")
	flag.Parse()

	if strings.TrimSpace(*email) == "" {
		fmt.Fprintln(os.Stderr, "usage: reset-password --email user@example.com [--disable-mfa]")
		os.Exit(2)
	}

	password, err := readNewPassword()
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pg, err := repository.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pg.Close()

	var sessions repository.SessionRepository = repository.NewSessionRepository(pg)
	if redisURL := strings.TrimSpace(cfg.RedisURL); redisURL != "" {
		redisCtx, redisCancel := context.WithTimeout(ctx, 5*time.Second)
		client, err := redis.NewClient(redisCtx, redisURL)
		redisCancel()
		if err != nil {
			log.Printf("redis unavailable; sessions will be revoked in the database only: %v", err)
		} else {
			defer client.Close()
			sessions = repository.NewCachedSessionRepository(sessions, sessioncache.New(client, sessioncache.DefaultTTL))
		}
	}

	auth := service.NewAuthService(
		repository.NewUserRepository(pg),
		sessions,
		cfg.JWTSecret,
		service.NewAuditService(repository.NewAuditLogRepository(pg)),
		nil,
	)

	user, err := auth.ResetAccountPassword(ctx, *email, password, *disableMFA)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			log.Fatalf("no account for %s", strings.TrimSpace(*email))
		}
		log.Fatal(err)
	}

	fmt.Printf("Reset the account password for %s.\n", user.Email)
	fmt.Println("Signed out existing sessions.")
	fmt.Println("The vault master password is unchanged.")
	if user.MfaEnabled {
		fmt.Println("The authenticator is still on. Sign-in needs a current code from that device.")
		fmt.Println("Run again with --disable-mfa if that device is gone and there is no recovery code.")
	} else if *disableMFA {
		fmt.Println("The authenticator was turned off.")
	}
}

func readNewPassword() (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", errors.New("enter the new password on stdin, or run this in a terminal")
		}
		password := scanner.Text()
		if strings.TrimSpace(password) == "" {
			return "", errors.New("password is empty")
		}
		return password, nil
	}

	fmt.Fprint(os.Stderr, "New password: ")
	first, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	fmt.Fprint(os.Stderr, "Confirm password: ")
	second, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	if string(first) != string(second) {
		return "", errors.New("passwords do not match")
	}
	if strings.TrimSpace(string(first)) == "" {
		return "", errors.New("password is empty")
	}
	return string(first), nil
}
