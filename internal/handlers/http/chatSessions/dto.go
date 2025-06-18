package chatSessions

import (
	"github.com/loukaspe/jedi-team-challenge/internal/core/domain"
)

type UserChatSessionsResponse struct {
	Sessions     []ChatSessionResponse `json:"sessions,omitempty"`
	ErrorMessage string                `json:"errorMessage,omitempty"`
}

type ChatSessionResponse struct {
	ID           string            `json:"id,omitempty"`
	Title        string            `json:"title,omitempty"`
	CreatedAt    string            `json:"createdAt,omitempty"`
	UpdatedAt    string            `json:"updatedAt,omitempty"`
	Messages     []MessageResponse `json:"messages,omitempty"`
	ErrorMessage string            `json:"errorMessage,omitempty"`
}

func ChatSessionResponseFromModel(domainChatSession *domain.ChatSession) *ChatSessionResponse {
	messages := make([]MessageResponse, len(domainChatSession.Messages))
	for i, msg := range domainChatSession.Messages {
		messages[i] = MessageResponse{
			ID:        msg.ID.String(),
			Sender:    msg.Sender,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.String(),
		}
	}

	return &ChatSessionResponse{
		ID:        domainChatSession.ID.String(),
		Title:     domainChatSession.Title,
		CreatedAt: domainChatSession.CreatedAt.String(),
		UpdatedAt: domainChatSession.UpdatedAt.String(),
		Messages:  messages,
	}
}

func UserChatSessionsResponseFromModel(domainChatSession []*domain.ChatSession) *UserChatSessionsResponse {
	sessions := make([]ChatSessionResponse, len(domainChatSession))
	for i, session := range domainChatSession {
		sessions[i] = *ChatSessionResponseFromModel(session)
	}

	return &UserChatSessionsResponse{
		Sessions: sessions,
	}
}

type MessageResponse struct {
	ID           string `json:"id,omitempty"`
	Sender       string `json:"sender,omitempty" enum:"USER,SYSTEM"`
	Content      string `json:"content,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func MessageResponseFromModel(msg *domain.Message) *MessageResponse {
	return &MessageResponse{
		ID:        msg.ID.String(),
		Sender:    msg.Sender,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt.String(),
	}
}

type SendMessageRequest struct {
	Content string `json:"content"`
}

type SendMessageResponse struct {
	UserMessage   *MessageResponse `json:"userMessage,omitempty"`
	SystemMessage *MessageResponse `json:"systemMessage,omitempty"`
	ErrorMessage  string           `json:"errorMessage,omitempty"`
}

type SubmitFeedbackRequest struct {
	Feedback string `json:"feedback"`
}

type SubmitFeedbackResponse struct {
	ErrorMessage string `json:"errorMessage,omitempty"`
}
