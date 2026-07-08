// Package http wires HTTP routes to their handlers.
package http

import (
	"net/http"

	"vibe-shop/internal/auth"
	"vibe-shop/internal/health"
	"vibe-shop/internal/product"
)

// NewRouter builds the application's HTTP handler with all routes registered.
// Product routes are public; auth exposes public register/login.
func NewRouter(products *product.Handler, authH *auth.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.Handler)

	mux.HandleFunc("GET /api/products", products.List)
	mux.HandleFunc("GET /api/products/{id}", products.GetByID)
	mux.HandleFunc("POST /api/products", products.Create)
	mux.HandleFunc("PUT /api/products/{id}", products.Update)
	mux.HandleFunc("DELETE /api/products/{id}", products.Delete)

	mux.HandleFunc("POST /api/register", authH.Register)
	mux.HandleFunc("POST /api/login", authH.Login)

	return mux
}
