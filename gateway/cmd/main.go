package main

import (
	"log"
	"net/http"

	"gateway/internal/app"
	"gateway/internal/db"
	"gateway/internal/repositories"
)

func main() {
	db, err := db.StartConn()
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	chatRepo := repositories.NewChatRepository(db)
	ragRepo := repositories.NewRagRepositoryHTTP()
	ollamaRepo := repositories.NewOllamaRepository()

	deps := app.ServerDependencies{
		ChatRepo:   chatRepo,
		RagRepo:    ragRepo,
		OllamaRepo: ollamaRepo,
	}

	srv, err := app.SetupServer(deps)
	if err != nil {
		log.Fatalf("failed to setup server: %v", err)
	}

	log.Printf("Starting server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
