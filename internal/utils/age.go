package utils

import (
	"errors"
	"strings"
	"time"
)

const (
	// MinSignupAge is the youngest age allowed to create an account (COPPA).
	MinSignupAge = 13
	// AdultAge is the age at which an account stops being flagged as a minor.
	AdultAge = 18

	dobLayout = "2006-01-02"
	maxAge    = 120
)

var (
	// ErrInvalidDOB is returned when a date of birth is missing, malformed or implausible.
	ErrInvalidDOB = errors.New("date of birth must be a valid date in YYYY-MM-DD format")
)

// ParseDOB parses a YYYY-MM-DD date of birth and rejects future or implausibly old dates.
func ParseDOB(s string, now time.Time) (time.Time, error) {
	dob, err := time.Parse(dobLayout, strings.TrimSpace(s))
	if err != nil {
		return time.Time{}, ErrInvalidDOB
	}
	if dob.After(now) || AgeOn(dob, now) > maxAge {
		return time.Time{}, ErrInvalidDOB
	}
	return dob, nil
}

// AgeOn returns the age in whole years of someone born on dob, as of now (UTC calendar dates).
func AgeOn(dob, now time.Time) int {
	dob = dob.UTC()
	now = now.UTC()
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	return age
}
