package unit

import (
	"testing"

	"github.com/URLshorter/url-shortener/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestBase62Encoding(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"Zero", 0, "0"},
		{"Small number", 61, "z"},
		{"Medium number", 62, "10"},
		{"Large number", 3844, "100"},
		{"Very large number", 238328, "zzz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.EncodeBase62(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBase62Decoding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"Zero", "0", 0},
		{"Single char", "z", 61},
		{"Two chars", "10", 62},
		{"Three chars", "100", 3844},
		{"Complex", "zzz", 238328},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.DecodeBase62(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBase62RoundTrip(t *testing.T) {
	testValues := []int64{0, 1, 61, 62, 3844, 238328, 14776336, 916132832}

	for _, value := range testValues {
		t.Run("", func(t *testing.T) {
			encoded := utils.EncodeBase62(value)
			decoded := utils.DecodeBase62(encoded)
			assert.Equal(t, value, decoded)
		})
	}
}

func TestIsValidBase62(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid alphanumeric", "abc123XYZ", true},
		{"Valid all lowercase", "abcdefghijk", true},
		{"Valid all uppercase", "ABCDEFGHIJK", true},
		{"Valid all numbers", "0123456789", true},
		{"Invalid with special chars", "abc-123", false},
		{"Invalid with space", "abc 123", false},
		{"Invalid with symbols", "abc@123", false},
		{"Empty string", "", true}, // Empty string is technically valid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsValidBase62(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkBase62Encode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		utils.EncodeBase62(int64(i))
	}
}

func BenchmarkBase62Decode(b *testing.B) {
	encoded := utils.EncodeBase62(123456789)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		utils.DecodeBase62(encoded)
	}
}