package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort       string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	JWTSecret     string
	JWTExpiration time.Duration
}

func LoadConfig() *Config {
	// Try to load .env file but don't fail if it doesn't exist
	_ = godotenv.Load()

	expiration, err := time.ParseDuration(os.Getenv("JWT_EXPIRATION"))
	if err != nil {
		log.Fatal("Invalid JWT_EXPIRATION format. Use format like '24h'")
	}

	return &Config{
		AppPort:       getEnv("APP_PORT", "8080"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", "cctv_db"),
		JWTSecret:     getEnv("JWT_SECRET", "default-secret"),
		JWTExpiration: expiration,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
