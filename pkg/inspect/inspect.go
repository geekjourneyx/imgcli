package inspect

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/topdf"
)

type Options struct {
	Inputs        []string
	InputDir      string
	IncludeHash   bool
	IncludeColors bool
	Limit         int
}

type Result struct {
	Images     []ImageInfo `json:"images"`
	Count      int         `json:"count"`
	DurationMS int64       `json:"duration_ms"`
}

type ImageInfo struct {
	Path            string `json:"path"`
	Format          string `json:"format"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	Orientation     string `json:"orientation"`
	SizeBytes       int64  `json:"size_bytes"`
	HasAlpha        bool   `json:"has_alpha"`
	EXIFOrientation int    `json:"exif_orientation,omitempty"`
	DominantColor   string `json:"dominant_color,omitempty"`
	AverageColor    string `json:"average_color,omitempty"`
	SHA256          string `json:"sha256,omitempty"`
	PHash           string `json:"phash,omitempty"`
}

func Run(opts Options) (Result, error) {
	start := time.Now()
	paths, err := topdf.CollectInputs(opts.Inputs, opts.InputDir)
	if err != nil {
		return Result{}, err
	}
	if opts.Limit > 0 && len(paths) > opts.Limit {
		paths = paths[:opts.Limit]
	}
	if len(paths) == 0 {
		return Result{}, apperr.New("INVALID_ARGUMENT", "no input images provided", 2)
	}

	items := make([]ImageInfo, 0, len(paths))
	for _, path := range paths {
		item, err := inspectFile(path, opts)
		if err != nil {
			return Result{}, err
		}
		items = append(items, item)
	}

	return Result{
		Images:     items,
		Count:      len(items),
		DurationMS: time.Since(start).Milliseconds(),
	}, nil
}

func inspectFile(path string, opts Options) (ImageInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return ImageInfo{}, apperr.Wrap("IO_ERROR", 5, err, "stat %q", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return ImageInfo{}, apperr.Wrap("IO_ERROR", 5, err, "open %q", path)
	}
	defer func() {
		_ = file.Close()
	}()

	cfg, format, err := image.DecodeConfig(file)
	if err != nil {
		return ImageInfo{}, apperr.Wrap("METADATA_EXTRACT_FAILED", 3, err, "decode config %q", path)
	}

	out := ImageInfo{
		Path:        path,
		Format:      strings.ToLower(format),
		Width:       cfg.Width,
		Height:      cfg.Height,
		Orientation: classifyOrientation(cfg.Width, cfg.Height),
		SizeBytes:   info.Size(),
	}

	if exifOrientation, err := readEXIFOrientation(path); err == nil && exifOrientation > 0 {
		out.EXIFOrientation = exifOrientation
	}

	alphaPossible := hasAlphaModel(cfg.ColorModel)
	if !opts.IncludeColors && !opts.IncludeHash && !alphaPossible {
		return out, nil
	}

	img, err := imaging.Open(path, imaging.AutoOrientation(false))
	if err != nil {
		return ImageInfo{}, apperr.Wrap("DECODE_FAILED", 3, err, "decode image %q", path)
	}

	if alphaPossible {
		out.HasAlpha = containsTransparency(img)
	}

	if opts.IncludeColors {
		avg, dom := sampleColors(img)
		out.AverageColor = avg
		out.DominantColor = dom
	}

	if opts.IncludeHash {
		sum, err := fileSHA256(path)
		if err != nil {
			return ImageInfo{}, err
		}
		out.SHA256 = sum
		out.PHash = averageHash(img)
	}

	return out, nil
}

func classifyOrientation(width, height int) string {
	switch {
	case width == height:
		return "square"
	case width > height:
		return "landscape"
	default:
		return "portrait"
	}
}

func hasAlphaModel(model color.Model) bool {
	switch model {
	case color.AlphaModel, color.Alpha16Model, color.RGBAModel, color.RGBA64Model, color.NRGBAModel, color.NRGBA64Model:
		return true
	default:
		return false
	}
}

func readEXIFOrientation(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = f.Close()
	}()
	meta, err := exif.Decode(f)
	if err != nil {
		return 0, err
	}
	tag, err := meta.Get(exif.Orientation)
	if err != nil {
		return 0, err
	}
	val, err := tag.Int(0)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func sampleColors(img image.Image) (string, string) {
	sample := imaging.Resize(img, 64, 0, imaging.Box)
	bounds := sample.Bounds()

	var sumR, sumG, sumB, count uint64
	histogram := make(map[int]int)
	bestKey := 0
	bestCount := -1

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r16, g16, b16, _ := sample.At(x, y).RGBA()
			r := uint8(r16 >> 8)
			g := uint8(g16 >> 8)
			b := uint8(b16 >> 8)
			sumR += uint64(r)
			sumG += uint64(g)
			sumB += uint64(b)
			count++

			key := quantizeColorKey(r, g, b)
			histogram[key]++
			if histogram[key] > bestCount {
				bestKey = key
				bestCount = histogram[key]
			}
		}
	}

	avgR := uint8(sumR / maxUint64(count, 1))
	avgG := uint8(sumG / maxUint64(count, 1))
	avgB := uint8(sumB / maxUint64(count, 1))
	domR, domG, domB := expandColorKey(bestKey)

	return hexColor(avgR, avgG, avgB), hexColor(domR, domG, domB)
}

func quantizeColorKey(r, g, b uint8) int {
	return (int(r>>4) << 8) | (int(g>>4) << 4) | int(b>>4)
}

func expandColorKey(key int) (uint8, uint8, uint8) {
	r := uint8(((key >> 8) & 0xF) * 17)
	g := uint8(((key >> 4) & 0xF) * 17)
	b := uint8((key & 0xF) * 17)
	return r, g, b
}

func hexColor(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", apperr.Wrap("IO_ERROR", 5, err, "open %q", path)
	}
	defer func() {
		_ = f.Close()
	}()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", apperr.Wrap("IO_ERROR", 5, err, "read %q", path)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func averageHash(img image.Image) string {
	resized := imaging.Resize(imaging.Grayscale(img), 8, 8, imaging.Lanczos)
	bounds := resized.Bounds()
	values := make([]uint8, 0, 64)
	var total uint64

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r16, _, _, _ := resized.At(x, y).RGBA()
			v := uint8(r16 >> 8)
			values = append(values, v)
			total += uint64(v)
		}
	}

	avg := float64(total) / math.Max(float64(len(values)), 1)
	var bits uint64
	for i, v := range values {
		if float64(v) >= avg {
			bits |= 1 << (63 - i)
		}
	}
	return fmt.Sprintf("%016x", bits)
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func containsTransparency(img image.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a != 0xffff {
				return true
			}
		}
	}
	return false
}
