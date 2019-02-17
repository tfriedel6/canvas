package goglbackend

import (
	"unsafe"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/tfriedel6/canvas/backend/backendbase"
)

/*
// FillRect fills a rectangle with the active fill style
func (b *GoGLBackend) FillRect(x, y, w, h float64) {
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

	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	data := [8]float32{float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), float32(p3[0]), float32(p3[1])}
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	vertex := cv.useShader(&cv.state.fill)
	gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(vertex)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(vertex)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)
}
*/

// ClearRect sets the color of the rectangle to transparent black
func (b *GoGLBackend) ClearRect(x, y, w, h int) {
	gl.Scissor(int32(x), int32(b.h-y-h), int32(w), int32(h))
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	// cv.applyScissor()
}

func (b *GoGLBackend) Clear(pts [4][2]float64) {
	data := [8]float32{
		float32(pts[0][0]), float32(pts[0][1]),
		float32(pts[1][0]), float32(pts[1][1]),
		float32(pts[2][0]), float32(pts[2][1]),
		float32(pts[3][0]), float32(pts[3][1])}

	gl.UseProgram(b.sr.ID)
	gl.Uniform2f(b.sr.CanvasSize, float32(b.fw), float32(b.fh))
	gl.Uniform4f(b.sr.Color, 0, 0, 0, 0)
	gl.Uniform1f(b.sr.GlobalAlpha, 1)

	gl.Disable(gl.BLEND)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.VertexAttribPointer(b.sr.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(b.sr.Vertex)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(b.sr.Vertex)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

	gl.Enable(gl.BLEND)
}

/*
func (b *GoGLBackend) Fill(style *backendbase.Style, pts [4][2]float64) {
	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	data := [8]float32{
		float32(pts[0][0]), float32(pts[0][1]),
		float32(pts[1][0]), float32(pts[1][1]),
		float32(pts[2][0]), float32(pts[2][1]),
		float32(pts[3][0]), float32(pts[3][1])}
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	vertex := b.useShader(style)
	gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(vertex)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(vertex)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)
}
*/

func (b *GoGLBackend) Fill(style *backendbase.Style, pts [][2]float64) {
	b.ptsBuf = b.ptsBuf[:0]
	for _, pt := range pts {
		b.ptsBuf = append(b.ptsBuf, float32(pt[0]), float32(pt[1]))
	}

	mode := uint32(gl.TRIANGLES)
	if len(pts) == 4 {
		mode = gl.TRIANGLE_FAN
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	gl.BufferData(gl.ARRAY_BUFFER, len(b.ptsBuf)*4, unsafe.Pointer(&b.ptsBuf[0]), gl.STREAM_DRAW)

	if style.GlobalAlpha >= 1 { // && cv.state.fill.isOpaque() {
		vertex := b.useShader(style)

		gl.EnableVertexAttribArray(vertex)
		gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, nil)
		gl.DrawArrays(mode, 0, int32(len(b.ptsBuf)/2))
		gl.DisableVertexAttribArray(vertex)
	} else {
		gl.ColorMask(false, false, false, false)
		gl.StencilFunc(gl.ALWAYS, 1, 0xFF)
		gl.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
		gl.StencilMask(0x01)

		gl.UseProgram(b.sr.ID)
		gl.Uniform4f(b.sr.Color, 0, 0, 0, 0)
		gl.Uniform2f(b.sr.CanvasSize, float32(b.fw), float32(b.fh))

		gl.EnableVertexAttribArray(b.sr.Vertex)
		gl.VertexAttribPointer(b.sr.Vertex, 2, gl.FLOAT, false, 0, nil)
		gl.DrawArrays(mode, 0, int32(len(b.ptsBuf)/2))
		gl.DisableVertexAttribArray(b.sr.Vertex)

		gl.ColorMask(true, true, true, true)

		gl.StencilFunc(gl.EQUAL, 1, 0xFF)

		vertex := b.useShader(style)
		gl.EnableVertexAttribArray(vertex)
		gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, nil)

		b.ptsBuf = append(b.ptsBuf[:0], 0, 0, float32(b.fw), 0, float32(b.fw), float32(b.fh), 0, float32(b.fh))
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
		gl.DisableVertexAttribArray(vertex)

		gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
		gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

		gl.Clear(gl.STENCIL_BUFFER_BIT)
		gl.StencilMask(0xFF)
	}
}
