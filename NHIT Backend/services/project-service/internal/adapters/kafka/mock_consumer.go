package kafka

import (
	"context"
	"log"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/ports"
)

// MockKafkaConsumer is a mock implementation for development/testing
type MockKafkaConsumer struct {
	logger *log.Logger
}

// NewMockKafkaConsumer creates a new mock Kafka consumer
func NewMockKafkaConsumer(logger *log.Logger) ports.KafkaConsumer {
	if logger == nil {
		logger = log.Default()
	}
	return &MockKafkaConsumer{logger: logger}
}

// Subscribe simulates consuming messages from Kafka
func (m *MockKafkaConsumer) Subscribe(ctx context.Context, topic string, handler func(message interface{}) error) error {
	m.logger.Printf("Mock Kafka: Subscribed to topic '%s'", topic)

	// Simulate receiving an organization created event for testing
	event := &domain.OrganizationCreatedEvent{
		EventID:   "test-event-12345678-1234-1234-1234-123456789abc",
		EventType: "organization.created",
		Timestamp: time.Now(),
		TenantID:  "12345678-1234-1234-1234-123456789abc",
		OrgID:     "87654321-4321-4321-4321-cba987654321",
		OrgName:   "Test Organization",
		Projects:  []string{"Test Project 1", "Test Project 2"},
		CreatedBy: "Test User",
	}

	m.logger.Printf("Mock Kafka: Simulating received event for org: %s", event.OrgID)
	return handler(event)
}

// Close closes the mock consumer
func (m *MockKafkaConsumer) Close() error {
	m.logger.Println("Mock Kafka: Consumer closed")
	return nil
}
