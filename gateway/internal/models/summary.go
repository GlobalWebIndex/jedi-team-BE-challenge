package models

import "time"

type ChatSummaryItem struct {
	ChatId      int    `json:"chat_id"`
	Title       string    `json:"title"`
	LastUpdated time.Time `json:"last_updated"`
}