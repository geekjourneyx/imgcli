package runbook

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestRunDryRunJSON(t *testing.T) {
	dir := t.TempDir()
	hero := filepath.Join(dir, "hero.jpg")
	logo := filepath.Join(dir, "logo.png")
	writeJPEG(t, hero, 1200, 900, color.RGBA{R: 180, G: 90, B: 50, A: 255})
	writePNG(t, logo, 200, 200, color.NRGBA{R: 20, G: 20, B: 220, A: 220})

	recipePath := filepath.Join(dir, "recipe.json")
	recipe := fmt.Sprintf(`{
  "version": "v1",
  "inputs": {
    "hero": %q,
    "logo": %q
  },
  "steps": [
    {
      "id": "card",
      "type": "compose",
      "input": "input:hero",
      "output": %q,
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
      "output": %q,
      "max_width": 720,
      "max_height": 720,
      "quality": 80,
      "strip_metadata": true
    },
    {
      "id": "social",
      "type": "variants",
      "input": "step:web",
      "output_dir": %q,
      "preset_set": "creator-basic"
    }
  ]
}`, hero, logo, filepath.Join(dir, "card.jpg"), filepath.Join(dir, "web.jpg"), filepath.Join(dir, "variants"))
	if err := os.WriteFile(recipePath, []byte(recipe), 0o644); err != nil {
		t.Fatalf("write recipe: %v", err)
	}

	result, err := Run(Options{RecipePath: recipePath, DryRun: true})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !result.DryRun {
		t.Fatalf("expected dry run result")
	}
	if len(result.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(result.Steps))
	}
	if len(result.Steps[2].Outputs) != 4 {
		t.Fatalf("expected 4 planned variant outputs, got %d", len(result.Steps[2].Outputs))
	}
}

func TestRunExecuteYAML(t *testing.T) {
	dir := t.TempDir()
	hero := filepath.Join(dir, "hero.jpg")
	logo := filepath.Join(dir, "logo.png")
	writeJPEG(t, hero, 900, 1200, color.RGBA{R: 210, G: 140, B: 80, A: 255})
	writePNG(t, logo, 160, 160, color.NRGBA{R: 255, G: 255, B: 255, A: 180})

	recipePath := filepath.Join(dir, "recipe.yaml")
	recipe := fmt.Sprintf(`version: v1
inputs:
  hero: %q
  logo: %q
steps:
  - id: inspect-source
    type: inspect
    inputs:
      - input:hero
    include_hash: true
  - id: card
    type: compose
    input: input:hero
    output: %q
    width: 1080
    height: 1440
    layout: poster
    title: "Launch Day"
    subtitle: "Recipe driven"
    logo: input:logo
  - id: web
    type: convert
    input: step:card
    output: %q
    max_width: 720
    max_height: 720
    quality: 78
    strip_metadata: true
  - id: social
    type: variants
    input: step:web
    output_dir: %q
    preset_set: creator-basic
`, hero, logo, filepath.Join(dir, "card.jpg"), filepath.Join(dir, "web.jpg"), filepath.Join(dir, "variants"))
	if err := os.WriteFile(recipePath, []byte(recipe), 0o644); err != nil {
		t.Fatalf("write recipe: %v", err)
	}

	result, err := Run(Options{RecipePath: recipePath})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.DryRun {
		t.Fatalf("expected execution result")
	}
	if len(result.Steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(result.Steps))
	}
	for _, path := range []string{
		filepath.Join(dir, "card.jpg"),
		filepath.Join(dir, "web.jpg"),
		filepath.Join(dir, "variants", "web_xiaohongshu.jpg"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected output %q: %v", path, err)
		}
	}
}

func TestRunRejectsOutputConflict(t *testing.T) {
	dir := t.TempDir()
	hero := filepath.Join(dir, "hero.jpg")
	writeJPEG(t, hero, 300, 200, color.RGBA{R: 220, A: 255})

	recipePath := filepath.Join(dir, "conflict.json")
	recipe := fmt.Sprintf(`{
  "version": "v1",
  "inputs": {"hero": %q},
  "steps": [
    {"id": "one", "type": "convert", "input": "input:hero", "output": %q},
    {"id": "two", "type": "convert", "input": "input:hero", "output": %q}
  ]
}`, hero, filepath.Join(dir, "dup.jpg"), filepath.Join(dir, "dup.jpg"))
	if err := os.WriteFile(recipePath, []byte(recipe), 0o644); err != nil {
		t.Fatalf("write recipe: %v", err)
	}

	_, err := Run(Options{RecipePath: recipePath, DryRun: true})
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if got := err.Error(); got == "" {
		t.Fatal("expected error message")
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
		t.Fatalf("create jpeg: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("close jpeg: %v", err)
		}
	}()
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
