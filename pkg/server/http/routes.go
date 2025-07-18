package http

import (
	"github.com/gorilla/websocket"
	"github.com/loukaspe/jedi-team-challenge/internal/core/services"
	httpHandlers "github.com/loukaspe/jedi-team-challenge/internal/handlers/http"
	chatSessionHandlers "github.com/loukaspe/jedi-team-challenge/internal/handlers/http/chatSessions"
	"github.com/loukaspe/jedi-team-challenge/internal/repositories"

	"github.com/loukaspe/jedi-team-challenge/pkg/auth"
	"net/http"
	"os"
)

//	@title			Louk Chatwalker
//	@version		1.0
//	@description	GWI's Jedi Team Challenge

//	@host		localhost:8080
//	@BasePath	/

//	@contact.name	Loukas Peteinaris
//	@contact.url	loukas.peteinaris@gmail.com

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Header value should be in the form of `Bearer <JWT access token>`

// @accept		json
// @produce	json
func (s *Server) initializeRoutes() {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for demo
		},
	}

	wsAddHandler := httpHandlers.NewAddHandler(&upgrader, s.logger)
	wsSubtractHandler := httpHandlers.NewSubtractHandler(&upgrader, s.logger)

	// health check
	healthCheckHandler := httpHandlers.NewHealthCheckHandler(s.DB)
	s.router.HandleFunc("/health-check", healthCheckHandler.HealthCheckController).Methods("GET")

	mcpSSEServer := s.mcpServer.InitialiseSSEServer()

	s.router.HandleFunc("/mcp", mcpSSEServer.ServeHTTP)

	// auth
	jwtMechanism := auth.NewAuthMechanism(
		os.Getenv("JWT_SECRET_KEY"),
		os.Getenv("JWT_SIGNING_METHOD"),
	)
	jwtService := services.NewJwtService(jwtMechanism)
	jwtMiddleware := httpHandlers.NewAuthenticationMw(jwtMechanism)
	jwtHandler := httpHandlers.NewJwtClaimsHandler(jwtService, s.logger)

	s.router.HandleFunc("/token", jwtHandler.JwtTokenController).Methods(http.MethodPost)

	protected := s.router.PathPrefix("/").Subrouter()
	protected.Use(jwtMiddleware.AuthenticationMW)

	chatSessionRepository := repositories.NewChatSessionRepository(s.DB)
	chatSessionService := services.NewChatSessionService(s.logger, chatSessionRepository)
	messageRepository := repositories.NewMessageRepository(s.DB)
	messageService := services.NewMessageService(s.logger, messageRepository, chatSessionRepository, s.embedder, s.pineconeVectorDB, s.openAIClient)

	createChatSessionHandler := chatSessionHandlers.NewCreateUserChatSessionHandler(chatSessionService, s.logger)
	getChatSessionHandler := chatSessionHandlers.NewGetChatSessionHandler(chatSessionService, s.logger)
	sendMessageHandler := chatSessionHandlers.NewSendMessageHandler(messageService, s.logger)
	wsSendMessageHandler := chatSessionHandlers.NewWsSendMessageHandler(messageService, s.logger, &upgrader)
	sseSendMessageHandler := chatSessionHandlers.NewSseSendMessageHandler(messageService, s.logger)
	submitFeedbackHandler := chatSessionHandlers.NewSubmitFeedbackHandler(messageService, s.logger)

	protected.HandleFunc("/users/{user_id}/chat-sessions", createChatSessionHandler.CreateUserChatSessionController).Methods("POST")
	protected.HandleFunc("/users/{user_id}/chat-sessions", getChatSessionHandler.GetUserChatSessionsController).Methods("GET")
	protected.HandleFunc("/users/{user_id}/chat-sessions/{session_id}/messages", sendMessageHandler.SendMessageController).Methods("POST")
	protected.HandleFunc("/users/{user_id}/chat-sessions/{session_id}/messages/{message_id}/feedback", submitFeedbackHandler.SubmitFeedbackController).Methods("POST")

	protected.HandleFunc("/chat-sessions/{session_id}", getChatSessionHandler.GetChatSessionController).Methods("GET")

	protected.HandleFunc("/ws/add", wsAddHandler.AddController)
	protected.HandleFunc("/ws/subtract", wsSubtractHandler.SubtractController)
	protected.HandleFunc("/ws/users/{user_id}/chat-sessions/{session_id}/messages", wsSendMessageHandler.WsSendMessageController)
	protected.HandleFunc("/sse/users/{user_id}/chat-sessions/{session_id}/messages", sseSendMessageHandler.SseSendMessageController)
}
