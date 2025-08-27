package models

import "time"

type ChatHistoryMessage struct {
	Message   string    `json:"message"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatHistoryResponse struct {
	ChatId   int               		`json:"chat_id"`
	Title    string               	`json:"title"`
	Messages []ChatHistoryMessage 	`json:"messages"`
}