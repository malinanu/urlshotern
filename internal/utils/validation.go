package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Phone number validation regex patterns
var (
	// E.164 format: +[country code][subscriber number]
	e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

	// North American format: +1XXXXXXXXXX or +1 (XXX) XXX-XXXX
	northAmericanRegex = regexp.MustCompile(`^\+1\d{10}$|^\+1\s?\(\d{3}\)\s?\d{3}-\d{4}$`)

	// International format with optional spaces and dashes
	internationalRegex = regexp.MustCompile(`^\+\d{1,3}[\s-]?\(?\d{1,4}\)?[\s-]?\d{1,4}[\s-]?\d{1,9}$`)
)

// IsValidPhone validates phone number format
// Accepts various international formats including E.164
func IsValidPhone(phone string) bool {
	if phone == "" {
		return true // Empty is valid (optional field)
	}

	// Remove whitespace for validation
	cleaned := strings.TrimSpace(phone)

	// Check E.164 format (preferred)
	if e164Regex.MatchString(cleaned) {
		return true
	}

	// Check North American format
	if northAmericanRegex.MatchString(cleaned) {
		return true
	}

	// Check international format with flexible spacing
	if internationalRegex.MatchString(cleaned) {
		return true
	}

	return false
}

// NormalizePhone normalizes phone number to E.164 format
// Example: (555) 123-4567 -> +15551234567
func NormalizePhone(phone string) string {
	if phone == "" {
		return ""
	}

	// Remove all non-digit characters except leading +
	var normalized strings.Builder
	for i, r := range phone {
		if r == '+' && i == 0 {
			normalized.WriteRune(r)
		} else if unicode.IsDigit(r) {
			normalized.WriteRune(r)
		}
	}

	result := normalized.String()

	// If no country code, assume US (+1)
	if !strings.HasPrefix(result, "+") {
		if len(result) == 10 {
			result = "+1" + result
		} else {
			result = "+" + result
		}
	}

	return result
}

// ValidatePasswordStrength validates password strength requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be less than 128 characters")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateName validates user name format
func ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)

	if len(trimmed) < 2 {
		return fmt.Errorf("name must be at least 2 characters long")
	}

	if len(trimmed) > 100 {
		return fmt.Errorf("name must be less than 100 characters")
	}

	// Check for valid characters (letters, spaces, hyphens, apostrophes)
	validNameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !validNameRegex.MatchString(trimmed) {
		return fmt.Errorf("name contains invalid characters")
	}

	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	trimmed := strings.TrimSpace(email)

	if len(trimmed) == 0 {
		return fmt.Errorf("email is required")
	}

	if len(trimmed) > 254 {
		return fmt.Errorf("email is too long")
	}

	// RFC 5322 compliant email regex (simplified)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(trimmed) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// SanitizeInput removes potentially dangerous characters from user input
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// ValidateURL validates URL format (for original URLs in shortener)
func ValidateURL(url string) error {
	if len(url) == 0 {
		return fmt.Errorf("URL is required")
	}

	if len(url) > 2048 {
		return fmt.Errorf("URL is too long (max 2048 characters)")
	}

	// Basic URL validation
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("invalid URL format (must start with http:// or https://)")
	}

	return nil
}

// ValidateSortOrder validates sort order parameter (asc or desc only)
// Fixes Issue #13 - Add Query Parameter Validation
func ValidateSortOrder(sortOrder string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(sortOrder))

	switch normalized {
	case "asc", "desc":
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid sort_order: must be 'asc' or 'desc', got '%s'", sortOrder)
	}
}

// ValidateSortBy validates sort_by parameter against an allowed whitelist
// Fixes Issue #13 - Add Query Parameter Validation
func ValidateSortBy(sortBy string, allowedFields []string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(sortBy))

	// Check against whitelist
	for _, allowed := range allowedFields {
		if normalized == strings.ToLower(allowed) {
			return normalized, nil
		}
	}

	return "", fmt.Errorf("invalid sort_by: field '%s' is not allowed (allowed: %v)", sortBy, allowedFields)
}

// ValidateLimit validates limit parameter within acceptable range
// Fixes Issue #13 - Add Query Parameter Validation
func ValidateLimit(limit int, minLimit, maxLimit int) (int, error) {
	if limit < minLimit {
		return 0, fmt.Errorf("limit too small: must be >= %d, got %d", minLimit, limit)
	}

	if limit > maxLimit {
		return 0, fmt.Errorf("limit too large: must be <= %d, got %d", maxLimit, limit)
	}

	return limit, nil
}

// ValidateOffset validates offset parameter (must be non-negative)
// Fixes Issue #13 - Add Query Parameter Validation
func ValidateOffset(offset int64) (int64, error) {
	if offset < 0 {
		return 0, fmt.Errorf("invalid offset: must be >= 0, got %d", offset)
	}

	return offset, nil
}
