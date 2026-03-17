package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
)

type EventHandler interface {
	Handle(ctx context.Context, event *ConsumedEvent) error
	EventType() EventType
	CanHandle(eventType EventType) bool
}

type ConsumedEvent struct {
	ID           uuid.UUID
	EventType    EventType
	SourceModule string
	Payload      []byte
	Timestamp    time.Time
	Headers      map[string]string
}

type EventConsumer interface {
	Subscribe(handler EventHandler) error
	Start(ctx context.Context) error
	Stop() error
}

type KafkaConsumer struct {
	reader      *kafka.Reader
	registry    *EventRegistry
	db          *gorm.DB
	config      *ConsumerConfig
	handlers    []EventHandler
	dlqProducer EventProducer
	stopChan    chan struct{}
}

type ConsumerConfig struct {
	Brokers           []string
	GroupID           string
	Topic             string
	MaxConcurrent     int
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
	CommitInterval    time.Duration
	MaxRetries        int
	RetryBackoff      time.Duration
}

func DefaultConsumerConfig(brokers []string, groupID, topic string) *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:           brokers,
		GroupID:           groupID,
		Topic:             topic,
		MaxConcurrent:     10,
		SessionTimeout:    30 * time.Second,
		HeartbeatInterval: 3 * time.Second,
		CommitInterval:    time.Second,
		MaxRetries:        3,
		RetryBackoff:      time.Second,
	}
}

func NewKafkaConsumer(
	config *ConsumerConfig,
	registry *EventRegistry,
	db *gorm.DB,
	dlqProducer EventProducer,
) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           config.Brokers,
		GroupID:           config.GroupID,
		Topic:             config.Topic,
		MinBytes:          10e3, // 10KB
		MaxBytes:          10e6, // 10MB
		MaxWait:           500 * time.Millisecond,
		SessionTimeout:    config.SessionTimeout,
		HeartbeatInterval: config.HeartbeatInterval,
		CommitInterval:    config.CommitInterval,
		StartOffset:       kafka.LastOffset,
	})

	return &KafkaConsumer{
		reader:      reader,
		registry:    registry,
		db:          db,
		config:      config,
		handlers:    make([]EventHandler, 0),
		dlqProducer: dlqProducer,
		stopChan:    make(chan struct{}),
	}
}

func (c *KafkaConsumer) Subscribe(handler EventHandler) error {
	c.handlers = append(c.handlers, handler)
	return nil
}

func (c *KafkaConsumer) Start(ctx context.Context) error {
	semaphore := make(chan struct{}, c.config.MaxConcurrent)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.stopChan:
			return nil
		default:
			semaphore <- struct{}{}

			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				<-semaphore
				if err == context.Canceled {
					return nil
				}
				fmt.Printf("Error fetching message: %v\n", err)
				continue
			}

			go func(message kafka.Message) {
				defer func() { <-semaphore }()

				if err := c.processMessage(ctx, message); err != nil {
					fmt.Printf("Error processing message: %v\n", err)
					return
				}

				if err := c.reader.CommitMessages(ctx, message); err != nil {
					fmt.Printf("Error committing message: %v\n", err)
				}
			}(msg)
		}
	}
}

func (c *KafkaConsumer) processMessage(ctx context.Context, msg kafka.Message) error {
	headers := make(map[string]string)
	for _, h := range msg.Headers {
		headers[h.Key] = string(h.Value)
	}

	eventIDStr, ok := headers["event_id"]
	if !ok {
		return fmt.Errorf("missing event_id header")
	}

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return fmt.Errorf("invalid event_id: %w", err)
	}

	eventTypeStr, ok := headers["event_type"]
	if !ok {
		return fmt.Errorf("missing event_type header")
	}

	eventType := EventType(eventTypeStr)

	if c.isProcessed(ctx, eventID) {
		fmt.Printf("Event %s already processed, skipping\n", eventID)
		return nil
	}

	consumedEvent := &ConsumedEvent{
		ID:           eventID,
		EventType:    eventType,
		SourceModule: headers["source_module"],
		Payload:      msg.Value,
		Timestamp:    time.Now(),
		Headers:      headers,
	}

	var handler EventHandler
	for _, h := range c.handlers {
		if h.CanHandle(eventType) {
			handler = h
			break
		}
	}

	if handler == nil {
		fmt.Printf("No handler for event type %s\n", eventType)
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.config.RetryBackoff * time.Duration(1<<uint(attempt-1))
			time.Sleep(backoff)
		}

		if err := handler.Handle(ctx, consumedEvent); err != nil {
			lastErr = err
			fmt.Printf("Attempt %d failed for event %s: %v\n", attempt+1, eventID, err)
			continue
		}

		if err := c.markProcessed(ctx, eventID); err != nil {
			fmt.Printf("Failed to mark event as processed: %v\n", err)
		}

		return nil
	}

	fmt.Printf("Max retries exceeded for event %s, sending to DLQ\n", eventID)
	if err := c.sendToDLQ(ctx, consumedEvent, lastErr); err != nil {
		fmt.Printf("Failed to send to DLQ: %v\n", err)
	}

	return lastErr
}

func (c *KafkaConsumer) isProcessed(ctx context.Context, eventID uuid.UUID) bool {
	var count int64
	err := c.db.WithContext(ctx).
		Model(&models.ProcessedEvent{}).
		Where("event_id = ? AND consumer_group = ?", eventID, c.config.GroupID).
		Count(&count).Error

	return err == nil && count > 0
}

func (c *KafkaConsumer) markProcessed(ctx context.Context, eventID uuid.UUID) error {
	processedEvent := &models.ProcessedEvent{
		EventID:       eventID,
		ConsumerGroup: c.config.GroupID,
	}

	return c.db.WithContext(ctx).Create(processedEvent).Error
}

func (c *KafkaConsumer) sendToDLQ(ctx context.Context, event *ConsumedEvent, processingErr error) error {
	dlqPayload := map[string]interface{}{
		"original_event":   event,
		"processing_error": processingErr.Error(),
		"consumer_group":   c.config.GroupID,
		"failed_at":        time.Now(),
	}

	return c.dlqProducer.PublishEvent(ctx, EventType(fmt.Sprintf("%s.dlq", event.EventType)), dlqPayload)
}

func (c *KafkaConsumer) Stop() error {
	close(c.stopChan)
	return c.reader.Close()
}
