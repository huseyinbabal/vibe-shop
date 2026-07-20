package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// newAdminMock fakes the two Keycloak endpoints AdminClient talks to: the
// realm token endpoint (client_credentials) and the admin users endpoint.
func newAdminMock(t *testing.T, usersHandler http.HandlerFunc) (issuer string, tokenCalls *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /realms/vibe-shop/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse token form: %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "client_credentials" {
			t.Errorf("grant_type = %q, want client_credentials", got)
		}
		if got := r.PostForm.Get("client_id"); got != "vibe-shop-backend" {
			t.Errorf("client_id = %q, want vibe-shop-backend", got)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{"access_token": "admin-token", "expires_in": 60}); err != nil {
			t.Errorf("encode token: %v", err)
		}
	})
	mux.HandleFunc("POST /admin/realms/vibe-shop/users", usersHandler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL + "/realms/vibe-shop", &calls
}

func TestAdminClient_CreateUser(t *testing.T) {
	var seenAuth string
	var seenBody map[string]any
	issuer, tokenCalls := newAdminMock(t, func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&seenBody); err != nil {
			t.Errorf("decode users body: %v", err)
		}
		w.Header().Set("Location", "/admin/realms/vibe-shop/users/new-user-id")
		w.WriteHeader(http.StatusCreated)
	})

	client := NewAdminClient(issuer, "vibe-shop-backend", "dev-backend-secret")
	id, err := client.CreateUser(t.Context(), "yeni@vibe.shop", "parola123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if id != "new-user-id" {
		t.Errorf("id = %q, want new-user-id", id)
	}
	if seenAuth != "Bearer admin-token" {
		t.Errorf("Authorization = %q, want Bearer admin-token", seenAuth)
	}
	if seenBody["email"] != "yeni@vibe.shop" || seenBody["username"] != "yeni@vibe.shop" {
		t.Errorf("body = %v, want email/username set", seenBody)
	}
	if seenBody["enabled"] != true || seenBody["emailVerified"] != true {
		t.Errorf("body = %v, want enabled+emailVerified", seenBody)
	}
	if seenBody["firstName"] == "" || seenBody["lastName"] == "" {
		t.Errorf("body = %v, want firstName+lastName (Keycloak refuses password grants otherwise)", seenBody)
	}
	creds, _ := seenBody["credentials"].([]any)
	if len(creds) != 1 {
		t.Fatalf("credentials = %v, want one password credential", seenBody["credentials"])
	}
	cred, _ := creds[0].(map[string]any)
	if cred["type"] != "password" || cred["value"] != "parola123" || cred["temporary"] != false {
		t.Errorf("credential = %v, want permanent password", cred)
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Errorf("token endpoint calls = %d, want 1", got)
	}
}

func TestAdminClient_CachesToken(t *testing.T) {
	issuer, tokenCalls := newAdminMock(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	client := NewAdminClient(issuer, "vibe-shop-backend", "s")
	for range 2 {
		if _, err := client.CreateUser(t.Context(), "a@b.c", "parola123"); err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Errorf("token endpoint calls = %d, want 1 (cached)", got)
	}
}

func TestAdminClient_DuplicateEmailReturnsErrEmailTaken(t *testing.T) {
	issuer, _ := newAdminMock(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"errorMessage":"User exists"}`, http.StatusConflict)
	})

	client := NewAdminClient(issuer, "vibe-shop-backend", "s")
	if _, err := client.CreateUser(t.Context(), "var@vibe.shop", "parola123"); !errors.Is(err, ErrEmailTaken) {
		t.Errorf("err = %v, want ErrEmailTaken", err)
	}
}

func TestAdminClient_KeycloakDownReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	client := NewAdminClient(srv.URL+"/realms/vibe-shop", "vibe-shop-backend", "s")
	if _, err := client.CreateUser(t.Context(), "a@b.c", "parola123"); err == nil {
		t.Fatal("CreateUser succeeded against broken Keycloak, want error")
	}
}
