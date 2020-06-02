package goglbackend

import (
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"github.com/tfriedel6/canvas/backend/goglbackend/gl"
)

func (b *GoGLBackend) ClearClip() {
	b.activate()

	gl.StencilMask(0xFF)
	gl.Clear(gl.STENCIL_BUFFER_BIT)
}

func (b *GoGLBackend) Clip(pts []backendbase.Vec) {
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

	mode := uint32(gl.TRIANGLES)
	if len(pts) == 4 {
		mode = gl.TRIANGLE_FAN
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	gl.BufferData(gl.ARRAY_BUFFER, len(b.ptsBuf)*4, unsafe.Pointer(&b.ptsBuf[0]), gl.STREAM_DRAW)
	gl.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, nil)

	gl.UseProgram(b.shd.ID)
	gl.Uniform4f(b.shd.Color, 1, 1, 1, 1)
	gl.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	gl.UniformMatrix3fv(b.shd.Matrix, 1, false, &mat3identity[0])
	gl.Uniform1f(b.shd.GlobalAlpha, 1)
	gl.Uniform1i(b.shd.UseAlphaTex, 0)
	gl.Uniform1i(b.shd.Func, shdFuncSolid)
	gl.EnableVertexAttribArray(b.shd.Vertex)
	gl.EnableVertexAttribArray(b.shd.TexCoord)

	gl.ColorMask(false, false, false, false)

	// set bit 2 in the stencil buffer in the given shape
	gl.StencilMask(0x04)
	gl.StencilFunc(gl.ALWAYS, 4, 0)
	gl.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
	gl.DrawArrays(mode, 4, int32(len(pts)))

	// on entire screen, where neither bit 1 or 2 are set, invert bit 1
	gl.StencilMask(0x02)
	gl.StencilFunc(gl.EQUAL, 0, 0x06)
	gl.StencilOp(gl.KEEP, gl.INVERT, gl.INVERT)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	// on entire screen, clear bit 2
	gl.StencilMask(0x04)
	gl.StencilFunc(gl.ALWAYS, 0, 0)
	gl.StencilOp(gl.ZERO, gl.ZERO, gl.ZERO)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	gl.DisableVertexAttribArray(b.shd.Vertex)
	gl.DisableVertexAttribArray(b.shd.TexCoord)

	gl.ColorMask(true, true, true, true)
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
	gl.StencilMask(0xFF)
	gl.StencilFunc(gl.EQUAL, 0, 0xFF)
}
