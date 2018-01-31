package canvas

import (
	"math"
	"unsafe"

	"github.com/barnex/fmath"
	"github.com/tfriedel6/lm"
)

func (cv *Canvas) BeginPath() {
	if cv.path == nil {
		cv.path = make([]pathPoint, 0, 100)
	}
	cv.path = cv.path[:0]
}

func (cv *Canvas) MoveTo(x, y float32) {
	cv.path = append(cv.path, pathPoint{pos: lm.Vec2{x, y}, move: true})
}

func (cv *Canvas) LineTo(x, y float32) {
	cv.path = append(cv.path, pathPoint{pos: lm.Vec2{x, y}, move: false})
}

func (cv *Canvas) Arc(x, y, radius, startAngle, endAngle float32, anticlockwise bool) {
	step := 6 / radius
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
	for a := startAngle; a < endAngle; a += step {
		s, c := fmath.Sincos(a)
		cv.LineTo(x+radius*c, y+radius*s)
	}
	s, c := fmath.Sincos(endAngle)
	cv.LineTo(x+radius*c, y+radius*s)
}

func (cv *Canvas) ClosePath() {
	if len(cv.path) == 0 {
		return
	}
	cv.path = append(cv.path, pathPoint{pos: cv.path[0].pos, move: false})
}

func (cv *Canvas) Stroke() {
	if len(cv.path) == 0 {
		return
	}

	cv.activate()

	gli.ColorMask(false, false, false, false)
	gli.StencilFunc(gl_ALWAYS, 1, 0xFF)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_REPLACE)
	gli.StencilMask(0x01)
	gli.Clear(gl_STENCIL_BUFFER_BIT)

	gli.UseProgram(sr.id)
	s := cv.state.stroke
	gli.Uniform4f(sr.color, s.r, s.g, s.b, s.a)
	gli.EnableVertexAttribArray(sr.vertex)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	var buf [1000]float32
	tris := buf[:0]
	tris = append(tris, -1, -1, -1, 1, 1, 1, -1, -1, 1, 1, 1, -1)
	p0 := cv.path[0].pos
	for _, p := range cv.path {
		if p.move {
			p0 = p.pos
			continue
		}
		p1 := p.pos

		v1 := p1.Sub(p0).Norm()
		v2 := lm.Vec2{v1[1], -v1[0]}.MulF(cv.state.stroke.lineWidth * 0.5)
		v1 = v1.MulF(cv.state.stroke.lineWidth * 0.5)

		x0f, y0f := cv.vecToGL(p0.Sub(v1).Add(v2))
		x1f, y1f := cv.vecToGL(p1.Add(v1).Add(v2))
		x2f, y2f := cv.vecToGL(p1.Add(v1).Sub(v2))
		x3f, y3f := cv.vecToGL(p0.Sub(v1).Sub(v2))

		tris = append(tris, x0f, y0f, x1f, y1f, x2f, y2f, x0f, y0f, x2f, y2f, x3f, y3f)

		p0 = p1
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
	gli.StencilMask(0xFF)
	gli.StencilFunc(gl_EQUAL, 0, 0xFF)
}

func (cv *Canvas) Fill() {
	lastMove := 0
	for i, p := range cv.path {
		if p.move {
			lastMove = i
		}
	}

	path := cv.path[lastMove:]

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
	if len(cv.path) < 3 {
		return
	}

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

	tris = triangulatePath(cv.path, tris)
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
}
