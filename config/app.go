package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	AppHost        string
	AppPort        string
	DatabaseURL    string
	GeminiAPIKey   string
	GeminiModel    string
	GeminiTTSModel string
	GeminiBase     string
}

// NewConfig loads .env and returns AppConfig
func NewConfig() AppConfig {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables or defaults")
	}

	return AppConfig{
		AppHost:        getEnv("APP_HOST", "0.0.0.0"),
		AppPort:        getEnv("APP_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:123456@localhost:5432/sentence_miner?sslmode=disable"),
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		GeminiModel:    getEnv("GEMINI_MODEL", "gemini-1.5-flash"),
		GeminiTTSModel: getEnv("GEMINI_TTS_MODEL", "gemini-3.1-flash-tts-preview"),
		GeminiBase:     getEnv("GEMINI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta"),
	}
}

// getEnv returns environment variable or fallback
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
