package chatSessions

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/loukaspe/jedi-team-challenge/internal/core/domain"
	"github.com/loukaspe/jedi-team-challenge/internal/core/services"
	"github.com/loukaspe/jedi-team-challenge/internal/repositories"
	customerrors "github.com/loukaspe/jedi-team-challenge/pkg/errors"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"net/http"
)

type WsSendMessageHandler struct {
	MessageService services.MessageServiceInterface
	logger         logger.LoggerInterface
	upgrader       *websocket.Upgrader
}

func NewWsSendMessageHandler(
	service services.MessageServiceInterface,
	logger logger.LoggerInterface,
	upgrader *websocket.Upgrader,
) *WsSendMessageHandler {
	return &WsSendMessageHandler{
		MessageService: service,
		logger:         logger,
		upgrader:       upgrader,
	}
}

// Request model from WebSocket client
type SendMessageWSRequest struct {
	Content string `json:"content"`
}

// WebSocket message handler
func (handler *WsSendMessageHandler) WsSendMessageController(w http.ResponseWriter, r *http.Request) {
	userIdStr := mux.Vars(r)["user_id"]
	if userIdStr == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "malformed user uuid", http.StatusBadRequest)
		return
	}

	sessionIdStr := mux.Vars(r)["session_id"]
	if sessionIdStr == "" {
		http.Error(w, "missing session id", http.StatusBadRequest)
		return
	}
	chatSessionID, err := uuid.Parse(sessionIdStr)
	if err != nil {
		http.Error(w, "malformed session uuid", http.StatusBadRequest)
		return
	}

	conn, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		handler.logger.Error("Upgrade error", map[string]interface{}{"errorMessage": err.Error()})
		return
	}
	defer conn.Close()

	for {
		var request SendMessageWSRequest
		if err := conn.ReadJSON(&request); err != nil {
			handler.logger.Error("Invalid WS JSON input", map[string]interface{}{"errorMessage": err.Error()})
			conn.WriteJSON(SendMessageResponse{ErrorMessage: "Invalid input format"})
			return
		}

		domainMessage := &domain.Message{
			ChatSessionID: chatSessionID,
			Content:       request.Content,
			Sender:        repositories.USER_SENDER,
		}

		insertedUUID, err := handler.MessageService.CreateMessage(r.Context(), userId, domainMessage)
		if notFound, ok := err.(customerrors.ResourceNotFoundErrorWrapper); ok {
			handler.logger.Error("Chat session not found", map[string]interface{}{"errorMessage": notFound.Unwrap()})
			conn.WriteJSON(SendMessageResponse{ErrorMessage: "chat session not found: " + notFound.Error()})
			continue
		}
		if mismatch, ok := err.(customerrors.UserMismatchError); ok {
			handler.logger.Error("User mismatch error", map[string]interface{}{"errorMessage": mismatch.Error()})
			conn.WriteJSON(SendMessageResponse{ErrorMessage: "user mismatch: " + mismatch.Error()})
			continue
		}
		if err != nil {
			handler.logger.Error("Error creating message", map[string]interface{}{"errorMessage": err.Error()})
			conn.WriteJSON(SendMessageResponse{ErrorMessage: "internal error: " + err.Error()})
			continue
		}

		domainMessage.ID = insertedUUID

		replyMessage, err := handler.MessageService.GetAnswerForMessage(r.Context(), insertedUUID)
		if notFound, ok := err.(customerrors.ResourceNotFoundErrorWrapper); ok {
			handler.logger.Error("Reply not found", map[string]interface{}{"errorMessage": notFound.Unwrap()})
			conn.WriteJSON(SendMessageResponse{ErrorMessage: notFound.Error()})
			continue
		}
		if err != nil {
			handler.logger.Error("Error fetching reply", map[string]interface{}{"errorMessage": err.Error()})
			conn.WriteJSON(SendMessageResponse{ErrorMessage: "failed to fetch reply"})
			continue
		}

		// Send back both user and system messages
		conn.WriteJSON(SendMessageResponse{
			UserMessage:   MessageResponseFromModel(domainMessage),
			SystemMessage: MessageResponseFromModel(replyMessage),
		})
	}
}
