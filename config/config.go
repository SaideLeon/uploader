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
	// Prioridade 1: Variável de ambiente DOMAIN (geral)
	if value, ok := os.LookupEnv("DOMAIN"); ok && value != "" {
		return value
	}

	// Prioridade 2: Variáveis de ambiente específicas do ambiente (production/local)
	env := os.Getenv("ENV")
	if env == "production" {
		if value, ok := os.LookupEnv("DOMAIN_PROD"); ok && value != "" {
			return value
		}
	} else { // Assume "local" ou qualquer outro valor
		if value, ok := os.LookupEnv("DOMAIN_LOCAL"); ok && value != "" {
			return value
		}
	}
	
	// Fallback final: localhost
	port := getEnv("PORT", "8002")
	return "http://localhost:" + port
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
