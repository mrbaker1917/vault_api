package api

import (
	"net/http"
	"vault_api/internal/service"
	"vault_api/internal/api/handlers"
	"vault_api/internal/repository"
)

type Deps struct {
	Users repository.UserRepository
	Sessions repository.SessionRepository
	JWTSecret string
}

func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	auth := service.NewAuthService(deps.Users, deps.Sessions, deps.JWTSecret)
	h := handlers.NewHandler(auth)
	mux.HandleFunc("POST /api/v1/auth/signup", h.Signup)

	return mux
}
