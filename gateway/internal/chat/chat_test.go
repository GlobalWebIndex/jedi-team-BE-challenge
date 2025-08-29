package chat

import (
	"database/sql"
	"fmt"
	"encoding/json"
	"regexp"
	"strings"
	"net/http"
	"bytes"
	"time"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"gateway/internal/models"

)

func initMockDb() (*sql.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	return db, mock, nil
}

func TestChat(t *testing.T) {
	path := "/chat"
	chatId := 123
	t.Run("POST__/chat__Missing_User_Id", func(t *testing.T) {
		expectedStatus := http.StatusBadRequest
		expectedResponse := "Missing required fields"
		body, err := json.Marshal(models.ChatMessageRequest{
			Message: "Some Message",
		})
		if err != nil {
			t.Fatalf("failed to marshal credentials: %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", path, bytes.NewBuffer(body))

		ChatHandler(rr, req, &sql.DB{}, NotExistingChat)

		if rr.Code != expectedStatus {
			t.Errorf("Expected status %d, got: %d", expectedStatus, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != expectedResponse {
			t.Errorf("Expected response %s, got: %s", expectedResponse, strings.TrimSpace(rr.Body.String()))
		}
	})
	t.Run("POST__/chat__Missing_Message", func(t *testing.T) {
		expectedStatus := http.StatusBadRequest
		expectedResponse := "Missing required fields"
		body, err := json.Marshal(models.ChatMessageRequest{
			UserId: 123,
		})
		if err != nil {
			t.Fatalf("failed to marshal credentials: %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", path, bytes.NewBuffer(body))

		ChatHandler(rr, req, &sql.DB{}, NotExistingChat)

		if rr.Code != expectedStatus {
			t.Errorf("Expected status %d, got: %d", expectedStatus, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != expectedResponse {
			t.Errorf("Expected response %s, got: %s", expectedResponse, strings.TrimSpace(rr.Body.String()))
		}
	})
	// t.Run("POST__/chat__Success", func(t *testing.T) {})
	t.Run("POST__/chat/:chatId__Missing_Non_Existent_Chat_Id", func(t *testing.T) {
		expectedResponse := "Chat not found"
		expectedStatusCode := http.StatusNotFound

		db, mock, err := initMockDb()
		if err != nil {
			t.Fatalf("Received unexpected error when initializing mock db: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT title FROM chats WHERE id = $1`)).
		WithArgs(chatId).
		WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT chat_id, message, response, created_at 
			FROM chat_messages 
			WHERE chat_id = $1 
			ORDER BY created_at ASC
		`)).
		WithArgs(chatId).
		WillReturnError(sql.ErrNoRows)

		body, err := json.Marshal(models.ChatMessageRequest{
			UserId: 123,
			Message: "Some Message",
		})
		if err != nil {
			t.Fatalf("failed to marshal credentials: %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d", path, chatId), bytes.NewBuffer(body))

		ChatHandler(rr, req, db, chatId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != expectedResponse {
			t.Errorf("Expected response %s, got: %s", expectedResponse, strings.TrimSpace(rr.Body.String()))
		}
	})
	// t.Run("POST__/chat/:chatId__Invalid_Ollama_Response", func(t *testing.T) {})
	// t.Run("POST__/chat/:chatId__Ollama_Delay_40_secs", func(t *testing.T) {})
	// t.Run("POST__/chat/:chatId__Success", func(t *testing.T) {})
}

func TestGetChat(t *testing.T) {
	path := "/chat"
	chatId := 123
	t.Run("GET__/chat/:chat_id__Non_Existent_Chat", func(t *testing.T) {
		expectedResponse := "Chat not found"
		expectedStatusCode := http.StatusNotFound

		db, mock, err := initMockDb()
		if err != nil {
			t.Fatalf("Received unexpected error when initializing mock db: %v", err)
		}
		defer db.Close()

		// eexpecting following queries :
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT title FROM chats WHERE id = $1`)).
		WithArgs(chatId).
		WillReturnError(sql.ErrNoRows)

		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT chat_id, message, response, created_at 
			FROM chat_messages 
			WHERE chat_id = $1 
			ORDER BY created_at ASC
		`)).
		WithArgs(chatId).
		WillReturnError(sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", path, chatId), nil)
		rr := httptest.NewRecorder()

		GetChatHandler(rr, req, db, chatId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != expectedResponse {
			t.Errorf("Expected response %s, got: %s", expectedResponse, strings.TrimSpace(rr.Body.String()))
		}
	})

	
	t.Run("GET__/chat/:chat_id__Success", func(t *testing.T) { 
		expectedTitle := "Best test ever"
		expectedStatusCode := http.StatusOK
		
		db, mock, err := initMockDb()
		if err != nil {
			t.Fatalf("Received unexpected error when initializing mock db: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT title FROM chats WHERE id = $1`)).
		WithArgs(chatId).
		WillReturnRows(sqlmock.NewRows([]string{"title"}).AddRow(expectedTitle))

		mockMsgs := []models.DBChatMessage{
			{ChatId: chatId, Message: "Hi", Response: "Hello", CreatedAt: time.Now()},
			{ChatId: chatId, Message: "How are you?", Response: "Fine", CreatedAt: time.Now()},
		}
		rows := sqlmock.NewRows([]string{"chat_id", "message", "response", "created_at"})
		for _, m := range mockMsgs {
			rows.AddRow(m.ChatId, m.Message, m.Response, m.CreatedAt)
		}
	
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT chat_id, message, response, created_at 
			FROM chat_messages 
			WHERE chat_id = $1 
			ORDER BY created_at ASC
		`)).
		WithArgs(chatId).
		WillReturnRows(rows)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", path, chatId), nil)
		rr := httptest.NewRecorder()

		GetChatHandler(rr, req, db, chatId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		var resp models.ChatHistoryResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.ChatId != chatId {
			t.Errorf("expected chatId %d, got %d", chatId, resp.ChatId)
		}
		if resp.Title == "" {
			t.Errorf("expected non-empty title")
		}
		for _, m := range resp.Messages {
			if m.Message == "" {
				t.Errorf("expected non-empty message")
			}
			if m.Response == "" {
				t.Errorf("expected non-empty response")
			}
		}
	})
}


func TestGetUsersChat(t *testing.T) {
	userId := 123
	path := fmt.Sprintf("/chat/users/%d", userId)
	t.Run("GET__/chat/users/:user_id__Non_Existent_Chats", func(t *testing.T) {
		expectedResponse := "Chats not found"
		expectedStatusCode := http.StatusNotFound

		db, mock, err := initMockDb()
		if err != nil {
			t.Fatalf("Received unexpected error when initializing mock db: %v", err)
		}
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, user_id, title, created_at
			FROM chats
			WHERE user_id = $1
			ORDER BY created_at ASC
		`)).
		WithArgs(userId).
		WillReturnError(sql.ErrNoRows)


		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		GetUsersChatHandler(rr, req, db, userId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != expectedResponse {
			t.Errorf("Expected response %s, got: %s", expectedResponse, strings.TrimSpace(rr.Body.String()))
		}
	})

	t.Run("GET__/chat/users/:user_id__Success", func(t *testing.T) {
		expectedStatusCode := http.StatusOK
		
		db, mock, err := initMockDb()
		if err != nil {
			t.Fatalf("Received unexpected error when initializing mock db: %v", err)
		}
		defer db.Close()

		mockMsgs := []models.DBChats{
			{Id: 1, UserId: userId, Title: "someTitle", CreatedAt: time.Now()},
			{Id: 2, UserId: userId, Title: "someTitle2", CreatedAt: time.Now()},
			{Id: 3, UserId: userId, Title: "someTitle3", CreatedAt: time.Now()},
		}
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "created_at"})
		for _, m := range mockMsgs {
			rows.AddRow(m.Id, m.UserId, m.Title, m.CreatedAt)
		}
	
		mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, user_id, title, created_at
			FROM chats
			WHERE user_id = $1
			ORDER BY created_at ASC
		`)).
		WithArgs(userId).
		WillReturnRows(rows)

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		GetUsersChatHandler(rr, req, db, userId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		var resp []models.ChatSummaryItem
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		for _, m := range resp {
			if m.UserId != userId {
				t.Errorf("expected non-empty response")
			}
			if m.Title == "" {
				t.Errorf("expected non-empty response")
			}
		}
	})
}
