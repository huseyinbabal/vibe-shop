package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// registerHandlerFor wires a RegisterHandler to an AdminClient talking to the
// given users endpoint behavior — real HTTP all the way, no hand mocks.
func registerHandlerFor(t *testing.T, usersHandler http.HandlerFunc) http.HandlerFunc {
	t.Helper()
	issuer, _ := newAdminMock(t, usersHandler)
	return NewRegisterHandler(NewAdminClient(issuer, "vibe-shop-backend", "s")).Register
}

func postRegister(t *testing.T, h http.HandlerFunc, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body)))
	return rec
}

func TestRegister_ValidReturns201WithoutPassword(t *testing.T) {
	h := registerHandlerFor(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/admin/realms/vibe-shop/users/u-1")
		w.WriteHeader(http.StatusCreated)
	})

	rec := postRegister(t, h, `{"email":"yeni@vibe.shop","password":"parola123"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["id"] != "u-1" || body["email"] != "yeni@vibe.shop" {
		t.Errorf("body = %v, want id+email", body)
	}
	if strings.Contains(rec.Body.String(), "parola123") {
		t.Error("response leaked the password")
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	h := registerHandlerFor(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("admin API must not be called for invalid input")
	})

	cases := map[string]string{
		"invalid json":    `{"email":`,
		"missing email":   `{"password":"parola123"}`,
		"email without @": `{"email":"yanlis","password":"parola123"}`,
		"short password":  `{"email":"a@b.c","password":"kisa"}`,
	}
	for name, body := range cases {
		if rec := postRegister(t, h, body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
	}
}

func TestRegister_DuplicateReturns409(t *testing.T) {
	h := registerHandlerFor(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errorMessage":"User exists"}`, http.StatusConflict)
	})

	rec := postRegister(t, h, `{"email":"var@vibe.shop","password":"parola123"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"error":"email already registered"}` {
		t.Errorf("body = %s", got)
	}
}

func TestRegister_KeycloakDownReturns503(t *testing.T) {
	h := registerHandlerFor(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	rec := postRegister(t, h, `{"email":"a@b.c","password":"parola123"}`)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"error":"registration unavailable"}` {
		t.Errorf("body = %s", got)
	}
}
