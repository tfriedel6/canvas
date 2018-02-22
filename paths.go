package canvas

import (
	"math"
	"unsafe"

	"github.com/barnex/fmath"
	"github.com/tfriedel6/lm"
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

func isSamePoint(a, b lm.Vec2, maxDist float32) bool {
	return fmath.Abs(b[0]-a[0]) <= maxDist && fmath.Abs(b[1]-a[1]) <= maxDist
}

func (cv *Canvas) MoveTo(x, y float32) {
	tf := cv.tf(lm.Vec2{x, y})
	if len(cv.linePath) > 0 && isSamePoint(cv.linePath[len(cv.linePath)-1].tf, tf, 0.1) {
		return
	}
	cv.linePath = append(cv.linePath, pathPoint{pos: lm.Vec2{x, y}, tf: tf, move: true})
	cv.polyPath = append(cv.polyPath, pathPoint{pos: lm.Vec2{x, y}, tf: tf, move: true})
}

func (cv *Canvas) LineTo(x, y float32) {
	cv.strokeLineTo(x, y)
	cv.fillLineTo(x, y)
}

func (cv *Canvas) strokeLineTo(x, y float32) {
	if len(cv.linePath) > 0 && isSamePoint(cv.linePath[len(cv.linePath)-1].tf, cv.tf(lm.Vec2{x, y}), 0.1) {
		return
	}
	if len(cv.linePath) == 0 {
		cv.linePath = append(cv.linePath, pathPoint{pos: lm.Vec2{x, y}, tf: cv.tf(lm.Vec2{x, y}), move: true})
		return
	}
	if len(cv.state.lineDash) > 0 {
		lp := cv.linePath[len(cv.linePath)-1].pos
		tp := lm.Vec2{x, y}
		v := tp.Sub(lp)
		vl := v.Len()
		prev := cv.state.lineDashOffset
		for vl > 0 {
			draw := cv.state.lineDashPoint%2 == 0
			p := tp
			cv.state.lineDashOffset += vl
			if cv.state.lineDashOffset > cv.state.lineDash[cv.state.lineDashPoint] {
				cv.state.lineDashOffset = 0
				dl := cv.state.lineDash[cv.state.lineDashPoint] - prev
				p = lp.Add(v.MulF(dl / vl))
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
			v = tp.Sub(lp)
		}
	} else {
		tf := cv.tf(lm.Vec2{x, y})
		cv.linePath[len(cv.linePath)-1].next = tf
		cv.linePath[len(cv.linePath)-1].attach = true
		cv.linePath = append(cv.linePath, pathPoint{pos: lm.Vec2{x, y}, tf: tf, move: false})
	}
}

func (cv *Canvas) fillLineTo(x, y float32) {
	if len(cv.polyPath) > 0 && isSamePoint(cv.polyPath[len(cv.polyPath)-1].tf, cv.tf(lm.Vec2{x, y}), 0.1) {
		return
	}
	if len(cv.polyPath) == 0 {
		cv.polyPath = append(cv.polyPath, pathPoint{pos: lm.Vec2{x, y}, tf: cv.tf(lm.Vec2{x, y}), move: true})
		return
	}
	tf := cv.tf(lm.Vec2{x, y})
	cv.polyPath[len(cv.polyPath)-1].next = tf
	cv.polyPath[len(cv.polyPath)-1].attach = true
	cv.polyPath = append(cv.polyPath, pathPoint{pos: lm.Vec2{x, y}, tf: tf, move: false})
}

func (cv *Canvas) Arc(x, y, radius, startAngle, endAngle float32, anticlockwise bool) {
	startAngle = fmath.Mod(startAngle, math.Pi*2)
	if startAngle < 0 {
		startAngle += math.Pi * 2
	}
	endAngle = fmath.Mod(endAngle, math.Pi*2)
	if endAngle < 0 {
		endAngle += math.Pi * 2
	}
	if !anticlockwise && endAngle <= startAngle {
		endAngle += math.Pi * 2
	} else if anticlockwise && endAngle >= startAngle {
		endAngle -= math.Pi * 2
	}
	tr := cv.tf(lm.Vec2{radius, radius})
	step := 6 / fmath.Max(tr[0], tr[1])
	if step > 0.8 {
		step = 0.8
	} else if step < 0.05 {
		step = 0.05
	}
	if anticlockwise {
		for a := startAngle; a > endAngle; a -= step {
			s, c := fmath.Sincos(a)
			cv.LineTo(x+radius*c, y+radius*s)
		}
	} else {
		for a := startAngle; a < endAngle; a += step {
			s, c := fmath.Sincos(a)
			cv.LineTo(x+radius*c, y+radius*s)
		}
	}
	s, c := fmath.Sincos(endAngle)
	cv.LineTo(x+radius*c, y+radius*s)
}

func (cv *Canvas) ArcTo(x1, y1, x2, y2, radius float32) {
	if len(cv.linePath) == 0 {
		return
	}
	p0, p1, p2 := cv.linePath[len(cv.linePath)-1].pos, lm.Vec2{x1, y1}, lm.Vec2{x2, y2}
	v0, v1 := p0.Sub(p1).Norm(), p2.Sub(p1).Norm()
	angle := fmath.Acos(v0.Dot(v1))
	// should be in the range [0-pi]. if parallel, use a straight line
	if angle <= 0 || angle >= math.Pi {
		cv.LineTo(x2, y2)
		return
	}
	// cv are the vectors orthogonal to the lines that point to the center of the circle
	cv0 := lm.Vec2{-v0[1], v0[0]}
	cv1 := lm.Vec2{v1[1], -v1[0]}
	x := cv1.Sub(cv0).Div(v0.Sub(v1))[0] * radius
	if x < 0 {
		cv0 = cv0.MulF(-1)
		cv1 = cv1.MulF(-1)
	}
	center := p1.Add(v0.MulF(fmath.Abs(x))).Add(cv0.MulF(radius))
	a0, a1 := cv0.MulF(-1).Atan2(), cv1.MulF(-1).Atan2()
	cv.Arc(center[0], center[1], radius, a0, a1, x > 0)
}

func (cv *Canvas) ClosePath() {
	if len(cv.linePath) < 2 {
		return
	}
	if isSamePoint(cv.linePath[len(cv.linePath)-1].tf, cv.linePath[0].tf, 0.1) {
		return
	}
	cv.LineTo(cv.linePath[0].pos[0], cv.linePath[0].pos[1])
	cv.linePath[len(cv.linePath)-1].next = cv.linePath[0].tf
	cv.polyPath[len(cv.polyPath)-1].next = cv.polyPath[0].tf
}

func (cv *Canvas) Stroke() {
	if len(cv.linePath) == 0 {
		return
	}

	cv.activate()

	gli.ColorMask(false, false, false, false)
	gli.StencilFunc(gl_ALWAYS, 1, 0xFF)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_REPLACE)
	gli.StencilMask(0x01)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	var buf [1000]float32
	tris := buf[:0]
	tris = append(tris, 0, 0, cv.fw, 0, cv.fw, cv.fh, 0, 0, cv.fw, cv.fh, 0, cv.fh)

	start := true
	var p0 lm.Vec2
	for _, p := range cv.linePath {
		if p.move {
			p0 = p.tf
			start = true
			continue
		}
		p1 := p.tf

		v0 := p1.Sub(p0).Norm()
		v1 := lm.Vec2{v0[1], -v0[0]}.MulF(cv.state.lineWidth * 0.5)
		v0 = v0.MulF(cv.state.lineWidth * 0.5)

		lp0 := p0.Add(v1)
		lp1 := p1.Add(v1)
		lp2 := p0.Sub(v1)
		lp3 := p1.Sub(v1)

		if start {
			switch cv.state.lineEnd {
			case Butt:
				// no need to do anything
			case Square:
				lp0 = lp0.Sub(v0)
				lp2 = lp2.Sub(v0)
			case Round:
				tris = cv.addCircleTris(p0, cv.state.lineWidth*0.5, tris)
			}
		}

		if !p.attach {
			switch cv.state.lineEnd {
			case Butt:
				// no need to do anything
			case Square:
				lp1 = lp1.Add(v0)
				lp3 = lp3.Add(v0)
			case Round:
				tris = cv.addCircleTris(p1, cv.state.lineWidth*0.5, tris)
			}
		}

		tris = append(tris,
			lp0[0], lp0[1], lp1[0], lp1[1], lp3[0], lp3[1],
			lp0[0], lp0[1], lp3[0], lp3[1], lp2[0], lp2[1])

		if p.attach {
			tris = cv.lineJoint(p, p0, p1, p.next, lp0, lp1, lp2, lp3, tris)
		}

		p0 = p1
		start = false
	}

	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)

	gli.UseProgram(sr.id)
	gli.Uniform4f(sr.color, 0, 0, 0, 0)
	gli.Uniform2f(sr.canvasSize, cv.fw, cv.fh)

	gli.ColorMask(false, false, false, false)

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
	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	gli.StencilMask(0x01)
	gli.Clear(gl_STENCIL_BUFFER_BIT)
	gli.StencilMask(0xFF)
}

func (cv *Canvas) lineJoint(p pathPoint, p0, p1, p2, l0p0, l0p1, l0p2, l0p3 lm.Vec2, tris []float32) []float32 {
	v2 := p1.Sub(p2).Norm()
	v3 := lm.Vec2{v2[1], -v2[0]}.MulF(cv.state.lineWidth * 0.5)

	switch cv.state.lineJoin {
	case Miter:
		l1p0 := p2.Sub(v3)
		l1p1 := p1.Sub(v3)
		l1p2 := p2.Add(v3)
		l1p3 := p1.Add(v3)

		var ip0, ip1 lm.Vec2
		if l0p1.Sub(l1p1).LenSqr() < 0.000000001 {
			ip0 = l0p1.Sub(l1p1).MulF(0.5).Add(l1p1)
		} else {
			ip0, _, _ = lineIntersection(l0p0, l0p1, l1p1, l1p0)
		}
		if l0p3.Sub(l1p3).LenSqr() < 0.000000001 {
			ip1 = l0p3.Sub(l1p3).MulF(0.5).Add(l1p3)
		} else {
			ip1, _, _ = lineIntersection(l0p2, l0p3, l1p3, l1p2)
		}

		tris = append(tris,
			p1[0], p1[1], l0p1[0], l0p1[1], ip0[0], ip0[1],
			p1[0], p1[1], ip0[0], ip0[1], l1p1[0], l1p1[1],
			p1[0], p1[1], l1p3[0], l1p3[1], ip1[0], ip1[1],
			p1[0], p1[1], ip1[0], ip1[1], l0p3[0], l0p3[1])
	case Bevel:
		l1p1 := p1.Sub(v3)
		l1p3 := p1.Add(v3)

		tris = append(tris,
			p1[0], p1[1], l0p1[0], l0p1[1], l1p1[0], l1p1[1],
			p1[0], p1[1], l1p3[0], l1p3[1], l0p3[0], l0p3[1])
	case Round:
		tris = cv.addCircleTris(p1, cv.state.lineWidth*0.5, tris)
	}

	return tris
}

func (cv *Canvas) addCircleTris(center lm.Vec2, radius float32, tris []float32) []float32 {
	p0 := lm.Vec2{center[0], center[1] + radius}
	step := 6 / radius
	if step > 0.8 {
		step = 0.8
	} else if step < 0.05 {
		step = 0.05
	}
	for angle := step; angle <= math.Pi*2+step; angle += step {
		s, c := fmath.Sincos(angle)
		p1 := lm.Vec2{center[0] + s*radius, center[1] + c*radius}
		tris = append(tris, center[0], center[1], p0[0], p0[1], p1[0], p1[1])
		p0 = p1
	}
	return tris
}

func lineIntersection(a0, a1, b0, b1 lm.Vec2) (lm.Vec2, float32, float32) {
	va := a1.Sub(a0)
	vb := b1.Sub(b0)

	if (va[0] == 0 && vb[0] == 0) || (va[1] == 0 && vb[1] == 0) || (va[0] == 0 && va[1] == 0) || (vb[0] == 0 && vb[1] == 0) {
		return lm.Vec2{}, float32(math.Inf(1)), float32(math.Inf(1))
	}
	p := (vb[1]*(a0[0]-b0[0]) - a0[1]*vb[0] + b0[1]*vb[0]) / (va[1]*vb[0] - va[0]*vb[1])
	var q float32
	if vb[0] == 0 {
		q = (a0[1] + p*va[1] - b0[1]) / vb[1]
	} else {
		q = (a0[0] + p*va[0] - b0[0]) / vb[0]
	}

	return a0.Add(va.MulF(p)), p, q
}

func (cv *Canvas) Fill() {
	lastMove := 0
	for i, p := range cv.polyPath {
		if p.move {
			lastMove = i
		}
	}

	path := cv.polyPath[lastMove:]
	if len(path) < 3 {
		return
	}
	path = cv.cutIntersections(path)

	cv.activate()

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	var buf [1000]float32
	tris := triangulatePath(path, buf[:0])
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

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	var buf [1000]float32
	tris := buf[:0]
	tris = append(tris, 0, 0, cv.fw, 0, cv.fw, cv.fh, 0, 0, cv.fw, cv.fh, 0, cv.fh)
	tris = triangulatePath(path, tris)
	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)
	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)

	gli.UseProgram(sr.id)
	gli.Uniform4f(sr.color, 1, 1, 1, 1)
	gli.Uniform2f(sr.canvasSize, cv.fw, cv.fh)
	gli.EnableVertexAttribArray(sr.vertex)

	gli.ColorMask(false, false, false, false)
	gli.StencilFunc(gl_ALWAYS, 2, 0xFF)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_REPLACE)
	gli.StencilMask(0x02)
	gli.Clear(gl_STENCIL_BUFFER_BIT)

	gli.DrawArrays(gl_TRIANGLES, 0, 6)
	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)
	gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))

	gli.DisableVertexAttribArray(sr.vertex)

	gli.ColorMask(true, true, true, true)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_KEEP)
	gli.StencilMask(0xFF)
	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	cv.state.clip = make([]pathPoint, len(cv.polyPath))
	copy(cv.state.clip, cv.polyPath)
}
