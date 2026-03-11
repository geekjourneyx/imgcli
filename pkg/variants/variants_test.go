package variants

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/disintegration/imaging"
)

func TestRunPresetSet(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "hero.jpg")
	outputDir := filepath.Join(dir, "dist")
	writeJPEG(t, input, 240, 160, color.RGBA{R: 220, G: 140, B: 90, A: 255})

	result, err := Run(Options{
		Input:     input,
		OutputDir: outputDir,
		PresetSet: "creator-basic",
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Count != 4 {
		t.Fatalf("expected 4 outputs, got %d", result.Count)
	}
	img, err := imaging.Open(filepath.Join(outputDir, "hero_xiaohongshu.jpg"))
	if err != nil {
		t.Fatalf("open variant: %v", err)
	}
	if img.Bounds().Dx() != 1080 || img.Bounds().Dy() != 1440 {
		t.Fatalf("unexpected xiaohongshu size: %v", img.Bounds())
	}
}

func TestValidateTemplateRejectsPaths(t *testing.T) {
	if err := ValidateTemplate("../bad/{preset}.jpg"); err == nil {
		t.Fatal("expected template validation error")
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
