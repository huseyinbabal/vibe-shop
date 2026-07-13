package auth

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// KeycloakVerifier verifies RS256 JWTs issued by a Keycloak realm using the
// realm's JWKS endpoint. It is the only trust anchor for protected routes.
type KeycloakVerifier struct {
	issuer string
	keys   keyfunc.Keyfunc
}

// NewKeycloakVerifier fetches the realm's JWKS (derived from issuerURL the
// way Keycloak lays it out) and returns a verifier bound to that issuer.
// Fetching eagerly means the server fails fast at boot when Keycloak is down.
func NewKeycloakVerifier(issuerURL string) (*KeycloakVerifier, error) {
	jwksURL := issuerURL + "/protocol/openid-connect/certs"
	// The library's default tolerates a failed first fetch; we want the
	// opposite so a missing/broken Keycloak stops the server at boot.
	failOnFirstError := false
	keys, err := keyfunc.NewDefaultOverrideCtx(context.Background(), []string{jwksURL}, keyfunc.Override{
		NoErrorReturnFirstHTTPReq: &failOnFirstError,
	})
	if err != nil {
		return nil, fmt.Errorf("auth: fetch jwks from %s: %w", jwksURL, err)
	}
	return &KeycloakVerifier{issuer: issuerURL, keys: keys}, nil
}

// Verify checks the token's signature (RS256 only), issuer and expiry, and
// returns its subject — the Keycloak user id. Any failure is reported as
// ErrInvalidToken so callers can respond with a single 401.
func (v *KeycloakVerifier) Verify(tokenStr string) (string, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(tokenStr, &claims, v.keys.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", ErrInvalidToken
	}
	if claims.Subject == "" {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}
