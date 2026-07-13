// Command server starts the vibe-shop HTTP API on port 8080.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"vibe-shop/internal/auth"
	"vibe-shop/internal/cart"
	"vibe-shop/internal/db"
	apphttp "vibe-shop/internal/http"
	"vibe-shop/internal/order"
	"vibe-shop/internal/product"
)

const (
	addr     = ":8080"
	tokenTTL = 24 * time.Hour
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	gormDB, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	tokens := auth.NewTokenManager(jwtSecret, tokenTTL)

	products := product.NewHandler(product.NewRepository(gormDB))
	authH := auth.NewHandler(auth.NewRepository(gormDB), tokens)
	cartH := cart.NewHandler(cart.NewRepository(gormDB))
	ordersH := order.NewHandler(order.NewRepository(gormDB))

	log.Printf("vibe-shop listening on %s", addr)
	router := apphttp.NewRouter(products, authH, cartH, ordersH, tokens.RequireAuth)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
