package models

import "time"

type DBChats struct {
	Id			int 		`json:"id"`
	UserId		int 		`json:"user_id"`
	Title		string 	 	`json:"title"`
	CreatedAt	time.Time 	`json:"created_at"`
}

type DBChatMessage struct {
	ChatId		int			`json:"chat_id"`
	Message		string		`json:"message"`
	RagContext	string		`json:"rag_context"`
	Response	string		`json:"response"`
	CreatedAt	time.Time	`json:"created_at"`
}