package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"vault_api/internal/api/middleware"
	"vault_api/internal/requestmeta"
	"vault_api/internal/service"
)

func (h *Handler) GenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	codes, err := h.recoveryService.GenerateRecoveryCodes(r.Context(), userID, auditContextFromRequest(r))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRecoveryRequiresMFA):
			http.Error(w, "mfa must be enabled before generating recovery codes", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"recovery_codes": codes,
	})
}

func (h *Handler) VerifyRecovery(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		RecoveryCode string `json:"recovery_code"`
		DeviceName   string `json:"device_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.RecoveryCode) == "" {
		http.Error(w, "email, password, and recovery_code are required", http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := h.recoveryService.VerifyRecoveryLogin(
		r.Context(),
		req.Email,
		req.Password,
		req.RecoveryCode,
		service.LoginDeviceInfo{
			DeviceName: req.DeviceName,
			IPAddress:  requestmeta.ClientIP(r),
			UserAgent:  r.Header.Get("User-Agent"),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
		case errors.Is(err, service.ErrInvalidRecoveryCode):
			http.Error(w, "invalid recovery code", http.StatusUnauthorized)
		case errors.Is(err, service.ErrRecoveryRequiresMFA):
			http.Error(w, "mfa must be enabled to use recovery codes", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
