package models

type OllamaMessage struct {
	Role    string `json:"role"` // system, user or assistant -> use system for context instructions
	Content string `json:"content"`
}
type OllamaOptions struct {
	NumPredict *int `json:"num_predict,omitempty"`
}

type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *OllamaOptions  `json:"options,omitempty"`
}

type OllamaResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
}
