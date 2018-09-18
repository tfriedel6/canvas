package canvas

import (
	"math"
	"unsafe"
)

type pathPoint struct {
	pos   vec
	tf    vec
	next  vec
	flags pathPointFlag
}

type pathPointFlag uint8

const (
	pathMove pathPointFlag = 1 << iota
	pathAttach
	pathIsRect
	pathIsConvex
)

// BeginPath clears the current path and starts a new one
func (cv *Canvas) BeginPath() {
	if cv.path == nil {
		cv.path = make([]pathPoint, 0, 100)
	}
	cv.path = cv.path[:0]
}

func isSamePoint(a, b vec, maxDist float64) bool {
	return math.Abs(b[0]-a[0]) <= maxDist && math.Abs(b[1]-a[1]) <= maxDist
}

// MoveTo adds a gap and moves the end of the path to x/y
func (cv *Canvas) MoveTo(x, y float64) {
	tf := cv.tf(vec{x, y})
	if len(cv.path) > 0 && isSamePoint(cv.path[len(cv.path)-1].tf, tf, 0.1) {
		return
	}
	cv.path = append(cv.path, pathPoint{pos: vec{x, y}, tf: tf, flags: pathMove})
}

// LineTo adds a line to the end of the path
func (cv *Canvas) LineTo(x, y float64) {
	// cv.strokeLineTo(x, y)
	// cv.fillLineTo(x, y)

	if len(cv.path) > 0 && isSamePoint(cv.path[len(cv.path)-1].tf, cv.tf(vec{x, y}), 0.1) {
		return
	}
	if len(cv.path) == 0 {
		cv.path = append(cv.path, pathPoint{pos: vec{x, y}, tf: cv.tf(vec{x, y}), flags: pathMove})
		return
	}
	tf := cv.tf(vec{x, y})
	cv.path[len(cv.path)-1].next = tf
	cv.path[len(cv.path)-1].flags |= pathAttach
	cv.path = append(cv.path, pathPoint{pos: vec{x, y}, tf: tf})
}

// Arc adds a circle segment to the end of the path. x/y is the center, radius
// is the radius, startAngle and endAngle are angles in radians, anticlockwise
// means that the line is added anticlockwise
func (cv *Canvas) Arc(x, y, radius, startAngle, endAngle float64, anticlockwise bool) {
	lastWasMove := len(cv.path) == 0 || cv.path[len(cv.path)-1].flags&pathMove != 0

	startAngle = math.Mod(startAngle, math.Pi*2)
	if startAngle < 0 {
		startAngle += math.Pi * 2
	}
	endAngle = math.Mod(endAngle, math.Pi*2)
	if endAngle < 0 {
		endAngle += math.Pi * 2
	}
	if !anticlockwise && endAngle <= startAngle {
		endAngle += math.Pi * 2
	} else if anticlockwise && endAngle >= startAngle {
		endAngle -= math.Pi * 2
	}
	tr := cv.tf(vec{radius, radius})
	step := 6 / math.Max(tr[0], tr[1])
	if step > 0.8 {
		step = 0.8
	} else if step < 0.05 {
		step = 0.05
	}
	if anticlockwise {
		for a := startAngle; a > endAngle; a -= step {
			s, c := math.Sincos(a)
			cv.LineTo(x+radius*c, y+radius*s)
		}
	} else {
		for a := startAngle; a < endAngle; a += step {
			s, c := math.Sincos(a)
			cv.LineTo(x+radius*c, y+radius*s)
		}
	}
	s, c := math.Sincos(endAngle)
	cv.LineTo(x+radius*c, y+radius*s)

	if lastWasMove {
		cv.path[len(cv.path)-1].flags |= pathIsConvex
	}
}

// ArcTo adds to the current path by drawing a line toward x1/y1 and a circle
// segment of a radius given by the radius parameter. The circle touches the
// lines from the end of the path to x1/y1, and from x1/y1 to x2/y2. The line
// will only go to where the circle segment would touch the latter line
func (cv *Canvas) ArcTo(x1, y1, x2, y2, radius float64) {
	if len(cv.path) == 0 {
		return
	}
	p0, p1, p2 := cv.path[len(cv.path)-1].pos, vec{x1, y1}, vec{x2, y2}
	v0, v1 := p0.sub(p1).norm(), p2.sub(p1).norm()
	angle := math.Acos(v0.dot(v1))
	// should be in the range [0-pi]. if parallel, use a straight line
	if angle <= 0 || angle >= math.Pi {
		cv.LineTo(x2, y2)
		return
	}
	// cv are the vectors orthogonal to the lines that point to the center of the circle
	cv0 := vec{-v0[1], v0[0]}
	cv1 := vec{v1[1], -v1[0]}
	x := cv1.sub(cv0).div(v0.sub(v1))[0] * radius
	if x < 0 {
		cv0 = cv0.mulf(-1)
		cv1 = cv1.mulf(-1)
	}
	center := p1.add(v0.mulf(math.Abs(x))).add(cv0.mulf(radius))
	a0, a1 := cv0.mulf(-1).atan2(), cv1.mulf(-1).atan2()
	cv.Arc(center[0], center[1], radius, a0, a1, x > 0)
}

// QuadraticCurveTo adds a quadratic curve to the path. It uses the current end
// point of the path, x1/y1 defines the curve, and x2/y2 is the end point
func (cv *Canvas) QuadraticCurveTo(x1, y1, x2, y2 float64) {
	if len(cv.path) == 0 {
		return
	}
	p0 := cv.path[len(cv.path)-1].pos
	p1 := vec{x1, y1}
	p2 := vec{x2, y2}
	v0 := p1.sub(p0)
	v1 := p2.sub(p1)

	tp0, tp1, tp2 := cv.tf(p0), cv.tf(p1), cv.tf(p2)
	tv0 := tp1.sub(tp0)
	tv1 := tp2.sub(tp1)

	step := 1 / math.Max(math.Max(tv0[0], tv0[1]), math.Max(tv1[0], tv1[1]))
	if step > 0.1 {
		step = 0.1
	} else if step < 0.005 {
		step = 0.005
	}

	for r := 0.0; r < 1; r += step {
		i0 := v0.mulf(r).add(p0)
		i1 := v1.mulf(r).add(p1)
		p := i1.sub(i0).mulf(r).add(i0)
		cv.LineTo(p[0], p[1])
	}
}

// BezierCurveTo adds a bezier curve to the path. It uses the current end point
// of the path, x1/y1 and x2/y2 define the curve, and x3/y3 is the end point
func (cv *Canvas) BezierCurveTo(x1, y1, x2, y2, x3, y3 float64) {
	if len(cv.path) == 0 {
		return
	}
	p0 := cv.path[len(cv.path)-1].pos
	p1 := vec{x1, y1}
	p2 := vec{x2, y2}
	p3 := vec{x3, y3}
	v0 := p1.sub(p0)
	v1 := p2.sub(p1)
	v2 := p3.sub(p2)

	tp0, tp1, tp2, tp3 := cv.tf(p0), cv.tf(p1), cv.tf(p2), cv.tf(p3)
	tv0 := tp1.sub(tp0)
	tv1 := tp2.sub(tp1)
	tv2 := tp3.sub(tp2)

	step := 1 / math.Max(math.Max(math.Max(tv0[0], tv0[1]), math.Max(tv1[0], tv1[1])), math.Max(tv2[0], tv2[1]))
	if step > 0.1 {
		step = 0.1
	} else if step < 0.005 {
		step = 0.005
	}

	for r := 0.0; r < 1; r += step {
		i0 := v0.mulf(r).add(p0)
		i1 := v1.mulf(r).add(p1)
		i2 := v2.mulf(r).add(p2)
		iv0 := i1.sub(i0)
		iv1 := i2.sub(i1)
		j0 := iv0.mulf(r).add(i0)
		j1 := iv1.mulf(r).add(i1)
		p := j1.sub(j0).mulf(r).add(j0)
		cv.LineTo(p[0], p[1])
	}
}

// ClosePath closes the path to the beginning of the path or the last point
// from a MoveTo call
func (cv *Canvas) ClosePath() {
	if len(cv.path) < 2 {
		return
	}
	if isSamePoint(cv.path[len(cv.path)-1].tf, cv.path[0].tf, 0.1) {
		return
	}
	closeIdx := 0
	for i := len(cv.path) - 1; i >= 0; i-- {
		if cv.path[i].flags&pathMove != 0 {
			closeIdx = i
			break
		}
	}
	cv.LineTo(cv.path[closeIdx].pos[0], cv.path[closeIdx].pos[1])
	cv.path[len(cv.path)-1].next = cv.path[closeIdx].next
	cv.path[len(cv.path)-1].flags |= pathAttach
}

// Stroke uses the current StrokeStyle to draw the path
func (cv *Canvas) Stroke() {
	cv.stroke(cv.path)
}

func (cv *Canvas) stroke(path []pathPoint) {
	if len(path) == 0 {
		return
	}

	cv.activate()

	path = cv.applyLineDash(path)

	var triBuf [1000]float32
	tris := triBuf[:0]
	tris = append(tris, 0, 0, float32(cv.fw), 0, float32(cv.fw), float32(cv.fh), 0, 0, float32(cv.fw), float32(cv.fh), 0, float32(cv.fh))

	start := true
	var p0 vec
	for _, p := range path {
		if p.flags&pathMove != 0 {
			p0 = p.tf
			start = true
			continue
		}
		p1 := p.tf

		v0 := p1.sub(p0).norm()
		v1 := vec{v0[1], -v0[0]}.mulf(cv.state.lineWidth * 0.5)
		v0 = v0.mulf(cv.state.lineWidth * 0.5)

		lp0 := p0.add(v1)
		lp1 := p1.add(v1)
		lp2 := p0.sub(v1)
		lp3 := p1.sub(v1)

		if start {
			switch cv.state.lineEnd {
			case Butt:
				// no need to do anything
			case Square:
				lp0 = lp0.sub(v0)
				lp2 = lp2.sub(v0)
			case Round:
				tris = cv.addCircleTris(p0, cv.state.lineWidth*0.5, tris)
			}
		}

		if p.flags&pathAttach == 0 {
			switch cv.state.lineEnd {
			case Butt:
				// no need to do anything
			case Square:
				lp1 = lp1.add(v0)
				lp3 = lp3.add(v0)
			case Round:
				tris = cv.addCircleTris(p1, cv.state.lineWidth*0.5, tris)
			}
		}

		tris = append(tris,
			float32(lp0[0]), float32(lp0[1]), float32(lp1[0]), float32(lp1[1]), float32(lp3[0]), float32(lp3[1]),
			float32(lp0[0]), float32(lp0[1]), float32(lp3[0]), float32(lp3[1]), float32(lp2[0]), float32(lp2[1]))

		if p.flags&pathAttach != 0 && cv.state.lineWidth > 1 {
			tris = cv.lineJoint(p, p0, p1, p.next, lp0, lp1, lp2, lp3, tris)
		}

		p0 = p1
		start = false
	}

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)

	cv.drawShadow(tris)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	if cv.state.globalAlpha >= 1 && cv.state.lineAlpha >= 1 && cv.state.stroke.isOpaque() {
		vertex := cv.useShader(&cv.state.stroke)

		gli.EnableVertexAttribArray(vertex)
		gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, 0)
		gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))
		gli.DisableVertexAttribArray(vertex)
	} else {
		gli.ColorMask(false, false, false, false)
		gli.StencilFunc(gl_ALWAYS, 1, 0xFF)
		gli.StencilOp(gl_REPLACE, gl_REPLACE, gl_REPLACE)
		gli.StencilMask(0x01)

		gli.UseProgram(sr.id)
		gli.Uniform4f(sr.color, 0, 0, 0, 0)
		gli.Uniform2f(sr.canvasSize, float32(cv.fw), float32(cv.fh))

		gli.EnableVertexAttribArray(sr.vertex)
		gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, 0)
		gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))
		gli.DisableVertexAttribArray(sr.vertex)

		gli.ColorMask(true, true, true, true)

		gli.StencilFunc(gl_EQUAL, 1, 0xFF)

		origAlpha := cv.state.globalAlpha
		if cv.state.lineAlpha < 1 {
			cv.state.globalAlpha *= cv.state.lineAlpha
		}
		vertex := cv.useShader(&cv.state.stroke)
		cv.state.globalAlpha = origAlpha

		gli.EnableVertexAttribArray(vertex)
		gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, 0)
		gli.DrawArrays(gl_TRIANGLES, 0, 6)
		gli.DisableVertexAttribArray(vertex)

		gli.StencilOp(gl_KEEP, gl_KEEP, gl_KEEP)
		gli.StencilFunc(gl_ALWAYS, 0, 0xFF)

		gli.Clear(gl_STENCIL_BUFFER_BIT)
		gli.StencilMask(0xFF)
	}
}

func (cv *Canvas) applyLineDash(path []pathPoint) []pathPoint {
	if len(cv.state.lineDash) < 2 || len(path) < 2 {
		return path
	}

	ldo := cv.state.lineDashOffset
	ldp := cv.state.lineDashPoint

	path2 := make([]pathPoint, 0, len(path)*2)

	var lp vec
	for i, pp := range path {
		if i == 0 || pp.flags&pathMove != 0 {
			path2 = append(path2, pp)
			lp = pp.tf
			continue
		}

		tp := pp.tf
		v := tp.sub(lp)
		vl := v.len()
		prev := ldo
		for vl > 0 {
			draw := ldp%2 == 0
			p := tp
			ldo += vl
			if ldo > cv.state.lineDash[ldp] {
				ldo = 0
				dl := cv.state.lineDash[ldp] - prev
				p = lp.add(v.mulf(dl / vl))
				vl -= dl
				ldp++
				ldp %= len(cv.state.lineDash)
				prev = 0
			} else {
				vl = 0
			}

			if draw {
				path2[len(path2)-1].next = cv.tf(p)
				path2[len(path2)-1].flags |= pathAttach
				path2 = append(path2, pathPoint{pos: p, tf: cv.tf(p)})
			} else {
				path2 = append(path2, pathPoint{pos: p, tf: cv.tf(p), flags: pathMove})
			}

			lp = p
			v = tp.sub(lp)
		}
		lp = tp
	}

	return path2
}

func (cv *Canvas) lineJoint(p pathPoint, p0, p1, p2, l0p0, l0p1, l0p2, l0p3 vec, tris []float32) []float32 {
	v2 := p1.sub(p2).norm()
	v3 := vec{v2[1], -v2[0]}.mulf(cv.state.lineWidth * 0.5)

	switch cv.state.lineJoin {
	case Miter:
		l1p0 := p2.sub(v3)
		l1p1 := p1.sub(v3)
		l1p2 := p2.add(v3)
		l1p3 := p1.add(v3)

		var ip0, ip1 vec
		if l0p1.sub(l1p1).lenSqr() < 0.000000001 {
			ip0 = l0p1.sub(l1p1).mulf(0.5).add(l1p1)
		} else {
			var q float64
			ip0, _, q = lineIntersection(l0p0, l0p1, l1p1, l1p0)
			if q >= 1 {
				ip0 = l0p1.add(l1p1).mulf(0.5)
			}
		}

		if dist := ip0.sub(l0p1).lenSqr(); dist > cv.state.miterLimitSqr {
			l1p1 := p1.sub(v3)
			l1p3 := p1.add(v3)

			tris = append(tris,
				float32(p1[0]), float32(p1[1]), float32(l0p1[0]), float32(l0p1[1]), float32(l1p1[0]), float32(l1p1[1]),
				float32(p1[0]), float32(p1[1]), float32(l1p3[0]), float32(l1p3[1]), float32(l0p3[0]), float32(l0p3[1]))
			return tris
		}

		if l0p3.sub(l1p3).lenSqr() < 0.000000001 {
			ip1 = l0p3.sub(l1p3).mulf(0.5).add(l1p3)
		} else {
			var q float64
			ip1, _, q = lineIntersection(l0p2, l0p3, l1p3, l1p2)
			if q >= 1 {
				ip1 = l0p3.add(l1p3).mulf(0.5)
			}
		}

		if dist := ip1.sub(l1p1).lenSqr(); dist > cv.state.miterLimitSqr {
			l1p1 := p1.sub(v3)
			l1p3 := p1.add(v3)

			tris = append(tris,
				float32(p1[0]), float32(p1[1]), float32(l0p1[0]), float32(l0p1[1]), float32(l1p1[0]), float32(l1p1[1]),
				float32(p1[0]), float32(p1[1]), float32(l1p3[0]), float32(l1p3[1]), float32(l0p3[0]), float32(l0p3[1]))
			return tris
		}

		tris = append(tris,
			float32(p1[0]), float32(p1[1]), float32(l0p1[0]), float32(l0p1[1]), float32(ip0[0]), float32(ip0[1]),
			float32(p1[0]), float32(p1[1]), float32(ip0[0]), float32(ip0[1]), float32(l1p1[0]), float32(l1p1[1]),
			float32(p1[0]), float32(p1[1]), float32(l1p3[0]), float32(l1p3[1]), float32(ip1[0]), float32(ip1[1]),
			float32(p1[0]), float32(p1[1]), float32(ip1[0]), float32(ip1[1]), float32(l0p3[0]), float32(l0p3[1]))
	case Bevel:
		l1p1 := p1.sub(v3)
		l1p3 := p1.add(v3)

		tris = append(tris,
			float32(p1[0]), float32(p1[1]), float32(l0p1[0]), float32(l0p1[1]), float32(l1p1[0]), float32(l1p1[1]),
			float32(p1[0]), float32(p1[1]), float32(l1p3[0]), float32(l1p3[1]), float32(l0p3[0]), float32(l0p3[1]))
	case Round:
		tris = cv.addCircleTris(p1, cv.state.lineWidth*0.5, tris)
	}

	return tris
}

func (cv *Canvas) addCircleTris(center vec, radius float64, tris []float32) []float32 {
	p0 := vec{center[0], center[1] + radius}
	step := 6 / radius
	if step > 0.8 {
		step = 0.8
	} else if step < 0.05 {
		step = 0.05
	}
	for angle := step; angle <= math.Pi*2+step; angle += step {
		s, c := math.Sincos(angle)
		p1 := vec{center[0] + s*radius, center[1] + c*radius}
		tris = append(tris, float32(center[0]), float32(center[1]), float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]))
		p0 = p1
	}
	return tris
}

func lineIntersection(a0, a1, b0, b1 vec) (vec, float64, float64) {
	va := a1.sub(a0)
	vb := b1.sub(b0)

	if (va[0] == 0 && vb[0] == 0) || (va[1] == 0 && vb[1] == 0) || (va[0] == 0 && va[1] == 0) || (vb[0] == 0 && vb[1] == 0) {
		return vec{}, float64(math.Inf(1)), float64(math.Inf(1))
	}
	d := va[1]*vb[0] - va[0]*vb[1]
	if d == 0 {
		return vec{}, float64(math.Inf(1)), float64(math.Inf(1))
	}
	p := (vb[1]*(a0[0]-b0[0]) - a0[1]*vb[0] + b0[1]*vb[0]) / d
	var q float64
	if vb[0] == 0 {
		q = (a0[1] + p*va[1] - b0[1]) / vb[1]
	} else {
		q = (a0[0] + p*va[0] - b0[0]) / vb[0]
	}

	return a0.add(va.mulf(p)), p, q
}

// Fill fills the current path with the given FillStyle
func (cv *Canvas) Fill() {
	if len(cv.path) < 3 {
		return
	}
	cv.activate()

	var triBuf [1000]float32
	tris := triBuf[:0]
	tris = append(tris, 0, 0, float32(cv.fw), 0, float32(cv.fw), float32(cv.fh), 0, 0, float32(cv.fw), float32(cv.fh), 0, float32(cv.fh))

	start := 0
	for i, p := range cv.path {
		if p.flags&pathMove == 0 {
			continue
		}
		if i >= start+3 {
			tris = cv.appendSubPathTriangles(tris, cv.path[start:i])
		}
		start = i
	}
	if len(cv.path) >= start+3 {
		tris = cv.appendSubPathTriangles(tris, cv.path[start:])
	}
	if len(tris) == 0 {
		return
	}

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)

	cv.drawShadow(tris)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	if cv.state.globalAlpha >= 1 && cv.state.lineAlpha >= 1 && cv.state.fill.isOpaque() {
		vertex := cv.useShader(&cv.state.fill)

		gli.EnableVertexAttribArray(vertex)
		gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, 0)
		gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))
		gli.DisableVertexAttribArray(vertex)
	} else {
		gli.ColorMask(false, false, false, false)
		gli.StencilFunc(gl_ALWAYS, 1, 0xFF)
		gli.StencilOp(gl_REPLACE, gl_REPLACE, gl_REPLACE)
		gli.StencilMask(0x01)

		gli.UseProgram(sr.id)
		gli.Uniform4f(sr.color, 0, 0, 0, 0)
		gli.Uniform2f(sr.canvasSize, float32(cv.fw), float32(cv.fh))

		gli.EnableVertexAttribArray(sr.vertex)
		gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, 0)
		gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))
		gli.DisableVertexAttribArray(sr.vertex)

		gli.ColorMask(true, true, true, true)

		gli.StencilFunc(gl_EQUAL, 1, 0xFF)

		vertex := cv.useShader(&cv.state.fill)
		gli.EnableVertexAttribArray(vertex)
		gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, 0)
		gli.DrawArrays(gl_TRIANGLES, 0, 6)
		gli.DisableVertexAttribArray(vertex)

		gli.StencilOp(gl_KEEP, gl_KEEP, gl_KEEP)
		gli.StencilFunc(gl_ALWAYS, 0, 0xFF)

		gli.Clear(gl_STENCIL_BUFFER_BIT)
		gli.StencilMask(0xFF)
	}
}

func (cv *Canvas) appendSubPathTriangles(tris []float32, path []pathPoint) []float32 {
	if path[len(path)-1].flags&pathIsConvex != 0 {
		p0, p1 := path[0].tf, path[1].tf
		last := len(path)
		for i := 2; i < last; i++ {
			p2 := path[i].tf
			tris = append(tris, float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]))
			p1 = p2
		}
	} else {
		path = cv.cutIntersections(path)
		tris = triangulatePath(path, tris)
	}
	return tris
}

// Clip uses the current path to clip any further drawing. Use Save/Restore to
// remove the clipping again
func (cv *Canvas) Clip() {
	if len(cv.path) < 3 {
		return
	}

	path := cv.path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i].flags&pathMove != 0 {
			path = path[i:]
			break
		}
	}

	cv.clip(path)
}

func (cv *Canvas) clip(path []pathPoint) {
	if len(path) < 3 {
		return
	}
	if path[len(path)-1].flags&pathIsRect != 0 {
		cv.scissor(path)
		return
	}

	cv.activate()

	var triBuf [1000]float32
	tris := triBuf[:0]
	tris = append(tris, 0, 0, float32(cv.fw), 0, float32(cv.fw), float32(cv.fh), 0, 0, float32(cv.fw), float32(cv.fh), 0, float32(cv.fh))
	baseLen := len(tris)
	tris = triangulatePath(path, tris)
	if len(tris) <= baseLen {
		return
	}

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)
	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, 0)

	gli.UseProgram(sr.id)
	gli.Uniform4f(sr.color, 1, 1, 1, 1)
	gli.Uniform2f(sr.canvasSize, float32(cv.fw), float32(cv.fh))
	gli.EnableVertexAttribArray(sr.vertex)

	gli.ColorMask(false, false, false, false)

	gli.StencilMask(0x04)
	gli.StencilFunc(gl_ALWAYS, 4, 0x04)
	gli.StencilOp(gl_REPLACE, gl_REPLACE, gl_REPLACE)
	gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))

	gli.StencilMask(0x02)
	gli.StencilFunc(gl_EQUAL, 0, 0x06)
	gli.StencilOp(gl_KEEP, gl_INVERT, gl_INVERT)
	gli.DrawArrays(gl_TRIANGLES, 0, 6)

	gli.StencilMask(0x04)
	gli.StencilFunc(gl_ALWAYS, 0, 0x04)
	gli.StencilOp(gl_ZERO, gl_ZERO, gl_ZERO)
	gli.DrawArrays(gl_TRIANGLES, 0, 6)

	gli.DisableVertexAttribArray(sr.vertex)

	gli.ColorMask(true, true, true, true)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_KEEP)
	gli.StencilMask(0xFF)
	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	cv.state.clip = make([]pathPoint, len(cv.path))
	copy(cv.state.clip, cv.path)
}

func (cv *Canvas) scissor(path []pathPoint) {
	tl, br := vec{math.MaxFloat64, math.MaxFloat64}, vec{}
	for _, p := range path {
		tl[0] = math.Min(p.tf[0], tl[0])
		tl[1] = math.Min(p.tf[1], tl[1])
		br[0] = math.Max(p.tf[0], br[0])
		br[1] = math.Max(p.tf[1], br[1])
	}

	if cv.state.scissor.on {
		tl[0] = math.Max(tl[0], cv.state.scissor.tl[0])
		tl[1] = math.Max(tl[1], cv.state.scissor.tl[1])
		br[0] = math.Min(br[0], cv.state.scissor.br[0])
		br[1] = math.Min(br[1], cv.state.scissor.br[1])
	}

	cv.state.scissor = scissor{tl: tl, br: br, on: true}
	cv.applyScissor()
}

func (cv *Canvas) applyScissor() {
	s := &cv.state.scissor
	if s.on {
		gli.Scissor(int32(s.tl[0]+0.5), int32(cv.fh-s.br[1]+0.5), int32(s.br[0]-s.tl[0]+0.5), int32(s.br[1]-s.tl[1]+0.5))
	} else {
		gli.Scissor(0, 0, int32(cv.w), int32(cv.h))
	}
}

// Rect creates a closed rectangle path for stroking or filling
func (cv *Canvas) Rect(x, y, w, h float64) {
	cv.MoveTo(x, y)
	cv.LineTo(x+w, y)
	cv.LineTo(x+w, y+h)
	cv.LineTo(x, y+h)
	cv.ClosePath()
	cv.path[len(cv.path)-1].flags |= pathIsRect
}

// StrokeRect draws a rectangle using the current stroke style
func (cv *Canvas) StrokeRect(x, y, w, h float64) {
	v0 := vec{x, y}
	v1 := vec{x + w, y}
	v2 := vec{x + w, y + h}
	v3 := vec{x, y + h}
	v0t, v1t, v2t, v3t := cv.tf(v0), cv.tf(v1), cv.tf(v2), cv.tf(v3)
	var path [5]pathPoint
	path[0] = pathPoint{pos: v0, tf: v0t, flags: pathMove | pathAttach, next: v1t}
	path[1] = pathPoint{pos: v1, tf: v1t, next: v2t, flags: pathAttach}
	path[2] = pathPoint{pos: v2, tf: v2t, next: v3t, flags: pathAttach}
	path[3] = pathPoint{pos: v3, tf: v3t, next: v0t, flags: pathAttach}
	path[4] = pathPoint{pos: v0, tf: v0t, next: v1t, flags: pathAttach}
	cv.stroke(path[:])
}

// FillRect fills a rectangle with the active fill style
func (cv *Canvas) FillRect(x, y, w, h float64) {
	cv.activate()

	p0 := cv.tf(vec{x, y})
	p1 := cv.tf(vec{x, y + h})
	p2 := cv.tf(vec{x + w, y + h})
	p3 := cv.tf(vec{x + w, y})

	if cv.state.shadowColor.a != 0 {
		tris := [24]float32{
			0, 0,
			float32(cv.fw), 0,
			float32(cv.fw), float32(cv.fh),
			0, 0,
			float32(cv.fw), float32(cv.fh),
			0, float32(cv.fh),
			float32(p0[0]), float32(p0[1]),
			float32(p3[0]), float32(p3[1]),
			float32(p2[0]), float32(p2[1]),
			float32(p0[0]), float32(p0[1]),
			float32(p2[0]), float32(p2[1]),
			float32(p1[0]), float32(p1[1]),
		}
		cv.drawShadow(tris[:])
	}

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [8]float32{float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), float32(p3[0]), float32(p3[1])}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	vertex := cv.useShader(&cv.state.fill)
	gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, 0)
	gli.EnableVertexAttribArray(vertex)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(vertex)

	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)
}

// ClearRect sets the color of the rectangle to transparent black
func (cv *Canvas) ClearRect(x, y, w, h float64) {
	cv.activate()

	if cv.state.transform == matIdentity() {
		gli.Scissor(int32(x+0.5), int32(cv.fh-y-h+0.5), int32(w+0.5), int32(h+0.5))
		gli.ClearColor(0, 0, 0, 0)
		gli.Clear(gl_COLOR_BUFFER_BIT)
		cv.applyScissor()
		return
	}

	gli.UseProgram(sr.id)
	gli.Uniform2f(sr.canvasSize, float32(cv.fw), float32(cv.fh))
	gli.Uniform4f(sr.color, 0, 0, 0, 0)
	gli.Uniform1f(sr.globalAlpha, 1)

	gli.Disable(gl_BLEND)

	p0 := cv.tf(vec{x, y})
	p1 := cv.tf(vec{x, y + h})
	p2 := cv.tf(vec{x + w, y + h})
	p3 := cv.tf(vec{x + w, y})

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [8]float32{float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), float32(p3[0]), float32(p3[1])}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, 0)
	gli.EnableVertexAttribArray(sr.vertex)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(sr.vertex)

	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)

	gli.Enable(gl_BLEND)
}
