package repositories

import (
	"database/sql"
	"fmt"
	"gateway/internal/models"
)

type chatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) ChatRepository {
	chatRepo := &chatRepository{db: db}
	chatRepo.InitializeTables()
	return chatRepo
}

func (r *chatRepository) InitializeChatsTable() error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS chats (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			title TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)
	`
	if _, err := r.db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("error creating chats table: %w", err)
	}

	createUserIndexQuery := `CREATE INDEX IF NOT EXISTS idx_chats_user_id ON chats (user_id)`
	if _, err := r.db.Exec(createUserIndexQuery); err != nil {
		return fmt.Errorf("error creating chats user_id index: %w", err)
	}
	return nil
}

func (r *chatRepository) InitializeChatMessagesTable() error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS chat_messages (
			id SERIAL PRIMARY KEY,
			chat_id INT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
			message TEXT NOT NULL,
			response TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)
	`
	if _, err := r.db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("error creating chat_messages table: %w", err)
	}

	createChatIdIndexQuery := `CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_id ON chat_messages (chat_id)`
	if _, err := r.db.Exec(createChatIdIndexQuery); err != nil {
		return fmt.Errorf("error creating chat_messages chat_id index: %w", err)
	}
	return nil
}

func (r *chatRepository) CreateChat(chat models.DBChats) (int, error) {
	query := `
		INSERT INTO chats (user_id, title, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int
	err := r.db.QueryRow(query, chat.UserId, chat.Title, chat.CreatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error inserting chat: %w", err)
	}
	return id, nil
}

func (r *chatRepository) AddChatMessage(chatMsg models.DBChatMessage) error {
	query := `
		INSERT INTO chat_messages (chat_id, message, response, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(query, chatMsg.ChatId, chatMsg.Message, chatMsg.Response, chatMsg.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert chat message: %w", err)
	}
	return nil
}

func (r *chatRepository) GetChats(chatId int) ([]models.DBChatMessage, string, error) {
	var chatTitle string
	err := r.db.QueryRow(`SELECT title FROM chats WHERE id = $1`, chatId).Scan(&chatTitle)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get chat title: %w", err)
	}

	rows, err := r.db.Query(`
		SELECT chat_id, message, response, created_at
		FROM chat_messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
	`, chatId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query chat messages: %w", err)
	}
	defer rows.Close()

	var messages []models.DBChatMessage
	for rows.Next() {
		var msg models.DBChatMessage
		if err := rows.Scan(&msg.ChatId, &msg.Message, &msg.Response, &msg.CreatedAt); err != nil {
			return nil, "", fmt.Errorf("failed to scan chat message: %w", err)
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("error iterating chat messages: %w", err)
	}
	return messages, chatTitle, nil
}

func (r *chatRepository) GetUserChats(userId int) ([]models.DBChats, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, title, created_at
		FROM chats
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to query chats: %w", err)
	}
	defer rows.Close()

	var chats []models.DBChats
	for rows.Next() {
		var chat models.DBChats
		if err := rows.Scan(&chat.Id, &chat.UserId, &chat.Title, &chat.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chats: %w", err)
		}
		chats = append(chats, chat)
	}
	return chats, nil
}

func ChatMsgDbOllama(dbChatMsgs []models.DBChatMessage) []models.OllamaMessage {
    var result []models.OllamaMessage
    for _, dbMsg := range dbChatMsgs {
        ollamaMsg := models.OllamaMessage{
            Role:    "user",
            Content: dbMsg.Message,
        }
        ollamaResponse := models.OllamaMessage{
            Role:    "assistant",
            Content: dbMsg.Response,
        }
        result = append(result, ollamaMsg, ollamaResponse)
    }
    return result
}
func ConvertDBChatsToSummaryResponse(dbChats []models.DBChats) []models.ChatSummaryItem {
    summaries := make([]models.ChatSummaryItem, len(dbChats))

    for i, chat := range dbChats {
        summaries[i] = models.ChatSummaryItem{
            ChatId:      chat.Id,
            UserId:      chat.UserId,
            Title:       chat.Title,
            LastUpdated: chat.CreatedAt,
        }
    }

    return summaries
}

func ConvertDBChatToHistoryResponse(dbChats []models.DBChatMessage, chatId int, chatTitle string) models.ChatHistoryResponse {
	msgHistory := make([]models.ChatHistoryMessage, len(dbChats))

    for i, chat := range dbChats {
        msgHistory[i] = models.ChatHistoryMessage{
            Message:     	chat.Message,
            Response:      	chat.Response,
            CreatedAt:      chat.CreatedAt,
        }
    }

	historyResponse := models.ChatHistoryResponse{
		ChatId:		chatId,
		Title:		chatTitle,
		Messages:	msgHistory,
	}

	return historyResponse
}

func (r *chatRepository) InitializeTables() error {
	if err := r.InitializeChatsTable(); err != nil {
		return err
	}
	if err := r.InitializeChatMessagesTable(); err != nil {
		return err
	}
	return nil
}