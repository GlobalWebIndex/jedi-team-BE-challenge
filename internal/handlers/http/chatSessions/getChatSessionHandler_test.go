package chatSessions

import (
	"context"
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
)

func TestGetChatSessionHandler_GetChatSessionController(t *testing.T) {
	logger := logger.NewLogger(context.Background())
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_services.NewMockChatSessionServiceInterface(mockCtrl)

	type args struct {
		chatSessionID uuid.UUID
	}

	tests := []struct {
		name                     string
		args                     args
		mockServiceResponseData  *domain.ChatSession
		mockServiceResponseError error
		expected                 string
		expectedStatusCode       int
	}{
		{
			name: "valid",
			args: args{
				chatSessionID: uuid.UUID{0x32, 0x34, 0x56, 0x78},
			},
			mockServiceResponseData: &domain.ChatSession{
				ID:     uuid.UUID{0x32, 0x34, 0x56, 0x78},
				UserID: uuid.UUID{0x42, 0x34, 0x56, 0x78},
				Title:  "title",
			},
			mockServiceResponseError: nil,
			expected: `{"id":"32345678-0000-0000-0000-000000000000","title":"title","createdAt":"0001-01-01 00:00:00 +0000 UTC","updatedAt":"0001-01-01 00:00:00 +0000 UTC"}
`,
			expectedStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequest := httptest.NewRequest(
				"GET",
				"/chat-sessions/"+tt.args.chatSessionID.String(),
				nil,
			)
			vars := map[string]string{
				"session_id": tt.args.chatSessionID.String(),
			}
			mockRequest = mux.SetURLVars(mockRequest, vars)
			mockResponseRecorder := httptest.NewRecorder()

			mockService.EXPECT().GetChatSession(
				gomock.Any(),
				tt.args.chatSessionID,
			).Return(tt.mockServiceResponseData, tt.mockServiceResponseError)

			handler := &GetChatSessionHandler{
				ChatSessionService: mockService,
				logger:             logger,
			}
			sut := handler.GetChatSessionController

			sut(mockResponseRecorder, mockRequest)

			mockResponse := mockResponseRecorder.Result()
			actual, err := io.ReadAll(mockResponse.Body)
			if err != nil {
				t.Errorf("error with response reading: %v", err)
				return
			}
			actualStatusCode := mockResponse.StatusCode

			assert.Equal(t, tt.expected, string(actual))
			assert.Equal(t, tt.expectedStatusCode, actualStatusCode)
		})
	}
}

func TestGetChatSessionHandler_GetUserChatSessionsController(t *testing.T) {
	logger := logger.NewLogger(context.Background())
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_services.NewMockChatSessionServiceInterface(mockCtrl)

	type args struct {
		userID uuid.UUID
	}

	tests := []struct {
		name                     string
		args                     args
		mockServiceResponseData  []*domain.ChatSession
		mockServiceResponseError error
		expected                 string
		expectedStatusCode       int
	}{
		{
			name: "valid",
			args: args{
				userID: uuid.UUID{0x42, 0x34, 0x56, 0x78},
			},
			mockServiceResponseData: []*domain.ChatSession{
				&domain.ChatSession{
					ID:     uuid.UUID{0x12, 0x34, 0x56, 0x78},
					UserID: uuid.UUID{0x42, 0x34, 0x56, 0x78},
					Title:  "title1",
				},
				&domain.ChatSession{
					ID:     uuid.UUID{0x32, 0x34, 0x56, 0x78},
					UserID: uuid.UUID{0x42, 0x34, 0x56, 0x78},
					Title:  "title2",
				},
			},
			mockServiceResponseError: nil,
			expected: `{"sessions":[{"id":"12345678-0000-0000-0000-000000000000","title":"title1","createdAt":"0001-01-01 00:00:00 +0000 UTC","updatedAt":"0001-01-01 00:00:00 +0000 UTC"},{"id":"32345678-0000-0000-0000-000000000000","title":"title2","createdAt":"0001-01-01 00:00:00 +0000 UTC","updatedAt":"0001-01-01 00:00:00 +0000 UTC"}]}
`,
			expectedStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequest := httptest.NewRequest(
				"GET",
				"/user/"+tt.args.userID.String()+"/chat-sessions",
				nil,
			)
			vars := map[string]string{
				"user_id": tt.args.userID.String(),
			}
			mockRequest = mux.SetURLVars(mockRequest, vars)
			mockResponseRecorder := httptest.NewRecorder()

			mockService.EXPECT().GetUserChatSessions(
				gomock.Any(),
				tt.args.userID,
			).Return(tt.mockServiceResponseData, tt.mockServiceResponseError)

			handler := &GetChatSessionHandler{
				ChatSessionService: mockService,
				logger:             logger,
			}
			sut := handler.GetUserChatSessionsController

			sut(mockResponseRecorder, mockRequest)

			mockResponse := mockResponseRecorder.Result()
			actual, err := io.ReadAll(mockResponse.Body)
			if err != nil {
				t.Errorf("error with response reading: %v", err)
				return
			}
			actualStatusCode := mockResponse.StatusCode

			assert.Equal(t, tt.expected, string(actual))
			assert.Equal(t, tt.expectedStatusCode, actualStatusCode)
		})
	}
}
