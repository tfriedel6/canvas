package backendbase

import (
	"image"
	"image/color"
	"math"
)

// Backend is used by the canvas to actually do the final
// drawing. This enables the backend to be implemented by
// various methods (OpenGL, but also other APIs or software)
type Backend interface {
	LoadImage(img image.Image) (Image, error)
	LoadLinearGradient(data *LinearGradientData) LinearGradient
	LoadRadialGradient(data *RadialGradientData) RadialGradient

	// ClearRect(x, y, w, h int)
	Clear(pts [4][2]float64)
	Fill(style *FillStyle, pts [][2]float64)
	DrawImage(dimg Image, sx, sy, sw, sh, dx, dy, dw, dh float64, alpha float64)
	FillImageMask(style *FillStyle, mask *image.Alpha, pts [][2]float64) // pts must have four points

	ClearClip()
	Clip(pts [][2]float64)

	GetImageData(x, y, w, h int) *image.RGBA
	PutImageData(img *image.RGBA, x, y int)
}

// FillStyle is the color and other details on how to fill
type FillStyle struct {
	Color          color.RGBA
	Blur           float64
	LinearGradient LinearGradient
	RadialGradient RadialGradient
	Image          Image
	FillMatrix     [9]float64
}

type LinearGradientData struct {
	X0, Y0 float64
	X1, Y1 float64
	Stops  Gradient
}

type RadialGradientData struct {
	X0, Y0  float64
	X1, Y1  float64
	RadFrom float64
	RadTo   float64
	Stops   Gradient
}

type Gradient []GradientStop

func (g Gradient) ColorAt(pos float64) color.RGBA {
	if len(g) == 0 {
		return color.RGBA{}
	} else if len(g) == 1 {
		return g[0].Color
	}
	beforeIdx, afterIdx := -1, -1
	for i, stop := range g {
		if stop.Pos > pos {
			afterIdx = i
			break
		}
		beforeIdx = i
	}
	if beforeIdx == -1 {
		return g[0].Color
	} else if afterIdx == -1 {
		return g[len(g)-1].Color
	}
	before, after := g[beforeIdx], g[afterIdx]
	p := (pos - before.Pos) / (after.Pos - before.Pos)
	var c [4]float64
	c[0] = (float64(after.Color.R)-float64(before.Color.R))*p + float64(before.Color.R)
	c[1] = (float64(after.Color.G)-float64(before.Color.G))*p + float64(before.Color.G)
	c[2] = (float64(after.Color.B)-float64(before.Color.B))*p + float64(before.Color.B)
	c[3] = (float64(after.Color.A)-float64(before.Color.A))*p + float64(before.Color.A)
	return color.RGBA{
		R: uint8(math.Round(c[0])),
		G: uint8(math.Round(c[1])),
		B: uint8(math.Round(c[2])),
		A: uint8(math.Round(c[3])),
	}
}

type GradientStop struct {
	Pos   float64
	Color color.RGBA
}

type LinearGradient interface {
	Delete()
	IsDeleted() bool
	IsOpaque() bool
	Replace(data *LinearGradientData)
}

type RadialGradient interface {
	Delete()
	IsDeleted() bool
	IsOpaque() bool
	Replace(data *RadialGradientData)
}

type Image interface {
	Width() int
	Height() int
	Size() (w, h int)
	Delete()
	IsDeleted() bool
	Replace(src image.Image) error
	IsOpaque() bool
}
