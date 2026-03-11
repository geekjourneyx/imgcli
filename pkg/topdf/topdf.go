package topdf

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/go-pdf/fpdf"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/fontutil"
	"github.com/geekjourneyx/imgcli/pkg/ioimg"
)

type WatermarkPosition string

const (
	PositionBottomRight WatermarkPosition = "br"
	PositionCenter      WatermarkPosition = "center"
	PositionTile        WatermarkPosition = "tile"
)

type Options struct {
	Inputs            []string
	InputDir          string
	Output            string
	WatermarkText     string
	WatermarkOpacity  float64
	WatermarkSize     float64
	WatermarkPosition WatermarkPosition
	Quality           int
}

type Result struct {
	Output          string   `json:"output"`
	PageCount       int      `json:"page_count"`
	Inputs          []string `json:"inputs"`
	WatermarkText   string   `json:"watermark_text,omitempty"`
	WatermarkMode   string   `json:"watermark_mode,omitempty"`
	OutputSizeBytes int64    `json:"output_size_bytes"`
	DurationMS      int64    `json:"duration_ms"`
}

func Run(opts Options) (Result, error) {
	start := time.Now()
	inputs, err := CollectInputs(opts.Inputs, opts.InputDir)
	if err != nil {
		return Result{}, err
	}
	if len(inputs) == 0 {
		return Result{}, apperr.New("INVALID_ARGUMENT", "no input images provided", 2)
	}
	if opts.Quality <= 0 {
		opts.Quality = 85
	}
	if opts.WatermarkOpacity <= 0 {
		opts.WatermarkOpacity = 0.25
	}
	if opts.WatermarkSize <= 0 {
		opts.WatermarkSize = 42
	}
	pdf := fpdf.NewCustom(&fpdf.InitType{UnitStr: "pt"})

	for idx, path := range inputs {
		img, err := ioimg.Open(path)
		if err != nil {
			return Result{}, err
		}
		if opts.WatermarkText != "" {
			img, err = applyWatermark(img, opts.WatermarkText, opts.WatermarkOpacity, opts.WatermarkSize, opts.WatermarkPosition)
			if err != nil {
				return Result{}, err
			}
		}

		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: opts.Quality}); err != nil {
			return Result{}, apperr.Wrap("ENCODE_FAILED", 4, err, "encode page %q", path)
		}

		bounds := img.Bounds()
		orientation := "P"
		if bounds.Dx() > bounds.Dy() {
			orientation = "L"
		}
		size := fpdf.SizeType{Wd: float64(bounds.Dx()), Ht: float64(bounds.Dy())}
		pdf.AddPageFormat(orientation, size)
		name := fmt.Sprintf("page-%d", idx)
		opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
		pdf.RegisterImageOptionsReader(name, opt, &buf)
		pdf.ImageOptions(name, 0, 0, size.Wd, size.Ht, false, opt, 0, "")
	}

	if err := pdf.OutputFileAndClose(opts.Output); err != nil {
		return Result{}, apperr.Wrap("PDF_WRITE_FAILED", 6, err, "write pdf %q", opts.Output)
	}
	info, err := os.Stat(opts.Output)
	if err != nil {
		return Result{}, apperr.Wrap("IO_ERROR", 5, err, "stat output %q", opts.Output)
	}

	return Result{
		Output:          opts.Output,
		PageCount:       len(inputs),
		Inputs:          inputs,
		WatermarkText:   opts.WatermarkText,
		WatermarkMode:   string(opts.WatermarkPosition),
		OutputSizeBytes: info.Size(),
		DurationMS:      time.Since(start).Milliseconds(),
	}, nil
}

func CollectInputs(inputs []string, inputDir string) ([]string, error) {
	if len(inputs) > 0 && inputDir != "" {
		return nil, apperr.New("INVALID_ARGUMENT", "use either --input or --input-dir, not both", 2)
	}
	if len(inputs) > 0 {
		for _, path := range inputs {
			if err := ioimg.EnsureReadableFile(path); err != nil {
				return nil, err
			}
		}
		return inputs, nil
	}
	if inputDir == "" {
		return nil, apperr.New("INVALID_ARGUMENT", "missing --input or --input-dir", 2)
	}
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, apperr.Wrap("IO_ERROR", 5, err, "read dir %q", inputDir)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		switch ext {
		case ".jpg", ".jpeg", ".png":
			files = append(files, filepath.Join(inputDir, entry.Name()))
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return naturalLess(filepath.Base(files[i]), filepath.Base(files[j]))
	})
	return files, nil
}

func naturalLess(a, b string) bool {
	ta := tokenize(a)
	tb := tokenize(b)
	for i := 0; i < len(ta) && i < len(tb); i++ {
		if ta[i] == tb[i] {
			continue
		}
		ad, bd := isDigits(ta[i]), isDigits(tb[i])
		if ad && bd {
			if len(ta[i]) != len(tb[i]) {
				return len(ta[i]) < len(tb[i])
			}
		}
		return ta[i] < tb[i]
	}
	return len(ta) < len(tb)
}

func tokenize(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	runes := []rune(s)
	start := 0
	curDigit := unicode.IsDigit(runes[0])
	for i := 1; i < len(runes); i++ {
		nextDigit := unicode.IsDigit(runes[i])
		if nextDigit == curDigit {
			continue
		}
		out = append(out, string(runes[start:i]))
		start = i
		curDigit = nextDigit
	}
	out = append(out, string(runes[start:]))
	return out
}

func isDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}

func applyWatermark(img image.Image, text string, opacity float64, size float64, position WatermarkPosition) (image.Image, error) {
	dc := gg.NewContextForImage(img)
	face, _, err := fontutil.LoadFace("", size)
	if err != nil {
		return nil, apperr.Wrap("INTERNAL_ERROR", 1, err, "load embedded font")
	}
	dc.SetFontFace(face)
	dc.SetRGBA(1, 1, 1, opacity)
	bounds := img.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	switch position {
	case "", PositionBottomRight:
		dc.DrawStringAnchored(text, w-24, h-24, 1, 1)
	case PositionCenter:
		dc.DrawStringAnchored(text, w/2, h/2, 0.5, 0.5)
	case PositionTile:
		stepX := maxInt(int(size*4), 120)
		stepY := maxInt(int(size*3), 100)
		for y := stepY / 2; y < bounds.Dy()+stepY; y += stepY {
			for x := stepX / 2; x < bounds.Dx()+stepX; x += stepX {
				dc.DrawStringAnchored(text, float64(x), float64(y), 0.5, 0.5)
			}
		}
	default:
		return nil, apperr.New("INVALID_ARGUMENT", fmt.Sprintf("unsupported watermark position %q", position), 2)
	}

	return imaging.Clone(dc.Image()), nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
