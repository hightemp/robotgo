// Copyright (c) 2016-2025 AtomAI, All rights reserved.
//
// Package robotgo - Screenshot annotation helpers.
//
// This file provides visual annotation overlays for screenshots, helping AI agents
// accurately determine screen coordinates for mouse clicks and interactions.

package robotgo

import (
	"fmt"
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// AnnotationOptions configures visual overlays drawn on a screenshot.
// All options are additive — you can enable any combination.
type AnnotationOptions struct {
	// ShowCursor draws a semi-transparent red crosshair centred on the current
	// mouse cursor position, with a "(x,y)" coordinate label next to it.
	// Useful so the agent always knows exactly where the cursor currently sits.
	ShowCursor bool

	// ShowGrid draws a blue-grey semi-transparent coordinate grid over the image.
	// Numeric labels are drawn at every grid-line intersection so the agent can
	// estimate the pixel position of any UI element without counting pixels.
	ShowGrid bool

	// GridSize is the distance in pixels between consecutive grid lines.
	// Only effective when ShowGrid is true.  Defaults to 100 if ≤ 0.
	GridSize int

	// ShowRulers draws pixel rulers along the top and left edges of the image,
	// with major tick marks and labels every 50 px, minor tick marks every 10 px.
	// Useful for precise distance measurement from screen edges.
	ShowRulers bool

	// OffsetX / OffsetY are the absolute screen coordinates (in pixels) of the
	// top-left corner of the captured region.  When capturing a sub-region, set
	// these to x/y so that grid labels and cursor coordinates show absolute values
	// rather than region-relative ones.
	OffsetX int
	OffsetY int
}

// CaptureImgWithAnnotations takes a screenshot and draws the requested overlays.
//
// args follow the same convention as CaptureImg: optional (x, y, w, h, displayId).
// OffsetX / OffsetY in opts should be set to the capture region x/y when capturing
// a sub-region so that coordinate labels are absolute rather than region-relative.
func CaptureImgWithAnnotations(opts AnnotationOptions, args ...int) (image.Image, error) {
	img, err := CaptureImg(args...)
	if err != nil {
		return nil, err
	}
	if img == nil {
		return nil, nil
	}

	rgba := toRGBA(img)

	if opts.GridSize <= 0 {
		opts.GridSize = 100
	}

	DrawAnnotations(rgba, opts)
	return rgba, nil
}

// DrawAnnotations applies all enabled annotations to img in-place.
// The image must be *image.RGBA; call toRGBA to convert if needed.
func DrawAnnotations(img *image.RGBA, opts AnnotationOptions) {
	if opts.GridSize <= 0 {
		opts.GridSize = 100
	}

	if opts.ShowGrid {
		drawGrid(img, opts)
	}

	if opts.ShowRulers {
		drawRulers(img, opts)
	}

	if opts.ShowCursor {
		mx, my := Location()
		cx := mx - opts.OffsetX
		cy := my - opts.OffsetY
		drawCursor(img, cx, cy, mx, my)
	}
}

// ─── private helpers ────────────────────────────────────────────────────────

// toRGBA converts any image.Image to *image.RGBA, re-using the buffer when
// the input is already *image.RGBA.
func toRGBA(img image.Image) *image.RGBA {
	if r, ok := img.(*image.RGBA); ok {
		return r
	}
	b := img.Bounds()
	r := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r.Set(x, y, img.At(x, y))
		}
	}
	return r
}

// blendPixel alpha-composites colour c onto the pixel at (x, y) of img.
// Out-of-bounds coordinates are silently ignored.
func blendPixel(img *image.RGBA, x, y int, c color.RGBA) {
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
		return
	}
	dst := img.RGBAAt(x, y)
	a := float32(c.A) / 255.0
	ia := 1.0 - a
	dst.R = uint8(float32(c.R)*a + float32(dst.R)*ia)
	dst.G = uint8(float32(c.G)*a + float32(dst.G)*ia)
	dst.B = uint8(float32(c.B)*a + float32(dst.B)*ia)
	dst.A = 255
	img.SetRGBA(x, y, dst)
}

// fillRect fills a rectangle with blended colour.
func fillRect(img *image.RGBA, rx, ry, rw, rh int, c color.RGBA) {
	b := img.Bounds()
	x0 := max2(rx, b.Min.X)
	y0 := max2(ry, b.Min.Y)
	x1 := min2(rx+rw, b.Max.X)
	y1 := min2(ry+rh, b.Max.Y)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			blendPixel(img, x, y, c)
		}
	}
}

func drawHLine(img *image.RGBA, y, x1, x2 int, c color.RGBA) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	b := img.Bounds()
	if y < b.Min.Y || y >= b.Max.Y {
		return
	}
	x0c := max2(x1, b.Min.X)
	x1c := min2(x2, b.Max.X-1)
	for x := x0c; x <= x1c; x++ {
		blendPixel(img, x, y, c)
	}
}

func drawVLine(img *image.RGBA, x, y1, y2 int, c color.RGBA) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	b := img.Bounds()
	if x < b.Min.X || x >= b.Max.X {
		return
	}
	y0c := max2(y1, b.Min.Y)
	y1c := min2(y2, b.Max.Y-1)
	for y := y0c; y <= y1c; y++ {
		blendPixel(img, x, y, c)
	}
}

// drawText renders text at (x,y) (baseline-left) with basicfont.Face7x13.
func drawText(img *image.RGBA, x, y int, text string, c color.RGBA) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}

// drawCursor draws the red crosshair + label at image-local coords (cx,cy).
// mx,my are the absolute screen coords shown in the label.
func drawCursor(img *image.RGBA, cx, cy, mx, my int) {
	b := img.Bounds()

	crossColor := color.RGBA{R: 220, G: 0, B: 0, A: 200} // semi-transparent red
	solidRed := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	armLen := 22
	thickness := 3

	// Horizontal arm
	for t := -(thickness / 2); t <= thickness/2; t++ {
		drawHLine(img, cy+t, cx-armLen, cx+armLen, crossColor)
	}
	// Vertical arm
	for t := -(thickness / 2); t <= thickness/2; t++ {
		drawVLine(img, cx+t, cy-armLen, cy+armLen, crossColor)
	}

	// Solid centre circle (radius 4)
	r := 4
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				blendPixel(img, cx+dx, cy+dy, solidRed)
			}
		}
	}

	// Coordinate label
	label := fmt.Sprintf("(%d,%d)", mx, my)
	charW := 7
	lw := len(label)*charW + 6
	lh := 16

	lx := cx + 10
	ly := cy - 10

	// Clamp label into image bounds
	if lx+lw >= b.Max.X {
		lx = cx - lw - 10
	}
	if ly-13 < b.Min.Y {
		ly = cy + 20
	}
	if lx < b.Min.X {
		lx = b.Min.X + 2
	}

	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 175}
	fillRect(img, lx-2, ly-13, lw, lh, bgColor)
	drawText(img, lx, ly, label, color.RGBA{R: 255, G: 255, B: 255, A: 255})
}

// drawGrid draws a semi-transparent coordinate grid with numeric labels.
func drawGrid(img *image.RGBA, opts AnnotationOptions) {
	b := img.Bounds()
	gs := opts.GridSize

	lineColor := color.RGBA{R: 80, G: 120, B: 255, A: 55}
	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 130}
	fgColor := color.RGBA{R: 190, G: 210, B: 255, A: 255}

	// Figure out the first grid line within the image for both axes.
	// We want lines at absolute coordinates that are multiples of gs.
	// The image's left edge corresponds to absolute X = opts.OffsetX.
	firstVX := gs - (opts.OffsetX % gs)
	if firstVX == gs {
		firstVX = 0
	}
	firstHY := gs - (opts.OffsetY % gs)
	if firstHY == gs {
		firstHY = 0
	}

	// Vertical lines (fixed X, spanning full height)
	for relX := firstVX; relX < b.Dx(); relX += gs {
		imgX := b.Min.X + relX
		// Draw line
		for y := b.Min.Y; y < b.Max.Y; y++ {
			blendPixel(img, imgX, y, lineColor)
		}
		// Label
		absX := relX + opts.OffsetX
		label := fmt.Sprintf("%d", absX)
		lw := len(label)*7 + 4
		lx := imgX - lw/2
		fillRect(img, lx, b.Min.Y+1, lw, 13, bgColor)
		drawText(img, lx+2, b.Min.Y+12, label, fgColor)
	}

	// Horizontal lines (fixed Y, spanning full width)
	for relY := firstHY; relY < b.Dy(); relY += gs {
		imgY := b.Min.Y + relY
		// Draw line
		for x := b.Min.X; x < b.Max.X; x++ {
			blendPixel(img, x, imgY, lineColor)
		}
		// Label
		absY := relY + opts.OffsetY
		label := fmt.Sprintf("%d", absY)
		lw := len(label)*7 + 4
		ly := imgY + 12 // baseline below the line
		if ly >= b.Max.Y {
			ly = imgY - 2
		}
		fillRect(img, b.Min.X+1, imgY+1, lw, 13, bgColor)
		drawText(img, b.Min.X+3, ly, label, fgColor)
	}
}

// drawRulers draws pixel rulers along the top and left edges.
func drawRulers(img *image.RGBA, opts AnnotationOptions) {
	b := img.Bounds()
	rulerSz := 20 // ruler width/height in pixels

	rulerBg := color.RGBA{R: 35, G: 35, B: 35, A: 210}
	tickMajor := color.RGBA{R: 210, G: 210, B: 210, A: 255}
	tickMinor := color.RGBA{R: 140, G: 140, B: 140, A: 255}
	labelFg := color.RGBA{R: 240, G: 240, B: 240, A: 255}

	// Top ruler background
	fillRect(img, b.Min.X, b.Min.Y, b.Dx(), rulerSz, rulerBg)
	// Left ruler background
	fillRect(img, b.Min.X, b.Min.Y, rulerSz, b.Dy(), rulerBg)

	// Top ruler — ticks and labels along X axis
	for px := b.Min.X; px < b.Max.X; px++ {
		absX := px - b.Min.X + opts.OffsetX
		if absX%50 == 0 {
			drawVLine(img, px, b.Min.Y, b.Min.Y+rulerSz-1, tickMajor)
			label := fmt.Sprintf("%d", absX)
			if px+len(label)*7 < b.Max.X {
				drawText(img, px+2, b.Min.Y+12, label, labelFg)
			}
		} else if absX%10 == 0 {
			drawVLine(img, px, b.Min.Y+rulerSz-7, b.Min.Y+rulerSz-1, tickMinor)
		}
	}

	// Left ruler — ticks and labels along Y axis
	for py := b.Min.Y; py < b.Max.Y; py++ {
		absY := py - b.Min.Y + opts.OffsetY
		if absY%50 == 0 {
			drawHLine(img, py, b.Min.X, b.Min.X+rulerSz-1, tickMajor)
			label := fmt.Sprintf("%d", absY)
			if py+13 < b.Max.Y {
				drawText(img, b.Min.X+2, py+12, label, labelFg)
			}
		} else if absY%10 == 0 {
			drawHLine(img, py, b.Min.X+rulerSz-7, b.Min.X+rulerSz-1, tickMinor)
		}
	}

	// Blank corner square so ruler labels don't collide
	fillRect(img, b.Min.X, b.Min.Y, rulerSz, rulerSz, rulerBg)
}

// ─── tiny math helpers to avoid import of "math" ────────────────────────────

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}
