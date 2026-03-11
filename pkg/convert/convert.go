package convert

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/ioimg"
)

type Options struct {
	Input             string
	Output            string
	Quality           int
	StripMetadata     bool
	FlattenBackground string
	MaxWidth          int
	MaxHeight         int
}

type Result struct {
	Input             string `json:"input"`
	Output            string `json:"output"`
	SourceFormat      string `json:"source_format"`
	OutputFormat      string `json:"output_format"`
	SourceWidth       int    `json:"source_width"`
	SourceHeight      int    `json:"source_height"`
	Width             int    `json:"width"`
	Height            int    `json:"height"`
	Quality           int    `json:"quality,omitempty"`
	StripMetadata     bool   `json:"strip_metadata"`
	FlattenBackground string `json:"flatten_background,omitempty"`
	Resized           bool   `json:"resized"`
	DurationMS        int64  `json:"duration_ms"`
}

func Run(opts Options) (Result, error) {
	start := time.Now()
	if err := validate(opts); err != nil {
		return Result{}, err
	}
	if opts.Quality <= 0 {
		opts.Quality = 85
	}

	src, err := ioimg.Open(opts.Input)
	if err != nil {
		return Result{}, err
	}
	sourceBounds := src.Bounds()
	sourceFormat := formatFromPath(opts.Input)
	outputFormat := formatFromPath(opts.Output)

	if outputFormat == "jpeg" && hasTransparency(src) && opts.FlattenBackground == "" {
		return Result{}, apperr.New("INVALID_ARGUMENT", "JPEG output for transparent images requires --flatten-background", 2)
	}

	work := src
	if opts.MaxWidth > 0 || opts.MaxHeight > 0 {
		work = resizeWithinLimits(work, opts.MaxWidth, opts.MaxHeight)
	}
	resized := work.Bounds().Dx() != sourceBounds.Dx() || work.Bounds().Dy() != sourceBounds.Dy()

	flattenBackground := ""
	if opts.FlattenBackground != "" {
		bg, _ := parseHexColor(opts.FlattenBackground)
		work = flattenOntoColor(work, bg)
		flattenBackground = normalizeHex(opts.FlattenBackground)
	}

	if err := ioimg.Save(opts.Output, work, opts.Quality); err != nil {
		return Result{}, err
	}

	return Result{
		Input:             opts.Input,
		Output:            opts.Output,
		SourceFormat:      sourceFormat,
		OutputFormat:      outputFormat,
		SourceWidth:       sourceBounds.Dx(),
		SourceHeight:      sourceBounds.Dy(),
		Width:             work.Bounds().Dx(),
		Height:            work.Bounds().Dy(),
		Quality:           qualityForFormat(outputFormat, opts.Quality),
		StripMetadata:     opts.StripMetadata,
		FlattenBackground: flattenBackground,
		Resized:           resized,
		DurationMS:        time.Since(start).Milliseconds(),
	}, nil
}

func validate(opts Options) error {
	if opts.Input == "" || opts.Output == "" {
		return apperr.New("INVALID_ARGUMENT", "--input and --output are required", 2)
	}
	if opts.Quality < 0 || opts.Quality > 100 {
		return apperr.New("INVALID_ARGUMENT", "--quality must be between 1 and 100", 2)
	}
	if opts.MaxWidth < 0 || opts.MaxHeight < 0 {
		return apperr.New("INVALID_ARGUMENT", "--max-width and --max-height must be non-negative", 2)
	}
	if _, err := parseHexColor(opts.FlattenBackground); opts.FlattenBackground != "" && err != nil {
		return apperr.New("INVALID_ARGUMENT", "--flatten-background must be a 6-digit hex color", 2)
	}
	switch formatFromPath(opts.Output) {
	case "jpeg", "png":
		return nil
	default:
		return apperr.New("UNSUPPORTED_FORMAT", fmt.Sprintf("unsupported output extension for %q", opts.Output), 2)
	}
}

func formatFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	default:
		return "unknown"
	}
}

func qualityForFormat(format string, quality int) int {
	if format != "jpeg" {
		return 0
	}
	return quality
}

func resizeWithinLimits(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	limitWidth := maxWidth
	limitHeight := maxHeight
	if limitWidth <= 0 {
		limitWidth = width
	}
	if limitHeight <= 0 {
		limitHeight = height
	}
	if width <= limitWidth && height <= limitHeight {
		return src
	}
	return imaging.Fit(src, limitWidth, limitHeight, imaging.Lanczos)
}

func flattenOntoColor(src image.Image, bg color.Color) image.Image {
	bounds := src.Bounds()
	canvas := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
	draw.Draw(canvas, canvas.Bounds(), src, bounds.Min, draw.Over)
	return canvas
}

func hasTransparency(src image.Image) bool {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := src.At(x, y).RGBA()
			if a != 0xffff {
				return true
			}
		}
	}
	return false
}

func parseHexColor(raw string) (color.Color, error) {
	raw = normalizeHex(raw)
	if len(raw) != 7 || raw[0] != '#' {
		return nil, fmt.Errorf("invalid hex color")
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(raw, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return nil, fmt.Errorf("invalid hex color")
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

func normalizeHex(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if raw[0] != '#' {
		raw = "#" + raw
	}
	return strings.ToLower(raw)
}
