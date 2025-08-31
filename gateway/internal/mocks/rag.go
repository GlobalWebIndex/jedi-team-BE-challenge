package mocks

import (
	"errors"
	"fmt"
	"strings"
	"gateway/internal/models"
)

type MockRagRepository struct {
	SendRagRequestError              error
	RetrieveAndAugmentUserPromptError error

	CustomRagResponse     []models.RagResponseItem
	CustomAugmentedPrompt []models.OllamaMessage
	
	SimulateTimeout         bool
	SimulateNetworkError    bool
	SimulateNoResults       bool
	SimulateLowScoreResults bool
	
	LastRagRequest          *models.RagRequest
	LastAugmentMessage      string
	CallCount               map[string]int
	
	DefaultTopK             int
	ScoreThreshold          float64
}

func NewMockRagRepository() *MockRagRepository {
	return &MockRagRepository{
		CallCount:      make(map[string]int),
		DefaultTopK:    5,
		ScoreThreshold: 0.5,
		CustomRagResponse: []models.RagResponseItem{
			{
				Sentence: "This is a relevant document about the query topic.",
				Score:    0.95,
			},
			{
				Sentence: "Additional context that helps answer the question.",
				Score:    0.87,
			},
		},
	}
}

func NewMockRagRepositoryWithData() *MockRagRepository {
	mock := NewMockRagRepository()
	
	mock.CustomRagResponse = []models.RagResponseItem{
		{
			Sentence: "Go is a programming language developed by Google in 2009.",
			Score:    0.98,
		},
		{
			Sentence: "Go features garbage collection, type safety, and CSP-style concurrency.",
			Score:    0.92,
		},
		{
			Sentence: "The Go compiler produces statically linked native binaries.",
			Score:    0.85,
		},
	}
	
	return mock
}

func (m *MockRagRepository) SendRagRequest(req models.RagRequest) ([]models.RagResponseItem, error) {
	m.CallCount["SendRagRequest"]++
	m.LastRagRequest = &req
	
	if m.SendRagRequestError != nil {
		return nil, m.SendRagRequestError
	}
	
	if m.SimulateTimeout {
		return nil, errors.New("request timeout")
	}
	
	if m.SimulateNetworkError {
		return nil, errors.New("network error: connection refused")
	}
	
	if m.SimulateNoResults {
		return []models.RagResponseItem{}, nil
	}
	
	if m.CustomRagResponse != nil {
		topK := req.TopK
		if topK == 0 {
			topK = m.DefaultTopK
		}
		
		results := make([]models.RagResponseItem, 0)
		for i, item := range m.CustomRagResponse {
			if i >= topK {
				break
			}
			
			if m.SimulateLowScoreResults && item.Score > m.ScoreThreshold {
				continue
			}
			
			results = append(results, item)
		}
		
		return results, nil
	}
	
	return m.generateRealisticRagResponse(req), nil
}

func (m *MockRagRepository) RetrieveAndAugmentUserPrompt(message string) ([]models.OllamaMessage, error) {
	m.CallCount["RetrieveAndAugmentUserPrompt"]++
	m.LastAugmentMessage = message
	
	if m.RetrieveAndAugmentUserPromptError != nil {
		return nil, m.RetrieveAndAugmentUserPromptError
	}
	
	if m.SimulateTimeout {
		return nil, errors.New("request timeout")
	}
	
	if m.SimulateNetworkError {
		return nil, errors.New("network error: connection refused")
	}
	
	if m.CustomAugmentedPrompt != nil {
		return m.CustomAugmentedPrompt, nil
	}
	
	return m.generateRealisticAugmentedPrompt(message), nil
}

func (m *MockRagRepository) generateRealisticRagResponse(req models.RagRequest) []models.RagResponseItem {
	message := strings.ToLower(req.Message)
	var results []models.RagResponseItem
	
	topK := req.TopK
	if topK == 0 {
		topK = m.DefaultTopK
	}
	
	if strings.Contains(message, "go") || strings.Contains(message, "golang") {
		results = append(results, models.RagResponseItem{
			Sentence: "Go is a statically typed, compiled programming language.",
			Score:    0.95,
		})
		results = append(results, models.RagResponseItem{
			Sentence: "Go was designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson.",
			Score:    0.88,
		})
	} else if strings.Contains(message, "python") {
		results = append(results, models.RagResponseItem{
			Sentence: "Python is a high-level, interpreted programming language.",
			Score:    0.93,
		})
	} else if strings.Contains(message, "database") || strings.Contains(message, "sql") {
		results = append(results, models.RagResponseItem{
			Sentence: "SQL databases store data in tables with rows and columns.",
			Score:    0.91,
		})
	} else {
		results = append(results, models.RagResponseItem{
			Sentence: fmt.Sprintf("Relevant information about %s from the knowledge base.", req.Message),
			Score:    0.75,
		})
	}
	
	if len(results) > topK {
		results = results[:topK]
	}
	
	return results
}

func (m *MockRagRepository) generateRealisticAugmentedPrompt(message string) []models.OllamaMessage {
	ragResults := m.generateRealisticRagResponse(models.RagRequest{
		Message: message,
		TopK:    3,
	})
	
	var contextParts []string
	for _, result := range ragResults {
		contextParts = append(contextParts, result.Sentence)
	}
	
	context := strings.Join(contextParts, " ")
	
	messages := []models.OllamaMessage{
		{
			Role:    "system",
			Content: fmt.Sprintf("Use the following context to answer the user's question: %s", context),
		},
		{
			Role:    "user",
			Content: message,
		},
	}
	
	return messages
}

func (m *MockRagRepository) SetSendRagRequestError(err error) {
	m.SendRagRequestError = err
}

func (m *MockRagRepository) SetRetrieveAndAugmentUserPromptError(err error) {
	m.RetrieveAndAugmentUserPromptError = err
}

func (m *MockRagRepository) SetCustomRagResponse(response []models.RagResponseItem) {
	m.CustomRagResponse = response
}

func (m *MockRagRepository) SetCustomAugmentedPrompt(messages []models.OllamaMessage) {
	m.CustomAugmentedPrompt = messages
}

func (m *MockRagRepository) SetSimulateTimeout(simulate bool) {
	m.SimulateTimeout = simulate
}

func (m *MockRagRepository) SetSimulateNetworkError(simulate bool) {
	m.SimulateNetworkError = simulate
}

func (m *MockRagRepository) SetSimulateNoResults(simulate bool) {
	m.SimulateNoResults = simulate
}

func (m *MockRagRepository) SetSimulateLowScoreResults(simulate bool) {
	m.SimulateLowScoreResults = simulate
}

func (m *MockRagRepository) SetScoreThreshold(threshold float64) {
	m.ScoreThreshold = threshold
}

func (m *MockRagRepository) GetCallCount(method string) int {
	return m.CallCount[method]
}

func (m *MockRagRepository) GetLastRagRequest() *models.RagRequest {
	return m.LastRagRequest
}

func (m *MockRagRepository) GetLastAugmentMessage() string {
	return m.LastAugmentMessage
}

func (m *MockRagRepository) Reset() {
	m.SendRagRequestError = nil
	m.RetrieveAndAugmentUserPromptError = nil
	m.CustomRagResponse = []models.RagResponseItem{
		{
			Sentence: "This is a relevant document about the query topic.",
			Score:    0.95,
		},
		{
			Sentence: "Additional context that helps answer the question.",
			Score:    0.87,
		},
	}
	m.CustomAugmentedPrompt = nil
	m.SimulateTimeout = false
	m.SimulateNetworkError = false
	m.SimulateNoResults = false
	m.SimulateLowScoreResults = false
	m.LastRagRequest = nil
	m.LastAugmentMessage = ""
	m.CallCount = make(map[string]int)
	m.ScoreThreshold = 0.5
}

func (m *MockRagRepository) WithHighScoreResults(sentences ...string) *MockRagRepository {
	m.CustomRagResponse = make([]models.RagResponseItem, 0, len(sentences))
	for i, sentence := range sentences {
		score := 0.95 - float64(i)*0.05
		m.CustomRagResponse = append(m.CustomRagResponse, models.RagResponseItem{
			Sentence: sentence,
			Score:    score,
		})
	}
	return m
}

func (m *MockRagRepository) WithLowScoreResults(sentences ...string) *MockRagRepository {
	m.CustomRagResponse = make([]models.RagResponseItem, 0, len(sentences))
	for i, sentence := range sentences {
		score := 0.4 - float64(i)*0.05
		m.CustomRagResponse = append(m.CustomRagResponse, models.RagResponseItem{
			Sentence: sentence,
			Score:    score,
		})
	}
	return m
}

func (m *MockRagRepository) WithSystemPrompt(systemContent string, userMessage string) *MockRagRepository {
	m.CustomAugmentedPrompt = []models.OllamaMessage{
		{
			Role:    "system",
			Content: systemContent,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}
	return m
}

func (m *MockRagRepository) WithEmptyContext(userMessage string) *MockRagRepository {
	m.CustomAugmentedPrompt = []models.OllamaMessage{
		{
			Role:    "user",
			Content: userMessage,
		},
	}
	return m
}