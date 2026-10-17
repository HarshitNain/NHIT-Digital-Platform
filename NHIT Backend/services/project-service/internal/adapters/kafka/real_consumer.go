package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/ports"
	"github.com/segmentio/kafka-go"
)

// RealKafkaConsumer is a real Kafka consumer implementation
type RealKafkaConsumer struct {
	reader *kafka.Reader
	logger *log.Logger
}

// NewRealKafkaConsumer creates a new real Kafka consumer
func NewRealKafkaConsumer(brokers []string, topic string, groupID string, logger *log.Logger) (ports.KafkaConsumer, error) {
	if logger == nil {
		logger = log.Default()
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := reader.FetchMessage(ctx); err != nil && err != context.DeadlineExceeded {
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}

	logger.Printf("✅ Real Kafka consumer connected to brokers: %v, topic: %s", brokers, topic)
	return &RealKafkaConsumer{
		reader: reader,
		logger: logger,
	}, nil
}

// Subscribe subscribes to Kafka topic and handles messages
func (r *RealKafkaConsumer) Subscribe(ctx context.Context, topic string, handler func(message interface{}) error) error {
	r.logger.Printf("🎯 Starting Kafka consumer for topic: %s", topic)

	for {
		select {
		case <-ctx.Done():
			r.logger.Println("🛑 Kafka consumer stopped due to context cancellation")
			return ctx.Err()
		default:
			// Fetch message
			msg, err := r.reader.FetchMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return nil
				}
				r.logger.Printf("⚠️ Error fetching message: %v", err)
				continue
			}

			// Parse event
			var event domain.OrganizationCreatedEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				r.logger.Printf("❌ Failed to unmarshal message: %v", err)
				r.reader.CommitMessages(ctx, msg) // Commit even on error to avoid reprocessing
				continue
			}

			// Handle event
			r.logger.Printf("📨 Processing event: %s for org: %s", event.EventType, event.OrgID)
			if err := handler(&event); err != nil {
				r.logger.Printf("❌ Failed to handle event: %v", err)
				// Don't commit on error, will retry
				continue
			}

			// Commit message
			if err := r.reader.CommitMessages(ctx, msg); err != nil {
				r.logger.Printf("⚠️ Failed to commit message: %v", err)
			} else {
				r.logger.Printf("✅ Successfully processed and committed event for org: %s", event.OrgID)
			}
		}
	}
}

// Close closes the Kafka consumer
func (r *RealKafkaConsumer) Close() error {
	if err := r.reader.Close(); err != nil {
		r.logger.Printf("Error closing Kafka reader: %v", err)
		return err
	}
	r.logger.Println("✅ Kafka consumer closed")
	return nil
}
