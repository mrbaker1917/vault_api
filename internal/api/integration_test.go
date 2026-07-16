//go:build integration

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"vault_api/internal/testutil/integration"
)

func TestIntegrationAuthSignupLoginLogout(t *testing.T) {
	handler, cleanup := integration.NewTestRouter(t)
	defer cleanup()

	email := fmt.Sprintf("auth-%d@example.com", time.Now().UnixNano())
	password := "secure-password-123"

	signup := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/auth/signup",
		Body: map[string]string{
			"email":    email,
			"password": password,
		},
	})
	if signup.Status != http.StatusCreated {
		t.Fatalf("signup status = %d, body = %s", signup.Status, signup.Body)
	}

	dupSignup := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/auth/signup",
		Body: map[string]string{
			"email":    email,
			"password": password,
		},
	})
	if dupSignup.Status != http.StatusConflict {
		t.Fatalf("duplicate signup status = %d, body = %s", dupSignup.Status, dupSignup.Body)
	}

	login := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/auth/login",
		Body: map[string]string{
			"email":       email,
			"password":    password,
			"device_name": "integration-test",
		},
	})
	if login.Status != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", login.Status, login.Body)
	}

	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	integration.DecodeJSON(t, login, &tokens)
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatalf("expected tokens, got %+v", tokens)
	}

	me := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodGet,
		Path:   "/api/v1/me",
		Token:  tokens.AccessToken,
	})
	if me.Status != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", me.Status, me.Body)
	}

	logout := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/auth/logout",
		Token:  tokens.AccessToken,
	})
	if logout.Status != http.StatusNoContent {
		t.Fatalf("logout status = %d, body = %s", logout.Status, logout.Body)
	}

	meAfterLogout := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodGet,
		Path:   "/api/v1/me",
		Token:  tokens.AccessToken,
	})
	if meAfterLogout.Status != http.StatusUnauthorized {
		t.Fatalf("me after logout status = %d, body = %s", meAfterLogout.Status, meAfterLogout.Body)
	}
}

func TestIntegrationVaultCRUD(t *testing.T) {
	handler, cleanup := integration.NewTestRouter(t)
	defer cleanup()

	token := signupAndLogin(t, handler, "vault")

	create := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/vault/items",
		Token:  token,
		Body: map[string]any{
			"encrypted_data": integration.ValidEncryptedBlob(),
			"item_type":      "login",
			"title":          "GitHub",
			"folder":         "Work",
			"tags":           []string{"dev", "login"},
		},
	})
	if create.Status != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", create.Status, create.Body)
	}

	var item struct {
		ID      string `json:"ID"`
		Title   string `json:"Title"`
		Version int32  `json:"Version"`
	}
	integration.DecodeJSON(t, create, &item)
	if item.ID == "" {
		t.Fatalf("expected item id in response: %s", create.Body)
	}
	if item.Version != 1 {
		t.Fatalf("expected version 1, got %d", item.Version)
	}

	list := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodGet,
		Path:   "/api/v1/vault/items?folder=Work&limit=10",
		Token:  token,
	})
	if list.Status != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", list.Status, list.Body)
	}

	var listResp struct {
		Items  []json.RawMessage `json:"items"`
		Total  int64             `json:"total"`
		Limit  int32             `json:"limit"`
		Offset int32             `json:"offset"`
	}
	integration.DecodeJSON(t, list, &listResp)
	if listResp.Total != 1 || len(listResp.Items) != 1 {
		t.Fatalf("expected one listed item, got total=%d len=%d body=%s", listResp.Total, len(listResp.Items), list.Body)
	}

	get := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodGet,
		Path:   "/api/v1/vault/items/" + item.ID,
		Token:  token,
	})
	if get.Status != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", get.Status, get.Body)
	}

	update := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPut,
		Path:   "/api/v1/vault/items/" + item.ID,
		Token:  token,
		Body: map[string]any{
			"encrypted_data": integration.ValidEncryptedBlob(0xCC),
			"item_type":      "login",
			"title":          "GitHub Updated",
			"version":        item.Version,
		},
	})
	if update.Status != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", update.Status, update.Body)
	}

	var updated struct {
		Title   string `json:"Title"`
		Version int32  `json:"Version"`
	}
	integration.DecodeJSON(t, update, &updated)
	if updated.Version != 2 {
		t.Fatalf("expected version 2, got %d", updated.Version)
	}

	deleteResp := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodDelete,
		Path:   "/api/v1/vault/items/" + item.ID,
		Token:  token,
		Body: map[string]any{
			"version": updated.Version,
		},
	})
	if deleteResp.Status != http.StatusOK {
		t.Fatalf("delete status = %d, body = %s", deleteResp.Status, deleteResp.Body)
	}

	var deleted struct {
		Version int32 `json:"Version"`
	}
	integration.DecodeJSON(t, deleteResp, &deleted)

	getAfterDelete := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodGet,
		Path:   "/api/v1/vault/items/" + item.ID,
		Token:  token,
	})
	if getAfterDelete.Status != http.StatusNotFound {
		t.Fatalf("get after delete status = %d, body = %s", getAfterDelete.Status, getAfterDelete.Body)
	}

	restore := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/vault/items/" + item.ID + "/restore",
		Token:  token,
		Body: map[string]any{
			"version": deleted.Version,
		},
	})
	if restore.Status != http.StatusOK {
		t.Fatalf("restore status = %d, body = %s", restore.Status, restore.Body)
	}

	getAfterRestore := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodGet,
		Path:   "/api/v1/vault/items/" + item.ID,
		Token:  token,
	})
	if getAfterRestore.Status != http.StatusOK {
		t.Fatalf("get after restore status = %d, body = %s", getAfterRestore.Status, getAfterRestore.Body)
	}
}

func TestIntegrationMFAFlow(t *testing.T) {
	handler, cleanup := integration.NewTestRouter(t)
	defer cleanup()

	email := fmt.Sprintf("mfa-%d@example.com", time.Now().UnixNano())
	password := "secure-password-123"
	token := signupAndLoginWithCredentials(t, handler, email, password)

	enable := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/mfa/enable",
		Token:  token,
	})
	if enable.Status != http.StatusOK {
		t.Fatalf("enable mfa status = %d, body = %s", enable.Status, enable.Body)
	}

	var setup struct {
		Secret string `json:"secret"`
	}
	integration.DecodeJSON(t, enable, &setup)
	if setup.Secret == "" {
		t.Fatal("expected mfa secret")
	}

	code, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}

	verify := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/mfa/verify",
		Token:  token,
		Body: map[string]string{
			"code": code,
		},
	})
	if verify.Status != http.StatusNoContent {
		t.Fatalf("verify mfa status = %d, body = %s", verify.Status, verify.Body)
	}

	loginNoTOTP := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/auth/login",
		Body: map[string]string{
			"email":    email,
			"password": password,
		},
	})
	if loginNoTOTP.Status != http.StatusUnauthorized {
		t.Fatalf("login without totp status = %d, body = %s", loginNoTOTP.Status, loginNoTOTP.Body)
	}

	var mfaRequired struct {
		MFARequired bool `json:"mfa_required"`
	}
	integration.DecodeJSON(t, loginNoTOTP, &mfaRequired)
	if !mfaRequired.MFARequired {
		t.Fatalf("expected mfa_required=true, body=%s", loginNoTOTP.Body)
	}

	loginCode, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatalf("generate login totp code: %v", err)
	}

	loginWithTOTP := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/auth/login",
		Body: map[string]string{
			"email":     email,
			"password":  password,
			"totp_code": loginCode,
		},
	})
	if loginWithTOTP.Status != http.StatusOK {
		t.Fatalf("login with totp status = %d, body = %s", loginWithTOTP.Status, loginWithTOTP.Body)
	}
}

func TestIntegrationAuditLogs(t *testing.T) {
	handler, cleanup := integration.NewTestRouter(t)
	defer cleanup()

	token := signupAndLogin(t, handler, "audit")

	create := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/vault/items",
		Token:  token,
		Body: map[string]any{
			"encrypted_data": integration.ValidEncryptedBlob(),
			"item_type":      "note",
			"title":          "Audit Me",
		},
	})
	if create.Status != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", create.Status, create.Body)
	}

	logs := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodGet,
		Path:   "/api/v1/audit/logs?limit=20",
		Token:  token,
	})
	if logs.Status != http.StatusOK {
		t.Fatalf("audit logs status = %d, body = %s", logs.Status, logs.Body)
	}

	var entries []struct {
		Action string `json:"action"`
	}
	integration.DecodeJSON(t, logs, &entries)
	if len(entries) == 0 {
		t.Fatalf("expected audit log entries, body=%s", logs.Body)
	}

	foundSignup := false
	foundLogin := false
	foundCreate := false
	for _, entry := range entries {
		switch entry.Action {
		case "auth.signup":
			foundSignup = true
		case "auth.login":
			foundLogin = true
		case "vault.item.create":
			foundCreate = true
		}
	}
	if !foundSignup || !foundLogin || !foundCreate {
		t.Fatalf("expected signup/login/create audit events, got %+v", entries)
	}
}

func signupAndLogin(t *testing.T, handler http.Handler, prefix string) string {
	t.Helper()
	email := fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano())
	return signupAndLoginWithCredentials(t, handler, email, "secure-password-123")
}

func signupAndLoginWithCredentials(t *testing.T, handler http.Handler, email, password string) string {
	t.Helper()

	signup := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/auth/signup",
		Body: map[string]string{
			"email":    email,
			"password": password,
		},
	})
	if signup.Status != http.StatusCreated {
		t.Fatalf("signup status = %d, body = %s", signup.Status, signup.Body)
	}

	login := integration.DoJSON(t, handler, integration.JSONRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/auth/login",
		Body: map[string]string{
			"email":    email,
			"password": password,
		},
	})
	if login.Status != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", login.Status, login.Body)
	}

	var tokens struct {
		AccessToken string `json:"access_token"`
	}
	integration.DecodeJSON(t, login, &tokens)
	if tokens.AccessToken == "" {
		t.Fatal("expected access token")
	}
	return tokens.AccessToken
}
