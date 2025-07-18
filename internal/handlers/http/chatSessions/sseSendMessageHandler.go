package chatSessions

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loukaspe/jedi-team-challenge/internal/core/domain"
	"github.com/loukaspe/jedi-team-challenge/internal/core/services"
	"github.com/loukaspe/jedi-team-challenge/internal/repositories"
	customerrors "github.com/loukaspe/jedi-team-challenge/pkg/errors"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"net/http"
	"strings"
)

type SseSendMessageHandler struct {
	MessageService services.MessageServiceInterface
	logger         logger.LoggerInterface
}

func NewSseSendMessageHandler(
	service services.MessageServiceInterface,
	logger logger.LoggerInterface,
) *SseSendMessageHandler {
	return &SseSendMessageHandler{
		MessageService: service,
		logger:         logger,
	}
}

func (handler *SseSendMessageHandler) SseSendMessageController(w http.ResponseWriter, r *http.Request) {
	userIdStr := mux.Vars(r)["user_id"]
	sessionIdStr := mux.Vars(r)["session_id"]

	// Validate user_id
	userId, err := uuid.Parse(userIdStr)
	if userIdStr == "" || err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	// Validate session_id
	chatSessionID, err := uuid.Parse(sessionIdStr)
	if sessionIdStr == "" || err != nil {
		http.Error(w, "invalid session_id", http.StatusBadRequest)
		return
	}

	// Decode prompt from POST body
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Create the user's message
	domainMessage := &domain.Message{
		ChatSessionID: chatSessionID,
		Content:       req.Content,
		Sender:        repositories.USER_SENDER,
	}

	insertedUUID, err := handler.MessageService.CreateMessage(r.Context(), userId, domainMessage)
	if err != nil {
		handler.logger.Error("Error creating message", map[string]interface{}{"errorMessage": err.Error()})

		switch e := err.(type) {
		case customerrors.ResourceNotFoundErrorWrapper:
			http.Error(w, "chat session not found: "+e.Error(), http.StatusNotFound)
		case customerrors.UserMismatchError:
			http.Error(w, "user mismatch: "+e.Error(), http.StatusForbidden)
		default:
			http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Prepare SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial user message
	userMessageJson, _ := json.Marshal(MessageResponseFromModel(domainMessage))
	fmt.Fprintf(w, "data: {\"user_message\": %s}\n\n", userMessageJson)
	flusher.Flush()

	// Stream system message token-by-token
	tokenChan, errChan := handler.MessageService.StreamAnswerForMessage(r.Context(), insertedUUID)

	for {
		select {
		case token, ok := <-tokenChan:
			if !ok {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", formatAsJsonToken(token))
			flusher.Flush()

		case err := <-errChan:
			if err != nil {
				handler.logger.Error("AI streaming error", map[string]interface{}{"errorMessage": err.Error()})
				errorJson, _ := json.Marshal(map[string]string{"error": err.Error()})
				fmt.Fprintf(w, "data: %s\n\n", errorJson)
				flusher.Flush()
				return
			}
		}
	}
}

func formatAsJsonToken(token string) string {
	escaped := strings.ReplaceAll(token, `"`, `\"`)
	return fmt.Sprintf("{\"token\": \"%s\"}", escaped)
}
