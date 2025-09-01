package app

import (
	"database/sql"
	"net/http"
	"strconv"

	"gateway/config"
	"gateway/internal/chat"
	"gateway/internal/repositories"

	"github.com/go-chi/chi/v5"
)

type DatabaseCallback func() (*sql.DB, error)

const NotExistingChat = -1

type ServerDependencies struct {
	ChatRepo   repositories.ChatRepository
	RagRepo    repositories.RagRepository
	OllamaRepo repositories.OllamaRepository
}

func SetupServer(deps ServerDependencies) (*http.Server, error) {
	cfg := config.LoadConfig()

	r := chi.NewRouter()

	// POST `/chat`: Starts a conversation and prompts chatbot
	// Request: `ChatMessageRequest` 
	// Response: `ChatMessageResponse` 
	r.Post("/chat", func(w http.ResponseWriter, r *http.Request) {
		chat.ChatHandler(w, r, deps.ChatRepo, deps.RagRepo, deps.OllamaRepo, NotExistingChat)
	})

	// POST `/chat/:chat_id`: Continues a conversation and prompts chatbot
	// Request: `ChatMessageRequest` 
	// Response: `ChatMessageResponse` 
	r.Post("/chat/{chat_id}", func(w http.ResponseWriter, r *http.Request) {
		chatIDStr := chi.URLParam(r, "chat_id")
		chatID, err := strconv.Atoi(chatIDStr)
		if err != nil {
			http.Error(w, "Invalid chat ID", http.StatusBadRequest)
			return
		}
		chat.ChatHandler(w, r, deps.ChatRepo, deps.RagRepo, deps.OllamaRepo, chatID)
	})

	// GET `/chat/users/:user_id`: Retrieves all conversations of a user
	// Response: `[]ChatSummaryItem` 
	r.Get("/chat/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		userIdStr := chi.URLParam(r, "user_id")
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			http.Error(w, "Invalid chat ID", http.StatusBadRequest)
			return
		}
		chat.GetUsersChatHandler(w, r, deps.ChatRepo, userId)
	})


	// GET `/chat/:chat_id/history`: Retrieves all messages of a conversation
	// Response: `ChatHistoryResponse` 
	r.Get("/chat/{chat_id}/history", func(w http.ResponseWriter, r *http.Request) {
		chatIDStr := chi.URLParam(r, "chat_id")
		chatID, err := strconv.Atoi(chatIDStr)
		if err != nil {
			http.Error(w, "Invalid chat ID", http.StatusBadRequest)
			return
		}
		chat.GetChatHandler(w, r, deps.ChatRepo, chatID)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return &http.Server{
		Addr:		cfg.Server.Address,
		Handler:	r,
	}, nil
}