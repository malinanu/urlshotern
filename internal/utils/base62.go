package utils

import (
	"math"
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Encode encodes an integer to a base62 string
func EncodeBase62(num int64) string {
	if num == 0 {
		return "0"
	}

	var result strings.Builder
	for num > 0 {
		remainder := num % 62
		result.WriteByte(base62Chars[remainder])
		num = num / 62
	}

	// Reverse the string
	encoded := result.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// Decode decodes a base62 string to an integer
func DecodeBase62(encoded string) int64 {
	var result int64
	for i, char := range encoded {
		charValue := getCharValue(char)
		if charValue == -1 {
			return -1 // Invalid character
		}
		
		power := len(encoded) - i - 1
		result += int64(charValue) * int64(math.Pow(62, float64(power)))
	}

	return result
}

// getCharValue returns the numeric value of a base62 character
func getCharValue(char rune) int {
	switch {
	case char >= '0' && char <= '9':
		return int(char - '0')
	case char >= 'A' && char <= 'Z':
		return int(char - 'A' + 10)
	case char >= 'a' && char <= 'z':
		return int(char - 'a' + 36)
	default:
		return -1
	}
}

// IsValidBase62 checks if a string contains only valid base62 characters
func IsValidBase62(str string) bool {
	for _, char := range str {
		if getCharValue(char) == -1 {
			return false
		}
	}
	return true
}