package canvas

import (
	"math"
	"sort"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

// Go port of https://github.com/mapbox/earcut.hpp

type node struct {
	i    int
	x, y float64

	// previous and next vertice nodes in a polygon ring
	prev *node
	next *node

	// z-order curve value
	z int

	// previous and next nodes in z-order
	prevZ *node
	nextZ *node

	// indicates whether this is a steiner point
	steiner bool
}

type earcut struct {
	indices    []int
	vertices   int
	hashing    bool
	minX, minY float64
	maxX, maxY float64
	invSize    float64
	nodes      []node
}

func (ec *earcut) run(points [][]backendbase.Vec) {
	if len(points) == 0 {
		return
	}

	var x, y float64
	threshold := 80
	ln := 0

	for i := 0; threshold >= 0 && i < len(points); i++ {
		threshold -= len(points[i])
		ln += len(points[i])
	}

	//estimate size of nodes and indices
	ec.nodes = make([]node, 0, ln*3/2)
	ec.indices = make([]int, 0, ln+len(points[0]))
	ec.vertices = 0

	outerNode := ec.linkedList(points[0], true)
	if outerNode == nil || outerNode.prev == outerNode.next {
		return
	}

	if len(points) > 1 {
		outerNode = ec.eliminateHoles(points, outerNode)
	}

	// if the shape is not too simple, we'll use z-order curve hash later; calculate polygon bbox
	ec.hashing = threshold < 0
	if ec.hashing {
		p := outerNode.next
		ec.minX, ec.maxX = outerNode.x, outerNode.x
		ec.minY, ec.maxY = outerNode.y, outerNode.y
		for {
			x = p.x
			y = p.y
			ec.minX = math.Min(ec.minX, x)
			ec.minY = math.Min(ec.minY, y)
			ec.maxX = math.Min(ec.maxX, x)
			ec.maxY = math.Min(ec.maxY, y)
			p = p.next
			if p != outerNode {
				break
			}
		}

		// minX, minY and size are later used to transform coords into integers for z-order calculation
		ec.invSize = math.Max(ec.maxX-ec.minX, ec.maxY-ec.minY)
		if ec.invSize != 0 {
			ec.invSize = 1 / ec.invSize
		}
	}

	ec.earcutLinked(outerNode, 0)

	ec.nodes = ec.nodes[:0]
}

// create a circular doubly linked list from polygon points in the specified winding order
func (ec *earcut) linkedList(points []backendbase.Vec, clockwise bool) *node {
	var sum float64
	ln := len(points)
	var i, j int
	var last *node

	// calculate original winding order of a polygon ring
	if ln > 0 {
		j = ln - 1
	}
	for i < ln {
		p1 := points[i]
		p2 := points[j]
		p20 := p2[0]
		p10 := p1[0]
		p11 := p1[1]
		p21 := p2[1]
		sum += (p20 - p10) * (p11 + p21)
		j = i
		i++
	}

	// link points into circular doubly-linked list in the specified winding order
	if clockwise == (sum > 0) {
		for i := 0; i < ln; i++ {
			last = ec.insertNode(ec.vertices+i, points[i], last)
		}
	} else {
		for i = ln - 1; i >= 0; i-- {
			last = ec.insertNode(ec.vertices+i, points[i], last)
		}
	}

	if last != nil && ec.equals(last, last.next) {
		ec.removeNode(last)
		last = last.next
	}

	ec.vertices += ln

	return last
}

// eliminate colinear or duplicate points
func (ec *earcut) filterPoints(start, end *node) *node {
	if end == nil {
		end = start
	}

	p := start
	var again bool
	for {
		again = false

		if !p.steiner && (ec.equals(p, p.next) || ec.area(p.prev, p, p.next) == 0) {
			ec.removeNode(p)
			p, end = p.prev, p.prev

			if p == p.next {
				break
			}
			again = true

		} else {
			p = p.next
		}
		if !again && p == end {
			break
		}
	}

	return end
}

// main ear slicing loop which triangulates a polygon (given as a linked list)
func (ec *earcut) earcutLinked(ear *node, pass int) {
	if ear == nil {
		return
	}

	// interlink polygon nodes in z-order
	if pass == 0 && ec.hashing {
		ec.indexCurve(ear)
	}

	stop := ear
	var prev, next *node

	iterations := 0

	// iterate through ears, slicing them one by one
	for ear.prev != ear.next {
		iterations++
		prev = ear.prev
		next = ear.next

		var e bool
		if ec.hashing {
			e = ec.isEarHashed(ear)
		} else {
			e = ec.isEar(ear)
		}
		if e {
			// cut off the triangle
			ec.indices = append(ec.indices, prev.i, ear.i, next.i)

			ec.removeNode(ear)

			// skipping the next vertice leads to less sliver triangles
			ear = next.next
			stop = next.next

			continue
		}

		ear = next

		// if we looped through the whole remaining polygon and can't find any more ears
		if ear == stop {
			// try filtering points and slicing again
			if pass == 0 {
				ec.earcutLinked(ec.filterPoints(ear, nil), 1)
			} else if pass == 1 {
				// if this didn't work, try curing all small self-intersections locally
				ear = ec.cureLocalIntersections(ec.filterPoints(ear, nil))
				ec.earcutLinked(ear, 2)
			} else if pass == 2 {
				// as a last resort, try splitting the remaining polygon into two
				ec.splitEarcut(ear)
			}

			break
		}
	}
}

// check whether a polygon node forms a valid ear with adjacent nodes
func (ec *earcut) isEar(ear *node) bool {
	a := ear.prev
	b := ear
	c := ear.next

	if ec.area(a, b, c) >= 0 {
		return false // reflex, can't be an ear
	}

	// now make sure we don't have other points inside the potential ear
	p := ear.next.next

	for p != ear.prev {
		if ec.pointInTriangle(a.x, a.y, b.x, b.y, c.x, c.y, p.x, p.y) &&
			ec.area(p.prev, p, p.next) >= 0 {
			return false
		}
		p = p.next
	}

	return true
}

func (ec *earcut) isEarHashed(ear *node) bool {
	a := ear.prev
	b := ear
	c := ear.next

	if ec.area(a, b, c) >= 0 {
		return false // reflex, can't be an ear
	}

	// triangle bbox; min & max are calculated like this for speed

	minTX := math.Min(a.x, math.Min(b.x, c.x))
	minTY := math.Min(a.y, math.Min(b.y, c.y))
	maxTX := math.Max(a.x, math.Max(b.x, c.x))
	maxTY := math.Max(a.y, math.Max(b.y, c.y))

	// z-order range for the current triangle bbox;
	minZ := ec.zOrder(minTX, minTY)
	maxZ := ec.zOrder(maxTX, maxTY)

	// first look for points inside the triangle in increasing z-order
	p := ear.nextZ

	for p != nil && p.z <= maxZ {
		if p != ear.prev && p != ear.next &&
			ec.pointInTriangle(a.x, a.y, b.x, b.y, c.x, c.y, p.x, p.y) &&
			ec.area(p.prev, p, p.next) >= 0 {
			return false
		}
		p = p.nextZ
	}

	// then look for points in decreasing z-order
	p = ear.prevZ

	for p != nil && p.z >= minZ {
		if p != ear.prev && p != ear.next &&
			ec.pointInTriangle(a.x, a.y, b.x, b.y, c.x, c.y, p.x, p.y) &&
			ec.area(p.prev, p, p.next) >= 0 {
			return false
		}
		p = p.prevZ
	}

	return true
}

// go through all polygon nodes and cure small local self-intersections
func (ec *earcut) cureLocalIntersections(start *node) *node {
	p := start
	for {
		a := p.prev
		b := p.next.next

		// a self-intersection where edge (v[i-1],v[i]) intersects (v[i+1],v[i+2])
		if !ec.equals(a, b) && ec.intersects(a, p, p.next, b) && ec.locallyInside(a, b) && ec.locallyInside(b, a) {
			ec.indices = append(ec.indices, a.i, p.i, b.i)

			// remove two nodes involved
			ec.removeNode(p)
			ec.removeNode(p.next)

			p, start = b, b
		}
		p = p.next
		if p == start {
			break
		}
	}

	return ec.filterPoints(p, nil)
}

// try splitting polygon into two and triangulate them independently
func (ec *earcut) splitEarcut(start *node) {
	// look for a valid diagonal that divides the polygon into two
	a := start
	for {
		b := a.next.next
		for b != a.prev {
			if a.i != b.i && ec.isValidDiagonal(a, b) {
				// split the polygon in two by the diagonal
				c := ec.splitPolygon(a, b)

				// filter colinear points around the cuts
				a = ec.filterPoints(a, a.next)
				c = ec.filterPoints(c, c.next)

				// run earcut on each half
				ec.earcutLinked(a, 0)
				ec.earcutLinked(c, 0)
				return
			}
			b = b.next
		}
		a = a.next
		if a == start {
			break
		}
	}
}

// link every hole into the outer loop, producing a single-ring polygon without holes
func (ec *earcut) eliminateHoles(points [][]backendbase.Vec, outerNode *node) *node {
	ln := len(points)

	queue := make([]*node, 0, ln)
	for i := 1; i < ln; i++ {
		list := ec.linkedList(points[i], false)
		if list != nil {
			if list == list.next {
				list.steiner = true
			}
			queue = append(queue, ec.getLeftmost(list))
		}
	}
	sort.Slice(queue, func(a, b int) bool {
		return queue[a].x < queue[b].x
	})

	// process holes from left to right
	for i := 0; i < len(queue); i++ {
		ec.eliminateHole(queue[i], outerNode)
		outerNode = ec.filterPoints(outerNode, outerNode.next)
	}

	return outerNode
}

// find a bridge between vertices that connects hole with an outer ring and and link it
func (ec *earcut) eliminateHole(hole, outerNode *node) {
	outerNode = ec.findHoleBridge(hole, outerNode)
	if outerNode != nil {
		b := ec.splitPolygon(outerNode, hole)

		// filter out colinear points around cuts
		ec.filterPoints(outerNode, outerNode.next)
		ec.filterPoints(b, b.next)
	}
}

// David Eberly's algorithm for finding a bridge between hole and outer polygon
func (ec *earcut) findHoleBridge(hole, outerNode *node) *node {
	p := outerNode
	hx := hole.x
	hy := hole.y
	qx := math.Inf(-1)
	var m *node

	// find a segment intersected by a ray from the hole's leftmost Vertex to the left;
	// segment's endpoint with lesser x will be potential connection Vertex
	for {
		if hy <= p.y && hy >= p.next.y && p.next.y != p.y {
			x := p.x + (hy-p.y)*(p.next.x-p.x)/(p.next.y-p.y)
			if x <= hx && x > qx {
				qx = x
				if x == hx {
					if hy == p.y {
						return p
					}
					if hy == p.next.y {
						return p.next
					}
				}
				if p.x < p.next.x {
					m = p
				} else {
					m = p.next
				}
			}
		}
		p = p.next
		if p == outerNode {
			break
		}
	}

	if m == nil {
		return nil
	}

	if hx == qx {
		return m // hole touches outer segment; pick leftmost endpoint
	}

	// look for points inside the triangle of hole Vertex, segment intersection and endpoint;
	// if there are no points found, we have a valid connection;
	// otherwise choose the Vertex of the minimum angle with the ray as connection Vertex

	stop := m
	tanMin := math.Inf(1)
	tanCur := 0.0

	p = m
	mx := m.x
	my := m.y

	for {
		var pt1, pt2 float64
		if hy < my {
			pt1 = hx
			pt2 = qx
		} else {
			pt1 = qx
			pt2 = hx
		}
		if hx >= p.x && p.x >= mx && hx != p.x &&
			ec.pointInTriangle(pt1, hy, mx, my, pt2, hy, p.x, p.y) {

			tanCur = math.Abs(hy-p.y) / (hx - p.x) // tangential

			if ec.locallyInside(p, hole) &&
				(tanCur < tanMin || (tanCur == tanMin && (p.x > m.x || ec.sectorContainsSector(m, p)))) {
				m = p
				tanMin = tanCur
			}
		}

		p = p.next
		if p == stop {
			break
		}
	}

	return m
}

// whether sector in vertex m contains sector in vertex p in the same coordinates
func (ec *earcut) sectorContainsSector(m, p *node) bool {
	return ec.area(m.prev, m, p.prev) < 0 && ec.area(p.next, m, m.next) < 0
}

// interlink polygon nodes in z-order
func (ec *earcut) indexCurve(start *node) {
	if start == nil {
		panic("start must not be nil")
	}
	p := start

	for {
		if p.z <= 0 {
			p.z = ec.zOrder(p.x, p.y)
		}
		p.prevZ = p.prev
		p.nextZ = p.next
		p = p.next
		if p == start {
			break
		}
	}

	p.prevZ.nextZ = nil
	p.prevZ = nil

	ec.sortLinked(p)
}

// Simon Tatham's linked list merge sort algorithm
// http://www.chiark.greenend.org.uk/~sgtatham/algorithms/listsort.html
func (ec *earcut) sortLinked(list *node) *node {
	if list == nil {
		panic("list must not be nil")
	}
	var p, q, e, tail *node
	var i, numMerges, pSize, qSize int
	inSize := 1

	for {
		p = list
		list = nil
		tail = nil
		numMerges = 0

		for p != nil {
			numMerges++
			q = p
			pSize = 0
			for i = 0; i < inSize; i++ {
				pSize++
				q = q.nextZ
				if q == nil {
					break
				}
			}

			qSize = inSize

			for pSize > 0 || (qSize > 0 && q != nil) {

				if pSize == 0 {
					e = q
					q = q.nextZ
					qSize--
				} else if qSize == 0 || q == nil {
					e = p
					p = p.nextZ
					pSize--
				} else if p.z <= q.z {
					e = p
					p = p.nextZ
					pSize--
				} else {
					e = q
					q = q.nextZ
					qSize--
				}

				if tail != nil {
					tail.nextZ = e
				} else {
					list = e
				}

				e.prevZ = tail
				tail = e
			}

			p = q
		}

		tail.nextZ = nil

		if numMerges <= 1 {
			return list
		}

		inSize *= 2
	}
}

// z-order of a Vertex given coords and size of the data bounding box
func (ec *earcut) zOrder(x, y float64) int {
	// coords are transformed into non-negative 15-bit integer range
	x2 := int(32767.0 * (x - ec.minX) * ec.invSize)
	y2 := int(32767.0 * (y - ec.minY) * ec.invSize)

	x2 = (x2 | (x2 << 8)) & 0x00FF00FF
	x2 = (x2 | (x2 << 4)) & 0x0F0F0F0F
	x2 = (x2 | (x2 << 2)) & 0x33333333
	x2 = (x2 | (x2 << 1)) & 0x55555555

	y2 = (y2 | (y2 << 8)) & 0x00FF00FF
	y2 = (y2 | (y2 << 4)) & 0x0F0F0F0F
	y2 = (y2 | (y2 << 2)) & 0x33333333
	y2 = (y2 | (y2 << 1)) & 0x55555555

	return x2 | (y2 << 1)
}

// find the leftmost node of a polygon ring
func (ec *earcut) getLeftmost(start *node) *node {
	p := start
	leftmost := start
	for {
		if p.x < leftmost.x || (p.x == leftmost.x && p.y < leftmost.y) {
			leftmost = p
		}
		p = p.next
		if p == start {
			break
		}
	}

	return leftmost
}

// check if a point lies within a convex triangle
func (ec *earcut) pointInTriangle(ax, ay, bx, by, cx, cy, px, py float64) bool {
	return (cx-px)*(ay-py)-(ax-px)*(cy-py) >= 0 &&
		(ax-px)*(by-py)-(bx-px)*(ay-py) >= 0 &&
		(bx-px)*(cy-py)-(cx-px)*(by-py) >= 0
}

// check if a diagonal between two polygon nodes is valid (lies in polygon interior)
func (ec *earcut) isValidDiagonal(a, b *node) bool {
	return a.next.i != b.i && a.prev.i != b.i && !ec.intersectsPolygon(a, b) && // dones't intersect other edges
		((ec.locallyInside(a, b) && ec.locallyInside(b, a) && ec.middleInside(a, b) && // locally visible
			(ec.area(a.prev, a, b.prev) != 0.0 || ec.area(a, b.prev, b) != 0.0)) || // does not create opposite-facing sectors
			(ec.equals(a, b) && ec.area(a.prev, a, a.next) > 0 && ec.area(b.prev, b, b.next) > 0)) // special zero-length case
}

// signed area of a triangle
func (ec *earcut) area(p, q, r *node) float64 {
	return (q.y-p.y)*(r.x-q.x) - (q.x-p.x)*(r.y-q.y)
}

// check if two points are equal
func (ec *earcut) equals(p1, p2 *node) bool {
	return p1.x == p2.x && p1.y == p2.y
}

// check if two segments intersect
func (ec *earcut) intersects(p1, q1, p2, q2 *node) bool {
	o1 := ec.sign(ec.area(p1, q1, p2))
	o2 := ec.sign(ec.area(p1, q1, q2))
	o3 := ec.sign(ec.area(p2, q2, p1))
	o4 := ec.sign(ec.area(p2, q2, q1))

	if o1 != o2 && o3 != o4 {
		return true // general case
	}

	if o1 == 0 && ec.onSegment(p1, p2, q1) {
		// p1, q1 and p2 are collinear and p2 lies on p1q1
		return true
	}
	if o2 == 0 && ec.onSegment(p1, q2, q1) {
		// p1, q1 and q2 are collinear and q2 lies on p1q1
		return true
	}
	if o3 == 0 && ec.onSegment(p2, p1, q2) {
		// p2, q2 and p1 are collinear and p1 lies on p2q2
		return true
	}
	if o4 == 0 && ec.onSegment(p2, q1, q2) {
		// p2, q2 and q1 are collinear and q1 lies on p2q2
		return true
	}

	return false
}

// for collinear points p, q, r, check if point q lies on segment pr
func (ec *earcut) onSegment(p, q, r *node) bool {
	return q.x <= math.Max(p.x, r.x) &&
		q.x >= math.Min(p.x, r.x) &&
		q.y <= math.Max(p.y, r.y) &&
		q.y >= math.Min(p.y, r.y)
}

func (ec *earcut) sign(val float64) int {
	if val < 0 {
		return -1
	} else if val > 0 {
		return 1
	}
	return 0
}

// check if a polygon diagonal intersects any polygon segments
func (ec *earcut) intersectsPolygon(a, b *node) bool {
	p := a
	for {
		if p.i != a.i && p.next.i != a.i && p.i != b.i && p.next.i != b.i &&
			ec.intersects(p, p.next, a, b) {
			return true
		}
		p = p.next
		if p == a {
			break
		}
	}

	return false
}

// check if a polygon diagonal is locally inside the polygon
func (ec *earcut) locallyInside(a, b *node) bool {
	if ec.area(a.prev, a, a.next) < 0 {
		return ec.area(a, b, a.next) >= 0 && ec.area(a, a.prev, b) >= 0
	}
	return ec.area(a, b, a.prev) < 0 || ec.area(a, a.next, b) < 0
}

// check if the middle Vertex of a polygon diagonal is inside the polygon
func (ec *earcut) middleInside(a, b *node) bool {
	p := a
	inside := false
	px := (a.x + b.x) / 2
	py := (a.y + b.y) / 2
	for {
		if ((p.y > py) != (p.next.y > py)) && p.next.y != p.y &&
			(px < (p.next.x-p.x)*(py-p.y)/(p.next.y-p.y)+p.x) {
			inside = !inside
		}
		p = p.next
		if p == a {
			break
		}
	}

	return inside
}

// link two polygon vertices with a bridge; if the vertices belong to the same ring, it splits
// polygon into two; if one belongs to the outer ring and another to a hole, it merges it into a
// single ring
func (ec *earcut) splitPolygon(a, b *node) *node {
	ec.nodes = append(ec.nodes, node{i: a.i, x: a.x, y: a.y})
	a2 := &ec.nodes[len(ec.nodes)-1]
	ec.nodes = append(ec.nodes, node{i: b.i, x: b.x, y: b.y})
	b2 := &ec.nodes[len(ec.nodes)-1]
	an := a.next
	bp := b.prev

	a.next = b
	b.prev = a

	a2.next = an
	an.prev = a2

	b2.next = a2
	a2.prev = b2

	bp.next = b2
	b2.prev = bp

	return b2
}

// create a node and util::optionally link it with previous one (in a circular doubly linked list)
func (ec *earcut) insertNode(i int, pt backendbase.Vec, last *node) *node {
	ec.nodes = append(ec.nodes, node{i: i, x: pt[0], y: pt[1]})
	p := &ec.nodes[len(ec.nodes)-1]

	if last == nil {
		p.prev = p
		p.next = p
	} else {
		if last == nil {
			panic("last must not be nil")
		}
		p.next = last.next
		p.prev = last
		last.next.prev = p
		last.next = p
	}
	return p
}

func (ec *earcut) removeNode(p *node) {
	p.next.prev = p.prev
	p.prev.next = p.next

	if p.prevZ != nil {
		p.prevZ.nextZ = p.nextZ
	}
	if p.nextZ != nil {
		p.nextZ.prevZ = p.prevZ
	}
}

// sortFontContours takes the contours of a font glyph
// and checks whether each contour is the outside or a
// hole, and returns an array that is sorted so that
// it contains an index of an outer contour followed by
// any number of indices of hole contours followed by
// a terminating -1
func sortFontContours(contours [][]backendbase.Vec) []int {
	type cut struct {
		idx   int
		count int
	}
	type info struct {
		cuts     []cut
		cutTotal int
		outer    bool
	}

	cutBuf := make([]cut, len(contours)*len(contours))
	cinf := make([]info, len(contours))
	for i := range contours {
		cinf[i].cuts = cutBuf[i*len(contours) : i*len(contours)]
	}

	// go through each contour, pick one point on it, and
	// project that point to the right. count the number of
	// other contours that it cuts
	for i, p1 := range contours {
		pt := p1[0]
		for j, p2 := range contours {
			if i == j {
				continue
			}

			for k := range p2 {
				a, b := p2[k], p2[(k+1)%len(p2)]
				if a == b {
					continue
				}

				minY := math.Min(a[1], b[1])
				maxY := math.Max(a[1], b[1])

				if pt[1] <= minY || pt[1] > maxY {
					continue
				}

				r := (pt[1] - a[1]) / (b[1] - a[1])
				x := (b[0]-a[0])*r + a[0]
				if x <= pt[0] {
					continue
				}

				found := false
				for l := range cinf[i].cuts {
					if cinf[i].cuts[l].idx == j {
						cinf[i].cuts[l].count++
						found = true
						break
					}
				}
				if !found {
					cinf[i].cuts = append(cinf[i].cuts, cut{idx: j, count: 1})
				}
				cinf[i].cutTotal++
			}
		}
	}

	// any contour with an even number of cuts is outer,
	// odd number of cuts means it is a hole
	for i := range cinf {
		cinf[i].outer = cinf[i].cutTotal%2 == 0
	}

	// go through them again, pick any outer contour, then
	// find any hole where the first outer contour it cuts
	// an odd number of times is the picked contour and add
	// it to the list of its holes
	result := make([]int, 0, len(contours)*2)
	for i := range cinf {
		if !cinf[i].outer {
			continue
		}
		result = append(result, i)

		for j := range cinf {
			if cinf[j].outer {
				continue
			}
			for _, cut := range cinf[j].cuts {
				if cut.count%2 == 0 {
					continue
				}
				if cut.idx == i {
					result = append(result, j)
					break
				}
			}
		}

		result = append(result, -1)
	}

	return result
}
