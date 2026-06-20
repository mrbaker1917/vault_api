package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessTokenClaims struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

func MakeAccessToken(userID uuid.UUID, sessionID uuid.UUID, secret string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		return "", errors.New("ttl must be greater than 0")
	}
	now := time.Now()
	newJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "vault_api",
		Subject:   userID.String(),
		ID:        sessionID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	})

	ss, err := newJWT.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashToken(token string) (string, error) {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:]), nil
}

func ValidateAccessToken(tokenString, secret string) (AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})

	if err != nil {
		return AccessTokenClaims{}, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return AccessTokenClaims{}, errors.New("invalid token")
	}

	if claims.ID == "" {
		return AccessTokenClaims{}, errors.New("session ID is required")
	}

	idStr := claims.Subject

	userID, err := uuid.Parse(idStr)
	if err != nil {
		return AccessTokenClaims{}, errors.New("unable to parse")
	}

	sessionID, err := uuid.Parse(claims.ID)
	if err != nil {
		return AccessTokenClaims{}, errors.New("unable to parse")
	}

	return AccessTokenClaims{
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if auth == "" {
		return "", errors.New("no authorization header found")
	}

	tokenString, found := strings.CutPrefix(auth, "Bearer ")
	if !found {
		return "", errors.New("no authorization prefix found")
	}

	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", errors.New("no token found")
	}

	return tokenString, nil
}
