// Command server starts the vibe-shop HTTP API on port 8080.
package main

import (
	"log"
	"net/http"
	"os"

	"vibe-shop/internal/db"
	apphttp "vibe-shop/internal/http"
	"vibe-shop/internal/product"
)

const addr = ":8080"

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	gormDB, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	products := product.NewHandler(product.NewRepository(gormDB))

	log.Printf("vibe-shop listening on %s", addr)
	if err := http.ListenAndServe(addr, apphttp.NewRouter(products)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
