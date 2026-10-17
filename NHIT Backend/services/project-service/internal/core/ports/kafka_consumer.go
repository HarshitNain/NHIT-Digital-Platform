package ports

import "context"

// KafkaConsumer defines the interface for consuming events from Kafka
type KafkaConsumer interface {
	Subscribe(ctx context.Context, topic string, handler func(message interface{}) error) error
	Close() error
}
