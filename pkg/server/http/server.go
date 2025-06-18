package http

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/loukaspe/jedi-team-challenge/pkg/embeddings"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	"github.com/loukaspe/jedi-team-challenge/pkg/vectordb"
	"github.com/openai/openai-go"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	DB               *gorm.DB
	httpServer       *http.Server
	router           *mux.Router
	logger           logger.LoggerInterface
	openAIClient     *openai.Client
	embedder         *embeddings.EmbeddingService
	pineconeVectorDB *vectordb.PineconeVectorDB
}

func NewServer(
	db *gorm.DB,
	router *mux.Router,
	httpServer *http.Server,
	logger logger.LoggerInterface,
	openAIClient *openai.Client,
	embedder *embeddings.EmbeddingService,
	pineconeVectorDB *vectordb.PineconeVectorDB,
) *Server {
	return &Server{
		DB:               db,
		router:           router,
		httpServer:       httpServer,
		logger:           logger,
		openAIClient:     openAIClient,
		embedder:         embedder,
		pineconeVectorDB: pineconeVectorDB,
	}
}

func (s *Server) Run() {
	s.initializeRoutes()

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	db, err := s.DB.DB()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Close()
	if err != nil {
		log.Fatal(err)
	}
}
