package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// Handler serves the product read endpoints over HTTP.
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
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, products)
}

// GetByID handles GET /api/products/{id}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	p, err := h.repo.GetByID(r.Context(), uint(id))
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
