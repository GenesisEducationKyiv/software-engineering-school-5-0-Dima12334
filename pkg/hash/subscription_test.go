package hash_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"weather_forecast_sub/pkg/hash"
)

func TestGenerateSubscriptionHash(t *testing.T) {
	hasher := &hash.SHA256Hasher{}

	email := "user@example.com"
	city := "Kyiv"
	frequency := "daily"

	got := hasher.GenerateSubscriptionHash(email, city, frequency)

	assert.Len(t, got, 64, "hash length should be 64 characters")
	assert.True(t, hash.IsValidSHA256Hex(got), "hash should be a valid lowercase SHA256 hex string")
}

func TestIsValidSHA256Hex(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", true},  // valid
		{"E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855", false}, // uppercase
		{"1234", false},                   // too short
		{"xyz123", false},                 // invalid chars
		{string(make([]byte, 64)), false}, // invalid characters (nulls)
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := hash.IsValidSHA256Hex(tt.input)
			assert.Equal(t, tt.expected, result, "input: %q", tt.input)
		})
	}
}
