package websocket

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"gorm.io/gorm"
)

type ConnectionMonitor struct {
	heartbeatInterval time.Duration
	heartbeatTimeout  time.Duration
	maxInactivity     time.Duration
	db                *gorm.DB
}

func NewConnectionMonitor(heartbeatInterval, heartbeatTimeout, maxInactivity time.Duration) *ConnectionMonitor {
	return &ConnectionMonitor{
		heartbeatInterval: heartbeatInterval,
		heartbeatTimeout:  heartbeatTimeout,
		maxInactivity:     maxInactivity,
	}
}

func NewConnectionMonitorWithDB(heartbeatInterval, heartbeatTimeout, maxInactivity time.Duration, db *gorm.DB) *ConnectionMonitor {
	return &ConnectionMonitor{
		heartbeatInterval: heartbeatInterval,
		heartbeatTimeout:  heartbeatTimeout,
		maxInactivity:     maxInactivity,
		db:                db,
	}
}

func (cm *ConnectionMonitor) IsConnectionHealthy(client *Client) bool {
	if client == nil {
		return false
	}

	lastHB := client.GetLastHeartbeat()
	timeSinceHeartbeat := time.Since(lastHB)

	if timeSinceHeartbeat > cm.heartbeatTimeout {
		logger.Warn("connection unhealthy - heartbeat timeout",
			"userId", client.UserID,
			"clientId", client.ID,
			"timeSinceHeartbeat", timeSinceHeartbeat,
			"timeout", cm.heartbeatTimeout,
		)
		return false
	}

	return true
}

func (cm *ConnectionMonitor) CheckInactiveConnections(clients []*Client) []string {
	var inactiveClientIDs []string

	for _, client := range clients {
		if client == nil {
			continue
		}

		if !cm.IsConnectionHealthy(client) {
			inactiveClientIDs = append(inactiveClientIDs, client.ID)
			logger.Info("marked client as inactive",
				"clientId", client.ID,
				"userId", client.UserID,
				"lastHeartbeat", client.GetLastHeartbeat(),
			)
		}
	}

	return inactiveClientIDs
}

func (cm *ConnectionMonitor) ValidateSessionActivity(session *SessionState) bool {
	timeSinceLastActivity := time.Since(session.LastHeartbeat)

	if timeSinceLastActivity > cm.maxInactivity {
		logger.Warn("session expired due to inactivity",
			"sessionId", session.SessionID,
			"userId", session.UserID,
			"timeSinceLastActivity", timeSinceLastActivity,
			"maxInactivity", cm.maxInactivity,
		)
		return false
	}

	return true
}

func (cm *ConnectionMonitor) ValidateClientRole(userID string, sessionRole models.UserRole) bool {
	if sessionRole == "" {
		logger.Warn("invalid role in session", "userId", userID, "role", sessionRole)
		return false
	}

	if cm.db == nil {
		logger.Warn("no database connection for role validation", "userId", userID)
		return true
	}

	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := cm.db.WithContext(ctx).Select("id", "role", "status", "deleted_at").Where("id = ?", userID).First(&user)
	if result.Error != nil {
		logger.Error("failed to validate user role", "userId", userID, "error", result.Error)
		return false
	}

	if user.DeletedAt.Valid {
		logger.Warn("user account deleted, rejecting reconnection", "userId", userID)
		return false
	}

	if user.Role != sessionRole {
		logger.Warn("user role changed since session creation", "userId", userID, "sessionRole", sessionRole, "currentRole", user.Role)
		return false
	}

	return true
}

func (cm *ConnectionMonitor) CalculateReconnectDelay(attemptNumber int) time.Duration {

	baseDelay := time.Second
	maxDelay := 60 * time.Second

	delay := baseDelay * (1 << uint(attemptNumber-1))
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

type ClientHealthStatus struct {
	ClientID           string
	UserID             string
	IsHealthy          bool
	TimeSinceHeartbeat time.Duration
	SessionID          string
	ReconnectionCount  int
	ConnectionDuration time.Duration
	LastHeartbeatTime  time.Time
}

func (cm *ConnectionMonitor) GetClientHealth(client *Client) *ClientHealthStatus {
	if client == nil {
		return nil
	}

	client.mu.RLock()
	sessionID := client.reconnectToken
	client.mu.RUnlock()

	return &ClientHealthStatus{
		ClientID:           client.ID,
		UserID:             client.UserID,
		IsHealthy:          cm.IsConnectionHealthy(client),
		TimeSinceHeartbeat: time.Since(client.GetLastHeartbeat()),
		SessionID:          sessionID,
		ConnectionDuration: client.GetConnectionDuration(),
		LastHeartbeatTime:  client.GetLastHeartbeat(),
	}
}
