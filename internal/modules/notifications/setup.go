package notifications

import (
	"context"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/modules/notifications/repository"
	"github.com/umar5678/go-backend/internal/modules/notifications/service"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// NotificationSystem manages all notification components
type NotificationSystem struct {
	producer            EventProducer
	consumers           []*KafkaConsumer
	pushService         service.PushService
	notificationService service.NotificationService
	db                  *gorm.DB
}

// NewNotificationSystem initializes all notification components
func NewNotificationSystem(
	ctx context.Context,
	db *gorm.DB,
	kafkaConfig config.KafkaConfig,
) (*NotificationSystem, error) {
	// Initialize repositories
	notifRepo := repository.NewNotificationRepository(db)

	// Initialize push service (local, no external dependencies)
	pushSvc := service.NewLocalPushService(db, notifRepo)

	// Initialize notification service
	notifSvc := service.NewNotificationService(notifRepo)

	// Initialize Kafka producer with default config
	producerConfig := DefaultProducerConfig(kafkaConfig.Brokers)
	registry := NewEventRegistry()
	producer := NewKafkaProducer(producerConfig, registry, db)

	topicsToListen := []string{
		"ride-events",
		"payment-events",
		"fraud-events",
		"sos-events",
		"user-events",
		"profile-events",
		"auth-events",
		"pricing-events",
	}

	var consumers []*KafkaConsumer
	for _, topic := range topicsToListen {
		consumerConfig := DefaultConsumerConfig(kafkaConfig.Brokers, "notification-consumer", topic)
		consumer := NewKafkaConsumer(
			consumerConfig,
			registry,
			db,
			producer, // Use same producer for DLQ
		)
		consumers = append(consumers, consumer)
	}

	ns := &NotificationSystem{
		producer:            producer,
		consumers:           consumers,
		pushService:         pushSvc,
		notificationService: notifSvc,
		db:                  db,
	}
	for _, consumer := range consumers {
		ns.registerEventHandlers(consumer)
	}

	logger.Info("notification system initialized successfully", "consumer_count", len(consumers))

	return ns, nil
}

func (ns *NotificationSystem) Start(ctx context.Context) error {
	for i, consumer := range ns.consumers {
		go func(idx int, c *KafkaConsumer) {
			if err := c.Start(ctx); err != nil {
				logger.Error("notification consumer error", "consumer_index", idx, "error", err)
			}
		}(i, consumer)
	}
	logger.Info("notification system started", "active_consumers", len(ns.consumers))
	return nil
}

// Stop gracefully shuts down the notification system
func (ns *NotificationSystem) Stop() error {
	logger.Info("stopping notification system...", "consumer_count", len(ns.consumers))
	var lastErr error
	for _, consumer := range ns.consumers {
		if err := consumer.Stop(); err != nil {
			logger.Error("failed to stop consumer", "error", err)
			lastErr = err
		}
	}
	logger.Info("notification system stopped")
	return lastErr
}

// GetProducer returns the event producer
func (ns *NotificationSystem) GetProducer() EventProducer {
	return ns.producer
}

// GetPushService returns the push service
func (ns *NotificationSystem) GetPushService() service.PushService {
	return ns.pushService
}

// GetNotificationService returns the notification service
func (ns *NotificationSystem) GetNotificationService() service.NotificationService {
	return ns.notificationService
}

// registerEventHandlers registers all event handlers with a consumer
func (ns *NotificationSystem) registerEventHandlers(consumer *KafkaConsumer) {
	// Handler for ride events
	rideHandler := NewRideEventHandler(ns.pushService, ns.db)
	if err := consumer.Subscribe(rideHandler); err != nil {
		logger.Error("failed to subscribe to ride handler", "error", err)
	}

	// Handler for payment events
	paymentHandler := NewPaymentEventHandler(ns.pushService)
	if err := consumer.Subscribe(paymentHandler); err != nil {
		logger.Error("failed to subscribe to payment handler", "error", err)
	}

	// Handler for SOS events
	sosHandler := NewSOSEventHandler(ns.pushService)
	if err := consumer.Subscribe(sosHandler); err != nil {
		logger.Error("failed to subscribe to SOS handler", "error", err)
	}

	// Handler for fraud events
	fraudHandler := NewFraudEventHandler(ns.pushService)
	if err := consumer.Subscribe(fraudHandler); err != nil {
		logger.Error("failed to subscribe to fraud handler", "error", err)
	}

	// Handler for user events
	userHandler := NewUserEventHandler(ns.pushService)
	if err := consumer.Subscribe(userHandler); err != nil {
		logger.Error("failed to subscribe to user handler", "error", err)
	}
}
