# Messages Module Development Guide

## Overview

The Messages Module manages in-app messaging, notifications, and communication channels between users and system-to-user communications.

## Key Responsibilities

1. Direct Messaging - User-to-user communication
2. System Notifications - Platform-to-user messages
3. Notification Delivery - Send and track delivery
4. Message History - Store and retrieve conversations
5. Push Notifications - External push notification delivery

## Data Transfer Objects

### SendMessageRequest

```go
type SendMessageRequest struct {
    RecipientID string `json:"recipient_id" binding:"required"`
    MessageType string `json:"message_type" binding:"required"` // text, image, file
    Content     string `json:"content" binding:"required"`
    AttachmentURL string `json:"attachment_url,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}
```

### MessageResponse

```go
type MessageResponse struct {
    ID           string    `json:"id"`
    SenderID     string    `json:"sender_id"`
    RecipientID  string    `json:"recipient_id"`
    MessageType  string    `json:"message_type"`
    Content      string    `json:"content"`
    AttachmentURL string   `json:"attachment_url,omitempty"`
    IsRead       bool      `json:"is_read"`
    ReadAt       *time.Time `json:"read_at,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
}
```

### NotificationRequest

```go
type NotificationRequest struct {
    UserID      string                 `json:"user_id"`
    Type        string                 `json:"type"` // ride_update, order_update, promo, etc
    Title       string                 `json:"title"`
    Body        string                 `json:"body"`
    Data        map[string]interface{} `json:"data,omitempty"`
    Priority    string                 `json:"priority"` // normal, high
    ActionURL   string                 `json:"action_url,omitempty"`
}
```

## Handler Methods

```
SendMessage(c *gin.Context)            // POST /messages/send
GetMessages(c *gin.Context)            // GET /messages/{userID}
GetConversation(c *gin.Context)        // GET /messages/conversation/{otherUserID}
MarkAsRead(c *gin.Context)             // POST /messages/{id}/read
DeleteMessage(c *gin.Context)          // DELETE /messages/{id}
SendNotification(c *gin.Context)       // POST /notifications/send
GetNotifications(c *gin.Context)       // GET /notifications
MarkNotificationRead(c *gin.Context)   // POST /notifications/{id}/read
```

## Service Methods

```
SendMessage(ctx context.Context, senderID, recipientID, content string) (*MessageResponse, error)
GetMessages(ctx context.Context, userID string, limit, offset int) ([]MessageResponse, error)
GetConversation(ctx context.Context, userID, otherUserID string) ([]MessageResponse, error)
MarkAsRead(ctx context.Context, messageID string) error
DeleteMessage(ctx context.Context, messageID string) error
SendNotification(ctx context.Context, req NotificationRequest) error
GetNotifications(ctx context.Context, userID string) ([]NotificationResponse, error)
MarkNotificationRead(ctx context.Context, notificationID string) error
BroadcastNotification(ctx context.Context, userIDs []string, notification NotificationRequest) error
SendPushNotification(ctx context.Context, deviceToken, title, body string) error
```

## Message Types

```
Direct Messages:
- User-to-user communication
- Real-time delivery
- Message history tracking
- Read receipts

System Notifications:
- Ride updates (accepted, arriving, completed)
- Order status changes
- Promotional messages
- Account updates
- Safety alerts

In-App vs Push:
- In-app: Delivered to logged-in users
- Push: Sent to mobile devices
- Both: Critical messages
```

## Notification Categories

```
HIGH PRIORITY:
- Emergency/SOS alerts
- Ride safety warnings
- Account security alerts

NORMAL PRIORITY:
- Ride status updates
- Order notifications
- Promotional offers

LOW PRIORITY:
- General announcements
- Tips and suggestions
```

## Typical Use Cases

### 1. Send Message

Request:
```
POST /messages/send
{
    "recipient_id": "user-456",
    "message_type": "text",
    "content": "Hi, how far are you?"
}
```

Flow:
1. Validate recipient exists
2. Create message record
3. Mark as unread for recipient
4. Send real-time notification if online
5. Store in message history
6. Return message confirmation

### 2. Get Conversation

Request:
```
GET /messages/conversation/user-456?limit=50
```

Response:
```json
{
    "messages": [
        {
            "id": "msg-1",
            "sender_id": "user-123",
            "recipient_id": "user-456",
            "content": "Where are you?",
            "is_read": true,
            "created_at": "2024-02-20T10:30:00Z"
        }
    ],
    "total": 1,
    "unread_count": 0
}
```

### 3. Send System Notification

Request:
```
POST /notifications/send
{
    "user_id": "rider-123",
    "type": "ride_update",
    "title": "Driver Arriving",
    "body": "Your driver is 2 minutes away",
    "data": {
        "ride_id": "ride-456",
        "eta": 120
    },
    "priority": "high"
}
```

Flow:
1. Create notification record
2. Store in database
3. Send in-app notification (WebSocket if online)
4. Send push notification if app installed
5. Queue for retry if delivery fails
6. Track delivery status

### 4. Broadcast Notification

Request:
```
POST /notifications/broadcast
{
    "user_ids": ["user-1", "user-2", "user-3"],
    "type": "promo",
    "title": "Special Offer",
    "body": "Get 50% off on next ride"
}
```

Flow:
1. Validate user list
2. Create notification for each user
3. Send in batch
4. Track delivery for each user
5. Return broadcast summary

## Push Notification Integration

### Supported Platforms

1. Firebase Cloud Messaging (Android/iOS)
2. Apple Push Notification Service (iOS)
3. Custom push service

### Device Token Management

```
Store device tokens when:
- User logs in from app
- Device token changes
- App updates

Remove device tokens when:
- User logs out
- Device token expires
- User uninstalls app
```

## Notification Preferences

Users can configure:

```
- Notification frequency
- Notification categories
- Time-based preferences (quiet hours)
- Push vs in-app
- Email preferences
```

## Database Schema

### Messages Table

```sql
CREATE TABLE messages (
    id VARCHAR(36) PRIMARY KEY,
    sender_id VARCHAR(36) NOT NULL,
    recipient_id VARCHAR(36) NOT NULL,
    message_type VARCHAR(50),
    content TEXT,
    attachment_url VARCHAR(500),
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,
    created_at TIMESTAMP,
    FOREIGN KEY (sender_id) REFERENCES users(id),
    FOREIGN KEY (recipient_id) REFERENCES users(id),
    INDEX (recipient_id, created_at),
    INDEX (sender_id, recipient_id)
);
```

### Notifications Table

```sql
CREATE TABLE notifications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(50),
    title VARCHAR(255),
    body TEXT,
    data JSON,
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,
    created_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX (user_id, created_at),
    INDEX (user_id, is_read)
);
```

### Device Tokens Table

```sql
CREATE TABLE device_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token VARCHAR(500),
    platform VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---
