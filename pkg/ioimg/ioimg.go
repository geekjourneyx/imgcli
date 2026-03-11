package ioimg

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
)

func Open(path string) (image.Image, error) {
	img, err := imaging.Open(path, imaging.AutoOrientation(true))
	if err != nil {
		return nil, apperr.Wrap("DECODE_FAILED", 3, err, "decode image %q", path)
	}
	return img, nil
}

func Save(path string, img image.Image, quality int) error {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		err := imaging.Save(img, path, imaging.JPEGQuality(quality))
		if err != nil {
			return apperr.Wrap("ENCODE_FAILED", 4, err, "save jpeg %q", path)
		}
		return nil
	case ".png":
		err := imaging.Save(img, path)
		if err != nil {
			return apperr.Wrap("ENCODE_FAILED", 4, err, "save png %q", path)
		}
		return nil
	default:
		return apperr.New("UNSUPPORTED_FORMAT", fmt.Sprintf("unsupported output extension for %q", path), 2)
	}
}

func EnsureReadableFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return apperr.Wrap("IO_ERROR", 5, err, "stat %q", path)
	}
	if info.IsDir() {
		return apperr.New("INVALID_ARGUMENT", fmt.Sprintf("%q is a directory", path), 2)
	}
	return nil
}

func FillBackground(bounds image.Rectangle, c color.Color) *image.RGBA {
	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	return canvas
}
