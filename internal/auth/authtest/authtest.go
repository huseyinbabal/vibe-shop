// Package authtest gives other packages' tests a Keycloak-shaped identity
// fixture: a JWKS endpoint backed by a fresh RSA key, a KeycloakVerifier
// trusting it, and a mint for RS256 tokens. Signature verification stays
// cryptographically real; only the identity provider is faked.
package authtest

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"vibe-shop/internal/auth"
)

const kid = "authtest-kid"

// New starts a JWKS server for a generated RSA key and returns a verifier
// bound to it plus a mint that signs valid one-hour tokens for a subject.
// The server is shut down when the test finishes.
func New(t *testing.T) (*auth.KeycloakVerifier, func(sub string) string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("authtest: generate rsa key: %v", err)
	}

	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"kid": kid,
				"use": "sig",
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
			},
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /realms/vibe-shop/protocol/openid-connect/certs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(jwks); err != nil {
			t.Errorf("authtest: encode jwks: %v", err)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	issuer := srv.URL + "/realms/vibe-shop"
	verifier, err := auth.NewKeycloakVerifier(issuer)
	if err != nil {
		t.Fatalf("authtest: new verifier: %v", err)
	}

	mint := func(sub string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		})
		token.Header["kid"] = kid
		signed, err := token.SignedString(key)
		if err != nil {
			t.Fatalf("authtest: sign token: %v", err)
		}
		return signed
	}
	return verifier, mint
}
