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
	p2 := b.mask.Pix
	for i := range p2 {
		p2[i] = 0
	}

	iterateTriangles(pts[:], func(tri [][2]float64) {
		b.fillTriangle(tri, func(x, y int) {
			if b.clip.AlphaAt(x, y).A == 0 {
				return
			}
			if b.mask.AlphaAt(x, y).A > 0 {
				return
			}
			b.mask.SetAlpha(x, y, color.Alpha{A: 255})
			b.Image.SetRGBA(x, y, mix(style.Color, b.Image.RGBAAt(x, y)))
		})
	})
}

func (b *SoftwareBackend) FillImageMask(style *backendbase.FillStyle, mask *image.Alpha, pts [4][2]float64) {
	mw := float64(mask.Bounds().Dx())
	mh := float64(mask.Bounds().Dy())
	b.fillQuad(pts, func(x, y int, sx2, sy2 float64) {
		sxi := int(mw * sx2)
		syi := int(mh * sy2)
		a := mask.AlphaAt(sxi, syi)
		if a.A == 0 {
			return
		}
		b.Image.SetRGBA(x, y, alphaColor(style.Color, a))
	})
}

func (b *SoftwareBackend) ClearClip() {
	p := b.clip.Pix
	for i := range p {
		p[i] = 255
	}
}

func (b *SoftwareBackend) Clip(pts [][2]float64) {
	p2 := b.mask.Pix
	for i := range p2 {
		p2[i] = 0
	}

	iterateTriangles(pts[:], func(tri [][2]float64) {
		b.fillTriangle(tri, func(x, y int) {
			b.mask.SetAlpha(x, y, color.Alpha{A: 255})
		})
	})

	p := b.clip.Pix
	for i := range p {
		if p2[i] == 0 {
			p[i] = 0
		}
	}
}
