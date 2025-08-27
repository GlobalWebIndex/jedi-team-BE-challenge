package db

import (
	"fmt"
	"database/sql"

	_ "github.com/lib/pq"
	"gateway/config"
	"gateway/internal/repositories"
)


func StartConn() (*sql.DB, error) {
	cfg := config.LoadConfig()
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
		cfg.DB.SSLMode,
	)
	var err error
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error pinging database: %w", err)
	}
	return db, nil
}

func InitDatabase() (*sql.DB, error) {
	db, err := StartConn()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := repositories.InitializeChatsTable(db); err != nil {
		return nil, err
	}

	if err := repositories.InitializeChatMessagesTable(db); err != nil {
		return nil, err
	}

	return db, nil
}