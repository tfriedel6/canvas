package canvas

import (
	"unsafe"

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

	gli.Enable(gl_STENCIL_TEST)
	gli.ColorMask(false, false, false, false)
	gli.StencilFunc(gl_ALWAYS, 1, 0xFF)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_REPLACE)
	gli.StencilMask(0xFF)
	gli.Clear(gl_STENCIL_BUFFER_BIT)

	gli.UseProgram(sr.id)
	gli.Uniform4f(sr.color, cv.stroke.r, cv.stroke.g, cv.stroke.b, cv.stroke.a)
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
		v2 := lm.Vec2{v1[1], -v1[0]}.MulF(cv.stroke.lineWidth * 0.5)
		v1 = v1.MulF(cv.stroke.lineWidth * 0.5)

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
	gli.StencilMask(0)

	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 6)

	gli.DisableVertexAttribArray(sr.vertex)

	gli.Disable(gl_STENCIL_TEST)
}

func (cv *Canvas) Fill() {
	if len(cv.path) < 3 {
		return
	}

	cv.activate()

	gli.UseProgram(sr.id)
	gli.Uniform4f(sr.color, cv.fill.r, cv.fill.g, cv.fill.b, cv.fill.a)
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
	gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))

	gli.DisableVertexAttribArray(sr.vertex)
}
