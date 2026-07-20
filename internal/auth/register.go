package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"vibe-shop/internal/httpx"
)

// RegisterHandler serves POST /api/register: it validates the input and asks
// Keycloak (via AdminClient) to create the user. Passwords pass through to
// Keycloak only and are never logged or stored here.
type RegisterHandler struct {
	admin *AdminClient
}

// NewRegisterHandler builds a RegisterHandler backed by the given AdminClient.
func NewRegisterHandler(admin *AdminClient) *RegisterHandler {
	return &RegisterHandler{admin: admin}
}

type registerInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (in registerInput) validate() string {
	if in.Email == "" || !strings.Contains(in.Email, "@") {
		return "email must be a valid address"
	}
	if len(in.Password) < 8 {
		return "password must be at least 8 characters"
	}
	return ""
}

// Register handles POST /api/register.
func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in registerInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if msg := in.validate(); msg != "" {
		httpx.WriteError(w, http.StatusBadRequest, msg)
		return
	}

	id, err := h.admin.CreateUser(r.Context(), in.Email, in.Password)
	if errors.Is(err, ErrEmailTaken) {
		httpx.WriteError(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "registration unavailable")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, map[string]string{"id": id, "email": in.Email})
}
