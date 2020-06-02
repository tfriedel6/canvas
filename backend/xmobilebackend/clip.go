package xmobilebackend

import (
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"golang.org/x/mobile/gl"
)

func (b *XMobileBackend) ClearClip() {
	b.activate()

	b.glctx.StencilMask(0xFF)
	b.glctx.Clear(gl.STENCIL_BUFFER_BIT)
}

func (b *XMobileBackend) Clip(pts []backendbase.Vec) {
	b.activate()

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
	b.glctx.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, 0)

	b.glctx.UseProgram(b.shd.ID)
	b.glctx.Uniform4f(b.shd.Color, 1, 1, 1, 1)
	b.glctx.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.UniformMatrix3fv(b.shd.Matrix, mat3identity[:])
	b.glctx.Uniform1f(b.shd.GlobalAlpha, 1)
	b.glctx.Uniform1i(b.shd.UseAlphaTex, 0)
	b.glctx.Uniform1i(b.shd.Func, shdFuncSolid)
	b.glctx.EnableVertexAttribArray(b.shd.Vertex)
	b.glctx.EnableVertexAttribArray(b.shd.TexCoord)

	b.glctx.ColorMask(false, false, false, false)

	// set bit 2 in the stencil buffer in the given shape
	b.glctx.StencilMask(0x04)
	b.glctx.StencilFunc(gl.ALWAYS, 4, 0)
	b.glctx.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
	b.glctx.DrawArrays(mode, 4, len(pts))

	// on entire screen, where neither bit 1 or 2 are set, invert bit 1
	b.glctx.StencilMask(0x02)
	b.glctx.StencilFunc(gl.EQUAL, 0, 0x06)
	b.glctx.StencilOp(gl.KEEP, gl.INVERT, gl.INVERT)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	// on entire screen, clear bit 2
	b.glctx.StencilMask(0x04)
	b.glctx.StencilFunc(gl.ALWAYS, 0, 0)
	b.glctx.StencilOp(gl.ZERO, gl.ZERO, gl.ZERO)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	b.glctx.DisableVertexAttribArray(b.shd.Vertex)
	b.glctx.DisableVertexAttribArray(b.shd.TexCoord)

	b.glctx.ColorMask(true, true, true, true)
	b.glctx.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
	b.glctx.StencilMask(0xFF)
	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)
}
