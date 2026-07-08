// Package http wires HTTP routes to their handlers.
package http

import (
	"net/http"

	"vibe-shop/internal/auth"
	"vibe-shop/internal/cart"
	"vibe-shop/internal/health"
	"vibe-shop/internal/product"
)

// Middleware wraps a handler, e.g. to require authentication.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// NewRouter builds the application's HTTP handler with all routes registered.
// Product routes and register/login are public; cart routes are wrapped with
// requireAuth so only authenticated users reach them.
func NewRouter(products *product.Handler, authH *auth.Handler, cartH *cart.Handler, requireAuth Middleware) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.Handler)

	mux.HandleFunc("GET /api/products", products.List)
	mux.HandleFunc("GET /api/products/{id}", products.GetByID)
	mux.HandleFunc("POST /api/products", products.Create)
	mux.HandleFunc("PUT /api/products/{id}", products.Update)
	mux.HandleFunc("DELETE /api/products/{id}", products.Delete)

	mux.HandleFunc("POST /api/register", authH.Register)
	mux.HandleFunc("POST /api/login", authH.Login)

	mux.HandleFunc("POST /api/cart", requireAuth(cartH.Add))
	mux.HandleFunc("GET /api/cart", requireAuth(cartH.Get))

	return mux
}
