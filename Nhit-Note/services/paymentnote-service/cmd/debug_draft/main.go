package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/adapters/repository"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/services"
)

func main() {
	connStr := "host=localhost port=5433 user=postgres password=shristi dbname=nhit_db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	repo := repository.NewPaymentNoteRepository(db, nil)
	// We only need repo for this test, nil clients are fine
	svc := services.NewPaymentNoteService(repo, nil, nil, nil) 

	// Create test payload
	gnID := "test-gn-" + time.Now().Format("20060102150405")
	data := map[string]interface{}{
		"order_no":    "GN-" + gnID,
		"po_number":   "PO-TEST-123",
		"net_amount":  5000.0,
		"approver":    "Admin User",
		"approved_at": time.Now().Format(time.RFC3339),
	}

	fmt.Printf("Creating Draft for GreenNote: %s\n", gnID)
	note, err := svc.CreateDraftFromGreenNote(context.Background(), gnID, data)
	if err != nil {
		log.Fatalf("Error creating draft: %v", err)
	}

	fmt.Printf("✅ Draft Created! ID: %d\n", note.ID)
	
	// manually check keys
	if note.GreenNoteNo == nil {
		fmt.Println("❌ GreenNoteNo is NIL in returned object")
	} else {
		fmt.Printf("✅ GreenNoteNo in Domain: '%s'\n", *note.GreenNoteNo)
	}

	// Verify DB state
	var dbGreenNoteNo sql.NullString
	err = db.QueryRow("SELECT green_note_no FROM payment_notes WHERE id = $1", note.ID).Scan(&dbGreenNoteNo)
	if err != nil {
		log.Fatalf("Error querying DB: %v", err)
	}

	if !dbGreenNoteNo.Valid {
		fmt.Println("❌ DB green_note_no is NULL")
	} else {
		fmt.Printf("✅ DB green_note_no: '%s'\n", dbGreenNoteNo.String)
	}
}
