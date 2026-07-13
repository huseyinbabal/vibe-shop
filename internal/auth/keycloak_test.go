package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testKID = "test-kid"

// newRSAKey generates a fresh RSA key pair for signing test tokens.
func newRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	return key
}

// newJWKSServer serves the given public key as a JWK Set the way Keycloak
// does, under <issuer>/protocol/openid-connect/certs. It returns the issuer
// URL to configure the verifier with.
func newJWKSServer(t *testing.T, pub *rsa.PublicKey) string {
	t.Helper()
	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"kid": testKID,
				"use": "sig",
				"alg": "RS256",
				"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
			},
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /realms/vibe-shop/protocol/openid-connect/certs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(jwks); err != nil {
			t.Errorf("encode jwks: %v", err)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL + "/realms/vibe-shop"
}

// signRS256 issues a token signed with key, carrying the given kid header.
func signRS256(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.RegisteredClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func validClaims(issuer string) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   "3f8e9c4a-1111-2222-3333-444455556666",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
}

func TestKeycloakVerifier_ValidToken(t *testing.T) {
	key := newRSAKey(t)
	issuer := newJWKSServer(t, &key.PublicKey)
	verifier, err := NewKeycloakVerifier(issuer)
	if err != nil {
		t.Fatalf("NewKeycloakVerifier: %v", err)
	}

	claims := validClaims(issuer)
	sub, err := verifier.Verify(signRS256(t, key, testKID, claims))
	if err != nil {
		t.Fatalf("Verify returned error for valid token: %v", err)
	}
	if sub != claims.Subject {
		t.Errorf("sub = %q, want %q", sub, claims.Subject)
	}
}

func TestKeycloakVerifier_RejectsBadTokens(t *testing.T) {
	key := newRSAKey(t)
	issuer := newJWKSServer(t, &key.PublicKey)
	verifier, err := NewKeycloakVerifier(issuer)
	if err != nil {
		t.Fatalf("NewKeycloakVerifier: %v", err)
	}

	otherKey := newRSAKey(t)

	expired := validClaims(issuer)
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))

	wrongIssuer := validClaims("https://evil.example.com/realms/vibe-shop")

	emptySub := validClaims(issuer)
	emptySub.Subject = ""

	noExpiry := validClaims(issuer)
	noExpiry.ExpiresAt = nil

	hs256 := func() string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims(issuer))
		token.Header["kid"] = testKID
		signed, err := token.SignedString([]byte("shared-secret"))
		if err != nil {
			t.Fatalf("sign hs256: %v", err)
		}
		return signed
	}()

	algNone := func() string {
		token := jwt.NewWithClaims(jwt.SigningMethodNone, validClaims(issuer))
		signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("sign none: %v", err)
		}
		return signed
	}()

	cases := map[string]string{
		"expired":                 signRS256(t, key, testKID, expired),
		"wrong issuer":            signRS256(t, key, testKID, wrongIssuer),
		"empty sub":               signRS256(t, key, testKID, emptySub),
		"no expiry":               signRS256(t, key, testKID, noExpiry),
		"signed with another key": signRS256(t, otherKey, testKID, validClaims(issuer)),
		"unknown kid":             signRS256(t, key, "unknown-kid", validClaims(issuer)),
		"hs256 (alg confusion)":   hs256,
		"alg none":                algNone,
		"malformed":               "not-a-jwt",
		"empty":                   "",
	}
	for name, token := range cases {
		if _, err := verifier.Verify(token); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("%s: err = %v, want ErrInvalidToken", name, err)
		}
	}
}

func TestNewKeycloakVerifier_UnreachableJWKS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	if _, err := NewKeycloakVerifier(srv.URL + "/realms/vibe-shop"); err == nil {
		t.Fatal("NewKeycloakVerifier succeeded against a broken JWKS endpoint, want error")
	}
}
