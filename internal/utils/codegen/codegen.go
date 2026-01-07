package codegen

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	// ReferralCodeCharset defines the characters used for generating referral codes
	ReferralCodeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// ReferralCodeLength is the length of generated referral codes
	ReferralCodeLength = 10
)

// GenerateReferralCode generates a unique referral code
// Format: XXXXX-XXXXX (10 alphanumeric characters with hyphen for readability)
func GenerateReferralCode() (string, error) {
	code, err := generateRandomString(ReferralCodeLength)
	if err != nil {
		return "", err
	}

	// Format: XXXXX-XXXXX for readability
	return fmt.Sprintf("%s-%s", code[:5], code[5:]), nil
}

// GenerateRidePIN generates a 4-digit PIN for ride verification
func GenerateRidePIN() (string, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", num.Int64()), nil
}

// GeneratePromoCode generates a promo code
func GeneratePromoCode(prefix string) (string, error) {
	randomPart, err := generateRandomString(8)
	if err != nil {
		return "", err
	}

	code := prefix
	if prefix != "" {
		code = prefix + "-"
	}

	// Add timestamp for uniqueness
	timestamp := fmt.Sprintf("%d", time.Now().Unix()%100000)
	code += randomPart + timestamp

	return strings.ToUpper(code), nil
}

// generateRandomString generates a random string of given length from the charset
func generateRandomString(length int) (string, error) {
	charset := ReferralCodeCharset
	b := make([]byte, length)

	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}

	return string(b), nil
}
