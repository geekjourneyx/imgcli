package presets

import (
	"fmt"
	"image"
	"sort"
)

type Preset struct {
	Name string
	Size image.Point
}

var builtins = map[string]Preset{
	"banner_16x9":    {Name: "banner_16x9", Size: image.Point{X: 1600, Y: 900}},
	"detail_long":    {Name: "detail_long", Size: image.Point{X: 1080, Y: 2160}},
	"product_square": {Name: "product_square", Size: image.Point{X: 1200, Y: 1200}},
	"square":         {Name: "square", Size: image.Point{X: 1080, Y: 1080}},
	"story_9x16":     {Name: "story_9x16", Size: image.Point{X: 1080, Y: 1920}},
	"wechat_cover":   {Name: "wechat_cover", Size: image.Point{X: 900, Y: 383}},
	"xiaohongshu":    {Name: "xiaohongshu", Size: image.Point{X: 1080, Y: 1440}},
}

var presetSets = map[string][]string{
	"creator-basic":   {"xiaohongshu", "wechat_cover", "square", "story_9x16"},
	"ecommerce-basic": {"product_square", "detail_long", "banner_16x9"},
}

func Get(name string) (image.Point, error) {
	preset, ok := builtins[name]
	if !ok {
		return image.Point{}, fmt.Errorf("preset %q not found", name)
	}
	return preset.Size, nil
}

func ByName(name string) (Preset, error) {
	preset, ok := builtins[name]
	if !ok {
		return Preset{}, fmt.Errorf("preset %q not found", name)
	}
	return preset, nil
}

func Names() []string {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func Set(name string) ([]Preset, error) {
	names, ok := presetSets[name]
	if !ok {
		return nil, fmt.Errorf("preset set %q not found", name)
	}
	out := make([]Preset, 0, len(names))
	for _, presetName := range names {
		preset, ok := builtins[presetName]
		if !ok {
			return nil, fmt.Errorf("preset %q not found in set %q", presetName, name)
		}
		out = append(out, preset)
	}
	return out, nil
}

func Resolve(names []string, setName string) ([]Preset, error) {
	if len(names) > 0 && setName != "" {
		return nil, fmt.Errorf("use either explicit presets or a preset set, not both")
	}
	if setName != "" {
		return Set(setName)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("missing preset selection")
	}
	out := make([]Preset, 0, len(names))
	for _, name := range names {
		preset, ok := builtins[name]
		if !ok {
			return nil, fmt.Errorf("preset %q not found", name)
		}
		out = append(out, preset)
	}
	return out, nil
}
