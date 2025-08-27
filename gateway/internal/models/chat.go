package models

import "time"

type ChatMessageRequest struct {
	UserId  int 	`json:"user_id"`
	Message string 	`json:"message"`
}

type ChatMessageResponse struct {
	ChatId    int    	`json:"chat_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
}
