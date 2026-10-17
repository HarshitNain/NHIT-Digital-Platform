package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect establishes a pgx connection pool to the PostgreSQL database with optimized limits
func Connect(dbURL string) (*pgxpool.Pool, error) {
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	// ✅ Optimized default settings for shared services
	config.MaxConns = 5
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to db: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("cannot ping db: %w", err)
	}

	log.Println("✅ Database connection established (Shared Pool: Max=25, Min=5)")
	return pool, nil
}
