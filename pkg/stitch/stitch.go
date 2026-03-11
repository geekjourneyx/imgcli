package stitch

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
	"github.com/geekjourneyx/imgcli/pkg/topdf"
)

const DefaultPartHeightLimit = 65535

type Options struct {
	Inputs          []string
	InputDir        string
	Output          string
	Width           int
	Quality         int
	PartHeightLimit int
	Background      color.Color
}

type Result struct {
	Outputs    []string `json:"outputs"`
	InputCount int      `json:"input_count"`
	Width      int      `json:"width"`
	PartCount  int      `json:"part_count"`
	DurationMS int64    `json:"duration_ms"`
}

type itemMeta struct {
	Path   string
	Height int
}

func Run(opts Options) (Result, error) {
	start := time.Now()
	if opts.Width <= 0 {
		return Result{}, apperr.New("INVALID_ARGUMENT", "width must be positive", 2)
	}
	if opts.Quality <= 0 {
		opts.Quality = 85
	}
	if opts.PartHeightLimit <= 0 {
		opts.PartHeightLimit = DefaultPartHeightLimit
	}
	if opts.Background == nil {
		opts.Background = color.White
	}

	inputs, err := topdf.CollectInputs(opts.Inputs, opts.InputDir)
	if err != nil {
		return Result{}, err
	}
	if len(inputs) == 0 {
		return Result{}, apperr.New("INVALID_ARGUMENT", "no input images provided", 2)
	}

	metas := make([]itemMeta, 0, len(inputs))
	for _, path := range inputs {
		img, err := ioimg.Open(path)
		if err != nil {
			return Result{}, err
		}
		bounds := img.Bounds()
		height := int(float64(bounds.Dy()) * (float64(opts.Width) / float64(bounds.Dx())))
		if height <= 0 {
			return Result{}, apperr.New("INVALID_ARGUMENT", fmt.Sprintf("invalid resize result for %q", path), 2)
		}
		if height > opts.PartHeightLimit {
			return Result{}, apperr.New("CANVAS_TOO_LARGE", fmt.Sprintf("single image exceeds part height limit after resize: %q", path), 2)
		}
		metas = append(metas, itemMeta{Path: path, Height: height})
	}

	parts := partition(metas, opts.PartHeightLimit)
	outputs := make([]string, 0, len(parts))
	for idx, part := range parts {
		outputPath := opts.Output
		if len(parts) > 1 {
			outputPath = numberedOutput(opts.Output, idx+1)
		}
		if err := renderPart(outputPath, opts.Width, opts.Quality, opts.Background, part); err != nil {
			return Result{}, err
		}
		outputs = append(outputs, outputPath)
	}

	return Result{
		Outputs:    outputs,
		InputCount: len(inputs),
		Width:      opts.Width,
		PartCount:  len(outputs),
		DurationMS: time.Since(start).Milliseconds(),
	}, nil
}

func partition(metas []itemMeta, limit int) [][]itemMeta {
	var parts [][]itemMeta
	var current []itemMeta
	currentHeight := 0
	for _, meta := range metas {
		if currentHeight+meta.Height > limit && len(current) > 0 {
			parts = append(parts, current)
			current = nil
			currentHeight = 0
		}
		current = append(current, meta)
		currentHeight += meta.Height
	}
	if len(current) > 0 {
		parts = append(parts, current)
	}
	return parts
}

func renderPart(path string, width, quality int, bg color.Color, items []itemMeta) error {
	totalHeight := 0
	for _, item := range items {
		totalHeight += item.Height
	}
	canvas := image.NewRGBA(image.Rect(0, 0, width, totalHeight))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	offsetY := 0
	for _, item := range items {
		img, err := ioimg.Open(item.Path)
		if err != nil {
			return err
		}
		resized := imaging.Resize(img, width, 0, imaging.Lanczos)
		rect := image.Rect(0, offsetY, width, offsetY+resized.Bounds().Dy())
		draw.Draw(canvas, rect, resized, image.Point{}, draw.Src)
		offsetY += resized.Bounds().Dy()
	}

	return ioimg.Save(path, canvas, quality)
}

func numberedOutput(base string, part int) string {
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return fmt.Sprintf("%s_part%02d%s", name, part, ext)
}
