// Command server starts the vibe-shop HTTP API on port 8080.
package main

import (
	"log"
	"net/http"
	"os"

	"vibe-shop/internal/auth"
	"vibe-shop/internal/cart"
	"vibe-shop/internal/db"
	apphttp "vibe-shop/internal/http"
	"vibe-shop/internal/order"
	"vibe-shop/internal/product"
)

const defaultAddr = ":8080"

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	issuerURL := os.Getenv("KEYCLOAK_ISSUER_URL")
	if issuerURL == "" {
		log.Fatal("KEYCLOAK_ISSUER_URL is not set")
	}

	adminSecret := os.Getenv("KEYCLOAK_ADMIN_CLIENT_SECRET")
	if adminSecret == "" {
		log.Fatal("KEYCLOAK_ADMIN_CLIENT_SECRET is not set")
	}

	// Where this process reaches the realm; defaults to the issuer itself.
	// Differs only in containerized setups (e.g. Dokploy) where the public
	// issuer address is not routable from inside the network.
	internalURL := os.Getenv("KEYCLOAK_INTERNAL_URL")
	if internalURL == "" {
		internalURL = issuerURL
	}

	gormDB, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	verifier, err := auth.NewKeycloakVerifier(issuerURL, internalURL)
	if err != nil {
		log.Fatalf("connect to keycloak (is it running? see docker compose): %v", err)
	}

	products := product.NewHandler(product.NewRepository(gormDB))
	registerH := auth.NewRegisterHandler(auth.NewAdminClient(internalURL, "vibe-shop-backend", adminSecret))
	cartH := cart.NewHandler(cart.NewRepository(gormDB))
	ordersH := order.NewHandler(order.NewRepository(gormDB))

	log.Printf("vibe-shop listening on %s", addr)
	router := apphttp.NewRouter(products, registerH, cartH, ordersH, verifier.RequireAuth)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
