package services

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAddDailyViews(t *testing.T) {
	today := time.Date(2026, 11, 27, 15, 4, 0, 0, time.UTC)
	series := make([]int64, 14)

	addDailyViews(series, today, time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC), 5) // today: last slot
	addDailyViews(series, today, time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC), 3) // yesterday
	addDailyViews(series, today, time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC), 2) // 13 days ago: first slot
	addDailyViews(series, today, time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC), 9) // too old: ignored
	addDailyViews(series, today, time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC), 9) // future: ignored
	addDailyViews(series, today, time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC), 1) // adds up

	want := []int64{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 6}
	if !reflect.DeepEqual(series, want) {
		t.Fatalf("series = %v, want %v", series, want)
	}
}

func TestAddDailyViewsAcrossMonthAndDST(t *testing.T) {
	today := time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)
	series := make([]int64, 14)
	addDailyViews(series, today, time.Date(2026, 3, 17, 0, 0, 0, 0, time.UTC), 1)
	addDailyViews(series, today, time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC), 1)
	if series[0] != 1 || series[12] != 1 {
		t.Fatalf("series = %v", series)
	}
}

func TestHasUnpublishedChanges(t *testing.T) {
	v := func(n int) *int { return &n }
	cases := []struct {
		status    string
		version   int
		published *int
		want      bool
	}{
		{"published", 5, v(5), false},
		{"published", 6, v(5), true},
		{"published", 5, nil, false}, // published before migration 048
		{"draft", 6, v(5), false},
		{"draft", 1, nil, false},
	}
	for _, c := range cases {
		if got := hasUnpublishedChanges(c.status, c.version, c.published); got != c.want {
			t.Errorf("hasUnpublishedChanges(%q, %d, %v) = %v, want %v", c.status, c.version, c.published, got, c.want)
		}
	}
}

func TestDuplicateSlugCandidates(t *testing.T) {
	got := duplicateSlugCandidates("malin")
	if got[0] != "malin-copy" || got[1] != "malin-copy-2" || got[19] != "malin-copy-20" || len(got) != 20 {
		t.Fatalf("candidates = %v", got)
	}

	long := strings.Repeat("a", 30) + "-b"
	for _, c := range duplicateSlugCandidates(long) {
		if len(c) > 32 {
			t.Errorf("%q is longer than 32 characters", c)
		}
		if _, err := NormalizeMicrositeSlug(c); err != nil {
			t.Errorf("%q is not a valid slug: %v", c, err)
		}
	}

	// A cut that would end on a hyphen is trimmed, so no "--".
	for _, c := range duplicateSlugCandidates("abcdefghijklmnopqrstuvwxy-z") {
		if strings.Contains(c, "--") {
			t.Errorf("%q has a double hyphen", c)
		}
	}
}
