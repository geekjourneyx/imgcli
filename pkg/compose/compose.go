package compose

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"golang.org/x/image/font"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/fontutil"
	"github.com/geekjourneyx/imgcli/pkg/ioimg"
)

type Layout string

const (
	LayoutPoster    Layout = "poster"
	LayoutCover     Layout = "cover"
	LayoutQuoteCard Layout = "quote-card"
	LayoutProduct   Layout = "product-card"
)

type Options struct {
	Input           string
	Output          string
	Width           int
	Height          int
	BackgroundColor string
	BackgroundImage string
	Title           string
	Subtitle        string
	TitleSize       float64
	SubtitleSize    float64
	TitleColor      string
	SubtitleColor   string
	FontPath        string
	Logo            string
	BannerBadge     string
	Padding         int
	Radius          float64
	SafeArea        string
	Layout          Layout
	Quality         int
}

type Result struct {
	Output     string   `json:"output"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Layout     string   `json:"layout"`
	Elements   []string `json:"elements_rendered"`
	FontSource string   `json:"font_source"`
	DurationMS int64    `json:"duration_ms"`
}

type inset struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

type layoutSpec struct {
	ImageRect     image.Rectangle
	LogoRect      image.Rectangle
	BadgeOrigin   image.Point
	TitleBox      image.Rectangle
	SubtitleBox   image.Rectangle
	TitleAlign    gg.Align
	SubtitleAlign gg.Align
	CenterText    bool
	DarkOverlay   float64
}

func Run(opts Options) (Result, error) {
	start := time.Now()
	if err := validate(opts); err != nil {
		return Result{}, err
	}

	src, err := ioimg.Open(opts.Input)
	if err != nil {
		return Result{}, err
	}

	var bg image.Image
	if opts.BackgroundImage != "" {
		bg, err = ioimg.Open(opts.BackgroundImage)
		if err != nil {
			return Result{}, err
		}
	}

	var logo image.Image
	if opts.Logo != "" {
		logo, err = ioimg.Open(opts.Logo)
		if err != nil {
			return Result{}, err
		}
	}

	if opts.Padding <= 0 {
		opts.Padding = 48
	}
	if opts.Quality <= 0 {
		opts.Quality = 85
	}
	if opts.Radius < 0 {
		opts.Radius = 0
	}
	if opts.Layout == "" {
		opts.Layout = LayoutPoster
	}

	safeArea, err := parseSafeArea(opts.SafeArea, opts.Padding)
	if err != nil {
		return Result{}, err
	}
	spec, err := buildLayout(opts.Layout, opts.Width, opts.Height, safeArea)
	if err != nil {
		return Result{}, err
	}

	titleColor, subtitleColor, badgeBg, badgeFg := resolveColors(opts, spec)
	titleSize, subtitleSize := resolveSizes(opts, spec)

	titleFace, fontSource, err := fontutil.LoadFace(opts.FontPath, titleSize)
	if err != nil {
		return Result{}, apperr.Wrap("FONT_LOAD_FAILED", 2, err, "load title font")
	}
	subtitleFace, _, err := fontutil.LoadFace(opts.FontPath, subtitleSize)
	if err != nil {
		return Result{}, apperr.Wrap("FONT_LOAD_FAILED", 2, err, "load subtitle font")
	}
	badgeFace, _, err := fontutil.LoadFace(opts.FontPath, math.Max(18, subtitleSize*0.7))
	if err != nil {
		return Result{}, apperr.Wrap("FONT_LOAD_FAILED", 2, err, "load badge font")
	}

	dc := gg.NewContext(opts.Width, opts.Height)
	drawBackground(dc, src, bg, opts, spec)
	drawForeground(dc, src, spec.ImageRect, opts.Radius)

	elements := []string{"background", "foreground"}
	if opts.BannerBadge != "" {
		drawBadge(dc, badgeFace, opts.BannerBadge, spec.BadgeOrigin, badgeBg, badgeFg)
		elements = append(elements, "badge")
	}
	if logo != nil {
		drawImageFit(dc, logo, spec.LogoRect, opts.Radius/2)
		elements = append(elements, "logo")
	}

	if opts.Title != "" {
		drawTextBlock(dc, titleFace, opts.Title, spec.TitleBox, titleColor, spec.TitleAlign, spec.CenterText, 2)
		elements = append(elements, "title")
	}
	if opts.Subtitle != "" {
		drawTextBlock(dc, subtitleFace, opts.Subtitle, spec.SubtitleBox, subtitleColor, spec.SubtitleAlign, spec.CenterText, 3)
		elements = append(elements, "subtitle")
	}

	if err := ioimg.Save(opts.Output, dc.Image(), opts.Quality); err != nil {
		return Result{}, err
	}

	return Result{
		Output:     opts.Output,
		Width:      opts.Width,
		Height:     opts.Height,
		Layout:     string(opts.Layout),
		Elements:   elements,
		FontSource: fontSource,
		DurationMS: time.Since(start).Milliseconds(),
	}, nil
}

func validate(opts Options) error {
	if opts.Input == "" || opts.Output == "" {
		return apperr.New("INVALID_ARGUMENT", "--input and --output are required", 2)
	}
	if opts.Width <= 0 || opts.Height <= 0 {
		return apperr.New("INVALID_ARGUMENT", "--width and --height must be positive", 2)
	}
	if opts.BackgroundColor != "" {
		if _, err := parseHexColor(opts.BackgroundColor); err != nil {
			return apperr.New("INVALID_ARGUMENT", "--background-color must be a 6-digit hex color", 2)
		}
	}
	if opts.TitleColor != "" {
		if _, err := parseHexColor(opts.TitleColor); err != nil {
			return apperr.New("INVALID_ARGUMENT", "--title-color must be a 6-digit hex color", 2)
		}
	}
	if opts.SubtitleColor != "" {
		if _, err := parseHexColor(opts.SubtitleColor); err != nil {
			return apperr.New("INVALID_ARGUMENT", "--subtitle-color must be a 6-digit hex color", 2)
		}
	}
	switch opts.Layout {
	case "", LayoutPoster, LayoutCover, LayoutQuoteCard, LayoutProduct:
		return nil
	default:
		return apperr.New("INVALID_ARGUMENT", fmt.Sprintf("unsupported layout %q", opts.Layout), 2)
	}
}

func parseSafeArea(raw string, padding int) (inset, error) {
	if raw == "" {
		return inset{Top: padding, Right: padding, Bottom: padding, Left: padding}, nil
	}
	parts := strings.Split(raw, ",")
	if len(parts) != 4 {
		return inset{}, apperr.New("INVALID_ARGUMENT", "safe-area must be top,right,bottom,left", 2)
	}
	out := inset{}
	values := []*int{&out.Top, &out.Right, &out.Bottom, &out.Left}
	for i, part := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || v < 0 {
			return inset{}, apperr.New("INVALID_ARGUMENT", "safe-area values must be non-negative integers", 2)
		}
		*values[i] = v
	}
	return out, nil
}

func buildLayout(layout Layout, width, height int, safe inset) (layoutSpec, error) {
	safeRect := image.Rect(safe.Left, safe.Top, width-safe.Right, height-safe.Bottom)
	gap := maxInt(width/40, 24)
	logoSize := minInt(maxInt(width/10, 72), 132)

	switch layout {
	case "", LayoutPoster:
		imageHeight := int(float64(safeRect.Dy()) * 0.58)
		imageRect := image.Rect(safeRect.Min.X, safeRect.Min.Y+logoSize/3, safeRect.Max.X, safeRect.Min.Y+logoSize/3+imageHeight)
		titleTop := imageRect.Max.Y + gap
		return layoutSpec{
			ImageRect:     imageRect,
			LogoRect:      image.Rect(safeRect.Max.X-logoSize, safeRect.Min.Y, safeRect.Max.X, safeRect.Min.Y+logoSize),
			BadgeOrigin:   image.Pt(safeRect.Min.X, safeRect.Min.Y),
			TitleBox:      image.Rect(safeRect.Min.X, titleTop, safeRect.Max.X, minInt(titleTop+height/6, safeRect.Max.Y)),
			SubtitleBox:   image.Rect(safeRect.Min.X, minInt(titleTop+height/7, safeRect.Max.Y-gap), safeRect.Max.X, safeRect.Max.Y),
			TitleAlign:    gg.AlignLeft,
			SubtitleAlign: gg.AlignLeft,
		}, nil
	case LayoutCover:
		imageWidth := int(float64(safeRect.Dx()) * 0.44)
		imageRect := image.Rect(safeRect.Max.X-imageWidth, safeRect.Min.Y, safeRect.Max.X, safeRect.Max.Y)
		textMaxX := imageRect.Min.X - gap
		return layoutSpec{
			ImageRect:     imageRect,
			LogoRect:      image.Rect(safeRect.Min.X, safeRect.Min.Y, safeRect.Min.X+logoSize, safeRect.Min.Y+logoSize),
			BadgeOrigin:   image.Pt(safeRect.Min.X, safeRect.Min.Y+logoSize+gap/2),
			TitleBox:      image.Rect(safeRect.Min.X, safeRect.Min.Y+safeRect.Dy()/3, textMaxX, safeRect.Min.Y+safeRect.Dy()/3+height/5),
			SubtitleBox:   image.Rect(safeRect.Min.X, safeRect.Min.Y+safeRect.Dy()/3+height/6, textMaxX, safeRect.Max.Y),
			TitleAlign:    gg.AlignLeft,
			SubtitleAlign: gg.AlignLeft,
		}, nil
	case LayoutQuoteCard:
		imgSize := minInt(int(float64(safeRect.Dx())*0.34), int(float64(safeRect.Dy())*0.30))
		imageRect := image.Rect(width/2-imgSize/2, safeRect.Min.Y+gap, width/2+imgSize/2, safeRect.Min.Y+gap+imgSize)
		titleTop := imageRect.Max.Y + gap
		return layoutSpec{
			ImageRect:     imageRect,
			LogoRect:      image.Rect(safeRect.Min.X, safeRect.Min.Y, safeRect.Min.X+logoSize, safeRect.Min.Y+logoSize),
			BadgeOrigin:   image.Pt(width/2-90, safeRect.Max.Y-gap-42),
			TitleBox:      image.Rect(safeRect.Min.X+gap, titleTop, safeRect.Max.X-gap, titleTop+height/4),
			SubtitleBox:   image.Rect(safeRect.Min.X+gap, titleTop+height/5, safeRect.Max.X-gap, safeRect.Max.Y-gap),
			TitleAlign:    gg.AlignCenter,
			SubtitleAlign: gg.AlignCenter,
			CenterText:    true,
			DarkOverlay:   0.38,
		}, nil
	case LayoutProduct:
		imageHeight := int(float64(safeRect.Dy()) * 0.50)
		imageWidth := int(float64(safeRect.Dx()) * 0.70)
		imageRect := image.Rect(width/2-imageWidth/2, safeRect.Min.Y+logoSize/3, width/2+imageWidth/2, safeRect.Min.Y+logoSize/3+imageHeight)
		titleTop := imageRect.Max.Y + gap
		return layoutSpec{
			ImageRect:     imageRect,
			LogoRect:      image.Rect(safeRect.Max.X-logoSize, safeRect.Min.Y, safeRect.Max.X, safeRect.Min.Y+logoSize),
			BadgeOrigin:   image.Pt(safeRect.Min.X, safeRect.Min.Y),
			TitleBox:      image.Rect(safeRect.Min.X, titleTop, safeRect.Max.X, titleTop+height/7),
			SubtitleBox:   image.Rect(safeRect.Min.X, titleTop+height/9, safeRect.Max.X, safeRect.Max.Y),
			TitleAlign:    gg.AlignCenter,
			SubtitleAlign: gg.AlignCenter,
			CenterText:    true,
		}, nil
	default:
		return layoutSpec{}, apperr.New("INVALID_ARGUMENT", fmt.Sprintf("unsupported layout %q", layout), 2)
	}
}

func resolveColors(opts Options, spec layoutSpec) (color.Color, color.Color, color.Color, color.Color) {
	titleDefault := "#111111"
	subtitleDefault := "#4a4a4a"
	badgeBg := mustHex("#111111")
	badgeFg := mustHex("#ffffff")
	if spec.CenterText && spec.DarkOverlay > 0 {
		titleDefault = "#ffffff"
		subtitleDefault = "#f0f0f0"
		badgeBg = mustRGBA(255, 255, 255, 220)
		badgeFg = mustHex("#111111")
	}

	return parseColorOrDefault(opts.TitleColor, titleDefault), parseColorOrDefault(opts.SubtitleColor, subtitleDefault), badgeBg, badgeFg
}

func resolveSizes(opts Options, spec layoutSpec) (float64, float64) {
	titleSize := opts.TitleSize
	subtitleSize := opts.SubtitleSize

	switch {
	case titleSize <= 0 && spec.CenterText:
		titleSize = 64
	case titleSize <= 0:
		titleSize = 58
	}
	switch {
	case subtitleSize <= 0 && spec.CenterText:
		subtitleSize = 28
	case subtitleSize <= 0:
		subtitleSize = 26
	}
	return titleSize, subtitleSize
}

func drawBackground(dc *gg.Context, src, bg image.Image, opts Options, spec layoutSpec) {
	canvas := image.Rect(0, 0, opts.Width, opts.Height)
	switch {
	case bg != nil:
		drawImageCover(dc, bg, canvas, 0)
	case opts.Layout == LayoutQuoteCard:
		blurred := imaging.Blur(imaging.Resize(src, maxInt(opts.Width/12, 1), 0, imaging.Linear), 4)
		cover := imaging.Fill(blurred, opts.Width, opts.Height, imaging.Center, imaging.Lanczos)
		dc.DrawImage(cover, 0, 0)
	default:
		dc.SetColor(lightenColor(dominantColor(src), 0.22))
		dc.DrawRectangle(0, 0, float64(opts.Width), float64(opts.Height))
		dc.Fill()
	}

	if opts.BackgroundColor != "" {
		dc.SetColor(parseColorOrDefault(opts.BackgroundColor, "#f3efe7"))
		dc.DrawRectangle(0, 0, float64(opts.Width), float64(opts.Height))
		dc.Fill()
	}
	if spec.DarkOverlay > 0 {
		dc.SetRGBA(0, 0, 0, spec.DarkOverlay)
		dc.DrawRectangle(0, 0, float64(opts.Width), float64(opts.Height))
		dc.Fill()
	}
}

func drawForeground(dc *gg.Context, img image.Image, rect image.Rectangle, radius float64) {
	drawImageFit(dc, img, rect, radius)
}

func drawImageFit(dc *gg.Context, img image.Image, rect image.Rectangle, radius float64) {
	if rect.Empty() {
		return
	}
	fitted := imaging.Fit(img, rect.Dx(), rect.Dy(), imaging.Lanczos)
	x := rect.Min.X + (rect.Dx()-fitted.Bounds().Dx())/2
	y := rect.Min.Y + (rect.Dy()-fitted.Bounds().Dy())/2
	dc.Push()
	if radius > 0 {
		dc.DrawRoundedRectangle(float64(rect.Min.X), float64(rect.Min.Y), float64(rect.Dx()), float64(rect.Dy()), radius)
		dc.Clip()
	}
	dc.DrawImage(fitted, x, y)
	dc.Pop()
}

func drawImageCover(dc *gg.Context, img image.Image, rect image.Rectangle, radius float64) {
	covered := imaging.Fill(img, rect.Dx(), rect.Dy(), imaging.Center, imaging.Lanczos)
	dc.Push()
	if radius > 0 {
		dc.DrawRoundedRectangle(float64(rect.Min.X), float64(rect.Min.Y), float64(rect.Dx()), float64(rect.Dy()), radius)
		dc.Clip()
	}
	dc.DrawImage(covered, rect.Min.X, rect.Min.Y)
	dc.Pop()
}

func drawBadge(dc *gg.Context, face font.Face, text string, origin image.Point, bg, fg color.Color) {
	dc.SetFontFace(face)
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	textWidth, _ := dc.MeasureString(text)
	height := maxInt(int(dc.FontHeight())+18, 36)
	width := maxInt(int(textWidth)+26, 88)
	radius := float64(minInt(height/2, 18))

	dc.SetColor(bg)
	dc.DrawRoundedRectangle(float64(origin.X), float64(origin.Y), float64(width), float64(height), radius)
	dc.Fill()

	dc.SetColor(fg)
	dc.DrawStringAnchored(text, float64(origin.X+width/2), float64(origin.Y+height/2), 0.5, 0.5)
}

func drawTextBlock(dc *gg.Context, face font.Face, text string, box image.Rectangle, c color.Color, align gg.Align, center bool, maxLines int) {
	text = strings.TrimSpace(text)
	if text == "" || box.Empty() {
		return
	}

	dc.SetFontFace(face)
	dc.SetColor(c)

	lines := wrapAndClamp(dc, text, float64(box.Dx()), maxLines)
	if len(lines) == 0 {
		return
	}

	lineSpacing := 1.35
	lineHeight := dc.FontHeight() * lineSpacing
	totalHeight := dc.FontHeight() + float64(len(lines)-1)*lineHeight

	x := float64(box.Min.X)
	ax := 0.0
	switch align {
	case gg.AlignCenter:
		x = float64(box.Min.X + box.Dx()/2)
		ax = 0.5
	case gg.AlignRight:
		x = float64(box.Max.X)
		ax = 1
	}
	y := float64(box.Min.Y)
	if center {
		y = float64(box.Min.Y) + math.Max(0, (float64(box.Dy())-totalHeight)/2)
	}

	for i, line := range lines {
		dc.DrawStringAnchored(line, x, y+float64(i)*lineHeight, ax, 0)
	}
}

func dominantColor(src image.Image) color.Color {
	sample := imaging.Resize(src, 1, 1, imaging.Box)
	return sample.At(0, 0)
}

func lightenColor(c color.Color, amount float64) color.Color {
	r16, g16, b16, a16 := c.RGBA()
	r := lightenByte(uint8(r16>>8), amount)
	g := lightenByte(uint8(g16>>8), amount)
	b := lightenByte(uint8(b16>>8), amount)
	return color.NRGBA{R: r, G: g, B: b, A: uint8(a16 >> 8)}
}

func lightenByte(v uint8, amount float64) uint8 {
	return uint8(math.Min(255, float64(v)+(255-float64(v))*amount))
}

func parseColorOrDefault(raw, fallback string) color.Color {
	if raw == "" {
		return mustHex(fallback)
	}
	c, _ := parseHexColor(raw)
	return c
}

func parseHexColor(raw string) (color.Color, error) {
	s := strings.TrimPrefix(strings.TrimSpace(raw), "#")
	if len(s) != 6 {
		return nil, fmt.Errorf("invalid hex color")
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return nil, err
	}
	return color.NRGBA{
		R: uint8(v >> 16),
		G: uint8((v >> 8) & 0xff),
		B: uint8(v & 0xff),
		A: 255,
	}, nil
}

func mustHex(raw string) color.Color {
	c, _ := parseHexColor(raw)
	return c
}

func mustRGBA(r, g, b, a uint8) color.Color {
	return color.NRGBA{R: r, G: g, B: b, A: a}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func wrapAndClamp(dc *gg.Context, text string, width float64, maxLines int) []string {
	lines := dc.WordWrap(text, width)
	if maxLines <= 0 || len(lines) <= maxLines {
		return lines
	}
	lines = append([]string(nil), lines[:maxLines]...)
	lines[maxLines-1] = fitEllipsis(dc, lines[maxLines-1], width)
	return lines
}

func fitEllipsis(dc *gg.Context, line string, width float64) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return "..."
	}
	if w, _ := dc.MeasureString(line); w <= width {
		if w2, _ := dc.MeasureString(line + "..."); w2 <= width {
			return line + "..."
		}
	}

	runes := []rune(line)
	for len(runes) > 1 {
		runes = runes[:len(runes)-1]
		candidate := strings.TrimSpace(string(runes)) + "..."
		if w, _ := dc.MeasureString(candidate); w <= width {
			return candidate
		}
	}
	return "..."
}
