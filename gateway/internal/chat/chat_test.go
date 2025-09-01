package chat

import (
	"fmt"
	"encoding/json"
	"strings"
	"errors"
	"net/http"
	"bytes"
	"time"
	"net/http/httptest"
	"testing"

	"gateway/internal/models"
	"gateway/internal/mocks"

)

func TestChat(t *testing.T) {
	path := "/chat"
	chatId := 123
	const NotExistingChat = -1

	t.Run("POST__/chat__Missing_User_Id", func(t *testing.T) {
		expectedStatus := http.StatusBadRequest
		expectedResponse := "Missing required fields"

		mockChatRepo := mocks.NewMockChatRepository()
		mockOllamaRepo := mocks.NewMockOllamaRepository()
		mockRagRepo := mocks.NewMockRagRepository()

		body, err := json.Marshal(models.ChatMessageRequest{
			Message: "Some Message",
		})
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", path, bytes.NewBuffer(body))

		ChatHandler(rr, req, mockChatRepo, mockRagRepo, mockOllamaRepo, NotExistingChat)

		if rr.Code != expectedStatus {
			t.Errorf("Expected status %d, got: %d", expectedStatus, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != expectedResponse {
			t.Errorf("Expected response %s, got: %s", expectedResponse, strings.TrimSpace(rr.Body.String()))
		}

		if mockChatRepo.GetCallCount("CreateChat") > 0 {
			t.Error("Expected no chat repository calls on validation failure")
		}
	})

	t.Run("POST__/chat__Missing_Message", func(t *testing.T) {
		expectedStatus := http.StatusBadRequest
		expectedResponse := "Missing required fields"

		mockChatRepo := mocks.NewMockChatRepository()
		mockOllamaRepo := mocks.NewMockOllamaRepository()
		mockRagRepo := mocks.NewMockRagRepository()

		body, err := json.Marshal(models.ChatMessageRequest{
			UserId: 123,
		})
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", path, bytes.NewBuffer(body))

		ChatHandler(rr, req, mockChatRepo, mockRagRepo, mockOllamaRepo, NotExistingChat)

		if rr.Code != expectedStatus {
			t.Errorf("Expected status %d, got: %d", expectedStatus, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != expectedResponse {
			t.Errorf("Expected response %s, got: %s", expectedResponse, strings.TrimSpace(rr.Body.String()))
		}
	})

	t.Run("POST__/chat__Success_New_Chat", func(t *testing.T) {
		expectedStatus := http.StatusOK

		mockChatRepo := mocks.NewMockChatRepository()
		mockOllamaRepo := mocks.NewMockOllamaRepository()
		mockRagRepo := mocks.NewMockRagRepository()

		mockRagRepo.WithHighScoreResults(
			"Go is a programming language developed by Google.",
			"Go features garbage collection and strong typing.",
		).WithSystemPrompt(
			"Use this context to answer: Go is a programming language developed by Google.",
			"What is Go programming language?",
		)

		mockOllamaRepo.WithCompletedResponse("Go is a statically typed, compiled programming language designed at Google.")

		mockOllamaRepo.SetCustomTitleResponse("Discussion about Go Programming")

		body, err := json.Marshal(models.ChatMessageRequest{
			UserId:  123,
			Message: "What is Go programming language?",
		})
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", path, bytes.NewBuffer(body))

		ChatHandler(rr, req, mockChatRepo, mockRagRepo, mockOllamaRepo, NotExistingChat)

		if rr.Code != expectedStatus {
			t.Errorf("Expected status %d, got: %d", expectedStatus, rr.Code)
		}

		if mockChatRepo.GetCallCount("CreateChat") != 1 {
			t.Error("Expected CreateChat to be called once for new chat")
		}

		if mockRagRepo.GetCallCount("RetrieveAndAugmentUserPrompt") != 1 {
			t.Error("Expected RAG augmentation to be called once")
		}

		if mockOllamaRepo.GetCallCount("SendOllamaRequest") != 1 {
			t.Error("Expected Ollama request to be called once")
		}

		if mockOllamaRepo.GetCallCount("GenerateTitle") != 1 {
			t.Error("Expected title generation to be called once for new chat")
		}

		var response models.ChatMessageResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.ChatId == 0 {
			t.Error("Expected non-zero chat ID in response")
		}

		if response.Response == "" {
			t.Error("Expected non-empty response")
		}
	})

	t.Run("POST__/chat/:chatId__Non_Existent_Chat_Id", func(t *testing.T) {
		// expectedResponse := "Chat not found"
		expectedStatusCode := http.StatusNotFound

		// Setup mock repositories
		mockChatRepo := mocks.NewMockChatRepository()
		mockOllamaRepo := mocks.NewMockOllamaRepository()
		mockRagRepo := mocks.NewMockRagRepository()

		mockChatRepo.SetSimulateNoRows(true)

		body, err := json.Marshal(models.ChatMessageRequest{
			UserId:  123,
			Message: "Some Message",
		})
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d", path, chatId), bytes.NewBuffer(body))

		ChatHandler(rr, req, mockChatRepo, mockRagRepo, mockOllamaRepo, chatId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		if mockChatRepo.GetCallCount("GetChats") != 1 {
			t.Error("Expected GetChats to be called to verify chat existence")
		}

		if mockOllamaRepo.GetCallCount("SendOllamaRequest") > 0 {
			t.Error("Expected no Ollama requests for non-existent chat")
		}
	})

	t.Run("POST__/chat/:chatId__Ollama_Network_Error", func(t *testing.T) {
		expectedStatusCode := http.StatusInternalServerError

		mockChatRepo := mocks.NewMockChatRepository()
		mockOllamaRepo := mocks.NewMockOllamaRepository()
		mockRagRepo := mocks.NewMockRagRepository()

		existingChat := models.DBChats{
			Id:        chatId,
			UserId:    123,
			Title:     "Existing Chat",
			CreatedAt: time.Now(),
		}
		mockChatRepo.AddTestChat(existingChat)

		mockOllamaRepo.SetSimulateNetworkError(true)

		mockRagRepo.WithSystemPrompt("Context from RAG", "User question")

		body, err := json.Marshal(models.ChatMessageRequest{
			UserId:  123,
			Message: "Test message",
		})
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d", path, chatId), bytes.NewBuffer(body))

		ChatHandler(rr, req, mockChatRepo, mockRagRepo, mockOllamaRepo, chatId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		if mockRagRepo.GetCallCount("RetrieveAndAugmentUserPrompt") != 1 {
			t.Error("Expected RAG to be called before Ollama failure")
		}

		if mockOllamaRepo.GetCallCount("SendOllamaRequest") != 1 {
			t.Error("Expected Ollama request to be attempted")
		}

		if mockChatRepo.GetCallCount("AddChatMessage") > 0 {
			t.Error("Expected no message to be saved on Ollama failure")
		}
	})

	t.Run("POST__/chat/:chatId__RAG_Service_Unavailable", func(t *testing.T) {
		expectedStatusCode := http.StatusInternalServerError

		mockChatRepo := mocks.NewMockChatRepository()
		mockOllamaRepo := mocks.NewMockOllamaRepository()
		mockRagRepo := mocks.NewMockRagRepository()

		existingChat := models.DBChats{
			Id:        chatId,
			UserId:    123,
			Title:     "Existing Chat",
			CreatedAt: time.Now(),
		}
		mockChatRepo.AddTestChat(existingChat)

		mockRagRepo.SetRetrieveAndAugmentUserPromptError(errors.New("RAG service unavailable"))

		body, err := json.Marshal(models.ChatMessageRequest{
			UserId:  123,
			Message: "Test message",
		})
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d", path, chatId), bytes.NewBuffer(body))

		ChatHandler(rr, req, mockChatRepo, mockRagRepo, mockOllamaRepo, chatId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		if mockRagRepo.GetCallCount("RetrieveAndAugmentUserPrompt") != 1 {
			t.Error("Expected RAG to be called and fail")
		}

		if mockOllamaRepo.GetCallCount("SendOllamaRequest") > 0 {
			t.Error("Expected no Ollama requests after RAG failure")
		}
	})

	t.Run("POST__/chat/:chatId__Success_Existing_Chat", func(t *testing.T) {
		expectedStatus := http.StatusOK

		mockChatRepo := mocks.NewMockChatRepository()
		mockOllamaRepo := mocks.NewMockOllamaRepository()
		mockRagRepo := mocks.NewMockRagRepository()

		existingChat := models.DBChats{
			Id:        chatId,
			UserId:    123,
			Title:     "Programming Discussion",
			CreatedAt: time.Now(),
		}
		mockChatRepo.AddTestChat(existingChat)

		mockChatRepo.AddTestMessage(chatId, models.DBChatMessage{
			ChatId:    chatId,
			Message:   "Previous question",
			Response:  "Previous answer",
			CreatedAt: time.Now(),
		})

		mockRagRepo.WithHighScoreResults(
			"Python is a high-level programming language.",
			"Python supports object-oriented programming.",
		).WithSystemPrompt(
			"Use this context: Python is a high-level programming language.",
			"Tell me about Python features",
		)

		mockOllamaRepo.WithCompletedResponse("Python is known for its simplicity and readability, making it great for beginners.")

		body, err := json.Marshal(models.ChatMessageRequest{
			UserId:  123,
			Message: "Tell me about Python features",
		})
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d", path, chatId), bytes.NewBuffer(body))

		ChatHandler(rr, req, mockChatRepo, mockRagRepo, mockOllamaRepo, chatId)

		if rr.Code != expectedStatus {
			t.Errorf("Expected status %d, got: %d", expectedStatus, rr.Code)
		}

		if mockChatRepo.GetCallCount("GetChats") != 1 {
			t.Error("Expected GetChats to verify chat exists")
		}

		if mockChatRepo.GetCallCount("CreateChat") > 0 {
			t.Error("Expected no CreateChat for existing chat")
		}

		if mockOllamaRepo.GetCallCount("GenerateTitle") > 0 {
			t.Error("Expected no title generation for existing chat")
		}

		if mockRagRepo.GetCallCount("RetrieveAndAugmentUserPrompt") != 1 {
			t.Error("Expected RAG augmentation to be called")
		}

		if mockOllamaRepo.GetCallCount("SendOllamaRequest") != 1 {
			t.Error("Expected Ollama request to be called")
		}

		if mockChatRepo.GetCallCount("AddChatMessage") != 1 {
			t.Error("Expected AddChatMessage to save the conversation")
		}

		var response models.ChatMessageResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.ChatId != chatId {
			t.Errorf("Expected chat ID %d, got %d", chatId, response.ChatId)
		}

		if response.Response == "" {
			t.Error("Expected non-empty response")
		}
	})

	t.Run("POST__/chat/:chatId__Database_Error_On_Save", func(t *testing.T) {
		expectedStatusCode := http.StatusInternalServerError

		mockChatRepo := mocks.NewMockChatRepository()
		mockOllamaRepo := mocks.NewMockOllamaRepository()
		mockRagRepo := mocks.NewMockRagRepository()

		existingChat := models.DBChats{
			Id:        chatId,
			UserId:    123,
			Title:     "Test Chat",
			CreatedAt: time.Now(),
		}
		mockChatRepo.AddTestChat(existingChat)

		mockRagRepo.WithSystemPrompt("Context", "Question")
		mockOllamaRepo.WithCompletedResponse("Good response")

		mockChatRepo.SetAddChatMessageError(errors.New("database connection failed"))

		body, err := json.Marshal(models.ChatMessageRequest{
			UserId:  123,
			Message: "Test message",
		})
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("%s/%d", path, chatId), bytes.NewBuffer(body))

		ChatHandler(rr, req, mockChatRepo, mockRagRepo, mockOllamaRepo, chatId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		if mockRagRepo.GetCallCount("RetrieveAndAugmentUserPrompt") != 1 {
			t.Error("Expected RAG to be called")
		}

		if mockOllamaRepo.GetCallCount("SendOllamaRequest") != 1 {
			t.Error("Expected Ollama to be called")
		}

		if mockChatRepo.GetCallCount("AddChatMessage") != 1 {
			t.Error("Expected AddChatMessage to be called (and fail)")
		}
	})
}

func TestGetChat(t *testing.T) {
	path := "/chat"
	chatId := 123

	t.Run("GET__/chat/:chat_id__Non_Existent_Chat", func(t *testing.T) {
		expectedResponse := "Chat not found"
		expectedStatusCode := http.StatusNotFound

		mockRepo := mocks.NewMockChatRepository()
		mockRepo.SetSimulateNoRows(true)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", path, chatId), nil)
		rr := httptest.NewRecorder()

		GetChatHandler(rr, req, mockRepo, chatId)

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

		mockRepo := mocks.NewMockChatRepository()

		testChat := models.DBChats{
			Id:        chatId,
			UserId:    456,
			Title:     expectedTitle,
			CreatedAt: time.Now(),
		}
		mockRepo.AddTestChat(testChat)

		testMessages := []models.DBChatMessage{
			{ChatId: chatId, Message: "Hi", Response: "Hello", CreatedAt: time.Now()},
			{ChatId: chatId, Message: "How are you?", Response: "Fine", CreatedAt: time.Now()},
		}

		for _, msg := range testMessages {
			mockRepo.AddTestMessage(chatId, msg)
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", path, chatId), nil)
		rr := httptest.NewRecorder()

		GetChatHandler(rr, req, mockRepo, chatId)

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

		if resp.Title != expectedTitle {
			t.Errorf("expected title '%s', got '%s'", expectedTitle, resp.Title)
		}

		if len(resp.Messages) != len(testMessages) {
			t.Errorf("expected %d messages, got %d", len(testMessages), len(resp.Messages))
		}

		for i, m := range resp.Messages {
			if m.Message != testMessages[i].Message {
				t.Errorf("expected message '%s', got '%s'", testMessages[i].Message, m.Message)
			}
			if m.Response != testMessages[i].Response {
				t.Errorf("expected response '%s', got '%s'", testMessages[i].Response, m.Response)
			}
		}
	})

	t.Run("GET__/chat/:chat_id__Chat_Exists_But_No_Messages", func(t *testing.T) {
		expectedTitle := "Empty chat"
		expectedStatusCode := http.StatusOK

		mockRepo := mocks.NewMockChatRepository()

		testChat := models.DBChats{
			Id:        chatId,
			UserId:    456,
			Title:     expectedTitle,
			CreatedAt: time.Now(),
		}
		mockRepo.AddTestChat(testChat)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", path, chatId), nil)
		rr := httptest.NewRecorder()

		GetChatHandler(rr, req, mockRepo, chatId)

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

		if resp.Title != expectedTitle {
			t.Errorf("expected title '%s', got '%s'", expectedTitle, resp.Title)
		}

		if len(resp.Messages) != 0 {
			t.Errorf("expected 0 messages, got %d", len(resp.Messages))
		}
	})

	t.Run("GET__/chat/:chat_id__Database_Error", func(t *testing.T) {
		expectedStatusCode := http.StatusInternalServerError

		mockRepo := mocks.NewMockChatRepository()
		mockRepo.SetGetChatsError(errors.New("database connection failed"))

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", path, chatId), nil)
		rr := httptest.NewRecorder()

		GetChatHandler(rr, req, mockRepo, chatId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}
	})

	t.Run("GET__/chat/:chat_id__Repository_Error_Simulation", func(t *testing.T) {
		expectedResponse := "Chat not found"
		expectedStatusCode := http.StatusNotFound

		mockRepo := mocks.NewMockChatRepository()
		mockRepo.SetGetChatsError(errors.New("chat not found"))

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", path, chatId), nil)
		rr := httptest.NewRecorder()

		GetChatHandler(rr, req, mockRepo, chatId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != expectedResponse {
			t.Errorf("Expected response %s, got: %s", expectedResponse, strings.TrimSpace(rr.Body.String()))
		}
	})
}


func TestGetUsersChat(t *testing.T) {
	userId := 123
	path := fmt.Sprintf("/chat/users/%d", userId)
	
	t.Run("GET__/chat/users/:user_id__Non_Existent_Chats", func(t *testing.T) {
		expectedResponse := "Chats not found"
		expectedStatusCode := http.StatusNotFound

		mockRepo := mocks.NewMockChatRepository()
		mockRepo.SetSimulateNoRows(true)

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		GetUsersChatHandler(rr, req, mockRepo, userId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		if strings.TrimSpace(rr.Body.String()) != expectedResponse {
			t.Errorf("Expected response %s, got: %s", expectedResponse, strings.TrimSpace(rr.Body.String()))
		}
	})

	t.Run("GET__/chat/users/:user_id__Success", func(t *testing.T) {
		expectedStatusCode := http.StatusOK
		

		mockRepo := mocks.NewMockChatRepository()
		
		testChats := []models.DBChats{
			{Id: 1, UserId: userId, Title: "someTitle", CreatedAt: time.Now()},
			{Id: 2, UserId: userId, Title: "someTitle2", CreatedAt: time.Now()},
			{Id: 3, UserId: userId, Title: "someTitle3", CreatedAt: time.Now()},
		}
		
		for _, chat := range testChats {
			mockRepo.AddTestChat(chat)
		}

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		GetUsersChatHandler(rr, req, mockRepo, userId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		var resp []models.ChatSummaryItem
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp) != len(testChats) {
			t.Errorf("Expected %d chats, got %d", len(testChats), len(resp))
		}

		for _, m := range resp {
			if m.UserId != userId {
				t.Errorf("expected userId %d, got %d", userId, m.UserId)
			}
			if m.Title == "" {
				t.Errorf("expected non-empty title")
			}
		}
	})

	t.Run("GET__/chat/users/:user_id__Empty_Results", func(t *testing.T) {
		expectedStatusCode := http.StatusOK

		mockRepo := mocks.NewMockChatRepository()
		mockRepo.SetSimulateEmptyResults(true)

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		GetUsersChatHandler(rr, req, mockRepo, userId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}

		var resp []models.ChatSummaryItem
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp) != 0 {
			t.Errorf("Expected empty response, got %d items", len(resp))
		}
	})

	t.Run("GET__/chat/users/:user_id__Database_Error", func(t *testing.T) {
		expectedStatusCode := http.StatusInternalServerError
		
		mockRepo := mocks.NewMockChatRepository()
		mockRepo.SetGetUserChatsError(errors.New("database connection failed"))

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()

		GetUsersChatHandler(rr, req, mockRepo, userId)

		if rr.Code != expectedStatusCode {
			t.Errorf("Expected status %d, got: %d", expectedStatusCode, rr.Code)
		}
	})
}