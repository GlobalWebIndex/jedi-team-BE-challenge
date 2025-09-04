package handler

import (
	"challenge/api/model"
	"challenge/database/database"
	"challenge/database/schema"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetSessionData(sessionID string) ([]model.UserHistory, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID is required")
	}

	collection := database.GetCollection("test", "userhistory")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"sessionId": sessionID}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query MongoDB: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.UserHistory
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("failed to decode results: %w", err)
	}

	return results, nil
}

// RemoveSessionData deletes all entries for a given session ID
func RemoveSessionData(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID is required")
	}

	collection := database.GetCollection("test", "userhistory")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.DeleteMany(ctx, bson.M{"sessionId": sessionID})
	if err != nil {
		return fmt.Errorf("failed to delete session data: %w", err)
	}

	return nil
}

func AddMessage(sessionID, role, text string) error {
	collection := database.GetCollection("test", "userhistory")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := schema.Message{
		Role:      role,
		Text:      text,
		Timestamp: time.Now(),
	}
	filter := bson.M{"sessionId": sessionID}
	update := bson.M{
		"$push":        bson.M{"messages": msg},
		"$setOnInsert": bson.M{"createdAt": time.Now()},
	}

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}
