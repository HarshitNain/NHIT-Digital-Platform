package ports

import "context"

// KafkaPublisher defines the interface for publishing events to Kafka
type KafkaPublisher interface {
	Publish(ctx context.Context, topic string, message interface{}) error
	PublishBatch(ctx context.Context, topic string, messages []interface{}) error
	Close() error
}
