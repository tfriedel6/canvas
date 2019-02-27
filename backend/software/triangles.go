package softwarebackend

import "math"

func triangleLR(tri [][2]float64, y float64) (l, r float64) {
	a, b, c := tri[0], tri[1], tri[2]

	// check general bounds
	if y < a[1] && y < b[1] && y < c[1] {
		return 0, 0
	}
	if y > a[1] && y > b[1] && y > c[1] {
		return 0, 0
	}

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

	// find left and right x at y
	if y >= a[1] && y <= b[1] {
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
		lf, rf := triangleLR(tri, float64(y)+0.5)
		l := int(math.Floor(lf))
		r := int(math.Ceil(rf))
		if l < 0 {
			l = 0
		} else if l >= b.w {
			continue
		}
		if r < 0 {
			continue
		} else if r >= b.w {
			r = b.w - 1
		}
		for x := l; x <= r; x++ {
			fn(x, y)
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
