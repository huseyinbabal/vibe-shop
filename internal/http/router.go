// Package http wires HTTP routes to their handlers.
package http

import (
	"net/http"

	"vibe-shop/internal/health"
	"vibe-shop/internal/product"
)

// NewRouter builds the application's HTTP handler with all routes registered.
func NewRouter(products *product.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.Handler)
	mux.HandleFunc("GET /api/products", products.List)
	mux.HandleFunc("GET /api/products/{id}", products.GetByID)
	return mux
}
