package db

import (
	"context"
	"database/sql"

	"github.com/yoshapihoff/bricks/auth/internal/config"
	postgresRepo "github.com/yoshapihoff/bricks/auth/internal/repository/postgres"
)

// Init initializes the database connection and returns a *sql.DB instance
func Init(cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	// Test the database connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	userRepo := postgresRepo.NewUserRepository(db)
	if err := userRepo.CreateTables(context.Background()); err != nil {
		return nil, err
	}

	return db, nil
}
