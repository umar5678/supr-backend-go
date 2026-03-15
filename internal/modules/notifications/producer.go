package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
)

type EventProducer interface {
	PublishEvent(ctx context.Context, eventType EventType, payload interface{}) error
	PublishEventWithKey(ctx context.Context, eventType EventType, key string, payload interface{}) error
	Close() error
}

type KafkaProducer struct {
	writers  map[string]*kafka.Writer
	registry *EventRegistry
	db       *gorm.DB
	config   *ProducerConfig
}

type ProducerConfig struct {
	Brokers       []string
	MaxRetries    int
	BatchSize     int
	FlushInterval time.Duration
	Timeout       time.Duration
	Compression   kafka.Compression
}

func DefaultProducerConfig(brokers []string) *ProducerConfig {
	return &ProducerConfig{
		Brokers:       brokers,
		MaxRetries:    3,
		BatchSize:     100,
		FlushInterval: time.Second,
		Timeout:       10 * time.Second,
		Compression:   kafka.Snappy,
	}
}

func NewKafkaProducer(config *ProducerConfig, registry *EventRegistry, db *gorm.DB) *KafkaProducer {
	return &KafkaProducer{
		writers:  make(map[string]*kafka.Writer),
		registry: registry,
		db:       db,
		config:   config,
	}
}

func (p *KafkaProducer) getWriter(topic string) *kafka.Writer {
	if writer, exists := p.writers[topic]; exists {
		return writer
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(p.config.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{}, // Consistent hashing for key-based partitioning
		MaxAttempts:  p.config.MaxRetries,
		BatchSize:    p.config.BatchSize,
		BatchTimeout: p.config.FlushInterval,
		WriteTimeout: p.config.Timeout,
		Compression:  p.config.Compression,
		Async:        false, // Synchronous for reliability
	}

	p.writers[topic] = writer
	return writer
}

func (p *KafkaProducer) PublishEvent(ctx context.Context, eventType EventType, payload interface{}) error {
	return p.PublishEventWithKey(ctx, eventType, "", payload)
}

func (p *KafkaProducer) PublishEventWithKey(ctx context.Context, eventType EventType, key string, payload interface{}) error {
	// Get topic from registry
	schema, err := p.registry.Get(eventType)
	if err != nil {
		return fmt.Errorf("failed to get event schema: %w", err)
	}

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create event record
	event := &models.Event{
		ID:           uuid.New(),
		EventType:    string(eventType),
		SourceModule: schema.Module,
		Payload:      payloadBytes,
		Status:       models.EventStatusPending,
		RetryCount:   0,
	}

	// Persist event to database for audit trail
	if err := p.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("failed to persist event: %w", err)
	}

	// Create Kafka message
	message := kafka.Message{
		Key:   []byte(key),
		Value: payloadBytes,
		Headers: []kafka.Header{
			{Key: "event_id", Value: []byte(event.ID.String())},
			{Key: "event_type", Value: []byte(eventType)},
			{Key: "source_module", Value: []byte(schema.Module)},
			{Key: "version", Value: []byte(schema.Version)},
			{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	}

	// Publish to Kafka
	writer := p.getWriter(schema.Topic)
	err = writer.WriteMessages(ctx, message)

	// Update event status
	now := time.Now()
	if err != nil {
		event.Status = models.EventStatusFailed
		event.FailedReason = err.Error()
		event.RetryCount++
	} else {
		event.Status = models.EventStatusPublished
		event.PublishedAt = &now
	}

	if updateErr := p.db.WithContext(ctx).Save(event).Error; updateErr != nil {
		// Log but don't fail - event was published
		fmt.Printf("Failed to update event status: %v\n", updateErr)
	}

	if err != nil {
		return fmt.Errorf("failed to publish event to Kafka: %w", err)
	}

	return nil
}

func (p *KafkaProducer) Close() error {
	var lastErr error
	for _, writer := range p.writers {
		if err := writer.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
