package middleware

import (
    "context"
    "net/http"

    "github.com/google/uuid"
    "vault_api/internal/crypto"
    "vault_api/internal/repository"
)

func RequireAuth(jwtSecret string, sessions repository.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwt, err := crypto.GetBearerToken(r.Header)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			claims, err := crypto.ValidateAccessToken(jwt, jwtSecret)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			session, err := sessions.GetByID(r.Context(), claims.SessionID)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if session.UserID != claims.UserID {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, sessionIDKey, claims.SessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type contextKey string

const (
    userIDKey    contextKey = "userID"
    sessionIDKey contextKey = "sessionID"
)

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
    id, ok := ctx.Value(userIDKey).(uuid.UUID)
    return id, ok
}

func SessionIDFromContext(ctx context.Context) (uuid.UUID, bool) {
    id, ok := ctx.Value(sessionIDKey).(uuid.UUID)
    return id, ok
}