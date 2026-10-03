package middleware

import (
	"net/http"
	"strings"

	"vault_api/internal/crypto"
)

// DefaultMaxRequestBodyBytes allows one MaxEncryptedBlobSize payload plus base64 and JSON framing.
const DefaultMaxRequestBodyBytes = 2 * crypto.MaxEncryptedBlobSize

// LimitRequestBody wraps POST/PUT/PATCH bodies with http.MaxBytesReader to prevent DoS from huge payloads.
func LimitRequestBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes <= 0 || !requestBodyAllowed(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			if r.ContentLength > maxBytes {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
				return
			}
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func requestBodyAllowed(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}
