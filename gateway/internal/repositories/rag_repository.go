package repositories

import "gateway/internal/models"

type RagRepository interface {
    SendRagRequest(req models.RagRequest) ([]models.RagResponseItem, error)
    RetrieveAndAugmentUserPrompt(message string) ([]models.OllamaMessage, error)
}
