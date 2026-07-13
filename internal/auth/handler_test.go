package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"vibe-shop/internal/auth"
)

func newHandler() *auth.Handler {
	return auth.NewHandler(auth.NewRepository(gormDB), auth.NewTokenManager("test-secret", time.Hour))
}

func postJSON(h http.HandlerFunc, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestRegister_ValidCreatesUserWithoutLeakingPassword(t *testing.T) {
	h := newHandler()

	rec := postJSON(h.Register, "/api/register", `{"email":"reg-valid@example.com","password":"parola123"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["email"] != "reg-valid@example.com" {
		t.Errorf("email = %v, want reg-valid@example.com", body["email"])
	}
	if _, ok := body["id"]; !ok {
		t.Error("expected an id in the response")
	}
	if _, leaked := body["password"]; leaked {
		t.Error("password must not appear in the response")
	}
	if _, leaked := body["password_hash"]; leaked {
		t.Error("password_hash must not appear in the response")
	}
}

func TestRegister_DuplicateEmailReturns409(t *testing.T) {
	h := newHandler()

	first := postJSON(h.Register, "/api/register", `{"email":"reg-dupe@example.com","password":"parola123"}`)
	if first.Code != http.StatusCreated {
		t.Fatalf("first register status = %d, want 201", first.Code)
	}

	second := postJSON(h.Register, "/api/register", `{"email":"reg-dupe@example.com","password":"parola123"}`)
	if second.Code != http.StatusConflict {
		t.Errorf("second register status = %d, want 409", second.Code)
	}
}

func TestRegister_InvalidInputReturns400(t *testing.T) {
	h := newHandler()

	cases := map[string]string{
		"no at-sign email": `{"email":"notanemail","password":"parola123"}`,
		"short password":   `{"email":"short@example.com","password":"short"}`,
		"invalid json":     `{"email":`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			rec := postJSON(h.Register, "/api/register", body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestLogin_ValidReturnsUsableToken(t *testing.T) {
	h := newHandler()
	tokens := auth.NewTokenManager("test-secret", time.Hour)

	reg := postJSON(h.Register, "/api/register", `{"email":"login-ok@example.com","password":"parola123"}`)
	if reg.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want 201", reg.Code)
	}

	rec := postJSON(h.Login, "/api/login", `{"email":"login-ok@example.com","password":"parola123"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, err := tokens.Parse(body["token"]); err != nil {
		t.Errorf("returned token did not parse: %v", err)
	}
}

func TestLogin_WrongCredentialsReturn401(t *testing.T) {
	h := newHandler()

	reg := postJSON(h.Register, "/api/register", `{"email":"login-bad@example.com","password":"parola123"}`)
	if reg.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want 201", reg.Code)
	}

	t.Run("wrong password", func(t *testing.T) {
		rec := postJSON(h.Login, "/api/login", `{"email":"login-bad@example.com","password":"wrongpass"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
	})
	t.Run("unknown email", func(t *testing.T) {
		rec := postJSON(h.Login, "/api/login", `{"email":"nobody@example.com","password":"parola123"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
	})
}
