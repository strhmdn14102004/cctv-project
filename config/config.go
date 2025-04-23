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
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using environment variables")
	}

	// Parse JWT expiration
	expiration, err := time.ParseDuration(os.Getenv("JWT_EXPIRATION"))
	if err != nil {
		log.Fatal("Invalid JWT_EXPIRATION format. Use format like '24h'")
	}

	return &Config{
		AppPort:       os.Getenv("APP_PORT"),
		DBHost:        os.Getenv("DB_HOST"),
		DBPort:        os.Getenv("DB_PORT"),
		DBUser:        os.Getenv("DB_USER"),
		DBPassword:    os.Getenv("DB_PASSWORD"),
		DBName:        os.Getenv("DB_NAME"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiration: expiration,
	}
}
