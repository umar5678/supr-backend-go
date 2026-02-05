package codegen

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	ReferralCodeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ReferralCodeLength = 10
)

func GenerateReferralCode() (string, error) {
	code, err := generateRandomString(ReferralCodeLength)
	if err != nil {
		return "", err
	}

	// Format: XXXXX-XXXXX for readability
	return fmt.Sprintf("%s-%s", code[:5], code[5:]), nil
}

func GenerateRidePIN() (string, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", num.Int64()), nil
}

func GeneratePromoCode(prefix string) (string, error) {
	randomPart, err := generateRandomString(8)
	if err != nil {
		return "", err
	}

	code := prefix
	if prefix != "" {
		code = prefix + "-"
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix()%100000)
	code += randomPart + timestamp

	return strings.ToUpper(code), nil
}

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
