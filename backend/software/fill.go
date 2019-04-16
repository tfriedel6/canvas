package softwarebackend

import (
	"image"
	"image/color"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

func (b *SoftwareBackend) Clear(pts [4][2]float64) {
	iterateTriangles(pts[:], func(tri [][2]float64) {
		b.fillTriangle(tri, func(x, y int) {
			if b.clip.AlphaAt(x, y).A == 0 {
				return
			}
			b.Image.SetRGBA(x, y, color.RGBA{})
		})
	})
}

func (b *SoftwareBackend) Fill(style *backendbase.FillStyle, pts [][2]float64) {
	iterateTriangles(pts[:], func(tri [][2]float64) {
		b.fillTriangle(tri, func(x, y int) {
			if b.clip.AlphaAt(x, y).A == 0 {
				return
			}
			b.Image.SetRGBA(x, y, style.Color)
		})
	})
}

func mix(col color.Color, alpha color.Alpha) color.RGBA {
	ir, ig, ib, _ := col.RGBA()
	a2 := float64(alpha.A) / 255.0
	r := float64(ir) * a2 / 65535.0
	g := float64(ig) * a2 / 65535.0
	b := float64(ib) * a2 / 65535.0
	return color.RGBA{
		R: uint8(r * 255.0),
		G: uint8(g * 255.0),
		B: uint8(b * 255.0),
		A: 255,
	}
}

func (b *SoftwareBackend) FillImageMask(style *backendbase.FillStyle, mask *image.Alpha, pts [4][2]float64) {
	mw := float64(mask.Bounds().Dx())
	mh := float64(mask.Bounds().Dy())
	b.fillQuad(pts, func(x, y int, sx2, sy2 float64) {
		sxi := int(mw*sx2 + 0.5)
		syi := int(mh*sy2 + 0.5)
		a := mask.AlphaAt(sxi, syi)
		if a.A == 0 {
			return
		}
		b.Image.SetRGBA(x, y, mix(style.Color, a))
	})
}

func (b *SoftwareBackend) ClearClip() {
	p := b.clip.Pix
	for i := range p {
		p[i] = 255
	}
}

func (b *SoftwareBackend) Clip(pts [][2]float64) {
	p2 := b.clip2.Pix
	for i := range p2 {
		p2[i] = 0
	}

	iterateTriangles(pts[:], func(tri [][2]float64) {
		b.fillTriangle(tri, func(x, y int) {
			b.clip2.SetAlpha(x, y, color.Alpha{A: 255})
		})
	})

	p := b.clip.Pix
	for i := range p {
		if p2[i] == 0 {
			p[i] = 0
		}
	}
}
