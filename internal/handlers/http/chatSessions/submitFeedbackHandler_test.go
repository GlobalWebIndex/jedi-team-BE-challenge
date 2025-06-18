package chatSessions

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loukaspe/jedi-team-challenge/internal/core/domain"
	mock_services "github.com/loukaspe/jedi-team-challenge/mocks/mock_internal/core/services"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http/httptest"
	"testing"
)

func TestUpdateMessageFeedbackHandler_UpdateMessageFeedbackController(t *testing.T) {
	logger := logger.NewLogger(context.Background())
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_services.NewMockMessageServiceInterface(mockCtrl)

	type args struct {
		userID        uuid.UUID
		chatSessionID uuid.UUID
		messageID     uuid.UUID
		feedback      string
	}

	tests := []struct {
		name               string
		args               args
		expectedStatusCode int
	}{
		{
			name: "valid",
			args: args{
				messageID:     uuid.UUID{0x12, 0x34, 0x56, 0x78},
				chatSessionID: uuid.UUID{0x32, 0x34, 0x56, 0x78},
				userID:        uuid.UUID{0x42, 0x34, 0x56, 0x78},
				feedback:      "abla",
			},
			expectedStatusCode: 201,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequest := httptest.NewRequest(
				"POST",
				"/users/"+tt.args.messageID.String()+"/chat-sessions/"+tt.args.chatSessionID.String()+"/messages/"+tt.args.messageID.String(),
				bytes.NewBuffer(
					json.RawMessage(`{"feedback":"abla"}`),
				),
			)

			vars := map[string]string{
				"user_id":    tt.args.userID.String(),
				"session_id": tt.args.chatSessionID.String(),
				"message_id": tt.args.messageID.String(),
			}
			mockRequest = mux.SetURLVars(mockRequest, vars)

			mockRequest.Header.Set("Content-Type", "application/json")
			mockResponseRecorder := httptest.NewRecorder()

			mockService.EXPECT().UpdateMessageFeedback(
				gomock.Any(),
				&domain.Message{
					ID:            tt.args.messageID,
					Feedback:      &tt.args.feedback,
					ChatSessionID: tt.args.chatSessionID,
				},
				tt.args.userID,
			).Return(nil)

			handler := &SubmitFeedbackHandler{
				MessageService: mockService,
				logger:         logger,
			}
			sut := handler.SubmitFeedbackController

			sut(mockResponseRecorder, mockRequest)

			mockResponse := mockResponseRecorder.Result()
			actualStatusCode := mockResponse.StatusCode

			assert.Equal(t, tt.expectedStatusCode, actualStatusCode)
		})
	}
}
