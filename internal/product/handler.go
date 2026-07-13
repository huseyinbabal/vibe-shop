package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"vibe-shop/internal/httpx"
)

// Handler serves the product endpoints over HTTP.
type Handler struct {
	repo Repository
}

// NewHandler builds a Handler backed by the given Repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// List handles GET /api/products.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.repo.List(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, products)
}

// GetByID handles GET /api/products/{id}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	p, err := h.repo.GetByID(r.Context(), uint(id))
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

// Create handles POST /api/products.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}

	p, err := h.repo.Create(r.Context(), in.product(0))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, p)
}

// Update handles PUT /api/products/{id} as a full update.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	in, ok := decodeInput(w, r)
	if !ok {
		return
	}

	p, err := h.repo.Update(r.Context(), in.product(uint(id)))
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, p)
}

// Delete handles DELETE /api/products/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	err = h.repo.Delete(r.Context(), uint(id))
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// decodeInput parses and validates the request body, writing the 400 response
// itself so handlers only proceed on valid input.
func decodeInput(w http.ResponseWriter, r *http.Request) (Input, bool) {
	var in Input
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return Input{}, false
	}
	if err := in.Validate(); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return Input{}, false
	}
	return in, true
}
