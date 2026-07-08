package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"vibe-shop/internal/httpx"
)

const minPasswordLength = 8

// Handler serves the registration and login endpoints.
type Handler struct {
	repo   Repository
	tokens *TokenManager
}

// NewHandler builds a Handler backed by the given Repository and TokenManager.
func NewHandler(repo Repository, tokens *TokenManager) *Handler {
	return &Handler{repo: repo, tokens: tokens}
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// validate enforces the registration rules: a plausible email and a password
// long enough to be worth hashing.
func (c credentials) validate() error {
	email := strings.TrimSpace(c.Email)
	if email == "" || !strings.Contains(email, "@") {
		return errors.New("email must be a valid address")
	}
	if len(c.Password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}
	return nil
}

// Register handles POST /api/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := in.validate(); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	user, err := h.repo.Create(r.Context(), User{
		Email:        strings.TrimSpace(in.Email),
		PasswordHash: string(hash),
	})
	if errors.Is(err, ErrEmailTaken) {
		httpx.WriteError(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, user)
}

// Login handles POST /api/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	user, err := h.repo.GetByEmail(r.Context(), strings.TrimSpace(in.Email))
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.tokens.Issue(user.ID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}
