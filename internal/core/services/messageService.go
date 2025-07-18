package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/loukaspe/jedi-team-challenge/internal/core/domain"
	"github.com/loukaspe/jedi-team-challenge/internal/core/ports"
	"github.com/loukaspe/jedi-team-challenge/internal/repositories"
	customerrors "github.com/loukaspe/jedi-team-challenge/pkg/errors"
	"github.com/loukaspe/jedi-team-challenge/pkg/helpers"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"github.com/openai/openai-go"
	"strings"
	"time"
)

type MessageServiceInterface interface {
	CreateMessage(context.Context, uuid.UUID, *domain.Message) (uuid.UUID, error)
	GetAnswerForMessage(context.Context, uuid.UUID) (*domain.Message, error)
	StreamAnswerForMessage(context.Context, uuid.UUID) (<-chan string, <-chan error)
	UpdateMessageFeedback(ctx context.Context, message *domain.Message, userID uuid.UUID) error
}

type Embedder interface {
	Embed(context.Context, []string) ([]*domain.Embeddings, error)
}

type VectorDB interface {
	SemanticSearch(ctx context.Context, embeddings []float32) ([]string, error)
}

type MessageService struct {
	logger                logger.LoggerInterface
	messageRepository     ports.MessageRepositoryInterface
	chatSessionRepository ports.ChatSessionRepositoryInterface
	embedder              Embedder
	vectorDB              VectorDB
	openAIClient          *openai.Client
}

func NewMessageService(
	logger logger.LoggerInterface,
	messageRepositoryInterface ports.MessageRepositoryInterface,
	chatSessionRepository ports.ChatSessionRepositoryInterface,
	embedder Embedder,
	vectorDB VectorDB,
	openAIClient *openai.Client,
) *MessageService {
	return &MessageService{
		logger:                logger,
		messageRepository:     messageRepositoryInterface,
		chatSessionRepository: chatSessionRepository,
		embedder:              embedder,
		vectorDB:              vectorDB,
		openAIClient:          openAIClient,
	}
}

func (s *MessageService) CreateMessage(ctx context.Context, userID uuid.UUID, message *domain.Message) (uuid.UUID, error) {
	chatSession, err := s.chatSessionRepository.GetChatSession(ctx, message.ChatSessionID)
	if err != nil {
		return uuid.Nil, err
	}

	if chatSession.UserID != userID {
		return uuid.Nil, customerrors.NewUserMismatchError(message.ChatSessionID.String(), userID.String())
	}

	return s.messageRepository.CreateMessage(ctx, message)
}

func (s *MessageService) UpdateMessageFeedback(ctx context.Context, message *domain.Message, userID uuid.UUID) error {
	chatSession, err := s.chatSessionRepository.GetChatSession(ctx, message.ChatSessionID)
	if err != nil {
		return err
	}

	_, err = s.chatSessionRepository.GetChatSession(ctx, message.ChatSessionID)
	if err != nil {
		return err
	}

	if chatSession.UserID != userID {
		return customerrors.NewUserMismatchError(message.ChatSessionID.String(), userID.String())
	}

	return s.messageRepository.UpdateMessageFeedback(ctx, message.ID, *message.Feedback)
}

func (s *MessageService) GetAnswerForMessage(ctx context.Context, initialMessageID uuid.UUID) (*domain.Message, error) {
	initialMessage, err := s.messageRepository.GetMessage(ctx, initialMessageID)
	if err != nil {
		return nil, err
	}

	chatSession, err := s.chatSessionRepository.GetChatSession(ctx, initialMessage.ChatSessionID)
	if err != nil {
		return nil, err
	}

	if chatSession.Title == "" {
		title, err := s.generateTitleFromOpenAI(ctx, initialMessage.Content)
		if err != nil {
			return nil, err
		}

		err = s.chatSessionRepository.UpdateChatSessionTitle(ctx, initialMessage.ChatSessionID, title)
		if err != nil {
			return nil, err
		}
	}

	domainEmbeddings, err := s.embedder.Embed(ctx, []string{initialMessage.Content})
	if err != nil {
		return nil, err
	}

	// we only have on text so we only care for the first embedding row
	vectorToFloat32 := helpers.Float64ToFloat32(domainEmbeddings[0].Embeddings)

	accumulatedTextFromSearch, err := s.vectorDB.SemanticSearch(ctx, vectorToFloat32)

	var answer string
	if len(accumulatedTextFromSearch) == 0 {
		answer = "The force is not strong enough for me to answer that question based on my context."
	} else {
		answer, err = s.generateAnswerFromOpenAI(ctx, accumulatedTextFromSearch, initialMessage.Content, chatSession.Messages)
		if err != nil {
			return nil, err
		}
	}

	replyMessage := &domain.Message{
		ChatSessionID: initialMessage.ChatSessionID,
		Content:       answer,
		Sender:        repositories.SYSTEM_SENDER,
	}

	insertedMessageID, err := s.messageRepository.CreateMessage(ctx, replyMessage)
	if err != nil {
		return nil, err
	}

	replyMessage.ID = insertedMessageID

	return replyMessage, nil
}

func (s *MessageService) StreamAnswerForMessage(ctx context.Context, initialMessageID uuid.UUID) (<-chan string, <-chan error) {
	tokenChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(tokenChan)
		defer close(errChan)

		response, err := s.GetAnswerForMessage(ctx, initialMessageID)
		if err != nil {
			errChan <- err
			return
		}

		// split into tokens (simulate streaming)
		tokens := strings.Split(response.String(), " ")

		for _, token := range tokens {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case tokenChan <- token:
				time.Sleep(100 * time.Millisecond) // simulate generation delay
			}
		}
	}()

	return tokenChan, errChan
}

func (s *MessageService) generateAnswerFromOpenAI(ctx context.Context, text []string, initialMessage string, previousMessages []*domain.Message) (string, error) {
	prompt := fmt.Sprintf(`Use the following context to answer the question.
		Context:
		%s
		
		Question:
		%s
		
		Answer:`,
		strings.Join(text, "\n"),
		initialMessage,
	)

	completionParams := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
			openai.SystemMessage("Use only the provided context for answering the question."),
		},
		Model: openai.ChatModelGPT4_1Nano,
	}

	if len(previousMessages) >= 0 {
		for _, msg := range previousMessages {
			completionParams.Messages = append(completionParams.Messages, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: openai.String(msg.Content),
					},
				},
			})
		}
	}

	chatCompletion, err := s.openAIClient.Chat.Completions.New(ctx, completionParams)
	if err != nil {
		return "", err
	}

	if chatCompletion.Choices[0].Message.Content == "" {
		return "", errors.New("received empty response from LLM")

	}

	return chatCompletion.Choices[0].Message.Content, nil
}

func (s *MessageService) generateTitleFromOpenAI(ctx context.Context, initialMessage string) (string, error) {
	prompt := fmt.Sprintf(`Summarize the following user message into a short, descriptive chat title (max 5 words):
		%s
		Answer:`,
		initialMessage,
	)

	chatCompletion, err := s.openAIClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4_1Nano,
	})
	if err != nil {
		return "", err
	}

	if chatCompletion.Choices[0].Message.Content == "" {
		return "", errors.New("received empty response from LLM")

	}

	return chatCompletion.Choices[0].Message.Content, nil
}
