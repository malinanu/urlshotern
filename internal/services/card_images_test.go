package services

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func testPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 120, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func newCardImages(t *testing.T) (*CardImageService, string) {
	t.Helper()
	dir := t.TempDir()
	store, err := NewLocalAssetStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	return NewCardImageService(store, 0), store.Root()
}

func TestCardImageUploadStoresContentAddressedFile(t *testing.T) {
	svc, root := newCardImages(t)
	url, err := svc.Upload(42, bytes.NewReader(testPNG(t, 64, 48)))
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^/uploads/mp/42/[a-f0-9]{64}\.png$`).MatchString(url) {
		t.Fatalf("unexpected url %q", url)
	}
	if _, err := os.Stat(filepath.Join(root, strings.TrimPrefix(url, "/uploads/"))); err != nil {
		t.Fatalf("file not stored: %v", err)
	}
	// An identical re-upload returns the same URL and adds no file.
	again, err := svc.Upload(42, bytes.NewReader(testPNG(t, 64, 48)))
	if err != nil || again != url {
		t.Fatalf("re-upload = %q, %v; want %q", again, err, url)
	}
	entries, _ := os.ReadDir(filepath.Join(root, "mp", "42"))
	if len(entries) != 1 {
		t.Fatalf("%d files stored, want 1", len(entries))
	}
}

func TestCardImageRejectsNonImagesAndScripts(t *testing.T) {
	svc, _ := newCardImages(t)
	for name, body := range map[string]string{
		"svg":  `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`,
		"html": `<html><script>alert(1)</script></html>`,
		"text": "hello",
	} {
		if _, err := svc.Upload(1, strings.NewReader(body)); !errors.Is(err, ErrAssetType) {
			t.Errorf("%s: err = %v, want ErrAssetType", name, err)
		}
	}
}

func TestCardImageQuotaAndUserFileCleanup(t *testing.T) {
	svc, root := newCardImages(t)
	for i := 0; i < maxCardImagesPerUser; i++ {
		if _, err := svc.Upload(7, bytes.NewReader(testPNG(t, 8+i, 8))); err != nil {
			t.Fatalf("upload %d: %v", i, err)
		}
	}
	if _, err := svc.Upload(7, bytes.NewReader(testPNG(t, 200, 8))); !errors.Is(err, ErrAssetQuota) {
		t.Fatalf("err = %v, want ErrAssetQuota", err)
	}
	// Another user is unaffected.
	if _, err := svc.Upload(8, bytes.NewReader(testPNG(t, 200, 8))); err != nil {
		t.Fatal(err)
	}
	store, _ := NewLocalAssetStore(root)
	if err := store.DeleteUserFiles(7); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "mp", "7")); !os.IsNotExist(err) {
		t.Fatalf("user 7 files not removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "mp", "8")); err != nil {
		t.Fatal("user 8 files must remain")
	}
}
