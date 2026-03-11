package smartpad

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/disintegration/imaging"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/ioimg"
)

type BackgroundMode string

const (
	BackgroundBlur  BackgroundMode = "blur"
	BackgroundSolid BackgroundMode = "solid"
)

type Options struct {
	Input      string
	Output     string
	Target     image.Point
	Background BackgroundMode
	BlurSigma  float64
	Quality    int
}

type Result struct {
	Input        string   `json:"input"`
	Output       string   `json:"output"`
	SourceWidth  int      `json:"source_width"`
	SourceHeight int      `json:"source_height"`
	TargetWidth  int      `json:"target_width"`
	TargetHeight int      `json:"target_height"`
	Background   string   `json:"background"`
	DurationMS   int64    `json:"duration_ms"`
	Warnings     []string `json:"warnings,omitempty"`
}

func Run(opts Options) (Result, error) {
	start := time.Now()
	if opts.Target.X <= 0 || opts.Target.Y <= 0 {
		return Result{}, apperr.New("INVALID_ARGUMENT", "target size must be positive", 2)
	}
	if opts.BlurSigma <= 0 {
		opts.BlurSigma = 5
	}
	if opts.Quality <= 0 {
		opts.Quality = 85
	}

	src, err := ioimg.Open(opts.Input)
	if err != nil {
		return Result{}, err
	}

	bounds := src.Bounds()
	bg, err := buildBackground(src, opts)
	if err != nil {
		return Result{}, err
	}
	fg := imaging.Fit(src, opts.Target.X, opts.Target.Y, imaging.Lanczos)
	result := imaging.OverlayCenter(bg, fg, 1.0)
	if err := ioimg.Save(opts.Output, result, opts.Quality); err != nil {
		return Result{}, err
	}

	return Result{
		Input:        opts.Input,
		Output:       opts.Output,
		SourceWidth:  bounds.Dx(),
		SourceHeight: bounds.Dy(),
		TargetWidth:  opts.Target.X,
		TargetHeight: opts.Target.Y,
		Background:   string(opts.Background),
		DurationMS:   time.Since(start).Milliseconds(),
	}, nil
}

func buildBackground(src image.Image, opts Options) (image.Image, error) {
	switch opts.Background {
	case "", BackgroundBlur:
		return buildBlurBackground(src, opts.Target, opts.BlurSigma), nil
	case BackgroundSolid:
		return buildSolidBackground(opts.Target, dominantColor(src)), nil
	default:
		return nil, apperr.New("INVALID_ARGUMENT", fmt.Sprintf("unsupported background mode %q", opts.Background), 2)
	}
}

func buildBlurBackground(src image.Image, target image.Point, sigma float64) image.Image {
	previewWidth := maxInt(target.X/10, 1)
	preview := imaging.Resize(src, previewWidth, 0, imaging.Linear)
	blurred := imaging.Blur(preview, sigma)
	return imaging.Fill(blurred, target.X, target.Y, imaging.Center, imaging.Lanczos)
}

func buildSolidBackground(target image.Point, c color.Color) image.Image {
	return ioimg.FillBackground(image.Rect(0, 0, target.X, target.Y), c)
}

func dominantColor(src image.Image) color.Color {
	sample := imaging.Resize(src, 1, 1, imaging.Box)
	return sample.At(0, 0)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
