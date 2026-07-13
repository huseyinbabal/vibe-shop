// Package http wires HTTP routes to their handlers.
package http

import (
	"net/http"

	"vibe-shop/internal/cart"
	"vibe-shop/internal/health"
	"vibe-shop/internal/order"
	"vibe-shop/internal/product"
)

// Middleware wraps a handler, e.g. to require authentication.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// NewRouter builds the application's HTTP handler with all routes registered.
// Health and product reads are public; product writes, cart and order routes
// are wrapped with requireAuth so only requests carrying a valid Keycloak
// token reach them.
func NewRouter(products *product.Handler, cartH *cart.Handler, ordersH *order.Handler, requireAuth Middleware) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.Handler)

	mux.HandleFunc("GET /api/products", products.List)
	mux.HandleFunc("GET /api/products/{id}", products.GetByID)
	mux.HandleFunc("POST /api/products", requireAuth(products.Create))
	mux.HandleFunc("PUT /api/products/{id}", requireAuth(products.Update))
	mux.HandleFunc("DELETE /api/products/{id}", requireAuth(products.Delete))

	mux.HandleFunc("POST /api/cart", requireAuth(cartH.Add))
	mux.HandleFunc("GET /api/cart", requireAuth(cartH.Get))

	mux.HandleFunc("POST /api/orders", requireAuth(ordersH.Create))

	return mux
}
