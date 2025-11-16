```go
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/auth/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/helpers"
	"github.com/umar5678/go-backend/internal/utils/jwt"
	"github.com/umar5678/go-backend/internal/utils/password"
	"github.com/umar5678/go-backend/internal/utils/response"

	"gorm.io/gorm"
)

type Service interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error)
	Logout(ctx context.Context, userID string) error
}

type service struct {
	repo Repository
	cfg  *config.Config
}

func NewService(repo Repository, cfg *config.Config) Service {
	return &service{repo: repo, cfg: cfg}
}

func (s *service) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if email exists
	_, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, response.ConflictError("Email already registered")
	}

	// Hash password
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, response.InternalServerError("Failed to process password", err)
	}

	// Create user
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
		Status:   "active",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, response.InternalServerError("Failed to create user", err)
	}

	// Generate tokens
	accessToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 15*time.Minute)
	refreshToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 7*24*time.Hour)

	// Store session in Redis
	cache.SetSession(ctx, user.ID, accessToken, 24*time.Hour)

	return &dto.AuthResponse{
		User:         dto.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes
	}, nil
}

func (s *service) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Find user
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.UnauthorizedError("Invalid credentials")
		}
		return nil, response.InternalServerError("Failed to find user", err)
	}

	// Verify password
	if !password.Verify(req.Password, user.Password) {
		return nil, response.UnauthorizedError("Invalid credentials")
	}

	// Update last login
	s.repo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	accessToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 15*time.Minute)
	refreshToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 7*24*time.Hour)

	// Store session in Redis
	cache.SetSession(ctx, user.ID, accessToken, 24*time.Hour)

	// Broadcast user online status (optional - user will be online when they connect via WebSocket)
	// helpers.SendPresenceUpdate(user.ID, true)

	return &dto.AuthResponse{
		User:         dto.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	claims, err := jwt.Verify(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return nil, response.UnauthorizedError("Invalid refresh token")
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, response.UnauthorizedError("User not found")
	}

	// Generate new tokens
	accessToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 15*time.Minute)
	newRefreshToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 7*24*time.Hour)

	// Update session in Redis
	cache.SetSession(ctx, user.ID, accessToken, 24*time.Hour)

	return &dto.AuthResponse{
		User:         dto.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    900,
	}, nil
}

func (s *service) GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("User")
	}
	return dto.ToUserResponse(user), nil
}

func (s *service) Logout(ctx context.Context, userID string) error {
	// Delete session from Redis
	if err := cache.DeleteSession(ctx, userID); err != nil {
		return response.InternalServerError("Failed to logout", err)
	}

	// Broadcast user offline status (optional - user will be offline when WebSocket disconnects)
	helpers.SendPresenceUpdate(userID, false)

	return nil
}
```


# WebSocket & Redis Usage Guide

## 🚀 Quick Start

### Sending Notifications

```go
import "github.com/umar5678/go-backend/internal/utils/helpers"

// In your notification service
func (s *NotificationService) CreateNotification(ctx context.Context, userID, title, message string) error {
    // 1. Save to database
    notification := &models.Notification{
        UserID:  userID,
        Title:   title,
        Message: message,
    }
    s.repo.Create(ctx, notification)
    
    // 2. Send via WebSocket (ONE LINE!)
    helpers.SendNotification(userID, notification)
    
    return nil
}
```

### Sending Chat Messages

```go
import "github.com/umar5678/go-backend/internal/utils/helpers"

// In your message service
func (s *MessageService) SendMessage(ctx context.Context, senderID, receiverID, content string) error {
    // 1. Save to database
    message := &models.Message{
        SenderID:   senderID,
        ReceiverID: receiverID,
        Content:    content,
    }
    s.repo.Create(ctx, message)
    
    // 2. Send to receiver
    helpers.SendChatMessage(receiverID, message)
    
    // 3. Send confirmation to sender
    helpers.SendChatMessageSent(senderID, message)
    
    return nil
}
```

### Check User Online Status

```go
import "github.com/umar5678/go-backend/internal/utils/helpers"

isOnline, _ := helpers.IsUserOnline(userID)
if isOnline {
    // Send real-time notification
    helpers.SendNotification(userID, notification)
} else {
    // Queue for later or send push notification
    queueNotification(userID, notification)
}
```

## 📡 WebSocket Connection (Frontend)

### JavaScript/TypeScript

```javascript
const token = localStorage.getItem('accessToken');
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = () => {
    console.log('✅ Connected to WebSocket');
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    
    switch(message.type) {
        case 'notification':
            showNotification(message.data.notification);
            break;
        case 'chat_message':
            displayChatMessage(message.data.message);
            break;
        case 'typing':
            showTypingIndicator(message.data.senderId);
            break;
        case 'user_online':
            updateUserStatus(message.data.userId, true);
            break;
        case 'user_offline':
            updateUserStatus(message.data.userId, false);
            break;
    }
};

ws.onerror = (error) => {
    console.error('❌ WebSocket error:', error);
};

ws.onclose = () => {
    console.log('🔴 Disconnected from WebSocket');
    // Implement reconnection logic
};
```

### React Hook

```typescript
import { useEffect, useState } from 'react';

export function useWebSocket(token: string) {
    const [ws, setWs] = useState<WebSocket | null>(null);
    const [isConnected, setIsConnected] = useState(false);

    useEffect(() => {
        const socket = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

        socket.onopen = () => {
            console.log('✅ Connected');
            setIsConnected(true);
        };

        socket.onmessage = (event) => {
            const message = JSON.parse(event.data);
            handleMessage(message);
        };

        socket.onclose = () => {
            console.log('🔴 Disconnected');
            setIsConnected(false);
        };

        setWs(socket);

        return () => socket.close();
    }, [token]);

    return { ws, isConnected };
}
```

## 🔧 Backend Usage Examples

### Notification Service

```go
package service

import (
    "context"
    "github.com/umar5678/go-backend/internal/utils/helpers"
)

// Send to one user
func (s *NotificationService) NotifyUser(ctx context.Context, userID string, title, message string) error {
    notification := map[string]interface{}{
        "title":   title,
        "message": message,
        "type":    "info",
    }
    
    return helpers.SendNotification(userID, notification)
}

// Send to multiple users
func (s *NotificationService) NotifyMultiple(ctx context.Context, userIDs []string, notification interface{}) error {
    return helpers.SendNotificationToMultiple(userIDs, notification)
}

// Broadcast to all
func (s *NotificationService) BroadcastAnnouncement(ctx context.Context, announcement string) error {
    return helpers.BroadcastNotification(map[string]interface{}{
        "type":    "announcement",
        "message": announcement,
    })
}
```

### Message Service

```go
package service

import (
    "context"
    "github.com/umar5678/go-backend/internal/utils/helpers"
)

// Send message
func (s *MessageService) SendMessage(ctx context.Context, senderID, receiverID, content string) error {
    message := &Message{
        SenderID:   senderID,
        ReceiverID: receiverID,
        Content:    content,
    }
    
    // Save to DB
    s.repo.Create(ctx, message)
    
    // Send via WebSocket
    helpers.SendChatMessage(receiverID, message)
    helpers.SendChatMessageSent(senderID, message)
    
    return nil
}

// Send typing indicator
func (s *MessageService) SendTyping(ctx context.Context, senderID, receiverID string, isTyping bool) error {
    return helpers.SendTypingIndicator(receiverID, senderID, isTyping)
}

// Mark messages as read
func (s *MessageService) MarkAsRead(ctx context.Context, senderID, receiverID string, messageIDs []string) error {
    // Update in DB
    s.repo.MarkAsRead(ctx, messageIDs)
    
    // Send read receipt
    return helpers.SendReadReceipt(senderID, receiverID, messageIDs)
}
```

## 🗄️ Redis Usage

### Cache Operations

```go
import "github.com/umar5678/go-backend/internal/services/cache"

// Store JSON
user := &User{ID: "123", Name: "John"}
cache.SetJSON(ctx, "user:123", user, 1*time.Hour)

// Retrieve JSON
var user User
cache.GetJSON(ctx, "user:123", &user)

// Simple key-value
cache.Set(ctx, "counter", 0, 5*time.Minute)
cache.Get(ctx, "counter")

// Counter
cache.Increment(ctx, "views")
cache.IncrementWithExpiry(ctx, "rate:user123", 1*time.Minute)

// Check existence
exists, _ := cache.Exists(ctx, "user:123")

// Delete
cache.Delete(ctx, "user:123")
```

### Session Management

```go
import "github.com/umar5678/go-backend/internal/services/cache"

// Store session
cache.SetSession(ctx, userID, accessToken, 24*time.Hour)

// Get session
token, _ := cache.GetSession(ctx, userID)

// Delete session (logout)
cache.DeleteSession(ctx, userID)

// Check if session exists
exists, _ := cache.SessionExists(ctx, userID)
```

### Presence (Online/Offline)

```go
import "github.com/umar5678/go-backend/internal/services/cache"

// Set user online (automatically done by WebSocket)
metadata := map[string]interface{}{
    "device": "mobile",
    "ip":     "192.168.1.1",
}
cache.SetPresence(ctx, userID, socketID, metadata)

// Check if online
isOnline, _ := cache.IsOnline(ctx, userID)

// Get device count
deviceCount, _ := cache.GetDeviceCount(ctx, userID)

// Get all devices
devices, _ := cache.GetUserDevices(ctx, userID)

// Remove presence (automatically done by WebSocket)
cache.RemovePresence(ctx, userID, socketID)
```

## 📊 Message Types

### Notification Messages

- `notification` - New notification
- `notification_read` - Notification marked as read
- `notification_bulk` - Multiple notifications

### Chat Messages

- `chat_message` - New chat message
- `chat_message_sent` - Message sent confirmation
- `chat_edit` - Message edited
- `chat_delete` - Message deleted
- `typing` - Typing indicator
- `read_receipt` - Read receipt

### Presence Messages

- `user_online` - User came online
- `user_offline` - User went offline

### System Messages

- `system` - System message
- `error` - Error message
- `ping`/`pong` - Connection health check

## 🔒 Security

### WebSocket Authentication

WebSocket connections are authenticated using JWT tokens:

```go
// Token can be passed via query parameter or Authorization header
ws://localhost:8080/ws?token=YOUR_JWT_TOKEN

// Or via header (recommended for production)
Authorization: Bearer YOUR_JWT_TOKEN
```

### Session Validation

All WebSocket connections validate sessions in Redis:

```go
// In websocket/auth.go
func AuthenticateWebSocket(ctx context.Context, token, jwtSecret string) (string, error) {
    // 1. Verify JWT
    claims, err := jwt.Verify(token, jwtSecret)
    
    // 2. Check session in Redis
    _, err = cache.GetSession(ctx, claims.UserID)
    
    return claims.UserID, err
}
```

## 🎯 Best Practices

1. **Always use helpers** - Don't publish to Redis directly
2. **Check online status** - Before sending real-time notifications
3. **Handle failures gracefully** - WebSocket may fail, have fallback
4. **Use message IDs** - For request/response correlation
5. **Implement reconnection** - On frontend
6. **Monitor connections** - Use health endpoints

## 📈 Monitoring

Check WebSocket health:

```bash
curl http://localhost:8080/health

{
  "status": "ok",
  "redis": "healthy",
  "websocket": {
    "users": 150,        // Unique users
    "connections": 200   // Total connections (multi-device)
  }
}
```

## 🐛 Debugging

Enable debug logging:

```bash
LOG_LEVEL=debug go run cmd/api/main.go
```

Test WebSocket connection:

```bash
# Using websocat
websocat ws://localhost:8080/ws?token=YOUR_TOKEN

# Send ping
{"type":"ping","data":{}}

# Expected response
{"type":"pong","data":{"timestamp":"2025-11-15T..."}}
```

