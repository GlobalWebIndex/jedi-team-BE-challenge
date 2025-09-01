
package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const gatewayURL = "http://gateway:8080"

type ChatRequest struct {
	UserID  int    `json:"user_id"`
	Message string `json:"message"`
}

type ChatResponse struct {
	ChatID    int       `json:"chat_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatHistoryResponse struct {
	ChatID  int `json:"chat_id"`
	Title   string `json:"title"`
	Messages []struct {
		Message   string    `json:"message"`
		Response  string    `json:"response"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"messages"`
}

type UserChatSummary struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	LastUpdated time.Time `json:"last_updated"`
}

func TestGatewayHealth(t *testing.T) {
    resp, err := http.Get("http://gateway:8080/health")
    if err != nil {
        t.Fatal(err)
    }
    if resp.StatusCode != 200 {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
}

func TestChatFlow(t *testing.T) {
	startReq := ChatRequest{UserID: 5, Message: "Hello Ollama!"}
	startResp := ChatResponse{}
	postJSON(t, "/chat", startReq, &startResp)

	if startResp.ChatID == 0 {
		t.Fatal("Expected a valid chat_id")
	}
	if startResp.Response == "" {
		t.Fatal("Expected a response from Ollama")
	}
	if startResp.Title == "" {
		t.Fatal("Expected a title from Ollama")
	}

	t.Logf("Started chat: %+v", startResp)

	continueReq := ChatRequest{UserID: 5, Message: "Continue the chat"}
	continueResp := ChatResponse{}
	postJSON(t, fmt.Sprintf("/chat/%d", startResp.ChatID), continueReq, &continueResp)

	if continueResp.ChatID != startResp.ChatID {
		t.Fatalf("Expected chat_id %d, got %d", startResp.ChatID, continueResp.ChatID)
	}
	t.Logf("Continued chat: %+v", continueResp)

	historyResp := ChatHistoryResponse{}
	getJSON(t, fmt.Sprintf("/chat/%d/history", startResp.ChatID), &historyResp)

	if len(historyResp.Messages) < 2 {
		t.Fatalf("Expected at least 2 messages in history, got %d", len(historyResp.Messages))
	}
	t.Logf("Chat history: %+v", historyResp.Messages)

	summaries := []UserChatSummary{}
	getJSON(t, fmt.Sprintf("/chat/users/%d", startReq.UserID), &summaries)

	if len(summaries) == 0 {
		t.Fatal("Expected at least one chat summary")
	}
	t.Logf("User chat summaries: %+v", summaries)
}

func postJSON(t *testing.T, path string, body interface{}, out interface{}) {
	t.Helper()
	data, _ := json.Marshal(body)
	resp, err := http.Post(gatewayURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST %s failed: %v", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s returned status %d", path, resp.StatusCode)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
	}
}

func getJSON(t *testing.T, path string, out interface{}) {
	t.Helper()
	resp, err := http.Get(gatewayURL + path)
	if err != nil {
		t.Fatalf("GET %s failed: %v", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned status %d", path, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}