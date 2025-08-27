package repositories

import (
	"fmt"
	"database/sql"

	"gateway/internal/models"
)

func InitializeChatsTable(db *sql.DB) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS chats (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			title TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)
	`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("error creating chats table: %w", err)
	}

	createUserIndexQuery := `
		CREATE INDEX IF NOT EXISTS idx_chats_user_id ON chats (user_id);
	`
	_, err = db.Exec(createUserIndexQuery)
	if err != nil {
		return fmt.Errorf("error creating chats user_id index: %w", err)
	}

	return nil
}


func InitializeChatMessagesTable(db *sql.DB) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS chat_messages (
			id SERIAL PRIMARY KEY,
			chat_id INT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
			message TEXT NOT NULL,
			response TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)
	`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("error creating chat_messages table: %w", err)
	}

	createChatIdIndexQuery := `
		CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_id ON chat_messages (chat_id);
	`
	_, err = db.Exec(createChatIdIndexQuery)
	if err != nil {
		return fmt.Errorf("error creating chat_messages chat_id index: %w", err)
	}

	return nil
}


func CreateChat(chat models.DBChats, db *sql.DB) (int, error) {
	query := `
		INSERT INTO chats (user_id, title, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int
	err := db.QueryRow(query, chat.UserId, chat.Title, chat.CreatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error inserting chat: %w", err)
	}

	return id, nil
}



func AddChatMessage(chatMsg models.DBChatMessage, db *sql.DB) error {
	query  := `
		INSERT INTO chat_messages (chat_id, message, response, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(
		query,
		chatMsg.ChatId,
		chatMsg.Message,
		chatMsg.Response,
		chatMsg.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert chat message: %w", err)
	}

	return nil
}


func GetChats(chatId int, db *sql.DB) ([]models.DBChatMessage, string, error) {
	var chatTitle string

	titleQuery := `SELECT title FROM chats WHERE id = $1`
	err := db.QueryRow(titleQuery, chatId).Scan(&chatTitle)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", fmt.Errorf("chat not found: %w", err)
		}
		return nil, "", fmt.Errorf("failed to get chat title: %w", err)
	}

	// Get all chat msgs for this chat ordered by date
	messagesQuery := `
		SELECT chat_id, message, response, created_at 
		FROM chat_messages 
		WHERE chat_id = $1 
		ORDER BY created_at ASC
	`
	
	rows, err := db.Query(messagesQuery, chatId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query chat messages: %w", err)
	}
	defer rows.Close()

	var messages []models.DBChatMessage
	
	for rows.Next() {
		var msg models.DBChatMessage
		err := rows.Scan(
			&msg.ChatId,
			&msg.Message,
			&msg.Response,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan chat message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("error iterating chat messages: %w", err)
	}

	return messages, chatTitle, nil
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
