package softwarebackend

import (
	"math"
)

func triangleLR(tri [][2]float64, y float64) (l, r float64, outside bool) {
	a, b, c := tri[0], tri[1], tri[2]

	// sort by y
	if a[1] > b[1] {
		a, b = b, a
	}
	if b[1] > c[1] {
		b, c = c, b
		if a[1] > b[1] {
			a, b = b, a
		}
	}

	// check general bounds
	if y <= a[1] {
		return a[0], a[0], true
	}
	if y > c[1] {
		return c[0], c[0], true
	}

	// find left and right x at y
	if y >= a[1] && y <= b[1] && a[1] < b[1] {
		r0 := (y - a[1]) / (b[1] - a[1])
		l = (b[0]-a[0])*r0 + a[0]
		r1 := (y - a[1]) / (c[1] - a[1])
		r = (c[0]-a[0])*r1 + a[0]
	} else {
		r0 := (y - b[1]) / (c[1] - b[1])
		l = (c[0]-b[0])*r0 + b[0]
		r1 := (y - a[1]) / (c[1] - a[1])
		r = (c[0]-a[0])*r1 + a[0]
	}
	if l > r {
		l, r = r, l
	}

	return
}

func (b *SoftwareBackend) fillTriangle(tri [][2]float64, fn func(x, y int)) {
	minY := int(math.Floor(math.Min(math.Min(tri[0][1], tri[1][1]), tri[2][1])))
	maxY := int(math.Ceil(math.Max(math.Max(tri[0][1], tri[1][1]), tri[2][1])))
	if minY < 0 {
		minY = 0
	} else if minY >= b.h {
		return
	}
	if maxY < 0 {
		return
	} else if maxY >= b.h {
		maxY = b.h - 1
	}
	for y := minY; y <= maxY; y++ {
		l, r, out := triangleLR(tri, float64(y)+0.5)
		if out {
			continue
		}
		if l < 0 {
			l = 0
		} else if l > float64(b.w) {
			continue
		}
		if r < 0 {
			continue
		} else if r > float64(b.w) {
			r = float64(b.w)
		}
		if l >= r {
			continue
		}
		fl, cr := int(math.Floor(l)), int(math.Ceil(r))
		for x := fl; x <= cr; x++ {
			fx := float64(x) + 0.5
			if fx < l || fx >= r {
				continue
			}
			fn(x, y)
		}
	}
}

func (b *SoftwareBackend) fillQuad(quad [4][2]float64, fn func(x, y int, sx, sy float64)) {
	minY := int(math.Floor(math.Min(math.Min(quad[0][1], quad[1][1]), math.Min(quad[2][1], quad[3][1]))))
	maxY := int(math.Ceil(math.Max(math.Max(quad[0][1], quad[1][1]), math.Max(quad[2][1], quad[3][1]))))
	if minY < 0 {
		minY = 0
	} else if minY >= b.h {
		return
	}
	if maxY < 0 {
		return
	} else if maxY >= b.h {
		maxY = b.h - 1
	}

	leftv := [2]float64{quad[1][0] - quad[0][0], quad[1][1] - quad[0][1]}
	leftLen := math.Sqrt(leftv[0]*leftv[0] + leftv[1]*leftv[1])
	leftv[0] /= leftLen
	leftv[1] /= leftLen
	topv := [2]float64{quad[3][0] - quad[0][0], quad[3][1] - quad[0][1]}
	topLen := math.Sqrt(topv[0]*topv[0] + topv[1]*topv[1])
	topv[0] /= topLen
	topv[1] /= topLen

	tri1 := [3][2]float64{quad[0], quad[1], quad[2]}
	tri2 := [3][2]float64{quad[0], quad[2], quad[3]}
	for y := minY; y <= maxY; y++ {
		lf1, rf1, out1 := triangleLR(tri1[:], float64(y)+0.5)
		lf2, rf2, out2 := triangleLR(tri2[:], float64(y)+0.5)
		if out1 && out2 {
			continue
		}
		l := math.Min(lf1, lf2)
		r := math.Max(rf1, rf2)
		if l < 0 {
			l = 0
		} else if l > float64(b.w) {
			continue
		}
		if r < 0 {
			continue
		} else if r > float64(b.w) {
			r = float64(b.w)
		}
		if l >= r {
			continue
		}

		v0 := [2]float64{float64(l) - quad[0][0], float64(y) - quad[0][1]}
		sx0 := topv[0]*v0[0] + topv[1]*v0[1]
		sy0 := leftv[0]*v0[0] + leftv[1]*v0[1]

		v1 := [2]float64{float64(r) - quad[0][0], float64(y) - quad[0][1]}
		sx1 := topv[0]*v1[0] + topv[1]*v1[1]
		sy1 := leftv[0]*v1[0] + leftv[1]*v1[1]

		sx, sy := sx0/topLen, sy0/leftLen
		sxStep := (sx1 - sx0) / float64(r-l) / topLen
		syStep := (sy1 - sy0) / float64(r-l) / leftLen

		fl, cr := int(math.Floor(l)), int(math.Ceil(r))
		for x := fl; x <= cr; x++ {
			fx := float64(x) + 0.5
			if fx < l || fx >= r {
				continue
			}
			fn(x, y, sx, sy)
			sx += sxStep
			sy += syStep
		}
	}
}

func iterateTriangles(pts [][2]float64, fn func(tri [][2]float64)) {
	if len(pts) == 4 {
		var buf [3][2]float64
		buf[0] = pts[0]
		buf[1] = pts[1]
		buf[2] = pts[2]
		fn(buf[:])
		buf[1] = pts[2]
		buf[2] = pts[3]
		fn(buf[:])
		return
	}
	for i := 3; i <= len(pts); i += 3 {
		fn(pts[i-3 : i])
	}
}
