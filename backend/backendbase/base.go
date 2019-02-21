package backendbase

import (
	"image"
	"image/color"
)

// Backend is used by the canvas to actually do the final
// drawing. This enables the backend to be implemented by
// various methods (OpenGL, but also other APIs or software)
type Backend interface {
	ClearRect(x, y, w, h int)
	Clear(pts [4][2]float64)
	Fill(style *FillStyle, pts [][2]float64)
	LoadImage(img image.Image) (Image, error)
	DrawImage(dimg Image, sx, sy, sw, sh, dx, dy, dw, dh float64, alpha float64)
}

// FillStyle is the color and other details on how to fill
type FillStyle struct {
	Color color.RGBA
	Blur  float64
	// radialGradient *RadialGradient
	// linearGradient *LinearGradient
	Image      Image
	FillMatrix [9]float64
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
