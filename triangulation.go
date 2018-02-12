package canvas

import (
	"sort"

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
				break
			}
			target = append(target, a[0], a[1], b[0], b[1], c[0], c[1])
			break
		}
		remove := (i + 1) % len(polygon)
		polygon = append(polygon[:remove], polygon[remove+1:]...)
	}
	return target
}

func (cv *Canvas) cutIntersections(path []pathPoint) []pathPoint {
	type cut struct {
		from, to int
		j        int
		b        bool
		ratio    float32
		point    lm.Vec2
	}

	var cutBuf [50]cut
	cuts := cutBuf[:0]

	for i := 0; i < len(cv.polyPath); i++ {
		ip := (i + len(cv.polyPath) - 1) % len(cv.polyPath)
		a0 := cv.polyPath[ip].pos
		a1 := cv.polyPath[i].pos
		for j := i + 1; j < len(cv.polyPath); j++ {
			jp := (j + len(cv.polyPath) - 1) % len(cv.polyPath)
			if ip == j || jp == i {
				continue
			}
			b0 := cv.polyPath[jp].pos
			b1 := cv.polyPath[j].pos
			p, r1, r2 := lineIntersection(a0, a1, b0, b1)
			if r1 <= 0 || r1 >= 1 || r2 <= 0 || r2 >= 1 {
				continue
			}
			cuts = append(cuts, cut{
				from:  ip,
				to:    i,
				ratio: r1,
				point: p,
				j:     j,
			})
			cuts = append(cuts, cut{
				from:  jp,
				to:    j,
				ratio: r2,
				point: p,
				j:     i,
				b:     true,
			})
		}
	}

	if len(cuts) == 0 {
		return path
	}

	sort.Slice(cuts, func(i, j int) bool {
		a, b := cuts[i], cuts[j]
		return a.to > b.to || (a.to == b.to && a.ratio > b.ratio)
	})

	newPath := make([]pathPoint, len(path)+len(cuts))
	copy(newPath[:len(path)], path)

	for _, cut := range cuts {
		copy(newPath[cut.to+1:], newPath[cut.to:])
		newPath[cut.to].next = newPath[cut.to+1].tf
		newPath[cut.to].pos = cut.point
		newPath[cut.to].tf = cv.tf(cut.point)
	}

	return newPath
}
