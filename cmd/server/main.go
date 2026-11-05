package main

import (
	"log"
	"net/http"

	"github.com/CatalinaCovarrubias/pingeso-grupo6-identity-provider/internal/handler"
)

func main() {
	mux := http.NewServeMux()

	// Route registration. 
	mux.HandleFunc("GET /api/v1/health", handler.Health)

	addr := ":8080"
	log.Printf("Identity Provider listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}