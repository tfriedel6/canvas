package softwarebackend

import (
	"image"
	"image/draw"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

type SoftwareBackend struct {
	Image *image.RGBA

	MSAA int

	blurSwap *image.RGBA

	clip    *image.Alpha
	stencil *image.Alpha
	w, h    int
}

func New(w, h int) *SoftwareBackend {
	b := &SoftwareBackend{}
	b.SetSize(w, h)
	return b
}

func (b *SoftwareBackend) SetSize(w, h int) {
	b.w, b.h = w, h
	b.Image = image.NewRGBA(image.Rect(0, 0, w, h))
	b.clip = image.NewAlpha(image.Rect(0, 0, w, h))
	b.stencil = image.NewAlpha(image.Rect(0, 0, w, h))
	b.ClearClip()
}

func (b *SoftwareBackend) Size() (int, int) {
	return b.w, b.h
}

func (b *SoftwareBackend) GetImageData(x, y, w, h int) *image.RGBA {
	return b.Image.SubImage(image.Rect(x, y, w, h)).(*image.RGBA)
}

func (b *SoftwareBackend) PutImageData(img *image.RGBA, x, y int) {
	draw.Draw(b.Image, image.Rect(x, y, img.Rect.Dx(), img.Rect.Dy()), img, image.ZP, draw.Src)
}

func (b *SoftwareBackend) CanUseAsImage(b2 backendbase.Backend) bool {
	return false
}

func (b *SoftwareBackend) AsImage() backendbase.Image {
	return nil
}

type LinearGradient struct {
	data backendbase.Gradient
}
type RadialGradient struct {
	data backendbase.Gradient
}

func (b *SoftwareBackend) LoadLinearGradient(data backendbase.Gradient) backendbase.LinearGradient {
	return &LinearGradient{data: data}
}

func (b *SoftwareBackend) LoadRadialGradient(data backendbase.Gradient) backendbase.RadialGradient {
	return &RadialGradient{data: data}
}

func (g *LinearGradient) Delete() {
}

func (g *LinearGradient) Replace(data backendbase.Gradient) {
	g.data = data
}

func (g *RadialGradient) Delete() {
}

func (g *RadialGradient) Replace(data backendbase.Gradient) {
	g.data = data
}
