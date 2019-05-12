package canvas

import (
	"math"
	"sort"
)

const samePointTolerance = 1e-20

func pointIsRightOfLine(a, b, p vec) (bool, bool) {
	if a[1] == b[1] {
		return false, false
	}
	dir := false
	if a[1] > b[1] {
		a, b = b, a
		dir = !dir
	}
	if p[1] < a[1] || p[1] > b[1] {
		return false, false
	}
	v := b.sub(a)
	r := (p[1] - a[1]) / v[1]
	x := a[0] + r*v[0]
	return p[0] > x, dir
}

func pointIsBelowLine(a, b, p vec) (bool, bool) {
	if a[0] == b[0] {
		return false, false
	}
	dir := false
	if a[0] > b[0] {
		a, b = b, a
		dir = !dir
	}
	if p[0] < a[0] || p[0] > b[0] {
		return false, false
	}
	v := b.sub(a)
	r := (p[0] - a[0]) / v[0]
	x := a[1] + r*v[1]
	return p[1] > x, dir
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
	if r, _ := pointIsRightOfLine(a, b, p); r {
		count++
	}
	if r, _ := pointIsRightOfLine(b, c, p); r {
		count++
	}
	if r, _ := pointIsRightOfLine(c, a, p); r {
		count++
	}
	return count == 1
}

func polygonContainsPoint(polygon []vec, p vec) bool {
	a := polygon[len(polygon)-1]
	count := 0
	for _, b := range polygon {
		if r, _ := pointIsRightOfLine(a, b, p); r {
			count++
		}
		a = b
	}
	return count%2 == 1
}

func triangulatePath(path []pathPoint, mat mat, target [][2]float64) [][2]float64 {
	if path[0].pos == path[len(path)-1].pos {
		path = path[:len(path)-1]
	}

	var buf [500]vec
	polygon := buf[:0]
	for _, p := range path {
		polygon = append(polygon, p.pos.mulMat(mat))
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
			target = append(target, a, b, c)
			break
		}
		remove := (i + 1) % len(polygon)
		polygon = append(polygon[:remove], polygon[remove+1:]...)
	}
	return target
}

/*
tesselation strategy:

- cut the path at the intersections
- build a network of connected vertices and edges
- find out which side of each edge is inside the polygon
- pick an edge with only one side in the polygon, then follow it
  along the side on which the polygon is, always picking the edge
  with the smallest angle, until the path goes back to the starting
  point. set each edge as no longer having the polygon on that side
- repeat until no more edges have a polygon on either side

*/

type tessNet struct {
	verts []tessVert
	edges []tessEdge
}

type tessVert struct {
	pos      vec
	attached []int
	count    int
}

type tessEdge struct {
	a, b        int
	leftInside  bool
	rightInside bool
}

func cutIntersections(path []pathPoint) tessNet {
	// eliminate adjacent duplicate points
	for i := 0; i < len(path); i++ {
		a := path[i]
		b := path[(i+1)%len(path)]
		if isSamePoint(a.pos, b.pos, samePointTolerance) {
			copy(path[i:], path[i+1:])
			path = path[:len(path)-1]
			i--
		}
	}

	// find all the cuts
	type cut struct {
		from, to int
		ratio    float64
		point    vec
	}

	var cutBuf [50]cut
	cuts := cutBuf[:0]

	ip := len(path) - 1
	for i := 0; i < len(path); i++ {
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
			})
			cuts = append(cuts, cut{
				from:  jp,
				to:    j,
				ratio: r2,
				point: p,
			})
		}
		ip = i
	}

	if len(cuts) == 0 {
		return tessNet{}
	}

	sort.Slice(cuts, func(i, j int) bool {
		a, b := cuts[i], cuts[j]
		return a.to > b.to || (a.to == b.to && a.ratio > b.ratio)
	})

	// build vertex and edge lists
	verts := make([]tessVert, len(path)+len(cuts))
	for i, pp := range path {
		verts[i] = tessVert{
			pos:   pp.pos,
			count: 2,
		}
	}
	for _, cut := range cuts {
		copy(verts[cut.to+1:], verts[cut.to:])
		verts[cut.to].pos = cut.point
	}
	edges := make([]tessEdge, 0, len(path)+len(cuts)*2)
	for i := range verts {
		next := (i + 1) % len(verts)
		edges = append(edges, tessEdge{a: i, b: next})
	}

	// eliminate duplicate points
	for i := 0; i < len(verts); i++ {
		a := verts[i]
		for j := i + 1; j < len(verts); j++ {
			b := verts[j]
			if isSamePoint(a.pos, b.pos, samePointTolerance) {
				copy(verts[j:], verts[j+1:])
				verts = verts[:len(verts)-1]
				for k, e := range edges {
					if e.a == j {
						edges[k].a = i
					} else if e.a > j {
						edges[k].a--
					}
					if e.b == j {
						edges[k].b = i
					} else if e.b > j {
						edges[k].b--
					}
				}
				verts[i].count += 2
				j--
			}
		}
	}

	// build the attached edge lists on all vertices
	total := 0
	for _, v := range verts {
		total += v.count
	}
	attachedBuf := make([]int, 0, total)
	pos := 0
	for i := range verts {
		for j, e := range edges {
			if e.a == i || e.b == i {
				attachedBuf = append(attachedBuf, j)
			}
		}
		verts[i].attached = attachedBuf[pos:len(attachedBuf)]
		pos = len(attachedBuf)
	}

	return tessNet{verts: verts, edges: edges}
}

func setPathLeftRightInside(net *tessNet) {
	for i, e1 := range net.edges {
		a1, b1 := net.verts[e1.a], net.verts[e1.b]
		diff := b1.pos.sub(a1.pos)
		mid := a1.pos.add(diff.mulf(0.5))
		num := 0

		if math.Abs(diff[1]) < math.Abs(diff[0]) {
			edir := diff[1] > 0

			for j, e2 := range net.edges {
				if i == j {
					continue
				}
				a2, b2 := net.verts[e2.a], net.verts[e2.b]
				r, dir := pointIsRightOfLine(a2.pos, b2.pos, mid)
				if !r {
					continue
				}
				if dir {
					num++
				} else {
					num--
				}
			}

			if edir {
				net.edges[i].leftInside = (num - 1) != 0
				net.edges[i].rightInside = num != 0
			} else {
				net.edges[i].leftInside = num != 0
				net.edges[i].rightInside = (num + 1) != 0
			}
		} else {
			edir := diff[0] > 0

			for j, e2 := range net.edges {
				if i == j {
					continue
				}
				a2, b2 := net.verts[e2.a], net.verts[e2.b]
				b, dir := pointIsBelowLine(a2.pos, b2.pos, mid)
				if !b {
					continue
				}
				if dir {
					num++
				} else {
					num--
				}
			}

			if edir {
				net.edges[i].leftInside = num != 0
				net.edges[i].rightInside = (num - 1) != 0
			} else {
				net.edges[i].leftInside = (num + 1) != 0
				net.edges[i].rightInside = num != 0
			}
		}
	}
}

func selfIntersectingPathParts(p []pathPoint, partFn func(sp []pathPoint) bool) {
	runSubPaths(p, false, func(sp1 []pathPoint) bool {
		net := cutIntersections(sp1)
		if net.verts == nil {
			partFn(sp1)
			return false
		}

		setPathLeftRightInside(&net)

		var sp2Buf [50]pathPoint
		sp2 := sp2Buf[:0]

		for {
			var start, from, cur, count int
			var left bool
			for i, e := range net.edges {
				if e.leftInside != e.rightInside {
					count++
					start = e.a
					from = i
					cur = e.b
					if e.leftInside {
						left = true
						net.edges[i].leftInside = false
					} else {
						net.edges[i].rightInside = false
					}
					break
				}
			}
			if count == 0 {
				break
			}

			// fmt.Println("start", start, from, cur, net.verts[cur], left)

			sp2 = append(sp2, pathPoint{
				pos:   net.verts[cur].pos,
				flags: pathMove,
			})

			for limit := 0; limit < len(net.edges); limit++ {
				ecur := net.edges[from]
				acur, bcur := net.verts[ecur.a], net.verts[ecur.b]
				dir := bcur.pos.sub(acur.pos)
				dirAngle := math.Atan2(dir[1], dir[0])
				minAngleDiff := math.Pi * 2
				var next, nextEdge int
				any := false
				for _, ei := range net.verts[cur].attached {
					if ei == from {
						continue
					}
					e := net.edges[ei]
					if (left && !e.leftInside) || (!left && !e.rightInside) {
						continue
					}
					na, nb := net.verts[e.a], net.verts[e.b]
					if e.b == cur {
						na, nb = nb, na
					}
					ndir := nb.pos.sub(na.pos)
					nextAngle := math.Atan2(ndir[1], ndir[0]) + math.Pi
					if nextAngle < dirAngle {
						nextAngle += math.Pi * 2
					} else if nextAngle > dirAngle+math.Pi*2 {
						nextAngle -= math.Pi * 2
					}
					var angleDiff float64
					if left {
						angleDiff = nextAngle - dirAngle
					} else {
						angleDiff = dirAngle - nextAngle
					}
					if angleDiff < minAngleDiff {
						minAngleDiff = angleDiff
						nextEdge = ei
						if e.a == cur {
							next = e.b
						} else {
							next = e.a
						}
						any = true
						// fmt.Println("-", e, nextEdge, next)
					}
				}
				if !any {
					break
				}
				// fmt.Println(start, from, cur, net.verts[cur], nextEdge, next, net.verts[next])
				if left {
					net.edges[nextEdge].leftInside = false
				} else {
					net.edges[nextEdge].rightInside = false
				}
				sp2 = append(sp2, pathPoint{
					pos: net.verts[next].pos,
				})
				from = nextEdge
				cur = next
				if next == start {
					break
				}
			}

			stop := partFn(sp2)
			if stop {
				return true
			}
			sp2 = sp2[:0]
		}

		return false
	})
}
