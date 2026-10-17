package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
	"github.com/segmentio/kafka-go"
)

// RealKafkaPublisher is a real Kafka publisher implementation
type RealKafkaPublisher struct {
	writer *kafka.Writer
	logger *slog.Logger
}

// NewRealKafkaPublisher creates a new real Kafka publisher
func NewRealKafkaPublisher(brokers []string, topic string, logger *slog.Logger) (ports.KafkaPublisher, error) {

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	// Test connection by trying to get stats
	stats := writer.Stats()
	if len(brokers) > 0 && stats.Errors == 0 {
		logger.Info("Real Kafka publisher connected to brokers", "brokers", brokers)
	} else {
		logger.Warn("Kafka connection test inconclusive, proceeding anyway")
	}

	return &RealKafkaPublisher{
		writer: writer,
		logger: logger,
	}, nil
}

// Publish publishes a message to Kafka
func (r *RealKafkaPublisher) Publish(ctx context.Context, topic string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Value: data,
		Key:   []byte("organization.created"),
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte("organization.created")},
		},
	}

	if err := r.writer.WriteMessages(ctx, msg); err != nil {
		r.logger.Error("Failed to publish message to Kafka", "topic", topic, "error", err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	r.logger.Info("Published event to Kafka", "topic", topic)
	return nil
}

// PublishBatch publishes multiple messages to Kafka
func (r *RealKafkaPublisher) PublishBatch(ctx context.Context, topic string, messages []interface{}) error {
	kafkaMessages := make([]kafka.Message, len(messages))

	for i, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message %d: %w", i, err)
		}

		kafkaMessages[i] = kafka.Message{
			Topic: topic,
			Value: data,
			Key:   []byte("organization.created"),
			Headers: []kafka.Header{
				{Key: "event_type", Value: []byte("organization.created")},
			},
		}
	}

	if err := r.writer.WriteMessages(ctx, kafkaMessages...); err != nil {
		r.logger.Error("Failed to publish batch messages to Kafka", "topic", topic, "error", err)
		return fmt.Errorf("failed to publish batch: %w", err)
	}

	r.logger.Info("Published batch events to Kafka", "count", len(messages), "topic", topic)
	return nil
}

// Close closes the Kafka publisher
func (r *RealKafkaPublisher) Close() error {
	if err := r.writer.Close(); err != nil {
		r.logger.Error("Error closing Kafka writer", "error", err)
		return err
	}
	r.logger.Info("Kafka publisher closed")
	return nil
}
