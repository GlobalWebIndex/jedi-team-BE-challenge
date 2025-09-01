package models

import "time"

type ChatSummaryItem struct {
	ChatId      int    		`json:"id"`
	UserId		int			`json:"user_id"`
	Title       string    	`json:"title"`
	LastUpdated time.Time 	`json:"last_updated"`
}
