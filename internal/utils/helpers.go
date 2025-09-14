package utils

import "time"

// stringPtr returns a pointer to the string value
func StringPtr(s string) *string {
	return &s
}

// timePtr returns a pointer to the time value
func TimePtr(t time.Time) *time.Time {
	return &t
}

// intPtr returns a pointer to the int value
func IntPtr(i int) *int {
	return &i
}

// int64Ptr returns a pointer to the int64 value
func Int64Ptr(i int64) *int64 {
	return &i
}

// boolPtr returns a pointer to the bool value
func BoolPtr(b bool) *bool {
	return &b
}