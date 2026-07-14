package api

import (
	"net/http"
	"vault_api/internal/service"
	"vault_api/internal/api/handlers"
	"vault_api/internal/repository"
	"vault_api/internal/api/middleware"
)

type Deps struct {
	Users              repository.UserRepository
	Sessions           repository.SessionRepository
	RecoveryCodes      repository.RecoveryCodeRepository
	JWTSecret          string
	VaultItems         repository.VaultItemRepository
	CORSAllowedOrigins []string
}

func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	auth := service.NewAuthService(deps.Users, deps.Sessions, deps.JWTSecret)
	vault := service.NewVaultService(deps.VaultItems)
	mfa := service.NewMFAService(deps.Users)
	recovery := service.NewRecoveryService(deps.Users, deps.RecoveryCodes, auth)
	h := handlers.NewHandler(auth, vault, mfa, recovery)
	mux.HandleFunc("POST /api/v1/auth/signup", h.Signup)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.Handle("GET /api/v1/me", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.Me)))
	mux.HandleFunc("POST /api/v1/auth/refresh", h.Refresh)
	mux.Handle("POST /api/v1/auth/logout", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.Logout)))
	mux.Handle("POST /api/v1/vault/items", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.CreateItem)))
	mux.Handle("GET /api/v1/vault/items", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.ListItems)))
	mux.Handle("GET /api/v1/vault/items/{id}", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.GetItem)))
	mux.Handle("PUT /api/v1/vault/items/{id}", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.UpdateItem)))
	mux.Handle("DELETE /api/v1/vault/items/{id}", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.DeleteItem)))
	mux.Handle("POST /api/v1/vault/items/{id}/restore", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.RestoreItem)))
	mux.Handle("GET /api/v1/auth/sessions", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.ListSessions)))
	mux.Handle("DELETE /api/v1/auth/sessions/{id}", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.RevokeSession)))
	mux.Handle("POST /api/v1/mfa/enable", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.EnableMFA)))
	mux.Handle("POST /api/v1/mfa/verify", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.VerifyMFA)))
	mux.Handle("POST /api/v1/mfa/disable", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.DisableMFA)))
	mux.Handle("POST /api/v1/recovery/generate", middleware.RequireAuth(deps.JWTSecret, deps.Sessions)(http.HandlerFunc(h.GenerateRecoveryCodes)))
	mux.HandleFunc("POST /api/v1/recovery/verify", h.VerifyRecovery)

	return middleware.Chain(mux,
		middleware.Recover,
		middleware.LogRequests,
		middleware.CORS(deps.CORSAllowedOrigins),
		middleware.AuthRateLimit,
	)
}
