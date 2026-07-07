package api

import (
	"net/http"
	"vault_api/internal/service"
	"vault_api/internal/api/handlers"
	"vault_api/internal/repository"
	"vault_api/internal/api/middleware"
)

type Deps struct {
	Users repository.UserRepository
	Sessions repository.SessionRepository
	JWTSecret string
	VaultItems repository.VaultItemRepository
}

func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	auth := service.NewAuthService(deps.Users, deps.Sessions, deps.JWTSecret)
	vault := service.NewVaultService(deps.VaultItems)
	h := handlers.NewHandler(auth, vault)
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
	
	return mux
}
