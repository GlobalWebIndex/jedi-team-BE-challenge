package chatSessions

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loukaspe/jedi-team-challenge/internal/core/domain"
	mock_services "github.com/loukaspe/jedi-team-challenge/mocks/mock_internal/core/services"
	customerrors "github.com/loukaspe/jedi-team-challenge/pkg/errors"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"io"
	"net/http/httptest"
	"testing"
)

func TestCreateUserChatSessionHandler_CreateUserChatSessionController(t *testing.T) {
	logger := logger.NewLogger(context.Background())
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_services.NewMockChatSessionServiceInterface(mockCtrl)

	type args struct {
		userId uuid.UUID
	}

	tests := []struct {
		name                  string
		args                  args
		mockServiceInsertedID uuid.UUID
		expected              []byte
		expectedStatusCode    int
	}{
		{
			name: "valid",
			args: args{
				userId: uuid.UUID{0x12, 0x34, 0x56, 0x78},
			},
			mockServiceInsertedID: uuid.UUID{0x22, 0x34, 0x56, 0x88},
			expected: json.RawMessage(`{"id":"22345688-0000-0000-0000-000000000000"}
`),
			expectedStatusCode: 201,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequest := httptest.NewRequest(
				"POST",
				"/users/"+tt.args.userId.String()+"/chat-sessions",
				nil,
			)

			vars := map[string]string{
				"user_id": tt.args.userId.String(),
			}
			mockRequest = mux.SetURLVars(mockRequest, vars)

			mockRequest.Header.Set("Content-Type", "application/json")
			mockResponseRecorder := httptest.NewRecorder()

			mockService.EXPECT().CreateChatSession(
				gomock.Any(),
				&domain.ChatSession{UserID: tt.args.userId},
			).Return(tt.mockServiceInsertedID, nil)

			handler := &CreateUserChatSessionHandler{
				ChatSessionService: mockService,
				logger:             logger,
			}
			sut := handler.CreateUserChatSessionController

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

func TestCreateUserChatSessionHandler_CreateUserChatSessionAssetControllerHasBadRequestError(t *testing.T) {
	logger := logger.NewLogger(context.Background())
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_services.NewMockChatSessionServiceInterface(mockCtrl)

	type args struct {
		userId         uuid.UUID
		userIdAsString string
	}

	tests := []struct {
		name                  string
		args                  args
		mockServiceInsertedID uuid.UUID
		expected              []byte
		expectedStatusCode    int
	}{
		{
			name: "empty user id",
			args: args{
				userIdAsString: "",
			},

			expected: json.RawMessage(`{"errorMessage":"missing user id"}
`),
			expectedStatusCode: 400,
		},
		{
			name: "user id not a uuid",
			args: args{
				userIdAsString: "55",
			},
			expected: json.RawMessage(`{"errorMessage":"malformed user uuid"}
`),
			expectedStatusCode: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRequest := httptest.NewRequest(
				"POST",
				"/users/"+tt.args.userIdAsString+"/chat-sessions",
				nil,
			)

			vars := map[string]string{
				"user_id": tt.args.userIdAsString,
			}
			mockRequest = mux.SetURLVars(mockRequest, vars)

			mockRequest.Header.Set("Content-Type", "application/json")
			mockResponseRecorder := httptest.NewRecorder()

			handler := &CreateUserChatSessionHandler{
				ChatSessionService: mockService,
				logger:             logger,
			}
			sut := handler.CreateUserChatSessionController

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

func TestCreateUserChatSessionHandler_CreateUserChatSessionAssetControllerHasServiceError(t *testing.T) {
	logger := logger.NewLogger(context.Background())
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockService := mock_services.NewMockChatSessionServiceInterface(mockCtrl)

	type args struct {
		userId uuid.UUID
	}

	tests := []struct {
		name                     string
		args                     args
		mockServiceInsertedID    uuid.UUID
		mockServiceResponseError error
		expected                 []byte
		expectedStatusCode       int
	}{
		{
			name: "service random error",
			args: args{
				userId: uuid.UUID{0x12, 0x34, 0x56, 0x78},
			},
			mockServiceInsertedID:    uuid.UUID{},
			mockServiceResponseError: errors.New("random error"),
			expected: json.RawMessage(`{"errorMessage":"error in creating chat session"}
`),
			expectedStatusCode: 500,
		},
		{
			name: "service user not found error",
			args: args{
				userId: uuid.UUID{0x12, 0x34, 0x56, 0x78},
			},
			mockServiceInsertedID: uuid.UUID{},
			mockServiceResponseError: customerrors.ResourceNotFoundErrorWrapper{
				OriginalError: errors.New("user id 667 not found"),
			},
			expected: json.RawMessage(`{}
`),
			expectedStatusCode: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRequest := httptest.NewRequest(
				"POST",
				"/users/"+tt.args.userId.String()+"/chat-sessions",
				nil,
			)

			vars := map[string]string{
				"user_id": tt.args.userId.String(),
			}
			mockRequest = mux.SetURLVars(mockRequest, vars)

			mockRequest.Header.Set("Content-Type", "application/json")
			mockResponseRecorder := httptest.NewRecorder()

			mockService.EXPECT().CreateChatSession(
				gomock.Any(),
				&domain.ChatSession{UserID: tt.args.userId},
			).Return(tt.mockServiceInsertedID, tt.mockServiceResponseError)

			handler := &CreateUserChatSessionHandler{
				ChatSessionService: mockService,
				logger:             logger,
			}
			sut := handler.CreateUserChatSessionController

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
