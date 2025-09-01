package config

import (
	"os"
	"strconv"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RagConfig struct {
	Url  string
	TopK int
}
type OllamaConfig struct {
	Model  string
	Url    string
	Stream bool
}

type ServerConfig struct {
	Address string
}

type Config struct {
	DomainOrigin string
	DB           DBConfig
	Rag          RagConfig
	Ollama       OllamaConfig
	Server       ServerConfig
}

func LoadConfig() *Config {
	return &Config{
		DomainOrigin: getEnv("DOMAIN_ORIGIN", "http://localhost:3001"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("POSTGRES_USER", "myuser"),
			Password: getEnv("POSTGRES_PASSWORD", "mypassword"),
			DBName:   getEnv("POSTGRES_DB", "userdb"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Ollama: OllamaConfig{
			Model:  getEnv("OLLAMA_MODEL", "granite3-dense:8b"),
			Url:    getEnv("OLLAMA_URL", "http://ollama:11434/api/chat"),
			Stream: getEnv("OLLAMA_STREAM", false),
		},
		Rag: RagConfig{
			Url:  getEnv("RAG_URL", "http://rag:8000/retrieve"),
			TopK: getEnv("RAG_TOPK", 30),
		},
		Server: ServerConfig{
			Address: getEnv("SERVER_ADDRESS", ":8080"),
		},
	}
}
func getEnv[T any](key string, fallback T) T {
	if value, exists := os.LookupEnv(key); exists {
		switch any(fallback).(type) {
		case string:
			return any(value).(T)
		case int:
			if v, err := strconv.Atoi(value); err == nil {
				return any(v).(T)
			}
		case bool:
			if v, err := strconv.ParseBool(value); err == nil {
				return any(v).(T)
			}
		}
	}
	return fallback
}
