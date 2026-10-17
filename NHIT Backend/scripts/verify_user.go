package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable"
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var userID, email, name, password string
	var emailVerifiedAt *time.Time
	
	// Check for the user
	err = conn.QueryRow(context.Background(), "SELECT user_id, email, name, password, email_verified_at FROM users WHERE email='nhit@gmail.com'").Scan(&userID, &email, &name, &password, &emailVerifiedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Println("User not found.")
		} else {
			fmt.Printf("Query failed: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("User Found: %s (%s)\n", name, email)
	if emailVerifiedAt != nil {
		fmt.Printf("Email Verified At: %v\n", emailVerifiedAt)
	} else {
		fmt.Println("Email NOT verified.")
	}
}
