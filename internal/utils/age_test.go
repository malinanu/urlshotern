package utils

import (
	"testing"
	"time"
)

func TestAgeOn(t *testing.T) {
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		dob  string
		want int
	}{
		{"2013-10-27", 13}, // 13th birthday today
		{"2013-10-28", 12}, // 13th birthday tomorrow
		{"2013-11-01", 12},
		{"2008-10-28", 17}, // 18 tomorrow
		{"2008-10-27", 18},
		{"2000-02-29", 26}, // leap-day birthday
	}
	for _, c := range cases {
		dob, _ := time.Parse("2006-01-02", c.dob)
		if got := AgeOn(dob, now); got != c.want {
			t.Errorf("AgeOn(%s) = %d, want %d", c.dob, got, c.want)
		}
	}
}

func TestParseDOB(t *testing.T) {
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	valid := []string{"2000-01-15", " 2013-10-27 ", "1910-01-01"}
	for _, s := range valid {
		if _, err := ParseDOB(s, now); err != nil {
			t.Errorf("ParseDOB(%q) unexpected error: %v", s, err)
		}
	}
	invalid := []string{"", "15/01/2000", "2000-13-01", "2026-10-28", "1890-01-01", "not-a-date"}
	for _, s := range invalid {
		if _, err := ParseDOB(s, now); err == nil {
			t.Errorf("ParseDOB(%q) expected error", s)
		}
	}
}
