package softwarebackend

import (
	"image/color"
	"math"
)

func toRGBA(src color.Color) color.RGBA {
	ir, ig, ib, ia := src.RGBA()
	return color.RGBA{
		R: uint8(ir >> 8),
		G: uint8(ig >> 8),
		B: uint8(ib >> 8),
		A: uint8(ia >> 8),
	}
}

func mix(src, dest color.Color) color.RGBA {
	ir1, ig1, ib1, ia1 := src.RGBA()
	r1 := float64(ir1) / 65535.0
	g1 := float64(ig1) / 65535.0
	b1 := float64(ib1) / 65535.0
	a1 := float64(ia1) / 65535.0

	ir2, ig2, ib2, ia2 := dest.RGBA()
	r2 := float64(ir2) / 65535.0
	g2 := float64(ig2) / 65535.0
	b2 := float64(ib2) / 65535.0
	a2 := float64(ia2) / 65535.0

	r := (r1-r2)*a1 + r2
	g := (g1-g2)*a1 + g2
	b := (b1-b2)*a1 + b2
	a := math.Max((a1-a2)*a1+a2, a2)

	return color.RGBA{
		R: uint8(math.Round(r * 255.0)),
		G: uint8(math.Round(g * 255.0)),
		B: uint8(math.Round(b * 255.0)),
		A: uint8(math.Round(a * 255.0)),
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

func lerp(col1, col2 color.Color, ratio float64) color.RGBA {
	ir1, ig1, ib1, ia1 := col1.RGBA()
	r1 := float64(ir1) / 65535.0
	g1 := float64(ig1) / 65535.0
	b1 := float64(ib1) / 65535.0
	a1 := float64(ia1) / 65535.0

	ir2, ig2, ib2, ia2 := col2.RGBA()
	r2 := float64(ir2) / 65535.0
	g2 := float64(ig2) / 65535.0
	b2 := float64(ib2) / 65535.0
	a2 := float64(ia2) / 65535.0

	r := (r1-r2)*ratio + r2
	g := (g1-g2)*ratio + g2
	b := (b1-b2)*ratio + b2
	a := (a1-a2)*ratio + a2

	return color.RGBA{
		R: uint8(math.Round(r * 255.0)),
		G: uint8(math.Round(g * 255.0)),
		B: uint8(math.Round(b * 255.0)),
		A: uint8(math.Round(a * 255.0)),
	}
}
