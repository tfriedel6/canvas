package canvas

import (
	"math"
	"unsafe"
)

func (cv *Canvas) BeginPath() {
	if cv.linePath == nil {
		cv.linePath = make([]pathPoint, 0, 100)
	}
	if cv.polyPath == nil {
		cv.polyPath = make([]pathPoint, 0, 100)
	}
	cv.linePath = cv.linePath[:0]
	cv.polyPath = cv.polyPath[:0]
}

func isSamePoint(a, b vec, maxDist float64) bool {
	return math.Abs(b[0]-a[0]) <= maxDist && math.Abs(b[1]-a[1]) <= maxDist
}

func (cv *Canvas) MoveTo(x, y float64) {
	tf := cv.tf(vec{x, y})
	if len(cv.linePath) > 0 && isSamePoint(cv.linePath[len(cv.linePath)-1].tf, tf, 0.1) {
		return
	}
	cv.linePath = append(cv.linePath, pathPoint{pos: vec{x, y}, tf: tf, move: true})
	cv.polyPath = append(cv.polyPath, pathPoint{pos: vec{x, y}, tf: tf, move: true})
}

func (cv *Canvas) LineTo(x, y float64) {
	cv.strokeLineTo(x, y)
	cv.fillLineTo(x, y)
}

func (cv *Canvas) strokeLineTo(x, y float64) {
	if len(cv.linePath) > 0 && isSamePoint(cv.linePath[len(cv.linePath)-1].tf, cv.tf(vec{x, y}), 0.1) {
		return
	}
	if len(cv.linePath) == 0 {
		cv.linePath = append(cv.linePath, pathPoint{pos: vec{x, y}, tf: cv.tf(vec{x, y}), move: true})
		return
	}
	if len(cv.state.lineDash) > 0 {
		lp := cv.linePath[len(cv.linePath)-1].pos
		tp := vec{x, y}
		v := tp.sub(lp)
		vl := v.len()
		prev := cv.state.lineDashOffset
		for vl > 0 {
			draw := cv.state.lineDashPoint%2 == 0
			p := tp
			cv.state.lineDashOffset += vl
			if cv.state.lineDashOffset > cv.state.lineDash[cv.state.lineDashPoint] {
				cv.state.lineDashOffset = 0
				dl := cv.state.lineDash[cv.state.lineDashPoint] - prev
				p = lp.add(v.mulf(dl / vl))
				vl -= dl
				cv.state.lineDashPoint++
				cv.state.lineDashPoint %= len(cv.state.lineDash)
				prev = 0
			} else {
				vl = 0
			}

			if draw {
				cv.linePath[len(cv.linePath)-1].next = cv.tf(p)
				cv.linePath[len(cv.linePath)-1].attach = true
				cv.linePath = append(cv.linePath, pathPoint{pos: p, tf: cv.tf(p), move: false})
			} else {
				cv.linePath = append(cv.linePath, pathPoint{pos: p, tf: cv.tf(p), move: true})
			}

			lp = p
			v = tp.sub(lp)
		}
	} else {
		tf := cv.tf(vec{x, y})
		cv.linePath[len(cv.linePath)-1].next = tf
		cv.linePath[len(cv.linePath)-1].attach = true
		cv.linePath = append(cv.linePath, pathPoint{pos: vec{x, y}, tf: tf, move: false})
	}
}

func (cv *Canvas) fillLineTo(x, y float64) {
	if len(cv.polyPath) > 0 && isSamePoint(cv.polyPath[len(cv.polyPath)-1].tf, cv.tf(vec{x, y}), 0.1) {
		return
	}
	if len(cv.polyPath) == 0 {
		cv.polyPath = append(cv.polyPath, pathPoint{pos: vec{x, y}, tf: cv.tf(vec{x, y}), move: true})
		return
	}
	tf := cv.tf(vec{x, y})
	cv.polyPath[len(cv.polyPath)-1].next = tf
	cv.polyPath[len(cv.polyPath)-1].attach = true
	cv.polyPath = append(cv.polyPath, pathPoint{pos: vec{x, y}, tf: tf, move: false})
}

func (cv *Canvas) Arc(x, y, radius, startAngle, endAngle float64, anticlockwise bool) {
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
}

func (cv *Canvas) ArcTo(x1, y1, x2, y2, radius float64) {
	if len(cv.linePath) == 0 {
		return
	}
	p0, p1, p2 := cv.linePath[len(cv.linePath)-1].pos, vec{x1, y1}, vec{x2, y2}
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

func (cv *Canvas) QuadraticCurveTo(x1, y1, x2, y2 float64) {
	if len(cv.linePath) == 0 {
		return
	}
	p0 := cv.linePath[len(cv.linePath)-1].pos
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

func (cv *Canvas) BezierCurveTo(x1, y1, x2, y2, x3, y3 float64) {
	if len(cv.linePath) == 0 {
		return
	}
	p0 := cv.linePath[len(cv.linePath)-1].pos
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

func (cv *Canvas) ClosePath() {
	if len(cv.linePath) < 2 {
		return
	}
	if isSamePoint(cv.linePath[len(cv.linePath)-1].tf, cv.linePath[0].tf, 0.1) {
		return
	}
	closeIdx := 0
	for i := len(cv.linePath) - 1; i >= 0; i-- {
		if cv.linePath[i].move {
			closeIdx = i
			break
		}
	}
	cv.LineTo(cv.linePath[closeIdx].pos[0], cv.linePath[closeIdx].pos[1])
	cv.linePath[len(cv.linePath)-1].next = cv.linePath[closeIdx].tf
	cv.polyPath[len(cv.polyPath)-1].next = cv.polyPath[closeIdx].tf
}

func (cv *Canvas) Stroke() {
	if len(cv.linePath) == 0 {
		return
	}

	cv.activate()

	var triBuf [1000]float32
	tris := triBuf[:0]
	tris = append(tris, 0, 0, float32(cv.fw), 0, float32(cv.fw), float32(cv.fh), 0, 0, float32(cv.fw), float32(cv.fh), 0, float32(cv.fh))

	start := true
	var p0 vec
	for _, p := range cv.linePath {
		if p.move {
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

		if !p.attach {
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

		if p.attach {
			tris = cv.lineJoint(p, p0, p1, p.next, lp0, lp1, lp2, lp3, tris)
		}

		p0 = p1
		start = false
	}

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)

	gli.ColorMask(false, false, false, false)
	gli.StencilFunc(gl_ALWAYS, 1, 0xFF)
	gli.StencilOp(gl_REPLACE, gl_REPLACE, gl_REPLACE)
	gli.StencilMask(0x01)

	gli.UseProgram(sr.id)
	gli.Uniform4f(sr.color, 0, 0, 0, 0)
	gli.Uniform2f(sr.canvasSize, float32(cv.fw), float32(cv.fh))

	gli.EnableVertexAttribArray(sr.vertex)
	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))
	gli.DisableVertexAttribArray(sr.vertex)

	gli.ColorMask(true, true, true, true)

	gli.StencilFunc(gl_EQUAL, 1, 0xFF)
	gli.StencilMask(0xFF)

	vertex := cv.useShader(&cv.state.stroke)
	gli.EnableVertexAttribArray(vertex)
	gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, nil)
	gli.DrawArrays(gl_TRIANGLES, 0, 6)
	gli.DisableVertexAttribArray(vertex)

	gli.StencilOp(gl_KEEP, gl_KEEP, gl_KEEP)
	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)

	gli.StencilMask(0x01)
	gli.Clear(gl_STENCIL_BUFFER_BIT)
	gli.StencilMask(0xFF)
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
		if l0p3.sub(l1p3).lenSqr() < 0.000000001 {
			ip1 = l0p3.sub(l1p3).mulf(0.5).add(l1p3)
		} else {
			var q float64
			ip1, _, q = lineIntersection(l0p2, l0p3, l1p3, l1p2)
			if q >= 1 {
				ip1 = l0p3.add(l1p3).mulf(0.5)
			}
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
	p := (vb[1]*(a0[0]-b0[0]) - a0[1]*vb[0] + b0[1]*vb[0]) / (va[1]*vb[0] - va[0]*vb[1])
	var q float64
	if vb[0] == 0 {
		q = (a0[1] + p*va[1] - b0[1]) / vb[1]
	} else {
		q = (a0[0] + p*va[0] - b0[0]) / vb[0]
	}

	return a0.add(va.mulf(p)), p, q
}

func (cv *Canvas) Fill() {
	if len(cv.polyPath) < 3 {
		return
	}
	cv.activate()
	start := 0
	for i, p := range cv.polyPath {
		if !p.move {
			continue
		}
		if i >= start+3 {
			cv.fillPoly(start, i)
		}
		start = i
	}
	if len(cv.polyPath) >= start+3 {
		cv.fillPoly(start, len(cv.polyPath))
	}
}

func (cv *Canvas) fillPoly(from, to int) {
	path := cv.polyPath[from:to]
	if len(path) < 3 {
		return
	}
	path = cv.cutIntersections(path)

	cv.activate()

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	var triBuf [1000]float32
	tris := triangulatePath(path, triBuf[:0])
	if len(tris) == 0 {
		return
	}
	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)

	vertex := cv.useShader(&cv.state.fill)
	gli.EnableVertexAttribArray(vertex)
	gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, nil)
	gli.DrawArrays(gl_TRIANGLES, 0, int32(len(tris)/2))
	gli.DisableVertexAttribArray(vertex)
}

func (cv *Canvas) Clip() {
	if len(cv.polyPath) < 3 {
		return
	}

	cv.clip(cv.polyPath)
}

func (cv *Canvas) clip(path []pathPoint) {
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
	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)

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

	cv.state.clip = make([]pathPoint, len(cv.polyPath))
	copy(cv.state.clip, cv.polyPath)
}

// Rect creates a closed rectangle path for stroking or filling
func (cv *Canvas) Rect(x, y, w, h float64) {
	cv.MoveTo(x, y)
	cv.LineTo(x+w, y)
	cv.LineTo(x+w, y+h)
	cv.LineTo(x, y+h)
	cv.ClosePath()
}
