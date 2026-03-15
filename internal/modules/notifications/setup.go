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
	consumer            *KafkaConsumer
	pushService         service.PushService
	notificationService service.NotificationService
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

	// Initialize Kafka consumer with DLQ producer
	// Subscribe to notification topic for all system events
	consumerConfig := DefaultConsumerConfig(kafkaConfig.Brokers, "notification-consumer", "notification-events")
	consumer := NewKafkaConsumer(
		consumerConfig,
		registry,
		db,
		producer, // Use same producer for DLQ
	)

	logger.Info("notification system initialized successfully")

	return &NotificationSystem{
		producer:            producer,
		consumer:            consumer,
		pushService:         pushSvc,
		notificationService: notifSvc,
	}, nil
}

// Start begins event consumption
func (ns *NotificationSystem) Start(ctx context.Context) error {
	go func() {
		if err := ns.consumer.Start(ctx); err != nil {
			logger.Error("notification consumer error", "error", err)
		}
	}()
	logger.Info("notification system started")
	return nil
}

// Stop gracefully shuts down the notification system
func (ns *NotificationSystem) Stop() error {
	logger.Info("stopping notification system...")
	if err := ns.consumer.Stop(); err != nil {
		logger.Error("failed to stop consumer", "error", err)
		return err
	}
	logger.Info("notification system stopped")
	return nil
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
