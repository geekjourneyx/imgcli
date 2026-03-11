package inspect

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestRunInspectsJPEGAndPNG(t *testing.T) {
	dir := t.TempDir()
	jpegPath := filepath.Join(dir, "img2.jpg")
	pngPath := filepath.Join(dir, "img10.png")

	writeJPEG(t, jpegPath, 120, 80, color.RGBA{R: 210, G: 100, B: 80, A: 255})
	writePNG(t, pngPath, 64, 64, color.NRGBA{R: 10, G: 20, B: 30, A: 120})

	result, err := Run(Options{
		InputDir:      dir,
		IncludeHash:   true,
		IncludeColors: true,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Count != 2 {
		t.Fatalf("unexpected count: %d", result.Count)
	}
	if result.Images[0].Path != jpegPath || result.Images[1].Path != pngPath {
		t.Fatalf("unexpected natural order: %+v", result.Images)
	}
	if result.Images[0].Format != "jpeg" {
		t.Fatalf("unexpected jpeg format: %+v", result.Images[0])
	}
	if result.Images[0].Orientation != "landscape" {
		t.Fatalf("unexpected orientation: %+v", result.Images[0])
	}
	if result.Images[0].SHA256 == "" || result.Images[0].PHash == "" {
		t.Fatalf("expected hash fields: %+v", result.Images[0])
	}
	if result.Images[1].Format != "png" || !result.Images[1].HasAlpha {
		t.Fatalf("expected png alpha: %+v", result.Images[1])
	}
	if result.Images[1].AverageColor == "" || result.Images[1].DominantColor == "" {
		t.Fatalf("expected color stats: %+v", result.Images[1])
	}
}

func TestRunLimit(t *testing.T) {
	dir := t.TempDir()
	writeJPEG(t, filepath.Join(dir, "a1.jpg"), 10, 10, color.RGBA{R: 255, A: 255})
	writeJPEG(t, filepath.Join(dir, "a2.jpg"), 10, 10, color.RGBA{G: 255, A: 255})

	result, err := Run(Options{InputDir: dir, Limit: 1})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Count != 1 {
		t.Fatalf("expected limit to apply, got %d", result.Count)
	}
}

func TestRunDetectsOpaquePNGWithoutColorStats(t *testing.T) {
	dir := t.TempDir()
	opaquePNG := filepath.Join(dir, "opaque.png")
	writePNG(t, opaquePNG, 12, 12, color.NRGBA{R: 50, G: 60, B: 70, A: 255})

	result, err := Run(Options{Inputs: []string{opaquePNG}})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Count != 1 {
		t.Fatalf("unexpected count: %d", result.Count)
	}
	if result.Images[0].HasAlpha {
		t.Fatalf("expected opaque png to report no alpha: %+v", result.Images[0])
	}
}

func writeJPEG(t *testing.T, path string, width, height int, c color.Color) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Fatalf("close image file: %v", err)
		}
	})
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
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
		t.Fatalf("create image: %v", err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Fatalf("close image file: %v", err)
		}
	})
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}
