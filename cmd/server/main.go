package main

import (
	"log"
	"net/http"

	"backend-assignment/internal/handlers"
	"backend-assignment/internal/store"
)

func main() {
	// Initialize thread-safe in-memory stores
	rateLimiter := store.NewRateLimiter()
	productStore := store.NewProductStore()

	// Initialize handlers
	rateLimitHandler := handlers.NewRateLimitHandler(rateLimiter)
	productHandler := handlers.NewProductHandler(productStore)

	// Set up routing
	mux := http.NewServeMux()

	// Part 1: Rate Limiting Endpoints
	mux.HandleFunc("/request", rateLimitHandler.HandleRequest)
	mux.HandleFunc("/stats", rateLimitHandler.HandleStats)

	// Part 2: Product Catalog Endpoints
	// /products handles GET (list) and POST (create)
	mux.HandleFunc("/products", productHandler.HandleProducts)
	// /products/ handles GET (detail) and POST (add media)
	mux.HandleFunc("/products/", productHandler.HandleProductDetail)

	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
