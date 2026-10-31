package services

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255}) // #nosec G115 -- test pattern
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func jpegBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w/2; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255}) // left half red
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// withOrientation inserts an EXIF APP1 segment with the given orientation.
func withOrientation(jpg []byte, o uint16) []byte {
	tiff := []byte{'I', 'I', 42, 0, 8, 0, 0, 0, 1, 0, 0x12, 0x01, 3, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint16(tiff[18:20], o)
	payload := append([]byte("Exif\x00\x00"), tiff...)
	seg := []byte{0xFF, 0xE1, 0, 0}
	binary.BigEndian.PutUint16(seg[2:4], uint16(len(payload)+2)) // #nosec G115 -- small test segment
	seg = append(seg, payload...)
	out := append([]byte{}, jpg[:2]...)
	out = append(out, seg...)
	return append(out, jpg[2:]...)
}

func TestProcessImageAcceptsAndReencodes(t *testing.T) {
	src := append(pngBytes(t, 40, 30), []byte("<script>alert(1)</script>")...) // trailing polyglot payload
	img, err := ProcessImage(bytes.NewReader(src), 2<<20)
	if err != nil {
		t.Fatal(err)
	}
	if img.MimeType != "image/png" || img.Width != 40 || img.Height != 30 {
		t.Errorf("unexpected result: %+v", img)
	}
	if bytes.Contains(img.Data, []byte("<script>")) {
		t.Error("re-encoded image still contains appended payload")
	}

	j, err := ProcessImage(bytes.NewReader(jpegBytes(t, 50, 20)), 2<<20)
	if err != nil || j.MimeType != "image/jpeg" || j.Ext != "jpg" {
		t.Fatalf("jpeg: %+v, %v", j, err)
	}
}

func TestProcessImageRejects(t *testing.T) {
	var gifBuf bytes.Buffer
	_ = gif.Encode(&gifBuf, image.NewPaletted(image.Rect(0, 0, 4, 4), []color.Color{color.Black, color.White}), nil)
	cases := map[string]struct {
		data []byte
		max  int64
		want error
	}{
		"svg":        {[]byte(`<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"></svg>`), 2 << 20, ErrAssetType},
		"html":       {[]byte(`<!DOCTYPE html><script>alert(1)</script>`), 2 << 20, ErrAssetType},
		"gif":        {gifBuf.Bytes(), 2 << 20, ErrAssetType},
		"too large":  {pngBytes(t, 64, 64), 100, ErrAssetTooLarge},
		"dimensions": {pngBytes(t, 4100, 1), 2 << 20, ErrAssetDimensions},
		"truncated":  {pngBytes(t, 64, 64)[:80], 2 << 20, ErrAssetCorrupt},
	}
	for name, c := range cases {
		if _, err := ProcessImage(bytes.NewReader(c.data), c.max); !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", name, err, c.want)
		}
	}
}

func TestProcessImageDownscalesAndOrients(t *testing.T) {
	big, err := ProcessImage(bytes.NewReader(pngBytes(t, 3200, 800)), 8<<20)
	if err != nil {
		t.Fatal(err)
	}
	if big.Width != maxStoredDimension || big.Height != 400 {
		t.Errorf("downscale = %dx%d, want %dx400", big.Width, big.Height, maxStoredDimension)
	}

	src := withOrientation(jpegBytes(t, 60, 20), 6) // rotate 90° CW to display upright
	if got := jpegOrientation(src); got != 6 {
		t.Fatalf("jpegOrientation = %d, want 6", got)
	}
	rot, err := ProcessImage(bytes.NewReader(src), 2<<20)
	if err != nil {
		t.Fatal(err)
	}
	if rot.Width != 20 || rot.Height != 60 {
		t.Fatalf("oriented size = %dx%d, want 20x60", rot.Width, rot.Height)
	}
	decoded, err := jpeg.Decode(bytes.NewReader(rot.Data))
	if err != nil {
		t.Fatal(err)
	}
	// The red left half of the source becomes the top half after a 90° CW rotation.
	if r, _, _, _ := decoded.At(10, 5).RGBA(); r>>8 < 200 {
		t.Error("expected red at top after orientation 6")
	}
	if r, _, _, _ := decoded.At(10, 55).RGBA(); r>>8 > 60 {
		t.Error("expected non-red at bottom after orientation 6")
	}
}

func TestLocalAssetStore(t *testing.T) {
	store, err := NewLocalAssetStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put("m/7/a.png", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Get("m/7/a.png"); err != nil || string(got) != "x" {
		t.Fatalf("Get = %q, %v", got, err)
	}
	// Traversal attempts are confined to the root.
	if err := store.Put("../../escape.txt", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(store.Root()), "escape.txt")); err == nil {
		t.Error("path traversal escaped the store root")
	}
	if err := store.DeleteUserFiles(7); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("m/7/a.png"); err == nil {
		t.Error("DeleteUserFiles did not remove the user's files")
	}
}

func TestLoadVCardPhoto(t *testing.T) {
	store, err := NewLocalAssetStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	name := "m/9/" + string(bytes.Repeat([]byte("a"), 64)) + ".png"
	if err := store.Put(name, pngBytes(t, 600, 300)); err != nil {
		t.Fatal(err)
	}
	svc := &MicrositeAssetService{store: store}
	photo := svc.LoadVCardPhoto("/uploads/" + name)
	if photo == nil || photo.MimeType != "image/jpeg" {
		t.Fatal("expected embedded JPEG photo")
	}
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(photo.Data))
	if err != nil || cfg.Width != vcardPhotoDimension {
		t.Errorf("photo width = %d, %v", cfg.Width, err)
	}
	if svc.LoadVCardPhoto("https://cdn.example.com/a.jpg") != nil {
		t.Error("remote URLs must never be fetched")
	}
}
