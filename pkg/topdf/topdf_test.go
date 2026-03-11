package topdf

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCollectInputsNaturalSort(t *testing.T) {
	dir := t.TempDir()
	writeJPEG(t, filepath.Join(dir, "page10.jpg"), 10, 10, color.RGBA{R: 255, A: 255})
	writeJPEG(t, filepath.Join(dir, "page2.jpg"), 10, 10, color.RGBA{G: 255, A: 255})
	writeJPEG(t, filepath.Join(dir, "page1.jpg"), 10, 10, color.RGBA{B: 255, A: 255})

	got, err := CollectInputs(nil, dir)
	if err != nil {
		t.Fatalf("CollectInputs returned error: %v", err)
	}

	want := []string{
		filepath.Join(dir, "page1.jpg"),
		filepath.Join(dir, "page2.jpg"),
		filepath.Join(dir, "page10.jpg"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected order\nwant=%v\ngot=%v", want, got)
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
