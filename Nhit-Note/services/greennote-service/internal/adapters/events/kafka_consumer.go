package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"strings"

	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/core/ports"
	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ApprovalEvent represents the structure of events received from the Approval Service.
type ApprovalEvent struct {
	SourceType       string `json:"source_type"`
	SourceID         string `json:"source_id"`
	Status           string `json:"status"`
	ActorID          string `json:"actor_id"`
	ActorName        string `json:"actor_name"`
	ActorDesignation string `json:"actor_designation"`
	ActorRole        string `json:"actor_role"`
	Level            int32  `json:"level"`
	Remarks          string `json:"remarks"`
	Comments         string `json:"comments"`
}

type ApprovalHistoryEntry struct {
	ReviewerName string `json:"reviewerName"`
	Level        int32  `json:"level"`
	Status       string `json:"status"`
	Comments     string `json:"comments"`
	UpdatedAt    string `json:"updatedAt"`
}

type ApprovalHistory struct {
	History []ApprovalHistoryEntry `json:"history"`
}

// KafkaConsumer listens to approval events and updates GreenNote status.
type KafkaConsumer struct {
	reader *kafka.Reader
	repo   ports.GreenNoteRepository
	events ports.EventPublisher
	users  userpb.UserManagementClient
}

// NewKafkaConsumer creates a new consumer for the given brokers and topic.
func NewKafkaConsumer(brokers []string, topic string, groupID string, repo ports.GreenNoteRepository, events ports.EventPublisher, users userpb.UserManagementClient) *KafkaConsumer {
	if len(brokers) == 0 || topic == "" {
		return nil
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,    // Receive messages immediately (was 10KB)
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaConsumer{
		reader: r,
		repo:   repo,
		events: events,
		users:  users,
	}
}

// Start begins consuming messages in a blocking loop. 
// It should be run in a goroutine.
func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Printf("🚀 Starting Kafka Consumer (Topic: %s, Group: %s)", c.reader.Config().Topic, c.reader.Config().GroupID)
	defer func() {
		if err := c.reader.Close(); err != nil {
			log.Printf("Failed to close Kafka reader: %v", err)
		}
	}()

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
			log.Printf("⚠️ Error reading Kafka message: %v", err)
			// Backoff slightly to avoid busy loop on transient errors
			time.Sleep(time.Second)
			continue
		}

		log.Printf("📨 Kafka consumer polling for messages...")
		c.handleMessage(ctx, m)
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, m kafka.Message) {
	log.Printf("\n╔═══════════════════════════════════════════════════════════════")
	log.Printf("║ 📨 GREENNOTE KAFKA CONSUMER - MESSAGE RECEIVED")
	log.Printf("║ Topic: %s | Partition: %d | Offset: %d", m.Topic, m.Partition, m.Offset)
	log.Printf("║ Key: %s", string(m.Key))
	log.Printf("║ Timestamp: %s", m.Time)
	log.Printf("╠═══════════════════════════════════════════════════════════════")
	log.Printf("║ RAW MESSAGE BODY:")
	log.Printf("║ %s", string(m.Value))
	log.Printf("╚═══════════════════════════════════════════════════════════════\n")

	// 1. Parse the Event Wrapper
	// We support both "event_type"/"data" (standard) and "type"/"payload" (from approval-service Event struct)
	var eventWrapper struct {
		EventType   string          `json:"event_type"`
		Type        string          `json:"type"` // Alias used by approval-service
		EventID     string          `json:"event_id"`
		ID          string          `json:"id"`   // Alias used by approval-service
		AggregateID string          `json:"aggregate_id"`
		Data        json.RawMessage `json:"data"`
		Payload     json.RawMessage `json:"payload"`
	}

	var event ApprovalEvent
	var data json.RawMessage
	if err := json.Unmarshal(m.Value, &eventWrapper); err != nil {
		log.Printf("⚠️  Failed to unmarshal event wrapper (might be unwrapped): %v", err)
		data = m.Value
	} else {
		log.Printf("🔍 [PARSE] EventWrapper parsed: Type='%s', AggregateID='%s'", 
			eventWrapper.EventType, eventWrapper.AggregateID)
		
		// Use Data or Payload
		data = eventWrapper.Data
		if len(data) == 0 || string(data) == "null" {
			data = eventWrapper.Payload
		}
	}

	// Try to unmarshal the event to get Level etc early
	if len(data) > 0 && string(data) != "null" {
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("⚠️ [KAFKA] Minor: Failed to unmarshal early event info: %v", err)
		} else {
			log.Printf("📥 [KAFKA] Early Event Info: Level=%d, Actor=%s", event.Level, event.ActorID)
		}
	}

	// Use EventType or Type
	rawType := eventWrapper.EventType
	if rawType == "" {
		rawType = eventWrapper.Type
	}
	eventType := strings.ToLower(strings.TrimSpace(rawType))
	
	log.Printf("📩 [ROUTE] Event Type determined: '%s' (from raw: '%s')", eventType, rawType)
	log.Printf("📩 [ROUTE] Aggregate ID: '%s'", eventWrapper.AggregateID)

	// --- Heuristic/Fallback for Unwrapped Events ---
	// If eventType is empty, try to see if 'data' itself contains an ApprovalEvent
	if eventType == "" {
		log.Printf("🔍 [FALLBACK] Event type is empty, attempting auto-detection...")
		var maybePayload ApprovalEvent
		if err := json.Unmarshal(data, &maybePayload); err == nil && maybePayload.SourceType != "" {
			log.Printf("💡 [FALLBACK] Auto-detected unwrapped ApprovalEvent: SourceType=%s, Status=%s", 
				maybePayload.SourceType, maybePayload.Status)
			eventType = "approvalstatuschanged"
		} else {
			log.Printf("⚠️  [FALLBACK] Could not auto-detect event type")
		}
	}

	var greenNoteId string
	var status string
	var comments string
	var actorID string
	var actorName string
	var actorDesignation string
	var actorRole string
	

	log.Printf("📥 [KAFKA-GREATER] Raw Message: %s", string(m.Value))

	// Handle different event types (case-insensitive)
	switch eventType {
	case "approvalstatuschanged", "approval_status_changed":
		log.Printf("\n🎯 [HANDLER] ==> APPROVALSTATUSCHANGED HANDLER INVOKED")
		log.Printf("🎯 [HANDLER] This is the GreenNote-specific approval event")
		
		var payload ApprovalEvent
		// Attempt direct unmarshal
		if err := json.Unmarshal(data, &payload); err != nil {
			// Fallback: Check for double-encoded JSON string
			var jsonString string
			if err2 := json.Unmarshal(data, &jsonString); err2 == nil {
				if err3 := json.Unmarshal([]byte(jsonString), &payload); err3 == nil {
					log.Printf("⚠️  [HANDLER] Detected double-encoded JSON in GreenNote consumer")
				} else {
					log.Printf("❌ [HANDLER] Failed to unmarshal ApprovalStatusChanged inner JSON: %v", err3)
					return
				}
			} else {
				log.Printf("❌ [HANDLER] Failed to unmarshal ApprovalStatusChanged data: %v", err)
				log.Printf("❌ [HANDLER] Raw Data (%d bytes): %s", len(data), string(data))
				return
			}
		}
		
		log.Printf("✅ [HANDLER] Successfully parsed ApprovalStatusChanged payload:")
		log.Printf("   ├─ SourceID: %s", payload.SourceID)
		log.Printf("   ├─ SourceType: %s", payload.SourceType)
		log.Printf("   ├─ Status: %s", payload.Status)
		log.Printf("   ├─ ActorName: %s", payload.ActorName)
		log.Printf("   └─ Remarks: %s", payload.Remarks)

		// Source Type Validation
		log.Printf("🔍 [FILTER] Checking source type: '%s' (expecting 'GREENNOTE')", payload.SourceType)
		if !strings.EqualFold(payload.SourceType, "GREENNOTE") && payload.SourceType != "" {
			log.Printf("⏭️  [FILTER] Skipping non-GREENNOTE event: %s", payload.SourceType)
			log.Printf("⏭️  [FILTER] This event is not for greennote-service (probably for paymentnote-service)")
			return
		}
		log.Printf("✅ [FILTER] Source type is GREENNOTE - proceeding with handler")
		
		// Extract GreenNote ID
		greenNoteId = payload.SourceID
		log.Printf("🔍 [ID] SourceID from payload: '%s'", greenNoteId)
		// Fallback to AggregateID if SourceID is missing
		if greenNoteId == "" {
			greenNoteId = eventWrapper.AggregateID
			log.Printf("💡 [ID] Fell back to AggregateID: %s", greenNoteId)
		}
		
		// Extract approval details
		status = payload.Status
		comments = payload.Remarks
		if comments == "" {
			comments = payload.Comments
		}
		actorID = payload.ActorID
		actorName = payload.ActorName
		actorDesignation = payload.ActorDesignation
		actorRole = payload.ActorRole
		
		log.Printf("📝 [EXTRACT] Extracted approval details: Status=%s, Actor=%s, Comments=%s", 
			status, actorName, comments)

	case "approval_vote_submitted":
		log.Printf("🔍 Handling approval_vote_submitted for %s", eventWrapper.AggregateID)
		greenNoteId = eventWrapper.AggregateID
		var voteData struct {
			Status   string `json:"status"`
			Reviewer string `json:"reviewer_id"`
			Remarks  string `json:"remarks"`
			ReviewerName string `json:"reviewer_name"`
			ReviewerDesignation string `json:"reviewer_designation"`
			ReviewerRole string `json:"reviewer_role"`
		}
		if err := json.Unmarshal(data, &voteData); err == nil && voteData.Status != "" {
			status = "pending"
			comments = voteData.Remarks
			actorID = voteData.Reviewer
			actorName = voteData.ReviewerName
			actorDesignation = voteData.ReviewerDesignation
			actorRole = voteData.ReviewerRole
			log.Printf("✅ Parsed vote data: Status=%s, Actor=%s", status, actorID)
		} else {
			status = "pending"
			log.Printf("⚠️  Could not parse vote data, defaulting to pending")
		}

	case "approval_completed":
		log.Printf("🔍 Handling approval_completed for %s", eventWrapper.AggregateID)
		greenNoteId = eventWrapper.AggregateID
		status = "APPROVED"

	case "":
		log.Printf("ℹ️  Ignoring empty event type")
		return

	default:
		log.Printf("⚠️ Unhandled Event Type: '%s'", eventWrapper.EventType)
		return
	}

	if greenNoteId == "" {
		// Final fallback: use the Kafka message Key
		if len(m.Key) > 0 {
			greenNoteId = string(m.Key)
			log.Printf("💡 Fell back to Kafka Message Key for ID: %s", greenNoteId)
		}
	}

	if greenNoteId == "" {
		log.Printf("⚠️ GreenNote ID missing in event %s and message key", eventType)
		return
	}

	// 5. Update the Database
	log.Printf("\n🔄 [DATABASE] Fetching GreenNote %s from database...", greenNoteId)
	existing, orgID, tenantID, err := c.repo.Get(ctx, greenNoteId)
	if err != nil {
		log.Printf("❌ [DATABASE] Failed to find GreenNote %s: %v", greenNoteId, err)
		return
	}
	log.Printf("✅ [DATABASE] GreenNote %s found (current status: %s)", greenNoteId, existing.DetailedStatus)

	// Double check Event Wrapper vs Payload for Actor details
	// (Real approval events might have them at top level or deep in Data)
	// We'll trust what we parsed in the switch.

	// Normalize Status:
	// - "approval_completed" -> APPROVED
	// - "REJECTED" -> REJECTED
	// - All others (including intermediate votes) -> PENDING
	dbStatus := strings.ToLower(strings.TrimSpace(existing.DetailedStatus))
	incomingStatus := strings.ToUpper(strings.TrimSpace(status))
	
	log.Printf("🔍 [STATUS] Current DB status: '%s', Incoming status: '%s', Event type: '%s'", 
		dbStatus, incomingStatus, eventType)
	
	if eventType == "approval_completed" || incomingStatus == "APPROVED" || incomingStatus == "a" {
		dbStatus = "approved"
	} else if incomingStatus == "REJECTED" || incomingStatus == "r" {
		dbStatus = "rejected"
	} else if eventType == "approval_vote_submitted" {
		// Ensure we stay pending during voting
		if dbStatus == "draft" {
			dbStatus = "pending"
		}
		// If already pending, stay pending
	} else {
		// Default fallback - handle PENDING_LEVEL_1, PENDING_LEVEL_2 etc
		if incomingStatus != "" {
			// Map PENDING_LEVEL_X to "pending"
			if strings.HasPrefix(incomingStatus, "PENDING") {
				dbStatus = "pending"
			} else {
				dbStatus = strings.ToLower(incomingStatus)
			}
		}
	}
	
	log.Printf("📊 [STATUS] Computed new DB status: '%s'", dbStatus)

	// Fetch Actor Details if missing
	if actorName == "" && actorID != "" && c.users != nil {
		log.Printf("🔍 [USER-SERVICE] Missing actor details, fetching for ID: %s", actorID)
		userCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		userResp, err := c.users.GetUser(userCtx, &userpb.GetUserRequest{UserId: actorID})
		if err != nil {
			log.Printf("⚠️ [USER-SERVICE] Failed to fetch actor details: %v", err)
		} else if userResp != nil {
			actorName = userResp.Name
			actorDesignation = userResp.DesignationName
			actorRole = strings.Join(userResp.Roles, ", ")
			log.Printf("✅ [USER-SERVICE] Fetched: %s (%s, %s)", actorName, actorDesignation, actorRole)
		}
	}

	// Capture Approver Details if available
	// Usually only on final approval or votes
	// In greennote-service, we'll store them in the record if they are provided.
	// We'll prioritize the latest actor info.

	// Construct Log Entry
	displayStatus := "APPROVED" // For the log entry itself, show what the actor did
	if incomingStatus == "R" || incomingStatus == "REJECTED" {
		displayStatus = "REJECTED"
	}
	// Add timestamp to the log explicitly as requested
	// Format: [2026-01-06 14:30:00] APPROVED by ActorID


	if actorName == "" && actorID != "" {
		actorName = "User " + actorID
	}

	// Update GreenNote fields for persistence
	existing.DetailedStatus = dbStatus
	existing.Status = mapDBStatusToProtoEnum(dbStatus)

	if actorName != "" {
		existing.ApproverName = actorName
	}
	if actorDesignation != "" {
		existing.ApproverDesignation = actorDesignation
	}
	if actorRole != "" {
		existing.ApproverRole = actorRole
	}
	if dbStatus == "approved" {
		existing.ApprovedAt = timestamppb.New(time.Now())
	}

	actorDisplay := actorID
	if actorName != "" {
		actorDisplay = actorName
	}

	// Determine Level
	level := event.Level
	if level == 0 {
		// Fallback derivation logic: Map current record status to the step that just occurred.
		// If status is PENDING_LEVEL_2, it means Step 1 was just processed.
		if strings.Contains(incomingStatus, "LEVEL_2") {
			level = 1
		} else if strings.Contains(incomingStatus, "LEVEL_3") {
			level = 2
		} else if strings.Contains(incomingStatus, "LEVEL_4") {
			level = 3
		} else if strings.EqualFold(dbStatus, "approved") {
			level = 4
		} else if strings.Contains(incomingStatus, "LEVEL_1") || strings.EqualFold(dbStatus, "submitted") {
			level = 1 
		}
	}

	// Message refinement: Add "fully approved" for the last step if requested
	if level == 4 && strings.EqualFold(dbStatus, "approved") {
		if comments == "" {
			comments = "Fully approved."
		} else if !strings.Contains(strings.ToLower(comments), "fully approved") {
			comments = strings.TrimSuffix(comments, ".") + ". Fully approved."
		}
	}

	// Prepare Structured JSON History
	var history ApprovalHistory
	history.History = make([]ApprovalHistoryEntry, 0)

	if existing.ApprovalLevels != "" && existing.ApprovalLevels != "[]" {
		if err := json.Unmarshal([]byte(existing.ApprovalLevels), &history); err != nil {
			log.Printf("⚠️ [JSON] Failed to parse existing history: %v", err)
			// Fallback: try to see if it's a raw array
			var rawArray []ApprovalHistoryEntry
			if err2 := json.Unmarshal([]byte(existing.ApprovalLevels), &rawArray); err2 == nil {
				history.History = rawArray
			}
		}
	}
	if history.History == nil {
		history.History = make([]ApprovalHistoryEntry, 0)
	}

	displayLogAction := "A"
	if displayStatus == "REJECTED" {
		displayLogAction = "R"
	}

	// Logic for de-duplication: One entry per actor per level
	found := false
	for i, entry := range history.History {
		if entry.Level == level && entry.ReviewerName == actorName {
			history.History[i].Status = displayLogAction
			// Preserve comments if the new event is empty
			if comments != "" {
				history.History[i].Comments = comments
			}
			history.History[i].UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
			found = true
			break
		}
	}

	if !found && actorName != "" && actorName != "Unknown" {
		history.History = append(history.History, ApprovalHistoryEntry{
			ReviewerName: actorName,
			Level:        level,
			Status:       displayLogAction,
			Comments:     comments,
			UpdatedAt:    time.Now().Format("2006-01-02 15:04:05"),
		})
	}

	// Marshal back to JSON
	historyBytes, _ := json.Marshal(history)
	existing.ApprovalLevels = string(historyBytes)

	// Update Status
	existing.DetailedStatus = dbStatus
	// Sync the proto Status enum
	existing.Status = mapDBStatusToProtoEnum(dbStatus)

	log.Printf("💾 [DATABASE] Updating GreenNote %s to status '%s'...", greenNoteId, dbStatus)
	if err := c.repo.Update(ctx, greenNoteId, existing, orgID, tenantID); err != nil {
		log.Printf("❌ [DATABASE] Failed to update GreenNote %s: %v", greenNoteId, err)
		return
	}

	log.Printf("\n╔═══════════════════════════════════════════════════════════════")
	log.Printf("║ ✅ SUCCESS - GREENNOTE UPDATED")
	log.Printf("║ GreenNote ID: %s", greenNoteId)
	log.Printf("║ New Status: %s", dbStatus)
	log.Printf("║ Actor: %s", actorDisplay)
	log.Printf("║ Approval Log Status: %d bytes", len(existing.ApprovalLevels))
	log.Printf("╚═══════════════════════════════════════════════════════════════\n")

	// 6. Trigger Downstream Logic (If Approved)
	if dbStatus == "approved" {
		c.publishApprovedEvent(ctx, greenNoteId, existing, ApprovalEvent{
			Status:   status,
			Comments: comments,
			ActorID:  actorID,
		})
	}
}

func mapDBStatusToProtoEnum(dbStatus string) greennotepb.Status {
	s := strings.ToLower(strings.TrimSpace(dbStatus))
	if strings.Contains(s, "approved") {
		return greennotepb.Status_STATUS_APPROVED
	}
	if strings.Contains(s, "rejected") {
		return greennotepb.Status_STATUS_REJECTED
	}
	if strings.Contains(s, "pending") {
		return greennotepb.Status_STATUS_PENDING
	}
	if strings.Contains(s, "cancelled") {
		return greennotepb.Status_STATUS_CANCELLED
	}
	if s == "draft" {
		return greennotepb.Status_STATUS_DRAFT
	}
	return greennotepb.Status_STATUS_PENDING
}

func (c *KafkaConsumer) publishApprovedEvent(ctx context.Context, id string, note *greennotepb.GreenNotePayload, approvalEvent ApprovalEvent) {
	evt := ports.GreenNoteApprovedEvent{
		GreenNoteID: id,
		OrderNo:     fmt.Sprintf("GN-%s", id),
		PoNumber:    note.GetPoNumber(),
		NetAmount:   note.TotalAmount,
		Status:      "approved",
		Comments:    approvalEvent.Comments,
		ApprovedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	
	if err := c.events.PublishGreenNoteApproved(ctx, evt); err != nil {
		log.Printf("❌ Failed to publish GreenNoteApprovedEvent: %v", err)
	} else {
		log.Printf("📢 Published GreenNoteApprovedEvent for %s", id)
	}
}
