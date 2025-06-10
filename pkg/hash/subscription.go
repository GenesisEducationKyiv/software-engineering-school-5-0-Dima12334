package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"unicode"
)

//go:generate mockgen -source=subscription.go -destination=mocks/mock_subscription.go

type SubscriptionHasher interface {
	GenerateSubscriptionHash(email, city, frequency string) string
}

type SHA256Hasher struct{}

func NewSHA256Hasher() *SHA256Hasher {
	return &SHA256Hasher{}
}

func (h *SHA256Hasher) GenerateSubscriptionHash(email, city, frequency string) string {
	data := []byte(email + city + frequency)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func IsValidSHA256Hex(s string) bool {
	if len(s) != sha256.BlockSize {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}
