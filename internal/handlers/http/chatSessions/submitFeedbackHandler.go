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

type SubmitFeedbackHandler struct {
	MessageService services.MessageServiceInterface
	logger         logger.LoggerInterface
}

func NewSubmitFeedbackHandler(
	service services.MessageServiceInterface,
	logger logger.LoggerInterface,
) *SubmitFeedbackHandler {
	return &SubmitFeedbackHandler{
		MessageService: service,
		logger:         logger,
	}
}

// @Summary		Submits a feedback to a message
// @Description	Submits a feedback to a message
// @Security		BearerAuth
// @Param			SubmitFeedbackRequest	body		SubmitFeedbackRequest	true	"request body"
// @Param			message_id				body		int						true	"message_id"
// @Success		201						{object}	SubmitFeedbackResponse
// @Failure		400						{object}	SubmitFeedbackResponse	"Error in message payload"
// @Failure		401						{object}	SubmitFeedbackResponse	"Authentication error"
// @Failure		500						{object}	SubmitFeedbackResponse	"Internal Server Error"
// @Router			/users/user_id/chat-sessions/session_id/messages/message_id/feedback [post]
func (handler *SubmitFeedbackHandler) SubmitFeedbackController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var err error
	response := &SubmitFeedbackResponse{}
	request := &SubmitFeedbackRequest{}

	messageIDAsString := mux.Vars(r)["message_id"]
	if messageIDAsString == "" {
		response.ErrorMessage = "missing message id"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	messageID, err := uuid.Parse(messageIDAsString)
	if err != nil {
		handler.logger.Error("Error in submitting feedback",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed message uuid"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	userIDAsString := mux.Vars(r)["user_id"]
	if userIDAsString == "" {
		response.ErrorMessage = "missing user id"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	userID, err := uuid.Parse(userIDAsString)
	if err != nil {
		handler.logger.Error("Error in submitting feedback",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed message uuid"

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
		handler.logger.Error("Error in submitting feedback",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed sessio  uuid"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	err = json.NewDecoder(r.Body).Decode(request)
	if err != nil {
		handler.logger.Error("Error in submitting feedback",
			map[string]interface{}{
				"errorMessage": err.Error(),
			})

		response.ErrorMessage = "malformed submitting feedback request"

		handler.JsonResponse(w, http.StatusBadRequest, response)

		return
	}

	err = handler.MessageService.UpdateMessageFeedback(
		ctx,
		&domain.Message{
			ID:            messageID,
			Feedback:      &request.Feedback,
			ChatSessionID: chatSessionID,
		},
		userID,
	)

	if resourceNotFound, ok := err.(customerrors.ResourceNotFoundErrorWrapper); ok {
		handler.logger.Error("Error in submitting feedback",
			map[string]interface{}{
				"errorMessage": resourceNotFound.Unwrap(),
			})

		response.ErrorMessage = err.Error()
		handler.JsonResponse(w, http.StatusNotFound, response)

		return
	}

	if userMismatchError, ok := err.(customerrors.UserMismatchError); ok {
		handler.logger.Error("Error in submitting feedback",
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

	w.WriteHeader(http.StatusCreated)
}

func (handler *SubmitFeedbackHandler) JsonResponse(
	w http.ResponseWriter,
	statusCode int,
	response *SubmitFeedbackResponse,
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
