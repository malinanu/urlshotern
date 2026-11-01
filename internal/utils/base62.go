package utils

import (
	"math"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Pre-allocated buffer pool to reduce allocations
var bufferPool = make(chan []byte, 100)

func getBuffer() []byte {
	select {
	case buf := <-bufferPool:
		return buf[:0] // Reset length but keep capacity
	default:
		return make([]byte, 0, 12) // Snowflake IDs typically need ~11 chars in base62
	}
}

func putBuffer(buf []byte) {
	if cap(buf) == 12 {
		select {
		case bufferPool <- buf:
		default:
			// Pool is full, let it be garbage collected
		}
	}
}

// EncodeBase62 encodes an integer to a base62 string (optimized version)
func EncodeBase62(num int64) string {
	if num == 0 {
		return "0"
	}

	// Get buffer from pool
	buf := getBuffer()
	defer putBuffer(buf)

	// Build string in reverse order (avoid reversing later)
	for num > 0 {
		buf = append(buf, base62Chars[num%62])
		num /= 62
	}

	// Reverse in-place
	length := len(buf)
	for i := 0; i < length/2; i++ {
		buf[i], buf[length-1-i] = buf[length-1-i], buf[i]
	}

	return string(buf)
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
