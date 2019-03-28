package xmobilebackend

import (
	"unsafe"

	"golang.org/x/mobile/gl"
)

func (b *XMobileBackend) ClearClip() {
	b.curClip = nil
	b.activate()

	b.glctx.StencilMask(0xFF)
	b.glctx.Clear(gl.STENCIL_BUFFER_BIT)
}

func (b *XMobileBackend) Clip(pts [][2]float64) {
	b.curClip = nil
	b.activate()
	b.curClip = pts

	b.ptsBuf = b.ptsBuf[:0]
	b.ptsBuf = append(b.ptsBuf,
		0, 0,
		0, float32(b.fh),
		float32(b.fw), float32(b.fh),
		float32(b.fw), 0)
	for _, pt := range pts {
		b.ptsBuf = append(b.ptsBuf, float32(pt[0]), float32(pt[1]))
	}

	mode := gl.Enum(gl.TRIANGLES)
	if len(pts) == 4 {
		mode = gl.TRIANGLE_FAN
	}

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&b.ptsBuf[0]), len(b.ptsBuf)*4), gl.STREAM_DRAW)
	b.glctx.VertexAttribPointer(b.sr.Vertex, 2, gl.FLOAT, false, 0, 0)

	b.glctx.UseProgram(b.sr.ID)
	b.glctx.Uniform4f(b.sr.Color, 1, 1, 1, 1)
	b.glctx.Uniform2f(b.sr.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.EnableVertexAttribArray(b.sr.Vertex)

	b.glctx.ColorMask(false, false, false, false)

	b.glctx.StencilMask(0x04)
	b.glctx.StencilFunc(gl.ALWAYS, 4, 0x04)
	b.glctx.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
	b.glctx.DrawArrays(mode, 4, len(pts))

	b.glctx.StencilMask(0x02)
	b.glctx.StencilFunc(gl.EQUAL, 0, 0x06)
	b.glctx.StencilOp(gl.KEEP, gl.INVERT, gl.INVERT)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	b.glctx.StencilMask(0x04)
	b.glctx.StencilFunc(gl.ALWAYS, 0, 0x04)
	b.glctx.StencilOp(gl.ZERO, gl.ZERO, gl.ZERO)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	b.glctx.DisableVertexAttribArray(b.sr.Vertex)

	b.glctx.ColorMask(true, true, true, true)
	b.glctx.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
	b.glctx.StencilMask(0xFF)
	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)
}
