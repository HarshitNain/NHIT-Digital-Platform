package kafka

import (
	"context"
	"log/slog"

	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
)

// MockKafkaPublisher is a mock implementation for development/testing
type MockKafkaPublisher struct {
	logger *slog.Logger
}

// NewMockKafkaPublisher creates a new mock Kafka publisher
func NewMockKafkaPublisher(logger *slog.Logger) ports.KafkaPublisher {
	return &MockKafkaPublisher{logger: logger}
}

// Publish logs the message instead of actually publishing to Kafka
func (m *MockKafkaPublisher) Publish(ctx context.Context, topic string, message interface{}) error {
	m.logger.Info("Mock Kafka: Publishing to topic", "topic", topic, "message", message)
	return nil
}

// PublishBatch logs the messages instead of actually publishing to Kafka
func (m *MockKafkaPublisher) PublishBatch(ctx context.Context, topic string, messages []interface{}) error {
	m.logger.Info("Mock Kafka: Publishing batch to topic", "count", len(messages), "topic", topic)
	for i, msg := range messages {
		m.logger.Debug("Message", "index", i+1, "content", msg)
	}
	return nil
}

// Close closes the mock publisher
func (m *MockKafkaPublisher) Close() error {
	m.logger.Info("Mock Kafka: Publisher closed")
	return nil
}
