package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/loukaspe/jedi-team-challenge/internal/repositories"
	"github.com/loukaspe/jedi-team-challenge/pkg/chunks"
	"github.com/loukaspe/jedi-team-challenge/pkg/embeddings"
	"github.com/loukaspe/jedi-team-challenge/pkg/logger"
	http2 "github.com/loukaspe/jedi-team-challenge/pkg/server/http"
	"github.com/loukaspe/jedi-team-challenge/pkg/vectordb"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
	"github.com/pkoukk/tiktoken-go"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
)

func main() {
	ctx := context.Background()
	getEnv()

	//encoder := getEncoder()
	client := getOpenAIClient()
	//chunker := getChunker(encoder)
	embedder := getEmbedder(&client)
	pineconeVectorDB := getPineconeVectorDB()

	//inputPeopleKnowledgeBase(ctx, chunker, embedder, pineconeVectorDB)

	logger := logger.NewLogger(ctx)
	router := mux.NewRouter()
	httpServer := &http.Server{
		Addr:    os.Getenv("SERVER_ADDR"),
		Handler: router,
	}
	db := getDB()

	server := http2.NewServer(db, router, httpServer, logger, &client, embedder, pineconeVectorDB)

	server.Run()
}

func getDB() *gorm.DB {
	dbDsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=disable password=%s TimeZone=Europe/Athens",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PASSWORD"),
	)
	db, err := gorm.Open(postgres.Open(dbDsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}

	// Extension for UUID autogeneration as primary keys of tables
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		log.Fatal("failed to create uuid-ossp extension:", err)
	}

	// Drops added in order to start with clean DB on App start for
	// assessment reasons
	db.Migrator().DropTable("users")
	db.Migrator().DropTable("messages")
	db.Migrator().DropTable("chat_sessions")

	err = db.AutoMigrate(&repositories.User{})
	if err != nil {
		log.Fatal("cannot migrate user table")
	}

	// hardcoded user just for testing purposes
	admin := repositories.User{
		ID:       uuid.UUID{0x12, 0x34, 0x56, 0x78},
		Username: "loukas",
		Password: "loukastest",
	}

	fmt.Printf("Seeding user: %s\n", admin.ID)
	err = db.Debug().Model(&repositories.User{}).Create(&admin).Error
	if err != nil {
		log.Fatalf("cannot seed users table: %v", err)
	}

	err = db.AutoMigrate(&repositories.ChatSession{})
	if err != nil {
		log.Fatal("cannot migrate chat sessions table")
	}

	err = db.AutoMigrate(&repositories.Message{})
	if err != nil {
		log.Fatal("cannot migrate messages table")
	}

	return db
}

func getEnv() {
	err := godotenv.Load("./config/.env")
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
	}
}

func getEncoder() *tiktoken.Tiktoken {
	chunkEncoding := os.Getenv("CHUNK_ENCODING_MODEL")

	tiktokenEncoder, err := tiktoken.GetEncoding(chunkEncoding)
	if err != nil {
		log.Fatal("Cannot create encoder: ", err)
	}

	return tiktokenEncoder
}

func getChunker(encoder *tiktoken.Tiktoken) *chunks.Chunker {
	maxTokensPerChunksAsString := os.Getenv("MAX_TOKENS_PER_CHUNKS")
	maxTokensPerChunks, err := strconv.Atoi(maxTokensPerChunksAsString)
	if err != nil {
		log.Fatal("Cannot read max token per chunks: ", err)
	}

	chunker, err := chunks.NewChunker(encoder, maxTokensPerChunks)
	if err != nil {
		log.Fatal("Cannot create chunker: ", err)
	}

	return chunker
}

func getEmbedder(client *openai.Client) *embeddings.EmbeddingService {
	return embeddings.NewEmbeddingService(client, openai.EmbeddingModel(os.Getenv("EMBEDDING_MODEL")))
}

func getOpenAIClient() openai.Client {
	return openai.NewClient(option.WithAPIKey(os.Getenv("OPENAI_API_KEY")))
}

func getPineconeVectorDB() *vectordb.PineconeVectorDB {
	topKResultsNumberAsString := os.Getenv("TOP_K_RESULTS_NUMBER")
	topKResultsNumber, err := strconv.Atoi(topKResultsNumberAsString)
	if err != nil {
		log.Fatal("Cannot read top k results number: ", err)
	}

	similaritySearchThresholdAsString := os.Getenv("SIMILARITY_SEARCH_THRESHOLD")
	similaritySearchThreshold, err := strconv.ParseFloat(similaritySearchThresholdAsString, 32)
	if err != nil {
		log.Fatal("Cannot read similarity search threshold: ", err)
	}

	pineconeClient, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: os.Getenv("PINECONE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create pinecone Client: %v", err)
	}
	return vectordb.NewPineconeVectorDB(
		float32(similaritySearchThreshold),
		topKResultsNumber,
		os.Getenv("PINECONE_INDEX"),
		pineconeClient,
	)

}

func inputVehiclesKnowledgeBase(ctx context.Context, chunker *chunks.Chunker, embedder *embeddings.EmbeddingService, pineconeVectorDB *vectordb.PineconeVectorDB) {
	textBytes, err := os.ReadFile("./dataVehicles.md")
	if err != nil {
		log.Fatal(err)
	}
	text := string(textBytes)

	chunks := chunker.Chunk(text)
	fmt.Printf("Generated %d chunks\n", len(chunks))

	domainEmbeddings, err := embedder.Embed(ctx, chunks)
	if err != nil {
		log.Fatalf("Embedding error: %v", err)
	}

	count, err := pineconeVectorDB.StoreEmbeddings(
		ctx,
		domainEmbeddings,
		map[string]interface{}{
			"type": "vehicles",
		})
	if err != nil {
		log.Fatalf("Failed to store embeddings: %v", err)
	}

	fmt.Sprintf("Stored %d embeddings in Pinecone index %s\n", count, os.Getenv("PINECONE_INDEX"))
}

func inputPeopleKnowledgeBase(ctx context.Context, chunker *chunks.Chunker, embedder *embeddings.EmbeddingService, pineconeVectorDB *vectordb.PineconeVectorDB) {
	textBytes, err := os.ReadFile("./dataPeople.md")
	if err != nil {
		log.Fatal(err)
	}
	text := string(textBytes)

	chunks := chunker.Chunk(text)
	fmt.Printf("Generated %d chunks\n", len(chunks))

	domainEmbeddings, err := embedder.Embed(ctx, chunks)
	if err != nil {
		log.Fatalf("Embedding error: %v", err)
	}

	count, err := pineconeVectorDB.StoreEmbeddings(
		ctx,
		domainEmbeddings,
		map[string]interface{}{
			"type": "people",
		})
	if err != nil {
		log.Fatalf("Failed to store embeddings: %v", err)
	}

	fmt.Sprintf("Stored %d embeddings in Pinecone index %s\n", count, os.Getenv("PINECONE_INDEX"))
}
