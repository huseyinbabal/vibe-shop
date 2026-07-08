// Command server starts the vibe-shop HTTP API on port 8080.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"vibe-shop/internal/auth"
	"vibe-shop/internal/db"
	apphttp "vibe-shop/internal/http"
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

	log.Printf("vibe-shop listening on %s", addr)
	if err := http.ListenAndServe(addr, apphttp.NewRouter(products, authH)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
