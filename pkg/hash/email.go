package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

//go:generate mockgen -source=email.go -destination=mocks/mock_email.go

type EmailHasher interface {
	GenerateEmailHash(email string) string
}

type SHA256Hasher struct{}

func NewSHA256Hasher() *SHA256Hasher {
	return &SHA256Hasher{}
}

func (h *SHA256Hasher) GenerateEmailHash(email string) string {
	hash := sha256.Sum256([]byte(email))
	return hex.EncodeToString(hash[:])
}

func IsValidSHA256Hex(s string) bool {
	if len(s) != sha256.BlockSize {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}
