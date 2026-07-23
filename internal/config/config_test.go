package config

import "testing"

func TestValidateRejectsWeakJWTSecretInProduction(t *testing.T) {
	cfg := Config{AppEnv: "production", JWTSecret: "change-me"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for change-me")
	}

	cfg.JWTSecret = "dev-secret"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for dev-secret")
	}

	cfg.JWTSecret = "short"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for short secret")
	}
}

func TestValidateAllowsWeakJWTSecretInDevelopment(t *testing.T) {
	cfg := Config{AppEnv: "development", JWTSecret: "dev-secret"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected development to allow weak secret: %v", err)
	}
}

func TestValidateAcceptsStrongJWTSecret(t *testing.T) {
	cfg := Config{AppEnv: "production", JWTSecret: "a-sufficiently-long-production-secret"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected strong secret to pass: %v", err)
	}
}
