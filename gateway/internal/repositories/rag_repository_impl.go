package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gateway/config"
	"gateway/internal/models"
)

var topK = 3 // top k ones to save in the db and return from here
type ragRepositoryHTTP struct {
	client *http.Client
	url    string
	topK   int
}

func NewRagRepositoryHTTP() RagRepository {
	cfg := config.LoadConfig()
	return &ragRepositoryHTTP{
		client: &http.Client{Timeout: 30 * time.Second},
		url:    cfg.Rag.Url,
		topK:   cfg.Rag.TopK,
	}
}

func (r *ragRepositoryHTTP) SendRagRequest(request models.RagRequest) ([]models.RagResponseItem, error) {
	if request.TopK == 0 {
		request.TopK = r.topK
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := r.client.Post(r.url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to RAG: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RAG API error: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var ragResp []models.RagResponseItem
	if err := json.Unmarshal(body, &ragResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal RAG response: %w. Body=%s", err, string(body))
	}

	return ragResp, nil
}

func (r *ragRepositoryHTTP) RetrieveAndAugmentUserPrompt(message string) ([]models.OllamaMessage, string, error) {
	ragResponse, err := r.SendRagRequest(models.RagRequest{Message: message})
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve relevant info: %w", err)
	}
	return ConvertRagResponseToOllamaMessages(message, ragResponse), ConvertRagResponseToString(ragResponse, topK), nil
}

func ConvertRagResponseToOllamaMessages(message string, response []models.RagResponseItem) []models.OllamaMessage {
	var b strings.Builder
	for i, v := range response {
		b.WriteString(fmt.Sprintf("Document %d: %s\n\n", i+1, v.Sentence))
	}
	
	return []models.OllamaMessage{
		{
			Role:    "system",
			Content: fmt.Sprintf("You must only answer based on these documents:\n\n%s", b.String()),
		},
		{
			Role:    "user",
			Content: message,
		},
	}
}

// Since we can't keep all top 30 say rag responses in the db to provide in next conversations for future context
// we need to find of a hack to only be able to keep say, top 3 in the db, but still provide top 30 for this message's
func ConvertRagResponseToString(response []models.RagResponseItem, topK int) string {
	var b strings.Builder
	for i, v := range response[:min(topK, len(response))] {
		b.WriteString(fmt.Sprintf("Document %d: %s\n\n", i+1, v.Sentence))
	}
	return b.String()
}