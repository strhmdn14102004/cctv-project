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

	// Email configuration
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	EmailFrom    string
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

		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUsername: getEnv("SMTP_USERNAME", "satsat1410@gmail.com"),
		SMTPPassword: getEnv("SMTP_PASSWORD", "ugzs vdly dptv aekc"),
		EmailFrom:    getEnv("EMAIL_FROM", "satsat1410@gmail.com"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
