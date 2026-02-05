package websocket

import (
	"context"
	"errors"
	"strings"

	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/jwt"
)

func AuthenticateWebSocket(ctx context.Context, token, jwtSecret string) (string, error) {
	if token == "" {
		return "", errors.New("authentication token required")
	}

	token = strings.TrimPrefix(token, "Bearer ")

	claims, err := jwt.ValidateToken(token, jwtSecret)
	if err != nil {
		return "", errors.New("invalid or expired token")
	}

	_, err = cache.GetSession(ctx, claims.UserID)
	if err != nil {
		return "", errors.New("session expired or invalid")
	}

	return claims.UserID, nil
}

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
