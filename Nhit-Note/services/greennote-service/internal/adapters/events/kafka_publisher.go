package events

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/core/ports"

	"github.com/segmentio/kafka-go"
)

// NewEventPublisher builds an EventPublisher based on runtime configuration.
// If no brokers/topic are configured, a noop publisher is returned.
func NewEventPublisher(brokers []string, topic string) ports.EventPublisher {
	if len(brokers) == 0 || strings.TrimSpace(topic) == "" {
		return noopEventPublisher{}
	}

	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &kafkaEventPublisher{writer: w}
}

// kafkaEventPublisher sends events to a Kafka topic using kafka-go.
type kafkaEventPublisher struct {
	writer *kafka.Writer
}

func (p *kafkaEventPublisher) PublishGreenNoteApproved(ctx context.Context, ev ports.GreenNoteApprovedEvent) error {
	if p == nil || p.writer == nil {
		return nil
	}

	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(ev.OrderNo),
		Value: payload,
		Time:  time.Now().UTC(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("failed to publish GreenNoteApproved event: %v", err)
		return err
	}

	return nil
}

type noopEventPublisher struct{}

func (noopEventPublisher) PublishGreenNoteApproved(ctx context.Context, ev ports.GreenNoteApprovedEvent) error {
	return nil
}
