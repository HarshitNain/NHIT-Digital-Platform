package events

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/ShristiRnr/Nhit-Note/api/pb/paymentnotepb"
	"github.com/ShristiRnr/Nhit-Note/services/potracking-service/internal/core/ports"

	"github.com/segmentio/kafka-go"
)

type GreenNoteApprovedEvent struct {
	GreenNoteID string  `json:"green_note_id"`
	OrderNo     string  `json:"order_no"`
	NetAmount   float64 `json:"net_amount"`
}

type KafkaConsumer struct {
	reader             *kafka.Reader
	service            ports.POTrackingService
	paymentNoteClient  paymentnotepb.PaymentNoteServiceClient
}

func NewKafkaConsumer(brokers []string, topic string, groupID string, service ports.POTrackingService, paymentNoteClient paymentnotepb.PaymentNoteServiceClient) *KafkaConsumer {
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
		reader:            r,
		service:           service,
		paymentNoteClient: paymentNoteClient,
	}
}

type ApprovalStatusChangedEvent struct {
	SourceType string      `json:"source_type"`
	SourceID   string      `json:"source_id"`
	Status     string      `json:"status"`
	ActorID    interface{} `json:"actor_id"`
	Remarks    string      `json:"remarks"`
}

func (c *KafkaConsumer) Start(ctx context.Context) {
	log.Printf("🚀 Starting Kafka Consumer for PO Tracking (Topic: %s)...", c.reader.Config().Topic)
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

		c.handleMessage(ctx, m)
	}
}

func (c *KafkaConsumer) StartApprovalTracking(ctx context.Context) {
	log.Printf("🚀 Starting Kafka Consumer for Approval Tracking (Topic: %s)...", c.reader.Config().Topic)
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

		c.handleApprovalMessage(ctx, m)
	}
}

func (c *KafkaConsumer) handleMessage(ctx context.Context, m kafka.Message) {
	var event GreenNoteApprovedEvent
	if err := json.Unmarshal(m.Value, &event); err != nil {
		log.Printf("❌ Failed to unmarshal GreenNoteApprovedEvent: %v", err)
		return
	}

	log.Printf("🔔 Received GreenNote Approval for Order %s (Amount: %.2f)", event.OrderNo, event.NetAmount)

	err := c.service.UpdatePOUtilizationByOrderNo(ctx, event.OrderNo, event.NetAmount)
	if err != nil {
		log.Printf("❌ Failed to update PO utilization for order %s: %v", event.OrderNo, err)
		return
	}

	log.Printf("✅ Success: Updated PO utilization for order %s", event.OrderNo)
}

func (c *KafkaConsumer) handleApprovalMessage(ctx context.Context, m kafka.Message) {
	var event ApprovalStatusChangedEvent
	if err := json.Unmarshal(m.Value, &event); err != nil {
		log.Printf("❌ Failed to unmarshal ApprovalStatusChangedEvent: %v", err)
		return
	}

	// Only process events related to PAYMENTNOTE and status APPROVED
	if event.SourceType != "PAYMENTNOTE" || event.Status != "APPROVED" {
		return
	}

	log.Printf("🔔 Received PAYMENTNOTE Approval for ID %s", event.SourceID)

	if c.paymentNoteClient == nil {
		log.Printf("⚠️ PaymentNoteClient not initialized, skipping automation")
		return
	}

	// Parse SourceID to int64 for the gRPC call
	sourceIDInt, err := strconv.ParseInt(event.SourceID, 10, 64)
	if err != nil {
		log.Printf("❌ Failed to parse PaymentNote SourceID %s: %v", event.SourceID, err)
		return
	}

	// Fetch Payment Note details to get PO Number and amount
	resp, err := c.paymentNoteClient.GetPaymentNote(ctx, &paymentnotepb.GetPaymentNoteRequest{
		Id: sourceIDInt,
	})
	if err != nil {
		log.Printf("❌ Failed to fetch PaymentNote details for ID %s: %v", event.SourceID, err)
		return
	}

	note := resp.GetNote()
	if note == nil || note.LoaPoNo == "" {
		log.Printf("⚠️ PaymentNote %s has no PO number, skipping", event.SourceID)
		return
	}

	log.Printf("📦 Updating PO Tracking for %s with amount %.2f", note.LoaPoNo, note.NetPayableAmount)

	err = c.service.UpdatePOUtilizationByOrderNo(ctx, note.LoaPoNo, note.NetPayableAmount)
	if err != nil {
		log.Printf("❌ Failed to update PO utilization: %v", err)
		return
	}

	log.Printf("✅ Success: Automatically updated PO %s from PaymentNote %s", note.LoaPoNo, event.SourceID)
}
