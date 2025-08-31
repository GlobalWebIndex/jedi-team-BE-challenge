package repositories

import "gateway/internal/models"

type OllamaRepository interface {
	GenerateTitle(content string) (string, error)

	SendOllamaRequest(url string, req models.OllamaRequest) (*models.OllamaResponse, error)
}
