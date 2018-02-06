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

func (cv *Canvas) MoveTo(x, y float32) {
	tf := cv.tf(lm.Vec2{x, y})
	cv.linePath = append(cv.linePath, pathPoint{pos: lm.Vec2{x, y}, tf: tf, move: true})
	cv.polyPath = append(cv.polyPath, pathPoint{pos: lm.Vec2{x, y}, tf: tf, move: true})
}

func (cv *Canvas) LineTo(x, y float32) {
	if len(cv.linePath) == 0 {
		cv.MoveTo(x, y)
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
	} else if anticlockwise && startAngle <= endAngle {
		startAngle += math.Pi * 2
		startAngle, endAngle = endAngle, startAngle
	}
	tr := cv.tf(lm.Vec2{radius, radius})
	step := 6 / fmath.Max(tr[0], tr[1])
	if step > 0.8 {
		step = 0.8
	} else if step < 0.01 {
		step = 0.01
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
	if len(cv.state.lineDash) > 0 {
		cv.LineTo(cv.linePath[0].pos[0], cv.linePath[0].pos[1])
		return
	}
	cv.linePath[len(cv.linePath)-1].next = cv.linePath[0].pos
	cv.linePath[len(cv.linePath)-1].attach = true
	cv.linePath = append(cv.linePath, pathPoint{pos: cv.linePath[0].pos, move: false, tf: cv.linePath[0].tf, next: cv.linePath[1].pos, attach: true})
	cv.polyPath[len(cv.polyPath)-1].next = cv.polyPath[0].pos
	cv.polyPath[len(cv.polyPath)-1].attach = true
	cv.polyPath = append(cv.polyPath, pathPoint{pos: cv.polyPath[0].pos, move: false, tf: cv.linePath[0].tf, next: cv.polyPath[1].pos, attach: true})
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

	gli.UseProgram(sr.id)
	s := cv.state.stroke
	gli.Uniform4f(sr.color, s.r, s.g, s.b, s.a)
	gli.EnableVertexAttribArray(sr.vertex)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	var buf [1000]float32
	tris := buf[:0]
	tris = append(tris, -1, -1, -1, 1, 1, 1, -1, -1, 1, 1, 1, -1)

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
		v1 := lm.Vec2{v0[1], -v0[0]}.MulF(cv.state.stroke.lineWidth * 0.5)
		v0 = v0.MulF(cv.state.stroke.lineWidth * 0.5)

		l0p0 := p0.Add(v1)
		l0p1 := p1.Add(v1)
		l0p2 := p0.Sub(v1)
		l0p3 := p1.Sub(v1)

		if start {
			switch cv.state.lineEnd {
			case Butt:
				// no need to do anything
			case Square:
				l0p0 = l0p0.Sub(v0)
				l0p2 = l0p2.Sub(v0)
			case Round:
				tris = cv.addCircleTris(p0, cv.state.stroke.lineWidth*0.5, tris)
			}
		}

		if !p.attach {
			switch cv.state.lineEnd {
			case Butt:
				// no need to do anything
			case Square:
				l0p1 = l0p1.Add(v0)
				l0p3 = l0p3.Add(v0)
			case Round:
				tris = cv.addCircleTris(p1, cv.state.stroke.lineWidth*0.5, tris)
			}
		}

		l0x0f, l0y0f := cv.vecToGL(l0p0)
		l0x1f, l0y1f := cv.vecToGL(l0p1)
		l0x2f, l0y2f := cv.vecToGL(l0p2)
		l0x3f, l0y3f := cv.vecToGL(l0p3)

		tris = append(tris,
			l0x0f, l0y0f, l0x1f, l0y1f, l0x3f, l0y3f,
			l0x0f, l0y0f, l0x3f, l0y3f, l0x2f, l0y2f)

		if p.attach {
			tris = cv.lineJoint(p, p0, p1, p.next, l0p0, l0p1, l0p2, l0p3, tris)
		}

		p0 = p1
		start = false
	}

	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)
	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))

	gli.ColorMask(true, true, true, true)
	gli.StencilFunc(gl_EQUAL, 1, 0xFF)
	gli.StencilMask(0xFF)

	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 6)

	gli.DisableVertexAttribArray(sr.vertex)

	gli.ColorMask(true, true, true, true)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_KEEP)
	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	gli.StencilMask(0x01)
	gli.Clear(gl_STENCIL_BUFFER_BIT)
	gli.StencilMask(0xFF)
}

func (cv *Canvas) lineJoint(p pathPoint, p0, p1, p2, l0p0, l0p1, l0p2, l0p3 lm.Vec2, tris []float32) []float32 {
	v2 := p1.Sub(p2).Norm()
	v3 := lm.Vec2{v2[1], -v2[0]}.MulF(cv.state.stroke.lineWidth * 0.5)

	l0x1f, l0y1f := cv.vecToGL(l0p1)
	l0x3f, l0y3f := cv.vecToGL(l0p3)

	switch cv.state.lineJoin {
	case Miter:
		l1p0 := p2.Sub(v3)
		l1p1 := p1.Sub(v3)
		l1p2 := p2.Add(v3)
		l1p3 := p1.Add(v3)

		l1x1f, l1y1f := cv.vecToGL(l1p1)
		l1x3f, l1y3f := cv.vecToGL(l1p3)

		ip0 := lineIntersection(l0p0, l0p1, l1p1, l1p0)
		ip1 := lineIntersection(l0p2, l0p3, l1p3, l1p2)

		cxf, cyf := cv.vecToGL(p1)
		ix0f, iy0f := cv.vecToGL(ip0)
		ix1f, iy1f := cv.vecToGL(ip1)

		tris = append(tris,
			cxf, cyf, l0x1f, l0y1f, ix0f, iy0f,
			cxf, cyf, ix0f, iy0f, l1x1f, l1y1f,
			cxf, cyf, l1x3f, l1y3f, ix1f, iy1f,
			cxf, cyf, ix1f, iy1f, l0x3f, l0y3f)
	case Bevel:
		l1p1 := p1.Sub(v3)
		l1p3 := p1.Add(v3)

		l1x1f, l1y1f := cv.vecToGL(l1p1)
		l1x3f, l1y3f := cv.vecToGL(l1p3)

		cxf, cyf := cv.vecToGL(p1)

		tris = append(tris,
			cxf, cyf, l0x1f, l0y1f, l1x1f, l1y1f,
			cxf, cyf, l1x3f, l1y3f, l0x3f, l0y3f)
	case Round:
		tris = cv.addCircleTris(p1, cv.state.stroke.lineWidth*0.5, tris)
	}

	return tris
}

func (cv *Canvas) addCircleTris(p lm.Vec2, radius float32, tris []float32) []float32 {
	cxf, cyf := cv.vecToGL(p)
	p0x, p0y := cv.vecToGL(lm.Vec2{p[0], p[1] + radius})
	step := 6 / radius
	if step > 0.8 {
		step = 0.8
	} else if step < 0.01 {
		step = 0.01
	}
	for angle := step; angle <= math.Pi*2+step; angle += step {
		s, c := fmath.Sincos(angle)
		p1x, p1y := cv.vecToGL(lm.Vec2{p[0] + s*radius, p[1] + c*radius})
		tris = append(tris, cxf, cyf, p0x, p0y, p1x, p1y)
		p0x, p0y = p1x, p1y
	}
	return tris
}

func lineIntersection(a0, a1, b0, b1 lm.Vec2) lm.Vec2 {
	va := a1.Sub(a0)
	vb := b1.Sub(b0)

	if vb[1] == 0 {
		q := (a0[0] + (b0[1]-a0[1])*(va[0]/va[1]) - b0[0]) / (vb[0] - vb[1]*(va[0]/va[1]))
		return b0.Add(vb.MulF(q))
	}

	p := (b0[0] + (a0[1]-b0[1])*(vb[0]/vb[1]) - a0[0]) / (va[0] - va[1]*(vb[0]/vb[1]))
	return a0.Add(va.MulF(p))
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

	cv.activate()

	gli.UseProgram(sr.id)
	f := cv.state.fill
	gli.Uniform4f(sr.color, f.r, f.g, f.b, f.a)
	gli.EnableVertexAttribArray(sr.vertex)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	var buf [1000]float32
	tris := buf[:0]
	tris = append(tris)

	tris = triangulatePath(path, tris)
	total := len(tris)
	for i := 0; i < total; i += 2 {
		x, y := tris[i], tris[i+1]
		tris[i], tris[i+1] = cv.ptToGL(x, y)
	}

	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)
	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.DrawArrays(gl_TRIANGLES, 0, int32(len(tris)/2))

	gli.DisableVertexAttribArray(sr.vertex)
}

func (cv *Canvas) Clip() {
	if len(cv.polyPath) < 3 {
		return
	}

	cv.clip(cv.polyPath)
}

func (cv *Canvas) clip(path []pathPoint) {
	cv.activate()

	gli.ColorMask(false, false, false, false)
	gli.StencilFunc(gl_ALWAYS, 2, 0xFF)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_REPLACE)
	gli.StencilMask(0x02)
	gli.Clear(gl_STENCIL_BUFFER_BIT)

	gli.UseProgram(sr.id)
	f := cv.state.fill
	gli.Uniform4f(sr.color, f.r, f.g, f.b, f.a)
	gli.EnableVertexAttribArray(sr.vertex)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	var buf [1000]float32
	tris := buf[:0]
	tris = append(tris, -1, -1, -1, 1, 1, 1, -1, -1, 1, 1, 1, -1)

	tris = triangulatePath(path, tris)
	total := len(tris)
	for i := 12; i < total; i += 2 {
		x, y := tris[i], tris[i+1]
		tris[i], tris[i+1] = cv.ptToGL(x, y)
	}

	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)
	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)

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
