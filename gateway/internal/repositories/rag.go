package repositories

import (
	"net/http"
	"fmt"
	"encoding/json"
	"bytes"
	"strings"
	"io"
	"time"

	"gateway/config"
	"gateway/internal/models"
)

func SendRagRequest(request models.RagRequest) ([]models.RagResponseItem, error) {
	cfg := config.LoadConfig()
	if request.TopK == 0 {
		request.TopK = cfg.Rag.TopK
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(cfg.Rag.Url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API error: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var ragResp []models.RagResponseItem
	if err := json.Unmarshal(body, &ragResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Ollama response: %w. Body=%s", err, string(body))
	}

	return ragResp, nil
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

func RetrieveAndAugmentUserPrompt(message string) ([]models.OllamaMessage, error) {
	ragResponse, err := SendRagRequest(models.RagRequest{Message: message})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve relevant info: %w", err)
	}

	return ConvertRagResponseToOllamaMessages(message, ragResponse), nil
}
