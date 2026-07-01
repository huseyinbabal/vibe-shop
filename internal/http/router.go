// Package http wires HTTP routes to their handlers.
package http

import (
	"net/http"

	"vibe-shop/internal/health"
)

// NewRouter builds the application's HTTP handler with all routes registered.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.Handler)
	return mux
}
