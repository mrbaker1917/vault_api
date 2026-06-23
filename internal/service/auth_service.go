package service

import (
	"context"
	"fmt"
	"vault_api/internal/crypto"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
)

type AuthService struct {
	users     repository.UserRepository
	sessions  repository.SessionRepository
	jwtSecret string
}

func NewAuthService(users repository.UserRepository, sessions repository.SessionRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		sessions:  sessions,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Signup(ctx context.Context, email, password string) (domain.User, error) {
	hashPassword, err := crypto.HashPassword(password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	user := domain.User{
		Email: email,
		PasswordHash: hashPassword,
	}
	user, err = s.users.Create(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

