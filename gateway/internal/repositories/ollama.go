package repositories

import (	
	"encoding/json"
	"bytes"
	"net/http"
	"time"
	"fmt"
	"io"

	"gateway/config"
	"gateway/internal/models"
)

func GenerateTitle(content string) (string, error) {
	cfg := config.LoadConfig()

	// Crafts OllamaChatRequest w/ system prompt to generate title for this
	systemPrompt := models.OllamaMessage{
		Role:			"system",
		Content:		"Create a title for a conversation that starts like this. Max 4 words. Output title only, without any explanation or introduction.",
	}
	
	ollamaMessages := []models.OllamaMessage{
		systemPrompt,
		{
			Role:    "user",
			Content: content,
		},
	}
	
	maxTokens := 7 //4-5 words for the title max
	ollamaRequest := models.OllamaChatRequest{
		Model:		cfg.Ollama.Model,
		Messages:	ollamaMessages,
		Options: &models.OllamaOptions{
			NumPredict: &maxTokens,
		},
	}

	ollamaResponse, err := SendOllamaRequest(cfg.Ollama.Url, ollamaRequest)
	if err != nil {
		return "", err
	}
	
	return ollamaResponse.Message.Content, nil
}

func SendOllamaRequest(url string, requestBody models.OllamaChatRequest) (*models.OllamaResponse, error) {
	cfg := config.LoadConfig()
	requestBody.Stream = cfg.Ollama.Stream

	payload, err := json.Marshal(requestBody)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request body: %w", err)
    }

    client := &http.Client{
        Timeout: 30 * time.Second,
    }

    resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
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

    var ollamaResp models.OllamaResponse
    if err := json.Unmarshal(body, &ollamaResp); err != nil {
        return nil, fmt.Errorf("failed to unmarshal Ollama response: %w", err)
    }

    return &ollamaResp, nil
}