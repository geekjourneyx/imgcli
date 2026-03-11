package variants

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/presets"
	"github.com/geekjourneyx/imgcli/pkg/smartpad"
)

type Options struct {
	Input            string
	OutputDir        string
	PresetSet        string
	Presets          []string
	Background       smartpad.BackgroundMode
	FilenameTemplate string
	BlurSigma        float64
	Quality          int
}

type Result struct {
	Input      string      `json:"input"`
	OutputDir  string      `json:"output_dir"`
	Generated  []Generated `json:"generated"`
	Count      int         `json:"count"`
	DurationMS int64       `json:"duration_ms"`
}

type Generated struct {
	Preset string `json:"preset"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Path   string `json:"path"`
}

func Run(opts Options) (Result, error) {
	start := time.Now()
	if opts.Input == "" || opts.OutputDir == "" {
		return Result{}, apperr.New("INVALID_ARGUMENT", "--input and --output-dir are required", 2)
	}
	if opts.Quality <= 0 {
		opts.Quality = 85
	}
	selected, err := PlanGenerated(opts)
	if err != nil {
		return Result{}, err
	}
	if err := os.MkdirAll(opts.OutputDir, 0o755); err != nil {
		return Result{}, apperr.Wrap("IO_ERROR", 5, err, "create output dir %q", opts.OutputDir)
	}

	for _, item := range selected {
		preset, err := presets.ByName(item.Preset)
		if err != nil {
			return Result{}, apperr.New("INVALID_ARGUMENT", err.Error(), 2)
		}
		if err := smartpadToPreset(opts, preset, item.Path); err != nil {
			return Result{}, err
		}
	}

	return Result{
		Input:      opts.Input,
		OutputDir:  opts.OutputDir,
		Generated:  selected,
		Count:      len(selected),
		DurationMS: time.Since(start).Milliseconds(),
	}, nil
}

func PlanGenerated(opts Options) ([]Generated, error) {
	if opts.Input == "" || opts.OutputDir == "" {
		return nil, apperr.New("INVALID_ARGUMENT", "--input and --output-dir are required", 2)
	}
	if opts.FilenameTemplate == "" {
		opts.FilenameTemplate = "{base}_{preset}{ext}"
	}
	if err := ValidateTemplate(opts.FilenameTemplate); err != nil {
		return nil, apperr.New("INVALID_ARGUMENT", err.Error(), 2)
	}
	selected, err := presets.Resolve(opts.Presets, opts.PresetSet)
	if err != nil {
		return nil, apperr.New("INVALID_ARGUMENT", err.Error(), 2)
	}

	base := strings.TrimSuffix(filepath.Base(opts.Input), filepath.Ext(opts.Input))
	ext := normalizeExt(filepath.Ext(opts.Input))
	generated := make([]Generated, 0, len(selected))
	for _, preset := range selected {
		filename := renderFilename(opts.FilenameTemplate, base, preset.Name, ext)
		outputPath := filepath.Join(opts.OutputDir, filename)
		generated = append(generated, Generated{
			Preset: preset.Name,
			Width:  preset.Size.X,
			Height: preset.Size.Y,
			Path:   outputPath,
		})
	}
	return generated, nil
}

func smartpadToPreset(opts Options, preset presets.Preset, outputPath string) error {
	_, err := smartpad.Run(smartpad.Options{
		Input:      opts.Input,
		Output:     outputPath,
		Target:     preset.Size,
		Background: opts.Background,
		BlurSigma:  opts.BlurSigma,
		Quality:    opts.Quality,
	})
	return err
}

func renderFilename(template, base, preset, ext string) string {
	name := strings.ReplaceAll(template, "{base}", base)
	name = strings.ReplaceAll(name, "{preset}", preset)
	name = strings.ReplaceAll(name, "{ext}", ext)
	if filepath.Ext(name) == "" {
		name += ext
	}
	return name
}

func normalizeExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return ".jpg"
	case ".png":
		return ".png"
	default:
		return ".jpg"
	}
}

func ValidateTemplate(template string) error {
	if strings.Contains(template, "/") || strings.Contains(template, string(filepath.Separator)) {
		return fmt.Errorf("filename template must not include path separators")
	}
	return nil
}
