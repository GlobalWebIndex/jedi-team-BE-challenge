package chat

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"gateway/config"
	"gateway/internal/models"
	"gateway/internal/repositories"
)

const NotExistingChat = -1

func ChatHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, chatId int) {
	cfg := config.LoadConfig()

	var chatMessage models.ChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&chatMessage); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if chatMessage.UserId == 0 || chatMessage.Message == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	systemPrompt := models.OllamaMessage{
		Role:    "system",
		Content: "You are a question-answering assistant. You must answer the user's question ONLY using the information provided in the retrieved documents below. Do not use any outside knowledge, assumptions, or prior training data. If the answer is not explicitly contained in the provided documents, respond exactly with: \"I don’t have relevant information in the retrieved context to answer that question.\". However, if the question is highly relevant but not exactly matching one of the provided documents, answer based on what is relevant but clarify.",
	}

	ollamaMessages := []models.OllamaMessage{systemPrompt}

	prompt, err := repositories.RetrieveAndAugmentUserPrompt(chatMessage.Message)
	if err != nil {
		http.Error(w, fmt.Sprintf("ncountered issue while retrieving relavant documents: %v", err), http.StatusInternalServerError)
		return
	}

	ollamaMessages = append(ollamaMessages, prompt...)


	var title string

	if chatId == NotExistingChat {
		// If chat doesn't exist - auto generate a title
		title, err = repositories.GenerateTitle(chatMessage.Message)
		if err != nil {
			http.Error(w, fmt.Sprintf("Encountered error while generating title: %v", err), http.StatusInternalServerError)
			return
		}
		// Create a new chat entry
		chatId, err = repositories.CreateChat(
			models.DBChats{
				UserId:    chatMessage.UserId,
				Title:     title,
				CreatedAt: time.Now(),
			}, db)

		if err != nil {
			http.Error(w, fmt.Sprintf("Encountered error while creating chat: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		// Retrieve chat from db if exists and append to ollamaMessages for context
		dbChats, chatTitle, err := repositories.GetChats(chatId, db)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Chat not found", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Encountered error while retrieving chat: %v", err), http.StatusInternalServerError)
			return
		}
		title = chatTitle

		// Convert from DB Model -> Ollama Model

		ollamaMessages = append(ollamaMessages, repositories.ChatMsgDbOllama(dbChats)...)
	}

	ollamaRequest := models.OllamaChatRequest{
		Model:    cfg.Ollama.Model,
		Messages: ollamaMessages,
	}

	ollamaResponse, err := repositories.SendOllamaRequest(cfg.Ollama.Url, ollamaRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Received unexpected error while querying model: %v", err), http.StatusInternalServerError)
		return
	}

	// Craft new db chat message
	dbChatMessage := models.DBChatMessage{
		ChatId:    chatId, //generated
		Message:   chatMessage.Message,
		Response:  ollamaResponse.Message.Content,
		CreatedAt: time.Now(),
	}
	// Add new chat to DB
	if err := repositories.AddChatMessage(dbChatMessage, db); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save chat message: %v", err), http.StatusInternalServerError)
		return
	}

	chatMessageResponse := models.ChatMessageResponse{
		ChatId:    chatId,
		Title:     title,
		Message:   chatMessage.Message,
		Response:  ollamaResponse.Message.Content,
		CreatedAt: time.Now(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatMessageResponse)
}

func GetUsersChatHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, userId int) {
	dbChats, err := repositories.GetUserChats(userId, db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Chats not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Encountered error while retrieving chats: %v", err), http.StatusInternalServerError)
		return
	}

	response := repositories.ConvertDBChatsToSummaryResponse(dbChats)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetChatHandler(w http.ResponseWriter, r *http.Request, db *sql.DB, chatId int) {
	dbChatMessages, chatTitle, err := repositories.GetChats(chatId, db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Chat not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Encountered error while retrieving chat messages: %v", err), http.StatusInternalServerError)
		return
	}

	historyResponse := repositories.ConvertDBChatToHistoryResponse(dbChatMessages, chatId, chatTitle)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(historyResponse)
}
