package schema

import "time"

type Message struct {
	Role      string    `bson:"role" json:"role"` // "question" or "reply"
	Text      string    `bson:"text" json:"text"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

type Conversation struct {
	SessionID string    `bson:"sessionId" json:"sessionId"`
	Messages  []Message `bson:"messages" json:"messages"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
