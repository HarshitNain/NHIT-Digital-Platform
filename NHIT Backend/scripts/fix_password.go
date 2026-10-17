package main

import (
	"context"
	"fmt"
	"log"
	
	"golang.org/x/crypto/bcrypt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Database connection string
	databaseURL := "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"
	ctx := context.Background()
	
	// Parse config
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	// Connect to DB
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	email := "google@gmail.com"
	rawPassword := "google@123"

	// Generate bcrypt hash
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	hashedPassword := string(hashedBytes)
	fmt.Printf("Generated hash for '%s': %s\n", rawPassword, hashedPassword)

	// Update user password
	// Note: We update both users table (assuming user-service uses this DB)
	// I'll update users table where email matches.
	
	query := `UPDATE users SET password = $1 WHERE email = $2`
	tag, err := pool.Exec(ctx, query, hashedPassword, email)
	if err != nil {
		log.Fatalf("Failed to update password: %v", err)
	}

	if tag.RowsAffected() == 0 {
		fmt.Printf("❌ No user found with email %s\n", email)
	} else {
		fmt.Printf("✅ Successfully updated password for user %s\n", email)
	}

	// Verify update
	var currentParam string
	err = pool.QueryRow(ctx, "SELECT password FROM users WHERE email = $1", email).Scan(&currentParam)
	if err != nil {
		log.Fatalf("Failed to verify update: %v", err)
	}
	fmt.Printf("Verified stored hash: %s\n", currentParam)
}
