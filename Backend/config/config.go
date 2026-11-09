package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string
	FrontendURL string
	DatabaseURL string
}

func LoadConfig() *Config {
	// Load environment variables from .env if available.
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return &Config{
		Port:        port,
		Env:         env,
		FrontendURL: frontendURL,
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
