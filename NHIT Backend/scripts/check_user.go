package main

import (
	"context"
	"fmt"
	"log"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	email := "google@gmail.com"
	var userID, name, password string
	var tenantID string

	err = pool.QueryRow(ctx, "SELECT user_id, tenant_id, name, password FROM users WHERE email = $1", email).Scan(&userID, &tenantID, &name, &password)
	if err != nil {
		log.Printf("❌ User with email %s NOT FOUND: %v", email, err)
		
		// List all users to see what was created
		rows, err := pool.Query(ctx, "SELECT email, name FROM users")
		if err != nil {
			log.Fatalf("Failed to list users: %v", err)
		}
		defer rows.Close()
		fmt.Println("\nExisting Users:")
		for rows.Next() {
			var e, n string
			rows.Scan(&e, &n)
			fmt.Printf("- %s (%s)\n", e, n)
		}
		return
	}

	fmt.Printf("✅ User Found:\nID: %s\nTenantID: %s\nName: %s\nPassword Hash: %s\n", userID, tenantID, name, password)
}
