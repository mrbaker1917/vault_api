package handlers

import (
	"net/http"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"vault_api/internal/service"
	"vault_api/internal/api/middleware"
	"vault_api/internal/requestmeta"
)

type Handler struct {
	authService     *service.AuthService
	vaultService    *service.VaultService
	mfaService      *service.MFAService
	recoveryService *service.RecoveryService
	auditService    *service.AuditService
}

func NewHandler(auth *service.AuthService, vault *service.VaultService, mfa *service.MFAService, recovery *service.RecoveryService, audit *service.AuditService) *Handler {
	return &Handler{
		authService:     auth,
		vaultService:    vault,
		mfaService:      mfa,
		recoveryService: recovery,
		auditService:    audit,
	}
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

	user, err := h.authService.Signup(r.Context(), req.Email, req.Password, auditContextFromRequest(r))
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
		Email      string `json:"email"`
		Password   string `json:"password"`
		DeviceName string `json:"device_name"`
		TotpCode   string `json:"totp_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.authService.Login(r.Context(), req.Email, req.Password, req.TotpCode, service.LoginDeviceInfo{
		DeviceName: req.DeviceName,
		IPAddress:  requestmeta.ClientIP(r),
		UserAgent:  r.Header.Get("User-Agent"),
	})

	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, service.ErrMFARequired) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error":        "mfa required",
				"mfa_required": true,
			})
			return
		}
		if errors.Is(err, service.ErrInvalidTOTPCode) {
			http.Error(w, "invalid totp code", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	sessionID, ok := middleware.SessionIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err := h.authService.Logout(r.Context(), sessionID, userID, auditContextFromRequest(r))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "logged out successfully",
	})
}
	
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	currentSessionID, ok := middleware.SessionIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessions, err := h.authService.ListSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type sessionResponse struct {
		ID         string `json:"id"`
		DeviceName string `json:"device_name"`
		IPAddress  string `json:"ip_address"`
		UserAgent  string `json:"user_agent"`
		CreatedAt  string `json:"created_at"`
		ExpiresAt  string `json:"expires_at"`
		IsCurrent  bool   `json:"is_current"`
	}

	response := make([]sessionResponse, 0, len(sessions))
	for _, session := range sessions {
		response = append(response, sessionResponse{
			ID:         session.ID.String(),
			DeviceName: session.DeviceName,
			IPAddress:  session.IPAddress,
			UserAgent:  session.UserAgent,
			CreatedAt:  session.CreatedAt.Format(time.RFC3339),
			ExpiresAt:  session.ExpiresAt.Format(time.RFC3339),
			IsCurrent:  session.ID == currentSessionID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	err = h.authService.RevokeSession(r.Context(), sessionID, userID, auditContextFromRequest(r))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}