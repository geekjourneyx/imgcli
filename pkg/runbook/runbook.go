package runbook

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/compose"
	"github.com/geekjourneyx/imgcli/pkg/convert"
	"github.com/geekjourneyx/imgcli/pkg/inspect"
	"github.com/geekjourneyx/imgcli/pkg/ioimg"
	"github.com/geekjourneyx/imgcli/pkg/presets"
	"github.com/geekjourneyx/imgcli/pkg/smartpad"
	"github.com/geekjourneyx/imgcli/pkg/stitch"
	"github.com/geekjourneyx/imgcli/pkg/topdf"
	"github.com/geekjourneyx/imgcli/pkg/variants"
)

type Options struct {
	RecipePath string
	DryRun     bool
}

type Result struct {
	Recipe     string       `json:"recipe"`
	Version    string       `json:"version"`
	DryRun     bool         `json:"dry_run"`
	Steps      []StepResult `json:"steps"`
	DurationMS int64        `json:"duration_ms"`
}

type StepResult struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Inputs    []string `json:"inputs,omitempty"`
	Output    string   `json:"output,omitempty"`
	OutputDir string   `json:"output_dir,omitempty"`
	Outputs   []string `json:"outputs,omitempty"`
	Data      any      `json:"data,omitempty"`
}

type Recipe struct {
	Version string            `json:"version" yaml:"version"`
	Inputs  map[string]string `json:"inputs" yaml:"inputs"`
	Steps   []Step            `json:"steps" yaml:"steps"`
	Outputs map[string]string `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

type Step struct {
	ID                string   `json:"id" yaml:"id"`
	Type              string   `json:"type" yaml:"type"`
	Layout            string   `json:"layout,omitempty" yaml:"layout,omitempty"`
	Input             string   `json:"input,omitempty" yaml:"input,omitempty"`
	Inputs            []string `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	InputDir          string   `json:"input_dir,omitempty" yaml:"input_dir,omitempty"`
	Output            string   `json:"output,omitempty" yaml:"output,omitempty"`
	OutputDir         string   `json:"output_dir,omitempty" yaml:"output_dir,omitempty"`
	FilenameTemplate  string   `json:"filename_template,omitempty" yaml:"filename_template,omitempty"`
	Preset            string   `json:"preset,omitempty" yaml:"preset,omitempty"`
	PresetSet         string   `json:"preset_set,omitempty" yaml:"preset_set,omitempty"`
	Presets           []string `json:"presets,omitempty" yaml:"presets,omitempty"`
	Background        string   `json:"background,omitempty" yaml:"background,omitempty"`
	BlurSigma         float64  `json:"blur_sigma,omitempty" yaml:"blur_sigma,omitempty"`
	Quality           int      `json:"quality,omitempty" yaml:"quality,omitempty"`
	WatermarkText     string   `json:"watermark_text,omitempty" yaml:"watermark_text,omitempty"`
	WatermarkOpacity  float64  `json:"watermark_opacity,omitempty" yaml:"watermark_opacity,omitempty"`
	WatermarkSize     float64  `json:"watermark_size,omitempty" yaml:"watermark_size,omitempty"`
	WatermarkPosition string   `json:"watermark_position,omitempty" yaml:"watermark_position,omitempty"`
	Width             int      `json:"width,omitempty" yaml:"width,omitempty"`
	Height            int      `json:"height,omitempty" yaml:"height,omitempty"`
	BackgroundColor   string   `json:"background_color,omitempty" yaml:"background_color,omitempty"`
	BackgroundImage   string   `json:"background_image,omitempty" yaml:"background_image,omitempty"`
	Title             string   `json:"title,omitempty" yaml:"title,omitempty"`
	Subtitle          string   `json:"subtitle,omitempty" yaml:"subtitle,omitempty"`
	TitleSize         float64  `json:"title_size,omitempty" yaml:"title_size,omitempty"`
	SubtitleSize      float64  `json:"subtitle_size,omitempty" yaml:"subtitle_size,omitempty"`
	TitleColor        string   `json:"title_color,omitempty" yaml:"title_color,omitempty"`
	SubtitleColor     string   `json:"subtitle_color,omitempty" yaml:"subtitle_color,omitempty"`
	Font              string   `json:"font,omitempty" yaml:"font,omitempty"`
	Logo              string   `json:"logo,omitempty" yaml:"logo,omitempty"`
	Badge             string   `json:"badge,omitempty" yaml:"badge,omitempty"`
	Padding           int      `json:"padding,omitempty" yaml:"padding,omitempty"`
	Radius            float64  `json:"radius,omitempty" yaml:"radius,omitempty"`
	SafeArea          string   `json:"safe_area,omitempty" yaml:"safe_area,omitempty"`
	FlattenBackground string   `json:"flatten_background,omitempty" yaml:"flatten_background,omitempty"`
	MaxWidth          int      `json:"max_width,omitempty" yaml:"max_width,omitempty"`
	MaxHeight         int      `json:"max_height,omitempty" yaml:"max_height,omitempty"`
	StripMetadata     bool     `json:"strip_metadata,omitempty" yaml:"strip_metadata,omitempty"`
	IncludeHash       bool     `json:"include_hash,omitempty" yaml:"include_hash,omitempty"`
	IncludeColors     bool     `json:"include_colors,omitempty" yaml:"include_colors,omitempty"`
	Limit             int      `json:"limit,omitempty" yaml:"limit,omitempty"`
	PartHeightLimit   int      `json:"part_height_limit,omitempty" yaml:"part_height_limit,omitempty"`
}

type plannedStep struct {
	step          Step
	resolvedInput string
	resolvedLogo  string
	resolvedBG    string
	resolvedList  []string
	outputs       []string
}

type planContext struct {
	inputs      map[string]string
	stepOutputs map[string][]string
}

func Run(opts Options) (Result, error) {
	start := time.Now()
	if opts.RecipePath == "" {
		return Result{}, apperr.New("INVALID_ARGUMENT", "--recipe is required", 2)
	}

	recipe, err := LoadRecipe(opts.RecipePath)
	if err != nil {
		return Result{}, err
	}
	planned, err := plan(recipe)
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Recipe:  opts.RecipePath,
		Version: recipe.Version,
		DryRun:  opts.DryRun,
		Steps:   make([]StepResult, 0, len(planned)),
	}
	if opts.DryRun {
		for _, item := range planned {
			result.Steps = append(result.Steps, StepResult{
				ID:        item.step.ID,
				Type:      item.step.Type,
				Inputs:    item.planInputs(),
				Output:    item.step.Output,
				OutputDir: item.step.OutputDir,
				Outputs:   append([]string(nil), item.outputs...),
			})
		}
		result.DurationMS = time.Since(start).Milliseconds()
		return result, nil
	}

	for _, item := range planned {
		stepResult, err := executeStep(item)
		if err != nil {
			return Result{}, err
		}
		result.Steps = append(result.Steps, stepResult)
	}
	result.DurationMS = time.Since(start).Milliseconds()
	return result, nil
}

func LoadRecipe(path string) (Recipe, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Recipe{}, apperr.Wrap("IO_ERROR", 5, err, "read recipe %q", path)
	}
	ext := strings.ToLower(filepath.Ext(path))
	var recipe Recipe
	switch ext {
	case ".json":
		dec := json.NewDecoder(strings.NewReader(string(raw)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&recipe); err != nil {
			return Recipe{}, apperr.Wrap("CONFIG_ERROR", 2, err, "decode recipe %q", path)
		}
	case ".yaml", ".yml":
		dec := yaml.NewDecoder(strings.NewReader(string(raw)))
		dec.KnownFields(true)
		if err := dec.Decode(&recipe); err != nil {
			return Recipe{}, apperr.Wrap("CONFIG_ERROR", 2, err, "decode recipe %q", path)
		}
	default:
		return Recipe{}, apperr.New("CONFIG_ERROR", "recipe must use .json, .yaml, or .yml", 2)
	}
	return recipe, validateRecipe(recipe)
}

func validateRecipe(recipe Recipe) error {
	switch recipe.Version {
	case "v1", "1", "1.0":
	default:
		return apperr.New("PLAN_INVALID", "recipe version must be v1", 2)
	}
	if len(recipe.Steps) == 0 {
		return apperr.New("PLAN_INVALID", "recipe must include at least one step", 2)
	}
	for name, path := range recipe.Inputs {
		if name == "" {
			return apperr.New("PLAN_INVALID", "recipe inputs must use non-empty names", 2)
		}
		if err := ioimg.EnsureReadableFile(path); err != nil {
			return err
		}
	}
	return nil
}

func plan(recipe Recipe) ([]plannedStep, error) {
	ctx := planContext{
		inputs:      recipe.Inputs,
		stepOutputs: make(map[string][]string, len(recipe.Steps)),
	}
	inputPaths := make(map[string]string, len(recipe.Inputs))
	for name, path := range recipe.Inputs {
		inputPaths[path] = name
	}
	ids := make(map[string]struct{}, len(recipe.Steps))
	declaredOutputs := make(map[string]string)
	out := make([]plannedStep, 0, len(recipe.Steps))
	for _, step := range recipe.Steps {
		if step.ID == "" {
			return nil, apperr.New("PLAN_INVALID", "every step must include an id", 2)
		}
		if _, exists := ids[step.ID]; exists {
			return nil, apperr.New("PLAN_INVALID", fmt.Sprintf("duplicate step id %q", step.ID), 2)
		}
		ids[step.ID] = struct{}{}

		item, err := planStep(step, ctx)
		if err != nil {
			return nil, err
		}
		for _, output := range item.outputs {
			if inputName, exists := inputPaths[output]; exists {
				return nil, apperr.New("OUTPUT_CONFLICT", fmt.Sprintf("output %q would overwrite recipe input %q", output, inputName), 2)
			}
			if prior, exists := declaredOutputs[output]; exists {
				return nil, apperr.New("OUTPUT_CONFLICT", fmt.Sprintf("output %q is produced by both %q and %q", output, prior, step.ID), 2)
			}
			declaredOutputs[output] = step.ID
		}
		ctx.stepOutputs[step.ID] = append([]string(nil), item.outputs...)
		out = append(out, item)
	}
	return out, nil
}

func planStep(step Step, ctx planContext) (plannedStep, error) {
	item := plannedStep{step: step}
	switch step.Type {
	case "inspect":
		var err error
		item.resolvedList, err = resolveInputSet(step.Input, step.Inputs, step.InputDir, ctx)
		if err != nil {
			return plannedStep{}, err
		}
	case "smartpad":
		var err error
		item.resolvedInput, err = resolveSingle(step.Input, ctx)
		if err != nil {
			return plannedStep{}, err
		}
		if step.Output == "" || step.Preset == "" {
			return plannedStep{}, apperr.New("PLAN_INVALID", "smartpad requires input, output, and preset", 2)
		}
		if _, err := presets.Get(step.Preset); err != nil {
			return plannedStep{}, apperr.New("PRESET_NOT_FOUND", fmt.Sprintf("preset %q not found", step.Preset), 2)
		}
		item.outputs = []string{step.Output}
	case "compose":
		var err error
		item.resolvedInput, err = resolveSingle(step.Input, ctx)
		if err != nil {
			return plannedStep{}, err
		}
		if step.BackgroundImage != "" {
			item.resolvedBG, err = resolveSingle(step.BackgroundImage, ctx)
			if err != nil {
				return plannedStep{}, err
			}
		}
		if step.Logo != "" {
			item.resolvedLogo, err = resolveSingle(step.Logo, ctx)
			if err != nil {
				return plannedStep{}, err
			}
		}
		if step.Output == "" || step.Width <= 0 || step.Height <= 0 {
			return plannedStep{}, apperr.New("PLAN_INVALID", "compose requires input, output, width, and height", 2)
		}
		item.outputs = []string{step.Output}
	case "convert":
		var err error
		item.resolvedInput, err = resolveSingle(step.Input, ctx)
		if err != nil {
			return plannedStep{}, err
		}
		if step.Output == "" {
			return plannedStep{}, apperr.New("PLAN_INVALID", "convert requires input and output", 2)
		}
		item.outputs = []string{step.Output}
	case "variants":
		var err error
		item.resolvedInput, err = resolveSingle(step.Input, ctx)
		if err != nil {
			return plannedStep{}, err
		}
		generated, err := variants.PlanGenerated(variants.Options{
			Input:            item.resolvedInput,
			OutputDir:        step.OutputDir,
			PresetSet:        step.PresetSet,
			Presets:          step.Presets,
			Background:       smartpad.BackgroundMode(step.Background),
			FilenameTemplate: step.outputTemplate(),
			BlurSigma:        step.BlurSigma,
			Quality:          step.Quality,
		})
		if err != nil {
			return plannedStep{}, err
		}
		item.outputs = make([]string, 0, len(generated))
		for _, generatedItem := range generated {
			item.outputs = append(item.outputs, generatedItem.Path)
		}
	case "topdf":
		var err error
		item.resolvedList, err = resolveInputSet(step.Input, step.Inputs, step.InputDir, ctx)
		if err != nil {
			return plannedStep{}, err
		}
		if step.Output == "" {
			return plannedStep{}, apperr.New("PLAN_INVALID", "topdf requires output", 2)
		}
		item.outputs = []string{step.Output}
	case "stitch":
		var err error
		item.resolvedList, err = resolveInputSet(step.Input, step.Inputs, step.InputDir, ctx)
		if err != nil {
			return plannedStep{}, err
		}
		if step.Output == "" || step.Width <= 0 {
			return plannedStep{}, apperr.New("PLAN_INVALID", "stitch requires output and width", 2)
		}
		item.outputs = []string{step.Output}
	default:
		return plannedStep{}, apperr.New("PLAN_INVALID", fmt.Sprintf("unsupported step type %q", step.Type), 2)
	}
	return item, nil
}

func executeStep(item plannedStep) (StepResult, error) {
	switch item.step.Type {
	case "inspect":
		result, err := inspect.Run(inspect.Options{
			Inputs:        item.resolvedList,
			IncludeHash:   item.step.IncludeHash,
			IncludeColors: item.step.IncludeColors,
			Limit:         item.step.Limit,
		})
		if err != nil {
			return StepResult{}, err
		}
		return StepResult{ID: item.step.ID, Type: item.step.Type, Inputs: item.resolvedList, Data: result}, nil
	case "smartpad":
		target, err := presets.Get(item.step.Preset)
		if err != nil {
			return StepResult{}, apperr.New("PRESET_NOT_FOUND", fmt.Sprintf("preset %q not found", item.step.Preset), 2)
		}
		result, err := smartpad.Run(smartpad.Options{
			Input:      item.resolvedInput,
			Output:     item.step.Output,
			Target:     target,
			Background: smartpad.BackgroundMode(item.step.Background),
			BlurSigma:  item.step.BlurSigma,
			Quality:    item.step.Quality,
		})
		if err != nil {
			return StepResult{}, err
		}
		return StepResult{ID: item.step.ID, Type: item.step.Type, Inputs: []string{item.resolvedInput}, Output: item.step.Output, Outputs: []string{item.step.Output}, Data: result}, nil
	case "compose":
		result, err := compose.Run(compose.Options{
			Input:           item.resolvedInput,
			Output:          item.step.Output,
			Width:           item.step.Width,
			Height:          item.step.Height,
			BackgroundColor: item.step.BackgroundColor,
			BackgroundImage: item.resolvedBG,
			Title:           item.step.Title,
			Subtitle:        item.step.Subtitle,
			TitleSize:       item.step.TitleSize,
			SubtitleSize:    item.step.SubtitleSize,
			TitleColor:      item.step.TitleColor,
			SubtitleColor:   item.step.SubtitleColor,
			FontPath:        item.step.Font,
			Logo:            item.resolvedLogo,
			BannerBadge:     item.step.Badge,
			Padding:         item.step.Padding,
			Radius:          item.step.Radius,
			SafeArea:        item.step.SafeArea,
			Layout:          compose.Layout(item.step.layoutOrDefault()),
			Quality:         item.step.Quality,
		})
		if err != nil {
			return StepResult{}, err
		}
		return StepResult{ID: item.step.ID, Type: item.step.Type, Inputs: item.resolvedInputs(), Output: item.step.Output, Outputs: []string{item.step.Output}, Data: result}, nil
	case "convert":
		result, err := convert.Run(convert.Options{
			Input:             item.resolvedInput,
			Output:            item.step.Output,
			Quality:           item.step.Quality,
			StripMetadata:     item.step.StripMetadata,
			FlattenBackground: item.step.FlattenBackground,
			MaxWidth:          item.step.MaxWidth,
			MaxHeight:         item.step.MaxHeight,
		})
		if err != nil {
			return StepResult{}, err
		}
		return StepResult{ID: item.step.ID, Type: item.step.Type, Inputs: []string{item.resolvedInput}, Output: item.step.Output, Outputs: []string{item.step.Output}, Data: result}, nil
	case "variants":
		result, err := variants.Run(variants.Options{
			Input:            item.resolvedInput,
			OutputDir:        item.step.OutputDir,
			PresetSet:        item.step.PresetSet,
			Presets:          item.step.Presets,
			Background:       smartpad.BackgroundMode(item.step.Background),
			FilenameTemplate: item.step.outputTemplate(),
			BlurSigma:        item.step.BlurSigma,
			Quality:          item.step.Quality,
		})
		if err != nil {
			return StepResult{}, err
		}
		return StepResult{ID: item.step.ID, Type: item.step.Type, Inputs: []string{item.resolvedInput}, OutputDir: item.step.OutputDir, Outputs: item.outputs, Data: result}, nil
	case "topdf":
		result, err := topdf.Run(topdf.Options{
			Inputs:            item.resolvedList,
			Output:            item.step.Output,
			WatermarkText:     item.step.WatermarkText,
			WatermarkOpacity:  item.step.WatermarkOpacity,
			WatermarkSize:     item.step.WatermarkSize,
			WatermarkPosition: topdf.WatermarkPosition(item.step.WatermarkPosition),
			Quality:           item.step.Quality,
		})
		if err != nil {
			return StepResult{}, err
		}
		return StepResult{ID: item.step.ID, Type: item.step.Type, Inputs: item.resolvedList, Output: item.step.Output, Outputs: []string{item.step.Output}, Data: result}, nil
	case "stitch":
		result, err := stitch.Run(stitch.Options{
			Inputs:          item.resolvedList,
			Output:          item.step.Output,
			Width:           item.step.Width,
			Quality:         item.step.Quality,
			PartHeightLimit: item.step.PartHeightLimit,
			Background:      color.White,
		})
		if err != nil {
			return StepResult{}, err
		}
		return StepResult{ID: item.step.ID, Type: item.step.Type, Inputs: item.resolvedList, Output: item.step.Output, Outputs: result.Outputs, Data: result}, nil
	default:
		return StepResult{}, apperr.New("PLAN_INVALID", fmt.Sprintf("unsupported step type %q", item.step.Type), 2)
	}
}

func resolveSingle(raw string, ctx planContext) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", apperr.New("PLAN_INVALID", "missing input reference", 2)
	}
	paths, err := resolveMany([]string{raw}, ctx)
	if err != nil {
		return "", err
	}
	if len(paths) != 1 {
		return "", apperr.New("PLAN_INVALID", fmt.Sprintf("reference %q resolved to %d outputs; expected exactly one", raw, len(paths)), 2)
	}
	return paths[0], nil
}

func resolveInputSet(single string, many []string, inputDir string, ctx planContext) ([]string, error) {
	if strings.TrimSpace(single) != "" && len(many) > 0 {
		return nil, apperr.New("PLAN_INVALID", "use either input or inputs, not both", 2)
	}
	if strings.TrimSpace(single) != "" && inputDir != "" {
		return nil, apperr.New("PLAN_INVALID", "use either input or input_dir, not both", 2)
	}
	if len(many) > 0 && inputDir != "" {
		return nil, apperr.New("PLAN_INVALID", "use either inputs or input_dir, not both", 2)
	}
	if strings.TrimSpace(single) != "" {
		return resolveMany([]string{single}, ctx)
	}
	if len(many) > 0 {
		return resolveMany(many, ctx)
	}
	if inputDir == "" {
		return nil, apperr.New("PLAN_INVALID", "step requires input, inputs, or input_dir", 2)
	}
	return topdf.CollectInputs(nil, inputDir)
}

func resolveMany(raw []string, ctx planContext) ([]string, error) {
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		switch {
		case strings.HasPrefix(item, "input:"):
			name := strings.TrimPrefix(item, "input:")
			path, ok := ctx.inputs[name]
			if !ok {
				return nil, apperr.New("PLAN_INVALID", fmt.Sprintf("unknown recipe input %q", name), 2)
			}
			out = append(out, path)
		case strings.HasPrefix(item, "step:"):
			name := strings.TrimPrefix(item, "step:")
			paths, ok := ctx.stepOutputs[name]
			if !ok {
				return nil, apperr.New("PLAN_INVALID", fmt.Sprintf("unknown step reference %q", name), 2)
			}
			if len(paths) == 0 {
				return nil, apperr.New("PLAN_INVALID", fmt.Sprintf("step %q does not produce file outputs", name), 2)
			}
			out = append(out, paths...)
		default:
			if err := ioimg.EnsureReadableFile(item); err != nil {
				return nil, err
			}
			out = append(out, item)
		}
	}
	return out, nil
}

func (s Step) outputTemplate() string {
	return s.FilenameTemplate
}

func (s Step) layoutOrDefault() string {
	if strings.TrimSpace(s.Layout) == "" {
		return string(compose.LayoutPoster)
	}
	return strings.TrimSpace(s.Layout)
}

func (p plannedStep) resolvedInputs() []string {
	out := make([]string, 0, 3)
	if p.resolvedInput != "" {
		out = append(out, p.resolvedInput)
	}
	if p.resolvedBG != "" {
		out = append(out, p.resolvedBG)
	}
	if p.resolvedLogo != "" {
		out = append(out, p.resolvedLogo)
	}
	return out
}

func (p plannedStep) planInputs() []string {
	if len(p.resolvedList) > 0 {
		return append([]string(nil), p.resolvedList...)
	}
	return p.resolvedInputs()
}
