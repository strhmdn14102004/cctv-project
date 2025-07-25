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

	// Tambahkan retry mechanism untuk koneksi database
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

		err = db.Ping()
		if err != nil {
			log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, maxRetries, err)
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

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}