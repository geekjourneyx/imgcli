package smartpad

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestRunSolidBackground(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.jpg")
	output := filepath.Join(dir, "output.jpg")
	writeJPEG(t, input, 80, 40, color.RGBA{R: 200, G: 50, B: 30, A: 255})

	result, err := Run(Options{
		Input:      input,
		Output:     output,
		Target:     image.Point{X: 120, Y: 120},
		Background: BackgroundSolid,
		Quality:    85,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.TargetWidth != 120 || result.TargetHeight != 120 {
		t.Fatalf("unexpected target size: %+v", result)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output file: %v", err)
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
