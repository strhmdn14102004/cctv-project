package database

import (
	"database/sql"
	"fmt"
	"log"

	"cctv-api/config"

	_ "github.com/lib/pq" // PostgreSQL driver
	migrate "github.com/rubenv/sql-migrate"
)

func InitDB(cfg *config.Config) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	// Run migrations
	err = runMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrations := &migrate.FileMigrationSource{
		Dir: "database/migrations",
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return err
	}

	if n > 0 {
		log.Printf("Applied %d database migrations\n", n)
	} else {
		log.Println("No database migrations to apply")
	}

	return nil
}
