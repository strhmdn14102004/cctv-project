package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"cctv-api/internal/config"

	_ "github.com/lib/pq"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	var db *sql.DB
	var err error
	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Printf("Failed to open database (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(retryDelay)
			continue
		}

		// Test connection with a simple query
		err = db.Ping()
		if err != nil {
			log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, maxRetries, err)
			db.Close()
			time.Sleep(retryDelay)
			continue
		}

		// Verify tables exist
		_, err = db.Exec("SELECT 1 FROM users LIMIT 1")
		if err != nil {
			log.Printf("Users table check failed (attempt %d/%d): %v", i+1, maxRetries, err)
			db.Close()
			time.Sleep(retryDelay)
			continue
		}
		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	log.Println("Successfully connected to database")

	// Optimize connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}