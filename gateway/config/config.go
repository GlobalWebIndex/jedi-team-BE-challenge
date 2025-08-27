package config

import (
	"os"
)

type DBConfig struct {
	Host     	string
	Port     	string
	User     	string
	Password 	string
	DBName   	string
	SSLMode  	string
}

type OllamaConfig struct {
	Model	string
	Url		string
}

type ServerConfig struct {
	Address		string
}

type Config struct {
	DomainOrigin 	string
	DB				DBConfig
	Ollama			OllamaConfig
	Server			ServerConfig
}

func LoadConfig() *Config {
	return &Config{
		DomainOrigin: 	getEnv("DOMAIN_ORIGIN", "http://localhost:3001"),
		DB: DBConfig{
			Host:     	getEnv("DB_HOST", "postgres"),
			Port:     	getEnv("DB_PORT", "5432"),
			User:     	getEnv("POSTGRES_USER", "myuser"),
			Password: 	getEnv("POSTGRES_PASSWORD", "mypassword"),
			DBName:   	getEnv("POSTGRES_DB", "userdb"),
			SSLMode:  	getEnv("DB_SSL_MODE", "disable"),
		},
		Ollama: OllamaConfig{
			Model:		getEnv("OLLAMA_MODEL", "gemma:2b"),
			Url:		getEnv("OLLAMA_URL", "http://ollama:11434/api/chat"),
		},
		Server: ServerConfig{
			Address: 	getEnv("SERVER_ADDRESS", ":8080"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}