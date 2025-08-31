package mocks

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"gateway/internal/models"
)

type MockOllamaRepository struct {
	GenerateTitleError      error
	SendOllamaRequestError  error

	CustomTitleResponse     string
	CustomOllamaResponse    *models.OllamaResponse
	
	SimulateTimeout         bool
	SimulateNetworkError    bool
	SimulateInvalidResponse bool

	LastGenerateTitleContent string
	LastSendRequestURL       string
	LastSendRequest          *models.OllamaRequest
	CallCount                map[string]int
}

func NewMockOllamaRepository() *MockOllamaRepository {
	return &MockOllamaRepository{
		CallCount: make(map[string]int),
		CustomOllamaResponse: &models.OllamaResponse{
			Model:     "llama2",
			CreatedAt: time.Now().Format(time.RFC3339),
			Message: models.OllamaMessage{
				Role:    "assistant",
				Content: "This is a mock response from Ollama",
			},
			Done: true,
		},
	}
}

func NewMockOllamaRepositoryWithDefaults() *MockOllamaRepository {
	mock := NewMockOllamaRepository()
	mock.CustomTitleResponse = "Generated Chat Title"
	return mock
}

func (m *MockOllamaRepository) GenerateTitle(content string) (string, error) {
	m.CallCount["GenerateTitle"]++
	m.LastGenerateTitleContent = content
	
	if m.GenerateTitleError != nil {
		return "", m.GenerateTitleError
	}
	
	if m.SimulateTimeout {
		return "", errors.New("request timeout")
	}
	
	if m.SimulateNetworkError {
		return "", errors.New("network error: connection refused")
	}

	if m.CustomTitleResponse != "" {
		return m.CustomTitleResponse, nil
	}
	if content == "" {
		return "New Chat", nil
	}
	words := strings.Fields(content)
	if len(words) > 3 {
		return fmt.Sprintf("Chat about %s %s %s", words[0], words[1], words[2]), nil
	} else if len(words) > 0 {
		return fmt.Sprintf("Chat about %s", words[0]), nil
	}
	
	return "Generated Chat Title", nil
}

func (m *MockOllamaRepository) SendOllamaRequest(url string, req models.OllamaRequest) (*models.OllamaResponse, error) {
	m.CallCount["SendOllamaRequest"]++
	m.LastSendRequestURL = url
	m.LastSendRequest = &req
	
	if m.SendOllamaRequestError != nil {
		return nil, m.SendOllamaRequestError
	}
	
	if m.SimulateTimeout {
		return nil, errors.New("request timeout")
	}
	
	if m.SimulateNetworkError {
		return nil, errors.New("network error: connection refused")
	}
	
	if m.SimulateInvalidResponse {
		return nil, errors.New("invalid response format")
	}

	if m.CustomOllamaResponse != nil {
		response := *m.CustomOllamaResponse
		if req.Model != "" {
			response.Model = req.Model
		}
		return &response, nil
	}
	
	var responseContent string
	if len(req.Messages) > 0 {
		lastMessage := req.Messages[len(req.Messages)-1]
		responseContent = fmt.Sprintf("Mock response to: %s", lastMessage.Content)
	} else {
		responseContent = "Mock response from Ollama"
	}
	
	return &models.OllamaResponse{
		Model:     req.Model,
		CreatedAt: time.Now().Format(time.RFC3339),
		Message: models.OllamaMessage{
			Role:    "assistant",
			Content: responseContent,
		},
		Done: true,
	}, nil
}


func (m *MockOllamaRepository) SetGenerateTitleError(err error) {
	m.GenerateTitleError = err
}

func (m *MockOllamaRepository) SetSendOllamaRequestError(err error) {
	m.SendOllamaRequestError = err
}

func (m *MockOllamaRepository) SetCustomTitleResponse(title string) {
	m.CustomTitleResponse = title
}

func (m *MockOllamaRepository) SetCustomOllamaResponse(response *models.OllamaResponse) {
	m.CustomOllamaResponse = response
}

func (m *MockOllamaRepository) SetSimulateTimeout(simulate bool) {
	m.SimulateTimeout = simulate
}

func (m *MockOllamaRepository) SetSimulateNetworkError(simulate bool) {
	m.SimulateNetworkError = simulate
}

func (m *MockOllamaRepository) SetSimulateInvalidResponse(simulate bool) {
	m.SimulateInvalidResponse = simulate
}

func (m *MockOllamaRepository) GetCallCount(method string) int {
	return m.CallCount[method]
}

func (m *MockOllamaRepository) GetLastGenerateTitleContent() string {
	return m.LastGenerateTitleContent
}

func (m *MockOllamaRepository) GetLastSendRequest() *models.OllamaRequest {
	return m.LastSendRequest
}

func (m *MockOllamaRepository) GetLastSendRequestURL() string {
	return m.LastSendRequestURL
}

func (m *MockOllamaRepository) Reset() {
	m.GenerateTitleError = nil
	m.SendOllamaRequestError = nil
	m.CustomTitleResponse = ""
	m.CustomOllamaResponse = &models.OllamaResponse{
		Model:     "llama2",
		CreatedAt: time.Now().Format(time.RFC3339),
		Message: models.OllamaMessage{
			Role:    "assistant",
			Content: "This is a mock response from Ollama",
		},
		Done: true,
	}
	m.SimulateTimeout = false
	m.SimulateNetworkError = false
	m.SimulateInvalidResponse = false
	m.LastGenerateTitleContent = ""
	m.LastSendRequestURL = ""
	m.LastSendRequest = nil
	m.CallCount = make(map[string]int)
}

func (m *MockOllamaRepository) WithStreamingResponse(content string) *MockOllamaRepository {
	m.CustomOllamaResponse = &models.OllamaResponse{
		Model:     "llama2",
		CreatedAt: time.Now().Format(time.RFC3339),
		Message: models.OllamaMessage{
			Role:    "assistant",
			Content: content,
		},
		Done: false,
	}
	return m
}
func (m *MockOllamaRepository) WithCompletedResponse(content string) *MockOllamaRepository {
	m.CustomOllamaResponse = &models.OllamaResponse{
		Model:     "llama2",
		CreatedAt: time.Now().Format(time.RFC3339),
		Message: models.OllamaMessage{
			Role:    "assistant",
			Content: content,
		},
		Done: true,
	}
	return m
}

func (m *MockOllamaRepository) WithModel(model string) *MockOllamaRepository {
	if m.CustomOllamaResponse == nil {
		m.CustomOllamaResponse = &models.OllamaResponse{}
	}
	m.CustomOllamaResponse.Model = model
	return m
}