package fontutil

import (
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

func LoadFace(path string, size float64) (font.Face, string, error) {
	if path == "" {
		face, err := newFace(goregular.TTF, size)
		if err != nil {
			return nil, "", err
		}
		return face, "embedded:goregular", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	face, err := newFace(data, size)
	if err != nil {
		return nil, "", err
	}
	return face, path, nil
}

func newFace(data []byte, size float64) (font.Face, error) {
	ft, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(ft, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingNone,
	})
}
