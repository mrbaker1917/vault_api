package middleware

import (
	"log/slog"
	"net/http"

	"vault_api/internal/ratelimit"
	"vault_api/internal/requestmeta"
)

func AuthRateLimit(limiter ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isAuthRateLimitedRoute(r.Method, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			allowed, err := limiter.Allow(r.Context(), requestmeta.ClientIP(r))
			if err != nil {
				slog.Warn("rate limit check failed; allowing request", "error", err)
				next.ServeHTTP(w, r)
				return
			}
			if !allowed {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAuthRateLimitedRoute(method, path string) bool {
	if method != http.MethodPost {
		return false
	}

	switch path {
	case "/api/v1/auth/signup", "/api/v1/auth/login", "/api/v1/auth/refresh", "/api/v1/recovery/verify":
		return true
	default:
		return false
	}
}
