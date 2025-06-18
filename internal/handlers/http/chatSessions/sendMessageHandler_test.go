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
	"io"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendMessageHandler_SendMessageController(t *testing.T) {
	logger := logger.NewLogger(context.Background())
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockMessageService := mock_services.NewMockMessageServiceInterface(mockCtrl)

	type args struct {
		userId    uuid.UUID
		sessionId uuid.UUID
	}

	tests := []struct {
		name                     string
		args                     args
		mockMessageInsertedID    uuid.UUID
		mockReplyMessageInserted *domain.Message
		expected                 []byte
		expectedStatusCode       int
	}{
		{
			name: "valid",
			args: args{
				userId:    uuid.UUID{0x12, 0x34, 0x56, 0x78},
				sessionId: uuid.UUID{0x32, 0x34, 0x56, 0x78},
			},
			mockMessageInsertedID: uuid.UUID{0x42, 0x34, 0x56, 0x88},
			mockReplyMessageInserted: &domain.Message{
				ID:            uuid.UUID{0x52, 0x34, 0x56, 0x88},
				ChatSessionID: uuid.UUID{0x32, 0x34, 0x56, 0x78},
				Sender:        "SYSTEM",
				Content:       "Reply",
				CreatedAt:     time.Time{},
			},
			expected: json.RawMessage(`{"userMessage":{"id":"42345688-0000-0000-0000-000000000000","sender":"USER","content":"Hello, this is a test message","created_at":"0001-01-01 00:00:00 +0000 UTC"},"systemMessage":{"id":"52345688-0000-0000-0000-000000000000","sender":"SYSTEM","content":"Reply","created_at":"0001-01-01 00:00:00 +0000 UTC"}}
`),
			expectedStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequest := httptest.NewRequest(
				"POST",
				"/users/"+tt.args.userId.String()+"/chat-sessions"+"/"+tt.args.sessionId.String()+"/messages",
				bytes.NewBuffer(
					json.RawMessage(`{"content":"Hello, this is a test message"}`),
				),
			)

			vars := map[string]string{
				"user_id":    tt.args.userId.String(),
				"session_id": tt.args.sessionId.String(),
			}
			mockRequest = mux.SetURLVars(mockRequest, vars)

			mockRequest.Header.Set("Content-Type", "application/json")
			mockResponseRecorder := httptest.NewRecorder()

			mockMessageService.EXPECT().CreateMessage(
				gomock.Any(),
				tt.args.userId,
				&domain.Message{
					ChatSessionID: tt.args.sessionId,
					Sender:        "USER",
					Content:       "Hello, this is a test message",
				},
			).Return(tt.mockMessageInsertedID, nil)

			mockMessageService.EXPECT().GetAnswerForMessage(
				gomock.Any(),
				tt.mockMessageInsertedID,
			).Return(tt.mockReplyMessageInserted, nil)

			handler := &SendMessageHandler{
				MessageService: mockMessageService,
				logger:         logger,
			}
			sut := handler.SendMessageController

			sut(mockResponseRecorder, mockRequest)

			mockResponse := mockResponseRecorder.Result()
			actual, err := io.ReadAll(mockResponse.Body)
			if err != nil {
				t.Errorf("error with response reading: %v", err)
				return
			}
			actualStatusCode := mockResponse.StatusCode

			assert.Equal(t, string(tt.expected), string(actual))
			assert.Equal(t, tt.expectedStatusCode, actualStatusCode)
		})
	}
}
