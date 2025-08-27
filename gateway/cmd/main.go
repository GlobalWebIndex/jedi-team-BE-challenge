package main

import (
	"log"
	"net/http"

	"gateway/internal/app"
	"gateway/internal/db"
)

func main() {
	srv, err := app.SetupServer(db.InitDatabase)
	if err != nil {
		log.Fatalf("failed to setup server: %v", err)
	}

	log.Printf("Starting server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
