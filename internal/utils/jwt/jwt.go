package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a JWT token with proper validation and issuer claim
func GenerateToken(userID, role, secret, issuer string, expiry time.Duration) (string, error) {
	// Validate inputs before token generation
	if userID == "" {
		return "", errors.New("userID cannot be empty")
	}
	if role == "" {
		return "", errors.New("role cannot be empty")
	}
	if secret == "" {
		return "", errors.New("JWT secret cannot be empty")
	}
	if len(secret) < 32 {
		return "", errors.New("JWT secret must be at least 32 characters for HS256 security")
	}
	if expiry <= 0 {
		return "", errors.New("token expiry must be greater than 0")
	}
	if issuer == "" {
		return "", errors.New("JWT issuer cannot be empty")
	}

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken validates a JWT token with issuer verification
func ValidateToken(tokenString, secret, expectedIssuer string) (*Claims, error) {
	// Validate inputs
	if tokenString == "" {
		return nil, errors.New("token cannot be empty")
	}
	if secret == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 characters")
	}
	if expectedIssuer == "" {
		return nil, errors.New("JWT issuer cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Prevent algorithm substitution attacks - only allow HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method: expected HMAC")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate token signature and standard claims (expiration, issued-at, not-before)
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Additional validation: ensure required claims are present
		if claims.UserID == "" {
			return nil, errors.New("token missing userID claim")
		}
		if claims.Role == "" {
			return nil, errors.New("token missing role claim")
		}
		// Validate issuer claim to prevent token substitution from different issuers
		if claims.Issuer != expectedIssuer {
			return nil, errors.New("invalid token issuer")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
