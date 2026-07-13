package order

import (
	"errors"
	"net/http"

	"vibe-shop/internal/auth"
	"vibe-shop/internal/httpx"
)

// Handler serves the order endpoint. It relies on auth middleware having placed
// the authenticated user id in the request context.
type Handler struct {
	repo Repository
}

// NewHandler builds a Handler backed by the given Repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// Create handles POST /api/orders: it turns the user's cart into an order and
// returns the placed order with its items and total.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	placed, err := h.repo.CreateFromCart(r.Context(), userID)
	if errors.Is(err, ErrCartEmpty) {
		httpx.WriteError(w, http.StatusBadRequest, "cart is empty")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, placed)
}
