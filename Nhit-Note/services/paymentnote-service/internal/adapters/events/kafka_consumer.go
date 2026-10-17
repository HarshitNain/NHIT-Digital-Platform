package events

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/ports"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/domain"

	"github.com/segmentio/kafka-go"
)

type GreenNoteApprovedEvent struct {
	GreenNoteID string  `json:"green_note_id"`
	OrderNo     string  `json:"order_no"`
	PoNumber    string  `json:"po_number"`
	NetAmount   float64 `json:"net_amount"`
	Status      string  `json:"status"`
	Comments    string  `json:"comments"`
	ApprovedAt  string  `json:"approved_at"`
}

type ApprovalStatusChangedEvent struct {
	SourceType string `json:"source_type"`
	SourceID   string `json:"source_id"`
	Status     string `json:"status"`
	ActorID    interface{}  `json:"actor_id"` 
	ActorName  string      `json:"actor_name"`
	ActorEmail string      `json:"actor_email"`
	ActorRole  string      `json:"actor_role"`
	ActorDesignation string `json:"actor_designation"`
	ActorSignature   string `json:"actor_signature"`
	Remarks    string      `json:"remarks"`
	Comments   string      `json:"comments"`
	Level      int32       `json:"level"`
}

type KafkaConsumer struct {
	reader  *kafka.Reader
	service ports.PaymentNoteService
}

func NewKafkaConsumer(brokers []string, topic string, groupID string, service ports.PaymentNoteService) *KafkaConsumer {
	if len(brokers) == 0 || topic == "" {
		return nil
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaConsumer{
		reader:  r,
		service: service,
	}
}

func (c *KafkaConsumer) Start(ctx context.Context, handler func(context.Context, kafka.Message)) {
	log.Printf("🚀 Starting Kafka Consumer for topic: %s", c.reader.Config().Topic)
	defer func() {
		if err := c.reader.Close(); err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("⚠️ Error reading Kafka message: %v", err)
			time.Sleep(time.Second)
			continue
		}

		handler(ctx, m)
	}
}

// GreenNote Approved Handler
func (c *KafkaConsumer) HandleGreenNoteApproved(ctx context.Context, m kafka.Message) {
	var event GreenNoteApprovedEvent
	if err := json.Unmarshal(m.Value, &event); err != nil {
		log.Printf("❌ Failed to unmarshal GreenNoteApprovedEvent: %v", err)
		return
	}

	log.Printf("🔔 Received GreenNote Approval for %s (Order: %s)", event.GreenNoteID, event.OrderNo)

	data := map[string]interface{}{
		"order_no":    event.OrderNo,
		"po_number":   event.PoNumber,
		"net_amount":  event.NetAmount,
		"comments":    event.Comments,
		"approved_at": event.ApprovedAt,
	}

	_, err := c.service.CreateDraftFromGreenNote(ctx, event.GreenNoteID, data)
	if err != nil {
		log.Printf("❌ Failed to create PaymentNote draft: %v", err)
		return
	}

	log.Printf("✅ Success: Draft created for GreenNote %s", event.GreenNoteID)
}

// Approval Status Changed Handler
func (c *KafkaConsumer) HandleApprovalStatusChanged(ctx context.Context, m kafka.Message) {
	// 1. Try to parse as the wrapped format first (used by greennote-service)
	var wrapper struct {
		EventType string          `json:"event_type"`
		Data      json.RawMessage `json:"data"`
		Payload   json.RawMessage `json:"payload"`
	}

	rawPayload := m.Value
	if err := json.Unmarshal(m.Value, &wrapper); err == nil && wrapper.EventType != "" {
		if len(wrapper.Data) > 0 && string(wrapper.Data) != "null" {
			rawPayload = wrapper.Data
		} else if len(wrapper.Payload) > 0 {
			rawPayload = wrapper.Payload
		}
	}

	var event ApprovalStatusChangedEvent
	
	// First attempt: try to unmarshal directly
	err := json.Unmarshal(rawPayload, &event)
	if err != nil {
		// Second attempt: the payload might be a JSON-encoded string (double-encoded)
		// Try to unmarshal as a string first, then unmarshal that string
		var jsonString string
		if err2 := json.Unmarshal(rawPayload, &jsonString); err2 == nil {
			// Successfully got a string, now try to unmarshal the string as JSON
			if err3 := json.Unmarshal([]byte(jsonString), &event); err3 == nil {
				// Success! The payload was double-encoded
				log.Printf("⚠️  Detected double-encoded JSON in Kafka message")
			} else {
				// Still failed even after unwrapping
				log.Printf("❌ Failed to unmarshal ApprovalStatusChangedEvent: %v (Raw: %s)", err, string(rawPayload))
				return
			}
		} else {
			// Not a string either, original error was correct
			log.Printf("❌ Failed to unmarshal ApprovalStatusChangedEvent: %v (Raw: %s)", err, string(rawPayload))
			return
		}
	}

	// Only process events related to PAYMENTNOTE
	if event.SourceType != "PAYMENTNOTE" {
		// Log but don't error as this topic might contain events for other services
		return
	}
	
	// Handle approval_vote_submitted where SourceID might be in ReviewerID or similar depending on payload mapping
	// But typical ApprovalStatusChangedEvent map handles it.
	// Just ensure we have ID.

	sourceID, _ := strconv.ParseInt(event.SourceID, 10, 64)
	log.Printf("🔔 Received Approval Status Change for PaymentNote %d -> %s", sourceID, event.Status)

	// Use Remarks if available, otherwise Comments
	remarks := event.Remarks
	if remarks == "" {
		remarks = event.Comments
	}

	// Normalize status for database enum (A, R, P, etc.)
	dbStatus := event.Status
	incoming := strings.ToUpper(strings.TrimSpace(event.Status))
	
	if incoming == "APPROVED" || incoming == "A" || wrapper.EventType == "approval_completed" {
		dbStatus = "A"
	} else if incoming == "REJECTED" || incoming == "R" {
		dbStatus = "R"
	} else if wrapper.EventType == "approval_vote_submitted" {
		// Keep as Pending for intermediate votes
		dbStatus = "P" 
	} else if incoming == "PENDING" || incoming == "P" || incoming == "PARTIALLY_APPROVED" {
		dbStatus = "P"
	}

	// For vote submitted, we might just want to log it and NOT change status if it's already P or D
	// But UpdateStatus usually sets the status provided.
	// If it's a vote, we should ensure we don't accidentally set it to A prematurely.
	// The check above maps vote_submitted -> P, so it stays pending.
	
	// Create approval log entry
	actorID := int64(0)
	switch v := event.ActorID.(type) {
	case string:
		actorID, _ = strconv.ParseInt(v, 10, 64)
	case float64:
		actorID = int64(v)
	case int64:
		actorID = v
	}

	logEntry := &domain.PaymentApprovalLog{
		PaymentNoteID: sourceID,
		Status:        dbStatus,
		Comments:      stringPtr(remarks),
		ReviewerID:    actorID,
		ReviewerName:  stringPtr(event.ActorName),
		ReviewerEmail: stringPtr(event.ActorEmail),
		ReviewerRole:        stringPtr(event.ActorRole),
		ReviewerDesignation: stringPtr(event.ActorDesignation),
		ReviewerSignature:   stringPtr(event.ActorSignature),
		ApproverLevel:       &event.Level,
		CreatedAt:     time.Now(),
	}

	_, err = c.service.UpdateStatus(ctx, sourceID, dbStatus, false, logEntry)
	if err != nil {
		log.Printf("❌ Failed to update PaymentNote status: %v", err)
		return
	}

	log.Printf("✅ Success: Updated PaymentNote %d to status %s (Remarks: %s)", sourceID, event.Status, remarks)
}

func stringPtr(s string) *string {
	return &s
}
