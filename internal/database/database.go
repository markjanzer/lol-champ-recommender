package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"lol-champ-recommender/db"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type DB struct {
	Conn    *pgx.Conn
	Queries *db.Queries
}

// Initialize creates a database connection, initializes the schema, and returns a DB struct
func Initialize(ctx context.Context) (*DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
	}

	// Connect to database
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Initialize schema
	schemaSQL, err := os.ReadFile("db/schema.sql")
	if err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	if _, err = conn.Exec(ctx, string(schemaSQL)); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("failed to execute schema SQL: %w", err)
	}

	return &DB{
		Conn:    conn,
		Queries: db.New(conn),
	}, nil
}

// Close closes the database connection
func (db *DB) Close(ctx context.Context) error {
	return db.Conn.Close(ctx)
}
