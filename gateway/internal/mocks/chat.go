package mocks

import (
	"database/sql"
	"errors"
	"time"
	"gateway/internal/models"
)

type MockChatRepository struct {
	chats        						map[int]models.DBChats
	chatMessages 						map[int][]models.DBChatMessage
	nextChatId   						int

	InitializeChatsTableError        	error
	InitializeChatMessagesTableError 	error
	InitializeTablesError           	error
	CreateChatError                 	error
	AddChatMessageError             	error
	GetChatsError                   	error
	GetUserChatsError               	error

	CallCount							map[string]int

	SimulateNoRows       				bool
	SimulateEmptyResults 				bool
}

func NewMockChatRepository() *MockChatRepository {
	return &MockChatRepository{
		CallCount:      make(map[string]int),
		chats:        	make(map[int]models.DBChats),
		chatMessages: 	make(map[int][]models.DBChatMessage),
		nextChatId:   	1,
	}
}
func NewMockChatRepositoryWithData() *MockChatRepository {
	mock := NewMockChatRepository()
	
	chat1 := models.DBChats{
		Id:        1,
		UserId:    101,
		Title:     "Test Chat 1",
		CreatedAt: time.Now(),
	}
	chat2 := models.DBChats{
		Id:        2,
		UserId:    101,
		Title:     "Test Chat 2",
		CreatedAt: time.Now(),
	}
	
	mock.chats[1] = chat1
	mock.chats[2] = chat2
	mock.nextChatId = 3
	
	mock.chatMessages[1] = []models.DBChatMessage{
		{
			ChatId:    1,
			Message:   "Hello",
			Response:  "Hi there!",
			CreatedAt: time.Now(),
		},
		{
			ChatId:    1,
			Message:   "How are you?",
			Response:  "I'm doing well, thanks!",
			CreatedAt: time.Now(),
		},
	}
	
	return mock
}

func (m *MockChatRepository) InitializeChatsTable() error {
	if m.InitializeChatsTableError != nil {
		return m.InitializeChatsTableError
	}
	return nil
}

func (m *MockChatRepository) InitializeChatMessagesTable() error {
	if m.InitializeChatMessagesTableError != nil {
		return m.InitializeChatMessagesTableError
	}
	return nil
}

func (m *MockChatRepository) InitializeTables() error {
	if m.InitializeTablesError != nil {
		return m.InitializeTablesError
	}
	return nil
}

func (m *MockChatRepository) CreateChat(chat models.DBChats) (int, error) {
	m.CallCount["CreateChat"]++

	if m.CreateChatError != nil {
		return 0, m.CreateChatError
	}
	
	chatId := m.nextChatId
	m.nextChatId++
	
	newChat := models.DBChats{
		Id:        chatId,
		UserId:    chat.UserId,
		Title:     chat.Title,
		CreatedAt: time.Now(),
	}
	
	m.chats[chatId] = newChat
	return chatId, nil
}

func (m *MockChatRepository) AddChatMessage(chatMsg models.DBChatMessage) error {
	m.CallCount["AddChatMessage"]++

	if m.AddChatMessageError != nil {
		return m.AddChatMessageError
	}
	
	if _, exists := m.chats[chatMsg.ChatId]; !exists {
		return errors.New("chat not found")
	}
	
	message := models.DBChatMessage{
		ChatId:    chatMsg.ChatId,
		Message:   chatMsg.Message,
		Response:  chatMsg.Response,
		CreatedAt: time.Now(),
	}
	
	m.chatMessages[chatMsg.ChatId] = append(m.chatMessages[chatMsg.ChatId], message)
	return nil
}

func (m *MockChatRepository) GetChats(chatId int) ([]models.DBChatMessage, string, error) {
	m.CallCount["GetChats"]++

	if m.GetChatsError != nil {
		return nil, "", m.GetChatsError
	}
	
	if m.SimulateNoRows {
		return nil, "", sql.ErrNoRows
	}
	
	if m.SimulateEmptyResults {
		return []models.DBChatMessage{}, "", nil
	}
	
	// Check if chat exists
	chat, exists := m.chats[chatId]
	if !exists {
		return nil, "", errors.New("chat not found")
	}
	
	messages, exists := m.chatMessages[chatId]
	if !exists {
		messages = []models.DBChatMessage{}
	}
	
	return messages, chat.Title, nil
}

func (m *MockChatRepository) GetUserChats(userId int) ([]models.DBChats, error) {
	m.CallCount["GetUserChats"]++

	if m.GetUserChatsError != nil {
		return nil, m.GetUserChatsError
	}
	
	if m.SimulateNoRows {
		return nil, sql.ErrNoRows
	}
	
	if m.SimulateEmptyResults {
		return []models.DBChats{}, nil
	}
	
	var userChats []models.DBChats
	for _, chat := range m.chats {
		if chat.UserId == userId {
			userChats = append(userChats, chat)
		}
	}
	
	return userChats, nil
}

func (m *MockChatRepository) SetCreateChatError(err error) {
	m.CreateChatError = err
}

func (m *MockChatRepository) SetAddChatMessageError(err error) {
	m.AddChatMessageError = err
}

func (m *MockChatRepository) SetGetChatsError(err error) {
	m.GetChatsError = err
}

func (m *MockChatRepository) SetGetUserChatsError(err error) {
	m.GetUserChatsError = err
}

func (m *MockChatRepository) SetSimulateNoRows(simulate bool) {
	m.SimulateNoRows = simulate
}

func (m *MockChatRepository) SetSimulateEmptyResults(simulate bool) {
	m.SimulateEmptyResults = simulate
}

func (m *MockChatRepository) GetCallCount(method string) int {
	return m.CallCount[method]
}

func (m *MockChatRepository) Reset() {
	m.chats = make(map[int]models.DBChats)
	m.chatMessages = make(map[int][]models.DBChatMessage)
	m.nextChatId = 1
	
	m.InitializeChatsTableError = nil
	m.InitializeChatMessagesTableError = nil
	m.InitializeTablesError = nil
	m.CreateChatError = nil
	m.AddChatMessageError = nil
	m.GetChatsError = nil
	m.GetUserChatsError = nil
	
	// Reset flags
	m.SimulateNoRows = false
	m.SimulateEmptyResults = false
}

func (m *MockChatRepository) AddTestChat(chat models.DBChats) {
	m.chats[chat.Id] = chat
	if chat.Id >= m.nextChatId {
		m.nextChatId = chat.Id + 1
	}
}

func (m *MockChatRepository) AddTestMessage(chatId int, message models.DBChatMessage) {
	m.chatMessages[chatId] = append(m.chatMessages[chatId], message)
}