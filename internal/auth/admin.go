package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

// ErrEmailTaken is returned when registering an email Keycloak already knows.
var ErrEmailTaken = errors.New("auth: email already registered")

// AdminClient creates realm users through the Keycloak Admin API using a
// confidential service client (client_credentials). The admin token is cached
// in memory and renewed shortly before it expires. The client secret never
// leaves the backend (SPEC §12.6).
type AdminClient struct {
	tokenURL     string
	usersURL     string
	clientID     string
	clientSecret string
	httpClient   *http.Client

	mu     sync.Mutex
	token  string
	expiry time.Time
}

// NewAdminClient derives the token and admin endpoints from the realm issuer
// URL the way Keycloak lays them out.
func NewAdminClient(issuerURL, clientID, clientSecret string) *AdminClient {
	return &AdminClient{
		tokenURL:     issuerURL + "/protocol/openid-connect/token",
		usersURL:     strings.Replace(issuerURL, "/realms/", "/admin/realms/", 1) + "/users",
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// adminToken returns a valid service-account token, fetching a fresh one only
// when the cached token is missing or about to expire.
func (c *AdminClient) adminToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.expiry) {
		return c.token, nil
	}

	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("auth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth: request admin token: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth: admin token request failed with status %d", res.StatusCode)
	}

	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("auth: decode admin token: %w", err)
	}

	c.token = body.AccessToken
	// Renew a little early so a token never expires mid-request.
	c.expiry = time.Now().Add(time.Duration(body.ExpiresIn)*time.Second - 10*time.Second)
	return c.token, nil
}

// CreateUser registers a new enabled user with a permanent password and
// returns the Keycloak user id. A duplicate email maps to ErrEmailTaken.
func (c *AdminClient) CreateUser(ctx context.Context, email, password string) (string, error) {
	token, err := c.adminToken(ctx)
	if err != nil {
		return "", err
	}

	// firstName/lastName must be present or Keycloak flags the account as
	// "not fully set up" and refuses password grants (same constraint as the
	// realm-import test users).
	payload, err := json.Marshal(map[string]any{
		"username":      email,
		"email":         email,
		"firstName":     strings.SplitN(email, "@", 2)[0],
		"lastName":      "User",
		"enabled":       true,
		"emailVerified": true,
		"credentials": []map[string]any{
			{"type": "password", "value": password, "temporary": false},
		},
	})
	if err != nil {
		return "", fmt.Errorf("auth: marshal create user: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.usersURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("auth: build create user request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth: create user: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusCreated:
		// Keycloak returns the new user's id as the Location's last segment.
		return path.Base(res.Header.Get("Location")), nil
	case http.StatusConflict:
		return "", ErrEmailTaken
	default:
		return "", fmt.Errorf("auth: create user failed with status %d", res.StatusCode)
	}
}
