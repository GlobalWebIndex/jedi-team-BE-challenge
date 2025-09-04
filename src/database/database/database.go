package database

import (
	"challenge/config"
	"context"
	"fmt"
	"time"

	"github.com/allegro/bigcache"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	GlobalCache *bigcache.BigCache
)

// Init connects to MongoDB and initializes global cache
func Init() {
	// MongoDB connection URI
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s",
		config.Appconfig.GetString("database.username"),
		config.Appconfig.GetString("database.password"),
		config.Appconfig.GetString("database.host"),
		config.Appconfig.GetString("database.port"),
	)

	// Set client options
	clientOpts := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	MongoClient, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		panic(fmt.Errorf("failed to connect to MongoDB: %w", err))
	}

	// Ping to verify connection
	if err = MongoClient.Ping(ctx, nil); err != nil {
		panic(fmt.Errorf("failed to ping MongoDB: %w", err))
	}

	fmt.Println("✅ Connected to MongoDB")
	fmt.Println(uri)

	// Initialize global cache
	GlobalCache, err = bigcache.NewBigCache(bigcache.DefaultConfig(30 * time.Minute))
	if err != nil {
		panic(fmt.Errorf("failed to initialize cache: %w", err))
	}
}

func GetCollection(dbName, collectionName string) *mongo.Collection {
	return MongoClient.Database(dbName).Collection(collectionName)
}
