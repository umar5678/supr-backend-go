package websocket

import (
	"context"
	"errors"
	"strings"

	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/jwt"
)

// AuthenticateWebSocket validates WebSocket connection and returns userID
func AuthenticateWebSocket(ctx context.Context, token, jwtSecret string) (string, error) {
	if token == "" {
		return "", errors.New("authentication token required")
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	// Verify JWT token
	claims, err := jwt.ValidateToken(token, jwtSecret)
	if err != nil {
		return "", errors.New("invalid or expired token")
	}

	// Check if session exists in Redis
	_, err = cache.GetSession(ctx, claims.UserID)
	if err != nil {
		return "", errors.New("session expired or invalid")
	}

	return claims.UserID, nil
}

// ValidateToken is a simpler token validation without session check
func ValidateToken(token, jwtSecret string) (string, error) {
	if token == "" {
		return "", errors.New("token required")
	}

	token = strings.TrimPrefix(token, "Bearer ")

	claims, err := jwt.ValidateToken(token, jwtSecret)
	if err != nil {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}
