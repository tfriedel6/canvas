package softwarebackend

import (
	"image"
	"image/color"
	"math"

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

func mix(src, dest color.Color) color.RGBA {
	ir1, ig1, ib1, ia1 := src.RGBA()
	r1 := float64(ir1) / 65535.0
	g1 := float64(ig1) / 65535.0
	b1 := float64(ib1) / 65535.0
	a1 := float64(ia1) / 65535.0

	ir2, ig2, ib2, _ := dest.RGBA()
	r2 := float64(ir2) / 65535.0
	g2 := float64(ig2) / 65535.0
	b2 := float64(ib2) / 65535.0

	r := (r1-r2)*a1 + r2
	g := (g1-g2)*a1 + g2
	b := (b1-b2)*a1 + b2

	return color.RGBA{
		R: uint8(math.Round(r * 255.0)),
		G: uint8(math.Round(g * 255.0)),
		B: uint8(math.Round(b * 255.0)),
		A: 255,
	}
}

func alphaColor(col color.Color, alpha color.Alpha) color.RGBA {
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
