package chatSessions

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loukaspe/jedi-team-challenge/internal/core/domain"
	"github.com/loukaspe/jedi-team-challenge/internal/core/services"
	"github.com/loukaspe/jedi-team-challenge/internal/repositories"
	customerrors "github.com/loukaspe/jedi-team-challenge/pkg/errors"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"net/http"
)

type SendMessageHandler struct {
	MessageService services.MessageServiceInterface
	logger         logger.LoggerInterface
}

func NewSendMessageHandler(
	service services.MessageServiceInterface,
	logger logger.LoggerInterface,
) *SendMessageHandler {
	return &SendMessageHandler{
		MessageService: service,
		logger:         logger,
	}
}

// @Summary		Sends message to a given chat session and gets response
// @Description	Sends message to a given chat session and gets response
// @Security		BearerAuth
// @Param			SendMessageRequest	body		SendMessageRequest	true	"request body"
// @Param			user_id				path		int					true	"user id"
// @Param			session_id			body		int					true	"session id"
// @Success		201					{object}	SendMessageResponse
// @Failure		400					{object}	SendMessageResponse	"Error in message payload"
// @Failure		401					{object}	SendMessageResponse	"Authentication error"
// @Failure		500					{object}	SendMessageResponse	"Internal Server Error"
// @Router			/users/user_id/chat-sessions/session_id/messages [post]
func (handler *SendMessageHandler) SendMessageController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var err error
	response := &SendMessageResponse{}
	request := &SendMessageRequest{}

	userIdAsString := mux.Vars(r)["user_id"]
	if userIdAsString == "" {
		response.ErrorMessage = "missing user id"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	userId, err := uuid.Parse(userIdAsString)
	if err != nil {
		handler.logger.Error("Error in sending message",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed user uuid"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	chatSessionIDAsString := mux.Vars(r)["session_id"]
	if chatSessionIDAsString == "" {
		response.ErrorMessage = "missing session id"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	chatSessionID, err := uuid.Parse(chatSessionIDAsString)
	if err != nil {
		handler.logger.Error("Error in sending message",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed session uuid"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	err = json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		handler.logger.Error("Error in sending message",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed sending message request"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	domainMessage := &domain.Message{
		ChatSessionID: chatSessionID,
		Content:       request.Content,
		Sender:        repositories.USER_SENDER,
	}

	insertedUUID, err := handler.MessageService.CreateMessage(
		ctx,
		userId,
		domainMessage,
	)

	if resourceNotFound, ok := err.(customerrors.ResourceNotFoundErrorWrapper); ok {
		handler.logger.Error("Error in sending message",
			map[string]interface{}{
				"errorMessage": resourceNotFound.Unwrap(),
			})

		response.ErrorMessage = err.Error()
		handler.JsonResponse(w, http.StatusNotFound, response)

		return
	}

	if userMismatchError, ok := err.(customerrors.UserMismatchError); ok {
		handler.logger.Error("Error in sending message",
			map[string]interface{}{
				"errorMessage": userMismatchError.Error(),
			})

		response.ErrorMessage = err.Error()
		handler.JsonResponse(w, http.StatusForbidden, response)

		return
	}

	if err != nil {
		handler.logger.Error("Error in sending message",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "error in sending message"
		handler.JsonResponse(w, http.StatusInternalServerError, response)

		return
	}

	domainMessage.ID = insertedUUID

	replyMessage, err := handler.MessageService.GetAnswerForMessage(ctx, insertedUUID)
	if resourceNotFound, ok := err.(customerrors.ResourceNotFoundErrorWrapper); ok {
		handler.logger.Error("Error in replying to message",
			map[string]interface{}{
				"errorMessage": resourceNotFound.Unwrap(),
			})

		response.ErrorMessage = err.Error()
		handler.JsonResponse(w, http.StatusNotFound, response)

		return
	}

	if err != nil {
		handler.logger.Error("Error in replying to message",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "error in replying to message"
		handler.JsonResponse(w, http.StatusInternalServerError, response)

		return
	}

	response.UserMessage = MessageResponseFromModel(domainMessage)
	response.SystemMessage = MessageResponseFromModel(replyMessage)

	handler.JsonResponse(w, http.StatusOK, response)
}

func (handler *SendMessageHandler) JsonResponse(
	w http.ResponseWriter,
	statusCode int,
	response *SendMessageResponse,
) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.ErrorMessage = "error in sending message - json response"

		handler.logger.Error("Error in sending message - json response",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})
	}
}
