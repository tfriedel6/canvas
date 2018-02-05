package canvas

import (
	"github.com/tfriedel6/lm"
)

func pointIsRightOfLine(a, b, p lm.Vec2) bool {
	if a[1] == b[1] {
		return false
	}
	if a[1] > b[1] {
		a, b = b, a
	}
	if p[1] < a[1] || p[1] > b[1] {
		return false
	}
	v := b.Sub(a)
	r := (p[1] - a[1]) / v[1]
	x := a[0] + r*v[0]
	return p[0] > x
}

func triangleContainsPoint(a, b, c, p lm.Vec2) bool {
	// if point is outside triangle bounds, return false
	if p[0] < a[0] && p[0] < b[0] && p[0] < c[0] {
		return false
	}
	if p[0] > a[0] && p[0] > b[0] && p[0] > c[0] {
		return false
	}
	if p[1] < a[1] && p[1] < b[1] && p[1] < c[1] {
		return false
	}
	if p[1] > a[1] && p[1] > b[1] && p[1] > c[1] {
		return false
	}
	// check whether the point is to the right of each triangle line.
	// if the total is 1, it is inside the triangle
	count := 0
	if pointIsRightOfLine(a, b, p) {
		count++
	}
	if pointIsRightOfLine(b, c, p) {
		count++
	}
	if pointIsRightOfLine(c, a, p) {
		count++
	}
	return count == 1
}

func polygonContainsPoint(polygon []lm.Vec2, p lm.Vec2) bool {
	a := polygon[len(polygon)-1]
	count := 0
	for _, b := range polygon {
		if pointIsRightOfLine(a, b, p) {
			count++
		}
		a = b
	}
	return count%2 == 1
}

func triangulatePath(path []pathPoint, target []float32) []float32 {
	var buf [500]lm.Vec2
	polygon := buf[:0]
	for _, p := range path {
		polygon = append(polygon, p.tf)
	}

	for len(polygon) > 2 {
		var i int
	triangles:
		for i = range polygon {
			a := polygon[i]
			b := polygon[(i+1)%len(polygon)]
			c := polygon[(i+2)%len(polygon)]
			for i2, p := range polygon {
				if i2 >= i && i2 <= i+2 {
					continue
				}
				if triangleContainsPoint(a, b, c, p) {
					continue triangles
				}
				center := a.Add(b).Add(c).DivF(3)
				if !polygonContainsPoint(polygon, center) {
					continue triangles
				}
			}
			target = append(target, a[0], a[1], b[0], b[1], c[0], c[1])
			break
		}
		remove := (i + 1) % len(polygon)
		polygon = append(polygon[:remove], polygon[remove+1:]...)
	}
	return target
}
