package domain

import (
	"fmt"
	"github.com/google/uuid"
	"time"
)

type ChatSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []*Message
}

type Message struct {
	ID            uuid.UUID
	ChatSessionID uuid.UUID
	Sender        string
	Content       string
	CreatedAt     time.Time
	Feedback      *string
}

func (m Message) String() string {
	return fmt.Sprintf(
		"Message(ID: %s, ChatSessionID: %s, Sender: %s, Content: %s, CreatedAt: %s, Feedback: %v)",
		m.ID.String(),
		m.ChatSessionID.String(),
		m.Sender,
		m.Content,
		m.CreatedAt.Format(time.RFC3339),
		m.Feedback,
	)
}
