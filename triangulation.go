package canvas

import (
	"math"
	"sort"
)

func pointIsRightOfLine(a, b, p vec) bool {
	if a[1] == b[1] {
		return false
	}
	if a[1] > b[1] {
		a, b = b, a
	}
	if p[1] < a[1] || p[1] > b[1] {
		return false
	}
	v := b.sub(a)
	r := (p[1] - a[1]) / v[1]
	x := a[0] + r*v[0]
	return p[0] > x
}

func triangleContainsPoint(a, b, c, p vec) bool {
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

func polygonContainsPoint(polygon []vec, p vec) bool {
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
	var buf [500]vec
	polygon := buf[:0]
	for _, p := range path {
		polygon = append(polygon, p.tf)
	}

	for len(polygon) > 2 {
		var i int
	triangles:
		for i = range polygon {
			ib := (i + 1) % len(polygon)
			ic := (i + 2) % len(polygon)
			a := polygon[i]
			b := polygon[ib]
			c := polygon[ic]
			if isSamePoint(a, c, math.SmallestNonzeroFloat64) {
				break
			}
			for i2, p := range polygon {
				if i2 == i || i2 == ib || i2 == ic {
					continue
				}
				if triangleContainsPoint(a, b, c, p) {
					continue triangles
				}
				center := a.add(b).add(c).divf(3)
				if !polygonContainsPoint(polygon, center) {
					continue triangles
				}
			}
			target = append(target, float32(a[0]), float32(a[1]), float32(b[0]), float32(b[1]), float32(c[0]), float32(c[1]))
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
		ratio    float64
		point    vec
	}

	var cutBuf [50]cut
	cuts := cutBuf[:0]

	for i := 0; i < len(path); i++ {
		ip := (i + len(path) - 1) % len(path)
		a0 := path[ip].pos
		a1 := path[i].pos
		for j := i + 1; j < len(path); j++ {
			jp := (j + len(path) - 1) % len(path)
			if ip == j || jp == i {
				continue
			}
			b0 := path[jp].pos
			b1 := path[j].pos
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
