package services

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// Upload limits
const (
	maxUploadDimension  = 4096       // reject larger source images (decompression-bomb guard)
	maxUploadPixels     = 16_777_216 // 4096 x 4096
	maxStoredDimension  = 1600       // images are downscaled to fit this box
	vcardPhotoDimension = 256
	maxAssetsPerUser    = 100
)

// Asset errors
var (
	ErrAssetTooLarge   = errors.New("image is too large")
	ErrAssetType       = errors.New("only JPEG and PNG images are allowed")
	ErrAssetDimensions = errors.New("image dimensions are too large (max 4096 x 4096)")
	ErrAssetCorrupt    = errors.New("image could not be read")
	ErrAssetQuota      = errors.New("image limit reached; delete some images first")
	ErrAssetNotFound   = errors.New("image not found")
)

// MicrositeAssetService validates, re-encodes and stores page images.
type MicrositeAssetService struct {
	db       micrositeDB
	store    AssetStore
	maxBytes int64
}

// NewMicrositeAssetService creates the asset service.
func NewMicrositeAssetService(db *storage.PostgresStorage, store AssetStore, maxBytes int64) *MicrositeAssetService {
	if maxBytes <= 0 {
		maxBytes = 2 * 1024 * 1024
	}
	return &MicrositeAssetService{db: db, store: store, maxBytes: maxBytes}
}

// MaxBytes is the upload size limit.
func (s *MicrositeAssetService) MaxBytes() int64 { return s.maxBytes }

// ProcessedImage is a validated, re-encoded image.
type ProcessedImage struct {
	Data     []byte
	MimeType string
	Ext      string
	Width    int
	Height   int
}

// ProcessImage reads at most maxBytes, sniffs the real content type (the client
// filename and Content-Type are ignored), enforces dimension limits, applies
// EXIF orientation, downscales and re-encodes. Re-encoding drops all metadata
// and any data appended to or embedded in the original file.
func ProcessImage(r io.Reader, maxBytes int64) (*ProcessedImage, error) {
	raw, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, ErrAssetCorrupt
	}
	if int64(len(raw)) > maxBytes {
		return nil, ErrAssetTooLarge
	}
	sniffed := http.DetectContentType(raw)
	if sniffed != "image/jpeg" && sniffed != "image/png" {
		return nil, ErrAssetType
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return nil, ErrAssetCorrupt
	}
	if (format != "jpeg" && format != "png") || "image/"+format != sniffed {
		return nil, ErrAssetType
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > maxUploadDimension || cfg.Height > maxUploadDimension ||
		cfg.Width*cfg.Height > maxUploadPixels {
		return nil, ErrAssetDimensions
	}

	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, ErrAssetCorrupt
	}
	if format == "jpeg" {
		img = applyOrientation(img, jpegOrientation(raw))
	}
	img = downscale(img, maxStoredDimension)

	var out bytes.Buffer
	p := &ProcessedImage{Width: img.Bounds().Dx(), Height: img.Bounds().Dy()}
	if format == "png" {
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		if err := enc.Encode(&out, img); err != nil {
			return nil, ErrAssetCorrupt
		}
		p.MimeType, p.Ext = "image/png", "png"
	} else {
		if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 85}); err != nil {
			return nil, ErrAssetCorrupt
		}
		p.MimeType, p.Ext = "image/jpeg", "jpg"
	}
	p.Data = out.Bytes()
	return p, nil
}

// Upload processes and stores an image for a user, returning the asset.
func (s *MicrositeAssetService) Upload(userID int64, micrositeID *int64, r io.Reader) (*models.MicrositeAsset, error) {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM microsite_assets WHERE user_id = $1`, userID).Scan(&count); err != nil {
		return nil, mapMicrositeDBError(err)
	}
	if count >= maxAssetsPerUser {
		return nil, ErrAssetQuota
	}

	img, err := ProcessImage(r, s.maxBytes)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(img.Data)
	digest := hex.EncodeToString(sum[:])
	relPath := fmt.Sprintf("m/%d/%s.%s", userID, digest, img.Ext)

	// Identical re-uploads reuse the existing asset.
	if existing, err := s.getByPath(userID, relPath); err == nil {
		return existing, nil
	}
	if err := s.store.Put(relPath, img.Data); err != nil {
		return nil, fmt.Errorf("store image: %w", err)
	}

	a := &models.MicrositeAsset{MicrositeID: micrositeID, MimeType: img.MimeType, SizeBytes: int64(len(img.Data)),
		Width: img.Width, Height: img.Height, URL: "/uploads/" + relPath}
	err = s.db.QueryRow(`
		INSERT INTO microsite_assets (user_id, microsite_id, path, mime_type, size_bytes, width, height, sha256)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`,
		userID, micrositeID, relPath, img.MimeType, a.SizeBytes, img.Width, img.Height, digest).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		_ = s.store.Delete(relPath)
		return nil, mapMicrositeDBError(err)
	}
	return a, nil
}

func (s *MicrositeAssetService) getByPath(userID int64, relPath string) (*models.MicrositeAsset, error) {
	var (
		a    models.MicrositeAsset
		msID sql.NullInt64
		path string
	)
	err := s.db.QueryRow(`SELECT id, microsite_id, path, mime_type, size_bytes, width, height, created_at
		FROM microsite_assets WHERE user_id = $1 AND path = $2 LIMIT 1`, userID, relPath).
		Scan(&a.ID, &msID, &path, &a.MimeType, &a.SizeBytes, &a.Width, &a.Height, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	if msID.Valid {
		a.MicrositeID = &msID.Int64
	}
	a.URL = "/uploads/" + path
	return &a, nil
}

// List returns a user's uploaded images, newest first.
func (s *MicrositeAssetService) List(userID int64) ([]models.MicrositeAsset, error) {
	rows, err := s.db.Query(`SELECT id, microsite_id, path, mime_type, size_bytes, width, height, created_at
		FROM microsite_assets WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, mapMicrositeDBError(err)
	}
	defer rows.Close()
	out := []models.MicrositeAsset{}
	for rows.Next() {
		var (
			a    models.MicrositeAsset
			msID sql.NullInt64
			path string
		)
		if err := rows.Scan(&a.ID, &msID, &path, &a.MimeType, &a.SizeBytes, &a.Width, &a.Height, &a.CreatedAt); err != nil {
			return nil, err
		}
		if msID.Valid {
			a.MicrositeID = &msID.Int64
		}
		a.URL = "/uploads/" + path
		out = append(out, a)
	}
	return out, rows.Err()
}

// Delete removes one of the user's images (row and file).
func (s *MicrositeAssetService) Delete(userID, assetID int64) error {
	var path string
	err := s.db.QueryRow(`DELETE FROM microsite_assets WHERE id = $1 AND user_id = $2 RETURNING path`, assetID, userID).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrAssetNotFound
	}
	if err != nil {
		return mapMicrositeDBError(err)
	}
	return s.store.Delete(path)
}

// DeleteAllForUser removes every stored image of a user (account deletion).
func (s *MicrositeAssetService) DeleteAllForUser(userID int64) error {
	assets, err := s.List(userID)
	if err != nil {
		return err
	}
	for _, a := range assets {
		_ = s.store.Delete(strings.TrimPrefix(a.URL, "/uploads/"))
	}
	_, err = s.db.Exec(`DELETE FROM microsite_assets WHERE user_id = $1`, userID)
	return mapMicrositeDBError(err)
}

// LoadVCardPhoto returns a small JPEG of an uploaded image for embedding in a
// vCard. Only local upload paths are loaded; remote URLs are never fetched.
func (s *MicrositeAssetService) LoadVCardPhoto(imageURL string) *VCardPhoto {
	if s == nil || s.store == nil || !localAssetPattern.MatchString(imageURL) {
		return nil
	}
	data, err := s.store.Get(strings.TrimPrefix(imageURL, "/uploads/"))
	if err != nil {
		return nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	img = downscale(img, vcardPhotoDimension)
	var out bytes.Buffer
	if err := jpeg.Encode(&out, flattenOnWhite(img), &jpeg.Options{Quality: 82}); err != nil {
		return nil
	}
	return &VCardPhoto{Data: out.Bytes(), MimeType: "image/jpeg"}
}

// ---- image helpers -------------------------------------------------------------

// downscale shrinks img to fit within maxDim x maxDim using box filtering.
func downscale(src image.Image, maxDim int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxDim && h <= maxDim {
		return src
	}
	nw, nh := maxDim, maxDim
	if w >= h {
		nh = h * maxDim / w
	} else {
		nw = w * maxDim / h
	}
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewNRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		sy0, sy1 := b.Min.Y+y*h/nh, b.Min.Y+(y+1)*h/nh
		if sy1 <= sy0 {
			sy1 = sy0 + 1
		}
		for x := 0; x < nw; x++ {
			sx0, sx1 := b.Min.X+x*w/nw, b.Min.X+(x+1)*w/nw
			if sx1 <= sx0 {
				sx1 = sx0 + 1
			}
			var r, g, bl, a, n uint64
			for sy := sy0; sy < sy1; sy++ {
				for sx := sx0; sx < sx1; sx++ {
					c := color.NRGBA64Model.Convert(src.At(sx, sy)).(color.NRGBA64)
					r += uint64(c.R)
					g += uint64(c.G)
					bl += uint64(c.B)
					a += uint64(c.A)
					n++
				}
			}
			// #nosec G115 -- averages of uint16 channel values always fit in uint16
			dst.Set(x, y, color.NRGBA64{R: uint16(r / n), G: uint16(g / n), B: uint16(bl / n), A: uint16(a / n)})
		}
	}
	return dst
}

// flattenOnWhite composites transparent pixels onto white (JPEG has no alpha).
func flattenOnWhite(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)
			a := uint32(c.A)
			// #nosec G115 -- weighted average of two values <= 255 is <= 255
			blend := func(v uint8) uint8 { return uint8((uint32(v)*a + 255*(255-a)) / 255) }
			dst.Set(x, y, color.RGBA{R: blend(c.R), G: blend(c.G), B: blend(c.B), A: 255})
		}
	}
	return dst
}

// jpegOrientation returns the EXIF orientation (1-8) of a JPEG, or 1.
func jpegOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 1
	}
	i := 2
	for i+4 <= len(data) {
		if data[i] != 0xFF {
			return 1
		}
		marker := data[i+1]
		if marker == 0xDA || marker == 0xD9 { // start of scan / end of image
			return 1
		}
		segLen := int(binary.BigEndian.Uint16(data[i+2 : i+4]))
		if segLen < 2 || i+2+segLen > len(data) {
			return 1
		}
		seg := data[i+4 : i+2+segLen]
		if marker == 0xE1 && len(seg) > 14 && string(seg[:6]) == "Exif\x00\x00" {
			return exifOrientation(seg[6:])
		}
		i += 2 + segLen
	}
	return 1
}

func exifOrientation(tiff []byte) int {
	if len(tiff) < 8 {
		return 1
	}
	var bo binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return 1
	}
	ifd := int(bo.Uint32(tiff[4:8]))
	if ifd < 8 || ifd+2 > len(tiff) {
		return 1
	}
	n := int(bo.Uint16(tiff[ifd : ifd+2]))
	for k := 0; k < n; k++ {
		e := ifd + 2 + k*12
		if e+12 > len(tiff) {
			return 1
		}
		if bo.Uint16(tiff[e:e+2]) == 0x0112 {
			v := int(bo.Uint16(tiff[e+8 : e+10]))
			if v >= 1 && v <= 8 {
				return v
			}
			return 1
		}
	}
	return 1
}

// applyOrientation transforms img so it displays upright for EXIF orientation o.
func applyOrientation(src image.Image, o int) image.Image {
	if o <= 1 || o > 8 {
		return src
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dw, dh := w, h
	if o >= 5 {
		dw, dh = h, w
	}
	dst := image.NewNRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var dx, dy int
			switch o {
			case 2:
				dx, dy = w-1-x, y
			case 3:
				dx, dy = w-1-x, h-1-y
			case 4:
				dx, dy = x, h-1-y
			case 5:
				dx, dy = y, x
			case 6:
				dx, dy = h-1-y, x
			case 7:
				dx, dy = h-1-y, w-1-x
			case 8:
				dx, dy = y, w-1-x
			}
			dst.Set(dx, dy, src.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}
