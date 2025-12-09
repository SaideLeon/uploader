package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Domain      string
	Port        string
	JWTSecret   string
	DatabaseURL string
}

var AppConfig *Config

// LoadConfig loads configuration from .env file or environment variables
func LoadConfig() {
	// Carrega variáveis do .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Não foi possível carregar .env, usando variáveis do sistema")
	}

	AppConfig = &Config{
		Domain:      getDomain(),
		Port:        getEnv("PORT", "8002"),
		JWTSecret:   getEnv("JWT_SECRET", "a-very-secret-key"), // Default for development
		DatabaseURL: getEnv("DATABASE_URL", "forge.db"),
	}
}

// getDomain obtém o domínio baseado no ambiente
func getDomain() string {
	env := os.Getenv("ENV")
	domain := os.Getenv("DOMAIN_LOCAL")
	if env == "production" {
		domain = os.Getenv("DOMAIN_PROD")
	}
	// Fallback if not set
	if domain == "" {
		port := getEnv("PORT", "8002")
		domain = "http://localhost:" + port
	}
	return domain
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
