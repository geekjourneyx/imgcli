package stitch

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestRunSplitsIntoParts(t *testing.T) {
	dir := t.TempDir()
	input1 := filepath.Join(dir, "a.jpg")
	input2 := filepath.Join(dir, "b.jpg")
	output := filepath.Join(dir, "out.jpg")
	writeJPEG(t, input1, 100, 80, color.RGBA{R: 255, A: 255})
	writeJPEG(t, input2, 100, 80, color.RGBA{G: 255, A: 255})

	result, err := Run(Options{
		Inputs:          []string{input1, input2},
		Output:          output,
		Width:           100,
		Quality:         85,
		PartHeightLimit: 100,
		Background:      color.White,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.Outputs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(result.Outputs))
	}
	for _, path := range result.Outputs {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected output file %q: %v", path, err)
		}
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
		t.Fatalf("encode image: %v", err)
	}
}
