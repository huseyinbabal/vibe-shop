package cart

import (
	"encoding/json"
	"errors"
	"net/http"

	"vibe-shop/internal/auth"
	"vibe-shop/internal/httpx"
)

// Handler serves the cart endpoints. It relies on auth middleware having placed
// the authenticated Keycloak subject in the request context.
type Handler struct {
	repo Repository
}

// NewHandler builds a Handler backed by the given Repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

type addInput struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type cartResponse struct {
	Items []LineView `json:"items"`
	Total float64    `json:"total"`
}

// Add handles POST /api/cart.
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var in addInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if in.ProductID == 0 {
		httpx.WriteError(w, http.StatusBadRequest, "product_id is required")
		return
	}
	if in.Quantity <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "quantity must be greater than zero")
		return
	}

	item, err := h.repo.AddOrIncrement(r.Context(), userID, in.ProductID, in.Quantity)
	if errors.Is(err, ErrProductNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

// Get handles GET /api/cart.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.SubjectFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	lines, err := h.repo.ListByUser(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if lines == nil {
		lines = []LineView{}
	}

	var total float64
	for _, l := range lines {
		total += l.LineTotal
	}
	httpx.WriteJSON(w, http.StatusOK, cartResponse{Items: lines, Total: total})
}
