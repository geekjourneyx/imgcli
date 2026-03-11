package convert

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/disintegration/imaging"
)

func TestRunFlattenAndResizeToJPEG(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "alpha.png")
	output := filepath.Join(dir, "out.jpg")
	writePNG(t, input, 400, 200, color.NRGBA{R: 12, G: 140, B: 220, A: 180})

	result, err := Run(Options{
		Input:             input,
		Output:            output,
		Quality:           82,
		StripMetadata:     true,
		FlattenBackground: "#ffffff",
		MaxWidth:          100,
		MaxHeight:         100,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.OutputFormat != "jpeg" {
		t.Fatalf("expected jpeg output format, got %q", result.OutputFormat)
	}
	if !result.Resized {
		t.Fatalf("expected resized result")
	}
	if result.Width != 100 || result.Height != 50 {
		t.Fatalf("unexpected output size: %dx%d", result.Width, result.Height)
	}
	if result.FlattenBackground != "#ffffff" {
		t.Fatalf("unexpected flatten background: %q", result.FlattenBackground)
	}
	img, err := imaging.Open(output)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	if img.Bounds().Dx() != 100 || img.Bounds().Dy() != 50 {
		t.Fatalf("unexpected saved size: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestRunTransparentJPEGRequiresFlatten(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "alpha.png")
	output := filepath.Join(dir, "out.jpg")
	writePNG(t, input, 40, 20, color.NRGBA{R: 255, A: 120})

	_, err := Run(Options{Input: input, Output: output})
	if err == nil {
		t.Fatal("expected error")
	}
	appErr := err.Error()
	if appErr == "" {
		t.Fatal("expected error message")
	}
}

func TestRunNoUpscaleWhenAlreadyWithinLimits(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "small.png")
	output := filepath.Join(dir, "small.png")
	writePNG(t, input, 80, 40, color.NRGBA{G: 200, A: 255})

	result, err := Run(Options{
		Input:     input,
		Output:    output,
		MaxWidth:  200,
		MaxHeight: 200,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Resized {
		t.Fatalf("expected no resize")
	}
	if result.Width != 80 || result.Height != 40 {
		t.Fatalf("unexpected size: %dx%d", result.Width, result.Height)
	}
}

func writePNG(t *testing.T, path string, width, height int, c color.Color) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("close png: %v", err)
		}
	}()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}
