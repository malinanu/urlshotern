package handlers

import (
	"testing"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
)

func TestThumbnailContentPrefersPublishedSnapshot(t *testing.T) {
	draft := models.MicrositeContent{Blocks: []models.MicrositeBlock{{ID: "draft"}}}
	live := models.MicrositeContent{Blocks: []models.MicrositeBlock{{ID: "live"}}}

	m := &models.Microsite{Status: "published", Draft: draft, Published: &live}
	if got := thumbnailContent(m); got.Blocks[0].ID != "live" {
		t.Fatalf("published page shows %q, want the live snapshot", got.Blocks[0].ID)
	}

	m = &models.Microsite{Status: "draft", Draft: draft}
	if got := thumbnailContent(m); got.Blocks[0].ID != "draft" {
		t.Fatalf("draft page shows %q, want the draft", got.Blocks[0].ID)
	}
}

func TestThumbnailETagChangesWithContent(t *testing.T) {
	pub := time.Date(2026, 11, 27, 10, 0, 0, 0, time.UTC)
	base := models.Microsite{ID: 7, Version: 3, Status: "published", PublishedAt: &pub}
	tag := thumbnailETag(&base)

	same := base
	if thumbnailETag(&same) != tag {
		t.Fatal("same page gave a different ETag")
	}

	edited := base
	edited.Version = 4
	republished := base
	later := pub.Add(time.Minute)
	republished.PublishedAt = &later
	unpublished := base
	unpublished.Status = "draft"
	other := base
	other.ID = 8

	for name, m := range map[string]models.Microsite{
		"draft saved": edited, "republished": republished, "unpublished": unpublished, "other page": other,
	} {
		m := m
		if thumbnailETag(&m) == tag {
			t.Errorf("%s: ETag did not change", name)
		}
	}
}
