package handlers

import (
	"net/http"
	"encoding/json"
	"errors"
	"vault_api/internal/service"
	"vault_api/internal/api/middleware"
	"strings"
)

type Handler struct {
	authService *service.AuthService
	vaultService *service.VaultService
}

func NewHandler(auth *service.AuthService, vault *service.VaultService) *Handler {
    return &Handler{authService: auth, vaultService: vault}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Signup(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			http.Error(w, "email already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":    user.ID.String(),
		"email": user.Email,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	
	accessToken, refreshToken, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)  // 401
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)  // 500
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "id": userID.String(),
    })
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		http.Error(w, "refresh token is required", http.StatusBadRequest)
		return
	}

	accessToken, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := middleware.SessionIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.authService.Logout(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "logged out successfully",
	})
}
	
