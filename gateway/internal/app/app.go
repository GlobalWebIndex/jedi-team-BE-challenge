package app

import (
	"database/sql"
	"net/http"
	"strconv"

	"gateway/config"
	"gateway/internal/chat"

	"github.com/go-chi/chi/v5"
)

type DatabaseCallback func() (*sql.DB, error)

const NotExistingChat = -1

func SetupServer(dbCallback DatabaseCallback) (*http.Server, error) {
	cfg := config.LoadConfig()

	db, err := dbCallback()
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()

	// POST `/chat`: Starts a conversation and prompts chatbot
	// Request: `ChatMessageRequest` 
	// Response: `ChatMessageResponse` 
	r.Post("/chat", func(w http.ResponseWriter, r *http.Request) {
		chat.ChatHandler(w, r, db, NotExistingChat)
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
		chat.ChatHandler(w, r, db, chatID)
	})

	// // GET `/chat/users/:user_id`: Retrieves all conversations of a user
	// // Response: `[]ChatSummaryItem` 
	// r.Get("/chat/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
	// 	userId := chi.URLParam(r, "user_id")
	// 	chat.GetUsersChatHandler(w, r, db, userId)
	// })


	// // GET `/chat/:chat_id/history`: Retrieves all messages of a conversation
	// // Response: `ChatHistoryResponse` 
	// r.Get("/chat/{chat_id}/history", func(w http.ResponseWriter, r *http.Request) {
	// 	chatID := chi.URLParam(r, "chat_id")
	// 	chat.GetChatHandler(w, r, db, chatID)
	// })

	return &http.Server{
		Addr:		cfg.Server.Address,
		Handler:	r,
	}, nil
}