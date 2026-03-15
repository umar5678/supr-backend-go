#!/bin/bash

# Kafka Topics Setup Script
# This script creates all required Kafka topics for the notification system

echo "=========================================="
echo "Creating Kafka Topics for Ride-Sharing App"
echo "=========================================="
echo ""

# Function to create topic (works with both Docker and native installations)
create_topic() {
    local topic=$1
    echo "Creating topic: $topic"
    
    # Try Docker first
    if command -v docker &> /dev/null && docker ps | grep -q kafka; then
        docker-compose exec -T kafka kafka-topics.sh --create --topic "$topic" \
            --bootstrap-server localhost:9092 \
            --partitions 3 \
            --replication-factor 1 \
            --if-not-exists 2>/dev/null
    # Try native installation
    elif [ -f "/opt/kafka/bin/kafka-topics.sh" ]; then
        /opt/kafka/bin/kafka-topics.sh --create --topic "$topic" \
            --bootstrap-server localhost:9092 \
            --partitions 3 \
            --replication-factor 1 \
            --if-not-exists 2>/dev/null
    else
        echo "❌ Kafka not found. Please ensure Kafka is running (Docker or native installation)"
        return 1
    fi
    
    if [ $? -eq 0 ]; then
        echo "✅ Topic '$topic' created successfully"
    else
        echo "⚠️  Topic '$topic' already exists or error occurred"
    fi
}

# Array of all required topics
TOPICS=(
    "ride-events"
    "payment-events"
    "vehicle-events"
    "user-events"
    "fraud-events"
    "sos-events"
    "promotion-events"
    "message-events"
    "profile-events"
    "auth-events"
    "rider-events"
    "pricing-events"
    "admin-events"
    "ridepin-events"
    "tracking-events"
    "food-order-events"
    "food-deals-events"
    "food-product-events"
    "document-events"
)

# Create all topics
for topic in "${TOPICS[@]}"; do
    create_topic "$topic"
done

echo ""
echo "=========================================="
echo "Listing all Kafka topics:"
echo "=========================================="

# List all topics
if command -v docker &> /dev/null && docker ps | grep -q kafka; then
    docker-compose exec kafka kafka-topics.sh --list --bootstrap-server localhost:9092
elif [ -f "/opt/kafka/bin/kafka-topics.sh" ]; then
    /opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092
fi

echo ""
echo "=========================================="
echo "✅ Kafka topics setup completed!"
echo "=========================================="
