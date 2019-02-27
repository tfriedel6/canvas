package softwarebackend

import (
	"image/color"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

func (b *SoftwareBackend) Clear(pts [4][2]float64) {
	iterateTriangles(pts[:], func(tri [][2]float64) {
		b.fillTriangle(tri, func(x, y int) {
			b.Image.SetRGBA(x, y, color.RGBA{})
		})
	})
}

func (b *SoftwareBackend) Fill(style *backendbase.FillStyle, pts [][2]float64) {
	iterateTriangles(pts[:], func(tri [][2]float64) {
		b.fillTriangle(tri, func(x, y int) {
			b.Image.SetRGBA(x, y, style.Color)
		})
	})
}
