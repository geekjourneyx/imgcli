package compose

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/disintegration/imaging"
)

func TestRunPosterLayout(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.jpg")
	logo := filepath.Join(dir, "logo.png")
	output := filepath.Join(dir, "poster.jpg")
	writeJPEG(t, input, 240, 160, color.RGBA{R: 210, G: 160, B: 120, A: 255})
	writePNG(t, logo, 64, 64, color.NRGBA{R: 40, G: 50, B: 60, A: 255})

	result, err := Run(Options{
		Input:       input,
		Output:      output,
		Width:       1080,
		Height:      1440,
		Layout:      LayoutPoster,
		Title:       "Spring Drop",
		Subtitle:    "Limited edition capsule for creators",
		Logo:        logo,
		BannerBadge: "NEW",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Layout != string(LayoutPoster) {
		t.Fatalf("unexpected layout: %+v", result)
	}
	img, err := imaging.Open(output)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	if img.Bounds().Dx() != 1080 || img.Bounds().Dy() != 1440 {
		t.Fatalf("unexpected output size: %v", img.Bounds())
	}
}

func TestRunRejectsInvalidColor(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.jpg")
	writeJPEG(t, input, 120, 80, color.RGBA{R: 255, A: 255})

	_, err := Run(Options{
		Input:           input,
		Output:          filepath.Join(dir, "out.jpg"),
		Width:           600,
		Height:          900,
		BackgroundColor: "bad",
	})
	if err == nil {
		t.Fatal("expected invalid color error")
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
