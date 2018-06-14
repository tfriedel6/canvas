package canvas

import (
	"image"
	"unsafe"
)

func (cv *Canvas) drawShadow(tris []float32) {
	if len(tris) == 0 || cv.state.shadowColor.a == 0 {
		return
	}

	ox, oy := float32(cv.state.shadowOffsetX), float32(cv.state.shadowOffsetY)

	count := len(tris)
	for i := 12; i < count; i += 2 {
		tris[i] += ox
		tris[i+1] += oy
	}

	gli.BindBuffer(gl_ARRAY_BUFFER, shadowBuf)
	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)

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

	var style drawStyle
	style.color = cv.state.shadowColor

	vertex := cv.useShader(&style)
	gli.EnableVertexAttribArray(vertex)
	gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, 0)
	gli.DrawArrays(gl_TRIANGLES, 0, 6)
	gli.DisableVertexAttribArray(vertex)

	gli.StencilOp(gl_KEEP, gl_KEEP, gl_KEEP)
	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)

	gli.Clear(gl_STENCIL_BUFFER_BIT)
	gli.StencilMask(0xFF)
}

func (cv *Canvas) drawTextShadow(offset image.Point, strWidth, strHeight int, x, y float64) {
	x += cv.state.shadowOffsetX
	y += cv.state.shadowOffsetY

	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	var style drawStyle
	style.color = cv.state.shadowColor

	vertex, alphaTexCoord := cv.useAlphaShader(&style, 1)

	gli.EnableVertexAttribArray(vertex)
	gli.EnableVertexAttribArray(alphaTexCoord)

	p0 := cv.tf(vec{float64(offset.X) + x, float64(offset.Y) + y})
	p1 := cv.tf(vec{float64(offset.X) + x, float64(offset.Y+strHeight) + y})
	p2 := cv.tf(vec{float64(offset.X+strWidth) + x, float64(offset.Y+strHeight) + y})
	p3 := cv.tf(vec{float64(offset.X+strWidth) + x, float64(offset.Y) + y})

	tw := float64(strWidth) / alphaTexSize
	th := float64(strHeight) / alphaTexSize
	data := [16]float32{float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), float32(p3[0]), float32(p3[1]),
		0, 1, 0, float32(1 - th), float32(tw), float32(1 - th), float32(tw), 1}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, 0)
	gli.VertexAttribPointer(alphaTexCoord, 2, gl_FLOAT, false, 0, 8*4)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)

	gli.DisableVertexAttribArray(vertex)
	gli.DisableVertexAttribArray(alphaTexCoord)

	gli.ActiveTexture(gl_TEXTURE0)

	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)
}
