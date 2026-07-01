// Command server starts the vibe-shop HTTP API on port 8080.
package main

import (
	"log"
	"net/http"

	apphttp "vibe-shop/internal/http"
)

const addr = ":8080"

func main() {
	log.Printf("vibe-shop listening on %s", addr)
	if err := http.ListenAndServe(addr, apphttp.NewRouter()); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
