package health

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type HealthChecker struct {
	db      *gorm.DB
	brokers []string
}

func NewHealthChecker(db *gorm.DB, brokers []string) *HealthChecker {
	return &HealthChecker{
		db:      db,
		brokers: brokers,
	}
}

type HealthStatus struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Services  map[string]ServiceHealth `json:"services"`
}

type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (h *HealthChecker) Check(ctx context.Context) *HealthStatus {
	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceHealth),
	}

	// Check database
	dbHealth := h.checkDatabase(ctx)
	status.Services["database"] = dbHealth
	if dbHealth.Status != "healthy" {
		status.Status = "unhealthy"
	}

	// Check Kafka
	kafkaHealth := h.checkKafka(ctx)
	status.Services["kafka"] = kafkaHealth
	if kafkaHealth.Status != "healthy" {
		status.Status = "degraded"
	}

	return status
}

func (h *HealthChecker) checkDatabase(ctx context.Context) ServiceHealth {
	sqlDB, err := h.db.DB()
	if err != nil {
		return ServiceHealth{Status: "unhealthy", Message: err.Error()}
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return ServiceHealth{Status: "unhealthy", Message: err.Error()}
	}

	return ServiceHealth{Status: "healthy"}
}

func (h *HealthChecker) checkKafka(ctx context.Context) ServiceHealth {
	conn, err := kafka.DialContext(ctx, "tcp", h.brokers[0])
	if err != nil {
		return ServiceHealth{Status: "unhealthy", Message: err.Error()}
	}
	defer conn.Close()

	return ServiceHealth{Status: "healthy"}
}
