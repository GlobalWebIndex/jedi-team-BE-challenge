package models

type RagRequest struct {
	Message		string 	`json:"message"`
	TopK		int		`json:"k"`
}

type RagResponseItem struct {
	Sentence	string 	`json:"sentence"`
	Score		float64	`json:"score"`
}
