package password

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MinPasswordLength defines minimum password length
	MinPasswordLength = 8

	// MaxPasswordLength defines maximum password length
	MaxPasswordLength = 128

	// DefaultCost is bcrypt default cost (14 = ~1 second on modern hardware)
	DefaultCost = 14
)

// Hash generates bcrypt hash from password
func Hash(password string) (string, error) {
	// Validate password length
	if len(password) < MinPasswordLength {
		return "", fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	if len(password) > MaxPasswordLength {
		return "", fmt.Errorf("password must not exceed %d characters", MaxPasswordLength)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

// HashWithCost generates bcrypt hash with custom cost
func HashWithCost(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return "", fmt.Errorf("cost must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(bytes), nil
}

// Verify checks if password matches hash
func Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// VerifyWithError checks password and returns detailed error
func VerifyWithError(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("invalid password")
		}
		return fmt.Errorf("password verification failed: %w", err)
	}
	return nil
}

// Strength checks password strength
type PasswordStrength int

const (
	StrengthWeak PasswordStrength = iota
	StrengthFair
	StrengthGood
	StrengthStrong
)

// CheckStrength evaluates password strength
func CheckStrength(password string) PasswordStrength {
	length := len(password)

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	score := 0
	if length >= 8 {
		score++
	}
	if length >= 12 {
		score++
	}
	if hasUpper && hasLower {
		score++
	}
	if hasDigit {
		score++
	}
	if hasSpecial {
		score++
	}

	switch {
	case score <= 2:
		return StrengthWeak
	case score == 3:
		return StrengthFair
	case score == 4:
		return StrengthGood
	default:
		return StrengthStrong
	}
}

// ValidatePassword performs comprehensive password validation
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	}

	if len(password) > MaxPasswordLength {
		return fmt.Errorf("password must not exceed %d characters", MaxPasswordLength)
	}

	strength := CheckStrength(password)
	if strength == StrengthWeak {
		return errors.New("password is too weak, must contain uppercase, lowercase, numbers, and special characters")
	}

	return nil
}

// GenerateRandomPassword generates cryptographically secure random password
func GenerateRandomPassword(length int) (string, error) {
	if length < MinPasswordLength {
		length = MinPasswordLength
	}
	if length > MaxPasswordLength {
		length = MaxPasswordLength
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random password: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// NeedsRehash checks if password hash needs to be updated
func NeedsRehash(hash string) bool {
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return true
	}
	return cost < DefaultCost
}
