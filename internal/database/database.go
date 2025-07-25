package database

import (
	"database/sql"
	"fmt"
	"log"

	"cctv-api/internal/config"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	// if err = runMigrations(db); err != nil {
	// 	return nil, fmt.Errorf("failed to run migrations: %w", err)
	// }

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}

// func runMigrations(db *sql.DB) error {
// 	migrations := &migrate.FileMigrationSource{
// 		Dir: "migrations",
// 	}

// 	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
// 	if err != nil {
// 		return err
// 	}

// 	log.Printf("Applied %d database migrations\n", n)
// 	return nil
// }
