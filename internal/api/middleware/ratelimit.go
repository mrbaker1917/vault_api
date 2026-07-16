package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	authRateLimitCount  = 10
	authRateLimitWindow = time.Minute
)

var authRateLimiter = newIPRateLimiter(authRateLimitCount, authRateLimitWindow)

// ResetAuthRateLimiterForTests clears auth rate limit state between tests.
func ResetAuthRateLimiterForTests() {
	authRateLimiter = newIPRateLimiter(authRateLimitCount, authRateLimitWindow)
}

func AuthRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAuthRateLimitedRoute(r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		if !authRateLimiter.allow(clientIP(r)) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAuthRateLimitedRoute(method, path string) bool {
	if method != http.MethodPost {
		return false
	}

	switch path {
	case "/api/v1/auth/signup", "/api/v1/auth/login", "/api/v1/auth/refresh":
		return true
	default:
		return false
	}
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i != -1 {
		return host[:i]
	}
	return host
}

type ipRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string][]time.Time
}

func newIPRateLimiter(limit int, window time.Duration) *ipRateLimiter {
	return &ipRateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string][]time.Time),
	}
}

func (l *ipRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	hits := l.buckets[key]
	active := hits[:0]
	for _, hit := range hits {
		if hit.After(cutoff) {
			active = append(active, hit)
		}
	}

	if len(active) >= l.limit {
		l.buckets[key] = active
		return false
	}

	active = append(active, now)
	l.buckets[key] = active
	return true
}
