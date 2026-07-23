package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"vault_api/internal/api/middleware"
	"vault_api/internal/service"
)

func (h *Handler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	setup, err := h.mfaService.EnableMFA(r.Context(), userID, auditContextFromRequest(r), req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAAlreadyEnabled):
			http.Error(w, "mfa already enabled", http.StatusConflict)
		case errors.Is(err, service.ErrInvalidCredentials):
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
		default:
			http.Error(w, "failed to enable mfa", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"secret":      setup.Secret,
		"otpauth_url": setup.OTPAuthURL,
	})
}

func (h *Handler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Code) == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	err := h.mfaService.VerifyMFA(r.Context(), userID, auditContextFromRequest(r), req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFAAlreadyEnabled):
			http.Error(w, "mfa already enabled", http.StatusConflict)
		case errors.Is(err, service.ErrMFANotEnabled):
			http.Error(w, "mfa not enabled", http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidTOTPCode):
			http.Error(w, "invalid totp code", http.StatusUnauthorized)
		default:
			http.Error(w, "failed to verify mfa", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Code) == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	err := h.mfaService.DisableMFA(r.Context(), userID, auditContextFromRequest(r), req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMFANotEnabled):
			http.Error(w, "mfa not enabled", http.StatusBadRequest)
		case errors.Is(err, service.ErrInvalidTOTPCode):
			http.Error(w, "invalid totp code", http.StatusUnauthorized)
		default:
			http.Error(w, "failed to disable mfa", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
