package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

func main() {
	topic := "approval_events"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatalf("failed to dial leader: %v", err)
	}
	defer conn.Close()

	// Connect to DB to find a GreenNote ID
	db, err := sql.Open("postgres", "postgres://postgres:shristi@localhost:5433/nhit_db?sslmode=disable")
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	var greenNoteID string
	err = db.QueryRow("SELECT id FROM green_notes LIMIT 1").Scan(&greenNoteID)
	if err != nil {
		log.Fatalf("failed to find a green note: %v. Please make sure at least one green note exists.", err)
	}

	log.Printf("Found GreenNote ID: %s", greenNoteID)

	// Mimic Approval Service's event
	eventData := map[string]interface{}{
		"source_id":   greenNoteID,
		"source_type": "GREENNOTE",
		"status":      "APPROVED",
		"remarks":     "MOCK APPROVAL FOR TESTING",
		"actor_id":    "123",
	}

	wrapper := map[string]interface{}{
		"event_type":   "ApprovalStatusChanged",
		"aggregate_id": greenNoteID,
		"data":         eventData,
	}

	wrapperBytes, _ := json.Marshal(wrapper)

	_, err = conn.WriteMessages(
		kafka.Message{Value: wrapperBytes},
	)
	if err != nil {
		log.Fatalf("failed to write messages: %v", err)
	}

	log.Println("✅ Successfully published mock ApprovalStatusChanged event to Kafka topic:", topic)
}
