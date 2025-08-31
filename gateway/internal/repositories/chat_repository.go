package repositories

import "gateway/internal/models"

type ChatRepository interface {
    InitializeChatsTable() error
    InitializeChatMessagesTable() error
    CreateChat(chat models.DBChats) (int, error)
    AddChatMessage(chatMsg models.DBChatMessage) error
    GetChats(chatId int) ([]models.DBChatMessage, string, error)
    GetUserChats(userId int) ([]models.DBChats, error)
	InitializeTables() error
}
