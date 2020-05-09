package canvas

import (
	"math"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

// Path2D is a type that holds a predefined path which can be drawn
// with a single call
type Path2D struct {
	cv    *Canvas
	p     []pathPoint
	move  backendbase.Vec
	cwSum float64

	standalone bool
	fillCache  []backendbase.Vec

	noSelfIntersection bool
}

type pathPoint struct {
	pos   backendbase.Vec
	next  backendbase.Vec
	flags pathPointFlag
}

type pathPointFlag uint8

const (
	pathMove pathPointFlag = 1 << iota
	pathAttach
	pathIsRect
	pathIsConvex
	pathIsClockwise
	pathSelfIntersects
)

// NewPath2D creates a new Path2D and returns it
func (cv *Canvas) NewPath2D() *Path2D {
	return &Path2D{cv: cv, p: make([]pathPoint, 0, 20), standalone: true}
}

func (p *Path2D) clearCache() {
	p.fillCache = nil
}

// func (p *Path2D) AddPath(p2 *Path2D) {
// }

// MoveTo (see equivalent function on canvas type)
func (p *Path2D) MoveTo(x, y float64) {
	if len(p.p) > 0 && isSamePoint(p.p[len(p.p)-1].pos, backendbase.Vec{x, y}, 0.1) {
		return
	}
	p.clearCache()
	p.p = append(p.p, pathPoint{pos: backendbase.Vec{x, y}, flags: pathMove | pathIsConvex})
	p.cwSum = 0
	p.move = backendbase.Vec{x, y}
}

// LineTo (see equivalent function on canvas type)
func (p *Path2D) LineTo(x, y float64) {
	p.lineTo(x, y, true)
}

func (p *Path2D) lineTo(x, y float64, checkSelfIntersection bool) {
	count := len(p.p)
	if count > 0 && isSamePoint(p.p[len(p.p)-1].pos, backendbase.Vec{x, y}, 0.1) {
		return
	}
	p.clearCache()
	if count == 0 {
		p.MoveTo(x, y)
		return
	}
	prev := &p.p[count-1]
	prev.next = backendbase.Vec{x, y}
	prev.flags |= pathAttach
	p.p = append(p.p, pathPoint{pos: backendbase.Vec{x, y}})
	newp := &p.p[count]

	if prev.flags&pathIsConvex > 0 {
		px, py := prev.pos[0], prev.pos[1]
		p.cwSum += (x - px) * (y + py)
		cwTotal := p.cwSum
		cwTotal += (p.move[0] - x) * (p.move[1] + y)
		if cwTotal <= 0 {
			newp.flags |= pathIsClockwise
		}
	}

	if len(p.p) < 4 || Performance.AssumeConvex {
		newp.flags |= pathIsConvex
	} else if prev.flags&pathIsConvex > 0 {
		prev2 := &p.p[count-2]
		cw := (prev.flags & pathIsClockwise) > 0

		ln := prev.pos.Sub(prev2.pos)
		lo := backendbase.Vec{ln[1], -ln[0]}
		dot := newp.pos.Sub(prev2.pos).Dot(lo)

		if (cw && dot <= 0) || (!cw && dot >= 0) {
			newp.flags |= pathIsConvex
		}
	}

	csi := checkSelfIntersection && !Performance.IgnoreSelfIntersections && !p.noSelfIntersection

	if prev.flags&pathSelfIntersects > 0 {
		newp.flags |= pathSelfIntersects
	} else if newp.flags&pathIsConvex == 0 && newp.flags&pathSelfIntersects == 0 && csi {

		cuts := false
		var cutPoint backendbase.Vec
		b0, b1 := prev.pos, backendbase.Vec{x, y}
		for i := 1; i < count; i++ {
			a0, a1 := p.p[i-1].pos, p.p[i].pos
			var r1, r2 float64
			cutPoint, r1, r2 = lineIntersection(a0, a1, b0, b1)
			if r1 > 0 && r1 < 1 && r2 > 0 && r2 < 1 {
				cuts = true
				break
			}
		}
		if cuts && !isSamePoint(cutPoint, backendbase.Vec{x, y}, samePointTolerance) {
			newp.flags |= pathSelfIntersects
		}
	}
}

// Arc (see equivalent function on canvas type)
func (p *Path2D) Arc(x, y, radius, startAngle, endAngle float64, anticlockwise bool) {
	p.arc(x, y, radius, startAngle, endAngle, anticlockwise, backendbase.MatIdentity, true)
}

func (p *Path2D) arc(x, y, radius, startAngle, endAngle float64, anticlockwise bool, m backendbase.Mat, ident bool) {
	checkSelfIntersection := len(p.p) > 0

	lastWasMove := len(p.p) == 0 || p.p[len(p.p)-1].flags&pathMove != 0

	if endAngle == startAngle {
		s, c := math.Sincos(endAngle)
		pt := backendbase.Vec{x + radius*c, y + radius*s}
		if !ident {
			pt = pt.MulMat(m)
		}
		p.lineTo(pt[0], pt[1], checkSelfIntersection)

		if lastWasMove {
			p.p[len(p.p)-1].flags |= pathIsConvex
		}

		return
	}

	if !anticlockwise && endAngle < startAngle {
		endAngle = startAngle + (2*math.Pi - math.Mod(startAngle-endAngle, math.Pi*2))
	} else if anticlockwise && endAngle > startAngle {
		endAngle = startAngle - (2*math.Pi - math.Mod(endAngle-startAngle, math.Pi*2))
	}

	if !anticlockwise {
		diff := endAngle - startAngle
		if diff >= math.Pi*4 {
			diff = math.Mod(diff, math.Pi*2) + math.Pi*2
			endAngle = startAngle + diff
		}
	} else {
		diff := startAngle - endAngle
		if diff >= math.Pi*4 {
			diff = math.Mod(diff, math.Pi*2) + math.Pi*2
			endAngle = startAngle - diff
		}
	}

	const step = math.Pi * 2 / 90
	if !anticlockwise {
		for a := startAngle; a < endAngle; a += step {
			s, c := math.Sincos(a)
			pt := backendbase.Vec{x + radius*c, y + radius*s}
			if !ident {
				pt = pt.MulMat(m)
			}
			p.lineTo(pt[0], pt[1], checkSelfIntersection)
		}
	} else {
		for a := startAngle; a > endAngle; a -= step {
			s, c := math.Sincos(a)
			pt := backendbase.Vec{x + radius*c, y + radius*s}
			if !ident {
				pt = pt.MulMat(m)
			}
			p.lineTo(pt[0], pt[1], checkSelfIntersection)
		}
	}
	s, c := math.Sincos(endAngle)
	pt := backendbase.Vec{x + radius*c, y + radius*s}
	if !ident {
		pt = pt.MulMat(m)
	}
	p.lineTo(pt[0], pt[1], checkSelfIntersection)

	if lastWasMove {
		p.p[len(p.p)-1].flags |= pathIsConvex
	}
}

// ArcTo (see equivalent function on canvas type)
func (p *Path2D) ArcTo(x1, y1, x2, y2, radius float64) {
	p.arcTo(x1, y1, x2, y2, radius, backendbase.MatIdentity, true)
}

func (p *Path2D) arcTo(x1, y1, x2, y2, radius float64, m backendbase.Mat, ident bool) {
	if len(p.p) == 0 {
		return
	}
	p0, p1, p2 := p.p[len(p.p)-1].pos, backendbase.Vec{x1, y1}, backendbase.Vec{x2, y2}
	if !ident {
		p0 = p0.MulMat(m.Invert())
	}
	v0, v1 := p0.Sub(p1).Norm(), p2.Sub(p1).Norm()
	angle := math.Acos(v0.Dot(v1))
	// should be in the range [0-pi]. if parallel, use a straight line
	if angle <= 0 || angle >= math.Pi {
		p.LineTo(x2, y2)
		return
	}
	// cv0 and cv1 are vectors that point to the center of the circle
	cv0 := backendbase.Vec{-v0[1], v0[0]}
	cv1 := backendbase.Vec{v1[1], -v1[0]}
	x := cv1.Sub(cv0).Div(v0.Sub(v1))[0] * radius
	if x < 0 {
		cv0 = cv0.Mulf(-1)
		cv1 = cv1.Mulf(-1)
	}
	center := p1.Add(v0.Mulf(math.Abs(x))).Add(cv0.Mulf(radius))
	a0, a1 := cv0.Mulf(-1).Atan2(), cv1.Mulf(-1).Atan2()
	if x > 0 {
		if a1-a0 > 0 {
			a0 += math.Pi * 2
		}
	} else {
		if a0-a1 > 0 {
			a1 += math.Pi * 2
		}
	}
	p.arc(center[0], center[1], radius, a0, a1, x > 0, m, ident)
}

// QuadraticCurveTo (see equivalent function on canvas type)
func (p *Path2D) QuadraticCurveTo(x1, y1, x2, y2 float64) {
	if len(p.p) == 0 {
		return
	}
	p0 := p.p[len(p.p)-1].pos
	p1 := backendbase.Vec{x1, y1}
	p2 := backendbase.Vec{x2, y2}
	v0 := p1.Sub(p0)
	v1 := p2.Sub(p1)

	const step = 0.01

	for r := 0.0; r < 1; r += step {
		i0 := v0.Mulf(r).Add(p0)
		i1 := v1.Mulf(r).Add(p1)
		pt := i1.Sub(i0).Mulf(r).Add(i0)
		p.LineTo(pt[0], pt[1])
	}
	p.LineTo(x2, y2)
}

// BezierCurveTo (see equivalent function on canvas type)
func (p *Path2D) BezierCurveTo(x1, y1, x2, y2, x3, y3 float64) {
	if len(p.p) == 0 {
		return
	}
	p0 := p.p[len(p.p)-1].pos
	p1 := backendbase.Vec{x1, y1}
	p2 := backendbase.Vec{x2, y2}
	p3 := backendbase.Vec{x3, y3}
	v0 := p1.Sub(p0)
	v1 := p2.Sub(p1)
	v2 := p3.Sub(p2)

	const step = 0.01

	for r := 0.0; r < 1; r += step {
		i0 := v0.Mulf(r).Add(p0)
		i1 := v1.Mulf(r).Add(p1)
		i2 := v2.Mulf(r).Add(p2)
		iv0 := i1.Sub(i0)
		iv1 := i2.Sub(i1)
		j0 := iv0.Mulf(r).Add(i0)
		j1 := iv1.Mulf(r).Add(i1)
		pt := j1.Sub(j0).Mulf(r).Add(j0)
		p.LineTo(pt[0], pt[1])
	}
	p.LineTo(x3, y3)
}

// Ellipse (see equivalent function on canvas type)
func (p *Path2D) Ellipse(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64, anticlockwise bool) {
	checkSelfIntersection := len(p.p) > 0

	rs, rc := math.Sincos(rotation)

	lastWasMove := len(p.p) == 0 || p.p[len(p.p)-1].flags&pathMove != 0

	if endAngle == startAngle {
		s, c := math.Sincos(endAngle)
		rx, ry := radiusX*c, radiusY*s
		rx, ry = rx*rc-ry*rs, rx*rs+ry*rc
		p.lineTo(x+rx, y+ry, checkSelfIntersection)

		if lastWasMove {
			p.p[len(p.p)-1].flags |= pathIsConvex
		}

		return
	}

	if !anticlockwise && endAngle < startAngle {
		endAngle = startAngle + (2*math.Pi - math.Mod(startAngle-endAngle, math.Pi*2))
	} else if anticlockwise && endAngle > startAngle {
		endAngle = startAngle - (2*math.Pi - math.Mod(endAngle-startAngle, math.Pi*2))
	}

	if !anticlockwise {
		diff := endAngle - startAngle
		if diff >= math.Pi*4 {
			diff = math.Mod(diff, math.Pi*2) + math.Pi*2
			endAngle = startAngle + diff
		}
	} else {
		diff := startAngle - endAngle
		if diff >= math.Pi*4 {
			diff = math.Mod(diff, math.Pi*2) + math.Pi*2
			endAngle = startAngle - diff
		}
	}

	const step = math.Pi * 2 / 90
	if !anticlockwise {
		for a := startAngle; a < endAngle; a += step {
			s, c := math.Sincos(a)
			rx, ry := radiusX*c, radiusY*s
			rx, ry = rx*rc-ry*rs, rx*rs+ry*rc
			p.lineTo(x+rx, y+ry, checkSelfIntersection)
		}
	} else {
		for a := startAngle; a > endAngle; a -= step {
			s, c := math.Sincos(a)
			rx, ry := radiusX*c, radiusY*s
			rx, ry = rx*rc-ry*rs, rx*rs+ry*rc
			p.lineTo(x+rx, y+ry, checkSelfIntersection)
		}
	}
	s, c := math.Sincos(endAngle)
	rx, ry := radiusX*c, radiusY*s
	rx, ry = rx*rc-ry*rs, rx*rs+ry*rc
	p.lineTo(x+rx, y+ry, checkSelfIntersection)

	if lastWasMove {
		p.p[len(p.p)-1].flags |= pathIsConvex
	}
}

// ClosePath (see equivalent function on canvas type)
func (p *Path2D) ClosePath() {
	if len(p.p) < 2 {
		return
	}
	closeIdx := 0
	for i := len(p.p) - 1; i >= 0; i-- {
		if p.p[i].flags&pathMove != 0 {
			closeIdx = i
			break
		}
	}
	if !isSamePoint(p.p[len(p.p)-1].pos, p.p[0].pos, 0.1) {
		p.LineTo(p.p[closeIdx].pos[0], p.p[closeIdx].pos[1])
	}
	p.p[len(p.p)-1].next = p.p[closeIdx].next
	p.p[len(p.p)-1].flags |= pathAttach
}

// Rect (see equivalent function on canvas type)
func (p *Path2D) Rect(x, y, w, h float64) {
	lastWasMove := len(p.p) == 0 || p.p[len(p.p)-1].flags&pathMove != 0
	p.MoveTo(x, y)
	p.LineTo(x+w, y)
	p.LineTo(x+w, y+h)
	p.LineTo(x, y+h)
	p.LineTo(x, y)
	if lastWasMove {
		p.p[len(p.p)-1].flags |= pathIsRect
		p.p[len(p.p)-1].flags |= pathIsConvex
	}
}

func runSubPaths(path []pathPoint, close bool, fn func(subPath []pathPoint) bool) {
	start := 0
	for i, p := range path {
		if p.flags&pathMove == 0 {
			continue
		}
		if i >= start+3 {
			if runSubPath(path[start:i], close, fn) {
				return
			}
		}
		start = i
	}
	if len(path) >= start+3 {
		runSubPath(path[start:], close, fn)
	}
}

func runSubPath(path []pathPoint, close bool, fn func(subPath []pathPoint) bool) bool {
	if !close || path[0].pos == path[len(path)-1].pos {
		return fn(path)
	}

	var buf [64]pathPoint
	path2 := Path2D{
		p:    append(buf[:0], path...),
		move: path[0].pos,
	}
	path2.lineTo(path[0].pos[0], path[0].pos[1], true)
	return fn(path2.p)
}

type pathRule uint8

// Path rule constants. See https://en.wikipedia.org/wiki/Nonzero-rule
// and https://en.wikipedia.org/wiki/Even%E2%80%93odd_rule
const (
	NonZero pathRule = iota
	EvenOdd
)

// IsPointInPath returns true if the point is in the path according
// to the given rule
func (p *Path2D) IsPointInPath(x, y float64, rule pathRule) bool {
	inside := false
	runSubPaths(p.p, false, func(sp []pathPoint) bool {
		num := 0
		prev := sp[len(sp)-1].pos
		for _, pt := range sp {
			r, dir := pointIsRightOfLine(prev, pt.pos, backendbase.Vec{x, y})
			prev = pt.pos
			if !r {
				continue
			}
			if dir {
				num++
			} else {
				num--
			}
		}

		if rule == NonZero {
			inside = num != 0
		} else {
			inside = num%2 == 0
		}

		return inside
	})
	return inside
}

// IsPointInStroke returns true if the point is in the stroke
func (p *Path2D) IsPointInStroke(x, y float64) bool {
	if len(p.p) == 0 {
		return false
	}

	var triBuf [500]backendbase.Vec
	tris := p.cv.strokeTris(p, p.cv.state.transform, backendbase.Mat{}, false, triBuf[:0])

	pt := backendbase.Vec{x, y}

	for i := 0; i < len(tris); i += 3 {
		a := backendbase.Vec{tris[i][0], tris[i][1]}
		b := backendbase.Vec{tris[i+1][0], tris[i+1][1]}
		c := backendbase.Vec{tris[i+2][0], tris[i+2][1]}
		if triangleContainsPoint(a, b, c, pt) {
			return true
		}
	}
	return false
}
