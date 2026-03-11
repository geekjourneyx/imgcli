package e2e

import (
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

type successEnvelope struct {
	OK      bool            `json:"ok"`
	Command string          `json:"command"`
	Data    json.RawMessage `json:"data"`
}

func TestCLICommands(t *testing.T) {
	dir := t.TempDir()
	input1 := filepath.Join(dir, "img1.jpg")
	input2 := filepath.Join(dir, "img2.jpg")
	logo := filepath.Join(dir, "logo.png")
	writeJPEG(t, input1, 120, 80, color.RGBA{R: 255, A: 255})
	writeJPEG(t, input2, 120, 80, color.RGBA{G: 255, A: 255})
	writeJPEG(t, logo, 60, 60, color.RGBA{B: 255, A: 255})

	runJSONCommand(t, "inspect", "--input", input1, "--input", input2, "--hash", "--color-stats")

	composeOut := filepath.Join(dir, "compose.jpg")
	runJSONCommand(t, "compose", "--input", input1, "--output", composeOut, "--width", "1080", "--height", "1440", "--layout", "poster", "--title", "Launch Day", "--subtitle", "A creator card", "--logo", logo, "--badge", "NEW")

	alphaPNG := filepath.Join(dir, "alpha.png")
	writePNG(t, alphaPNG, 90, 60, color.NRGBA{R: 200, G: 10, B: 80, A: 180})
	convertOut := filepath.Join(dir, "convert.jpg")
	runJSONCommand(t, "convert", "--input", alphaPNG, "--output", convertOut, "--flatten-background", "#ffffff", "--max-width", "45", "--quality", "80", "--strip-metadata")

	recipePath := filepath.Join(dir, "recipe.json")
	recipe := `{
  "version": "v1",
  "inputs": {
    "hero": "` + input1 + `",
    "logo": "` + logo + `"
  },
  "steps": [
    {
      "id": "card",
      "type": "compose",
      "input": "input:hero",
      "output": "` + filepath.Join(dir, "run-card.jpg") + `",
      "width": 1080,
      "height": 1440,
      "layout": "poster",
      "title": "Launch Day",
      "logo": "input:logo"
    },
    {
      "id": "web",
      "type": "convert",
      "input": "step:card",
      "output": "` + filepath.Join(dir, "run-web.jpg") + `",
      "max_width": 720,
      "max_height": 720,
      "quality": 80,
      "strip_metadata": true
    }
  ]
}`
	if err := os.WriteFile(recipePath, []byte(recipe), 0o644); err != nil {
		t.Fatalf("write recipe: %v", err)
	}
	runJSONCommand(t, "run", "--recipe", recipePath, "--dry-run")
	runJSONCommand(t, "run", "--recipe", recipePath)

	variantsDir := filepath.Join(dir, "variants")
	runJSONCommand(t, "variants", "--input", composeOut, "--output-dir", variantsDir, "--preset-set", "creator-basic")

	smartpadOut := filepath.Join(dir, "smartpad.jpg")
	runJSONCommand(t, "smartpad", "--input", input1, "--output", smartpadOut, "--preset", "xiaohongshu")

	pdfOut := filepath.Join(dir, "bundle.pdf")
	runJSONCommand(t, "topdf", "--input", input1, "--input", input2, "--output", pdfOut, "--watermark-text", "demo")

	stitchOut := filepath.Join(dir, "stitched.jpg")
	runJSONCommand(t, "stitch", "--input", input1, "--input", input2, "--output", stitchOut, "--width", "120")
}

func runJSONCommand(t *testing.T, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"run", "."}, args...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, string(out))
	}
	var envelope successEnvelope
	if err := json.Unmarshal(out, &envelope); err != nil {
		t.Fatalf("invalid json output: %v\n%s", err, string(out))
	}
	if !envelope.OK {
		t.Fatalf("expected ok response: %s", string(out))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), ".."))
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
		t.Fatalf("encode image: %v", err)
	}
}
