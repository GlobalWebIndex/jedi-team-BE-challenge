package chatSessions

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loukaspe/jedi-team-challenge/internal/core/services"
	customerrors "github.com/loukaspe/jedi-team-challenge/pkg/errors"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"net/http"
)

type GetChatSessionHandler struct {
	ChatSessionService services.ChatSessionServiceInterface
	logger             logger.LoggerInterface
}

func NewGetChatSessionHandler(
	service services.ChatSessionServiceInterface,
	logger logger.LoggerInterface,
) *GetChatSessionHandler {
	return &GetChatSessionHandler{
		ChatSessionService: service,
		logger:             logger,
	}
}

// @Summary		Gets all User's chat sessions
// @Description	Gets all User's chat sessions
// @Security		BearerAuth
// @Param			user_id	path		int	true	"user id"
// @Success		201		{object}	UserChatSessionsResponse
// @Failure		400		{object}	UserChatSessionsResponse	"Error in message payload"
// @Failure		401		{object}	UserChatSessionsResponse	"Authentication error"
// @Failure		500		{object}	UserChatSessionsResponse	"Internal Server Error"
// @Router			/users/user_id/chat-sessions [get]
func (handler *GetChatSessionHandler) GetUserChatSessionsController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	response := &UserChatSessionsResponse{}

	ctx := r.Context()

	userIdAsString := mux.Vars(r)["user_id"]
	if userIdAsString == "" {
		response.ErrorMessage = "missing user id"

		handler.JsonUserChatSessionResponse(w, http.StatusBadRequest, response)

		return
	}

	userId, err := uuid.Parse(userIdAsString)
	if err != nil {
		handler.logger.Error("Error in getting chat session",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed user uuid"

		handler.JsonUserChatSessionResponse(w, http.StatusBadRequest, response)

		return
	}

	usersChatSessions, err := handler.ChatSessionService.GetUserChatSessions(ctx, userId)
	if userNotFoundError, ok := err.(customerrors.ResourceNotFoundErrorWrapper); ok {
		handler.logger.Error("Error in getting users chat sessions",
			map[string]interface{}{
				"errorMessage": userNotFoundError.Unwrap(),
			})

		response.ErrorMessage = err.Error()
		handler.JsonUserChatSessionResponse(w, http.StatusNotFound, response)

		return
	}

	if err != nil {
		handler.logger.Error("Error in getting user's chat sessions",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "error in getting user chat sessions"
		handler.JsonUserChatSessionResponse(w, http.StatusInternalServerError, response)

		return
	}

	response = UserChatSessionsResponseFromModel(usersChatSessions)
	handler.JsonUserChatSessionResponse(w, http.StatusOK, response)
}

// @Summary		Gets chat session
// @Description	Gets chat session
// @Security		BearerAuth
// @Param			session_id	path		int	true	"session id"
// @Success		201			{object}	ChatSessionResponse
// @Failure		400			{object}	ChatSessionResponse	"Error in message payload"
// @Failure		401			{object}	ChatSessionResponse	"Authentication error"
// @Failure		500			{object}	ChatSessionResponse	"Internal Server Error"
// @Router			/chat-sessions/session_id [get]
func (handler *GetChatSessionHandler) GetChatSessionController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	response := &ChatSessionResponse{}

	ctx := r.Context()

	sessionIDAsString := mux.Vars(r)["session_id"]
	if sessionIDAsString == "" {
		response.ErrorMessage = "missing session id"

		handler.JsonChatSessionResponse(w, http.StatusBadRequest, response)

		return
	}

	sessionID, err := uuid.Parse(sessionIDAsString)
	if err != nil {
		handler.logger.Error("Error in getting chat session",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed session uuid"

		handler.JsonChatSessionResponse(w, http.StatusBadRequest, response)

		return
	}

	chatSession, err := handler.ChatSessionService.GetChatSession(ctx, sessionID)
	if chatSessionNotFound, ok := err.(customerrors.ResourceNotFoundErrorWrapper); ok {
		handler.logger.Error("Error in getting chat session",
			map[string]interface{}{
				"errorMessage": chatSessionNotFound.Unwrap(),
			})

		response.ErrorMessage = err.Error()
		handler.JsonChatSessionResponse(w, http.StatusNotFound, response)

		return
	}

	if err != nil {
		handler.logger.Error("Error in getting chat session",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "error in getting chat session"
		handler.JsonChatSessionResponse(w, http.StatusInternalServerError, response)

		return
	}

	response = ChatSessionResponseFromModel(chatSession)
	handler.JsonChatSessionResponse(w, http.StatusOK, response)
}

func (handler *GetChatSessionHandler) JsonChatSessionResponse(
	w http.ResponseWriter,
	statusCode int,
	response *ChatSessionResponse,
) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.ErrorMessage = "error in getting chat session - json response"

		handler.logger.Error("Error in getting chat session - json response",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})
	}
}

func (handler *GetChatSessionHandler) JsonUserChatSessionResponse(
	w http.ResponseWriter,
	statusCode int,
	response *UserChatSessionsResponse,
) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.ErrorMessage = "error in getting user chat sessions - json response"

		handler.logger.Error("Error in getting user chat sessions - json response",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})
	}
}
