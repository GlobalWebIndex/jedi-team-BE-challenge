package chatSessions

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loukaspe/jedi-team-challenge/internal/core/domain"
	"github.com/loukaspe/jedi-team-challenge/internal/core/services"
	customerrors "github.com/loukaspe/jedi-team-challenge/pkg/errors"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"net/http"
)

type CreateUserChatSessionHandler struct {
	ChatSessionService services.ChatSessionServiceInterface
	logger             logger.LoggerInterface
}

func NewCreateUserChatSessionHandler(
	service services.ChatSessionServiceInterface,
	logger logger.LoggerInterface,
) *CreateUserChatSessionHandler {
	return &CreateUserChatSessionHandler{
		ChatSessionService: service,
		logger:             logger,
	}
}

// @Summary		Creates chat session
// @Description	Creates a chat session for User
// @Security		BearerAuth
// @Param			user_id	path		int	true	"user id"
// @Success		201		{object}	ChatSessionResponse
// @Failure		400		{object}	ChatSessionResponse	"Error in message payload"
// @Failure		401		{object}	ChatSessionResponse	"Authentication error"
// @Failure		500		{object}	ChatSessionResponse	"Internal Server Error"
// @Router			/users/user_id/chat-sessions [post]
func (handler *CreateUserChatSessionHandler) CreateUserChatSessionController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var err error
	response := &ChatSessionResponse{}

	userIdAsString := mux.Vars(r)["user_id"]
	if userIdAsString == "" {
		response.ErrorMessage = "missing user id"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	userId, err := uuid.Parse(userIdAsString)
	if err != nil {
		handler.logger.Error("Error in creating chat session",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed user uuid"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	insertedUUID, err := handler.ChatSessionService.CreateChatSession(
		ctx,
		&domain.ChatSession{
			UserID: userId,
		},
	)
	if userNotFoundError, ok := err.(customerrors.ResourceNotFoundErrorWrapper); ok {
		handler.logger.Error("Error in creating chat session",
			map[string]interface{}{
				"errorMessage": userNotFoundError.Unwrap(),
			})

		response.ErrorMessage = err.Error()
		handler.JsonResponse(w, http.StatusNotFound, response)

		return
	}

	if err != nil {
		handler.logger.Error("Error in creating chat session",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "error in creating chat session"
		handler.JsonResponse(w, http.StatusInternalServerError, response)

		return
	}

	response.ID = insertedUUID.String()
	handler.JsonResponse(w, http.StatusCreated, response)
}

func (handler *CreateUserChatSessionHandler) JsonResponse(
	w http.ResponseWriter,
	statusCode int,
	response *ChatSessionResponse,
) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.ErrorMessage = "error in creating chat session - json response"

		handler.logger.Error("Error in creating chat session - json response",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})
	}
}
