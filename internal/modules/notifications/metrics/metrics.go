package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    EventsPublished = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "kafka_events_published_total",
            Help: "Total number of events published to Kafka",
        },
        []string{"event_type", "topic", "status"},
    )
    
    EventsConsumed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "kafka_events_consumed_total",
            Help: "Total number of events consumed from Kafka",
        },
        []string{"event_type", "topic", "status"},
    )
    
    EventProcessingDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "event_processing_duration_seconds",
            Help:    "Duration of event processing",
            Buckets: prometheus.DefBuckets,
        },
        []string{"event_type"},
    )
    
    ConsumerLag = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "kafka_consumer_lag",
            Help: "Current consumer lag",
        },
        []string{"topic", "partition", "consumer_group"},
    )
    
    NotificationsSent = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "notifications_sent_total",
            Help: "Total number of notifications sent",
        },
        []string{"channel", "status"},
    )
    
    DLQMessages = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "dlq_messages_total",
            Help: "Total number of messages sent to DLQ",
        },
        []string{"event_type", "reason"},
    )
)

func RecordEventPublished(eventType, topic, status string) {
    EventsPublished.WithLabelValues(eventType, topic, status).Inc()
}

func RecordEventConsumed(eventType, topic, status string) {
    EventsConsumed.WithLabelValues(eventType, topic, status).Inc()
}

func RecordProcessingDuration(eventType string, duration float64) {
    EventProcessingDuration.WithLabelValues(eventType).Observe(duration)
}

func RecordNotificationSent(channel, status string) {
    NotificationsSent.WithLabelValues(channel, status).Inc()
}

func RecordDLQMessage(eventType, reason string) {
    DLQMessages.WithLabelValues(eventType, reason).Inc()
}