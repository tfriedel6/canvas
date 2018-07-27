package canvas

import (
	"image"
	"math"
	"unsafe"
)

func (cv *Canvas) drawShadow(tris []float32) {
	if len(tris) == 0 || cv.state.shadowColor.a == 0 {
		return
	}

	if cv.state.shadowBlur > 0 {
		cv.enableTextureRenderTarget(&offscr1)
		gli.ClearColor(0, 0, 0, 0)
		gli.Clear(gl_COLOR_BUFFER_BIT | gl_STENCIL_BUFFER_BIT)
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

	if cv.state.shadowBlur > 0 {
		cv.drawBlurredShadow()
	}
}

func (cv *Canvas) drawTextShadow(offset image.Point, strWidth, strHeight int, x, y float64) {
	x += cv.state.shadowOffsetX
	y += cv.state.shadowOffsetY

	if cv.state.shadowBlur > 0 {
		cv.enableTextureRenderTarget(&offscr1)
		gli.ClearColor(0, 0, 0, 0)
		gli.Clear(gl_COLOR_BUFFER_BIT | gl_STENCIL_BUFFER_BIT)
	}

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

	if cv.state.shadowBlur > 0 {
		cv.drawBlurredShadow()
	}
}

func (cv *Canvas) drawBlurredShadow() {
	gli.BlendFunc(gl_ONE, gl_ONE_MINUS_SRC_ALPHA)

	var kernel []float32
	var kernelBuf [255]float32
	var gs *gaussianShader
	if cv.state.shadowBlur < 3 {
		gs = gauss15r
		kernel = kernelBuf[:15]
	} else if cv.state.shadowBlur < 12 {
		gs = gauss63r
		kernel = kernelBuf[:63]
	} else {
		gs = gauss255r
		kernel = kernelBuf[:255]
	}

	gaussianKernel(cv.state.shadowBlur, kernel)

	cv.enableTextureRenderTarget(&offscr2)
	gli.ClearColor(0, 0, 0, 0)
	gli.Clear(gl_COLOR_BUFFER_BIT | gl_STENCIL_BUFFER_BIT)

	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	gli.BindBuffer(gl_ARRAY_BUFFER, shadowBuf)
	data := [16]float32{0, 0, 0, float32(cv.h), float32(cv.w), float32(cv.h), float32(cv.w), 0, 0, 0, 0, 1, 1, 1, 1, 0}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, offscr1.tex)

	gli.UseProgram(gs.id)
	gli.Uniform1i(gs.image, 0)
	gli.Uniform2f(gs.canvasSize, float32(cv.fw), float32(cv.fh))
	gli.Uniform2f(gs.kernelScale, 1.0/float32(cv.fw), 0.0)
	gli.Uniform1fv(gs.kernel, int32(len(kernel)), &kernel[0])
	gli.VertexAttribPointer(gs.vertex, 2, gl_FLOAT, false, 0, 0)
	gli.VertexAttribPointer(gs.texCoord, 2, gl_FLOAT, false, 0, 8*4)
	gli.EnableVertexAttribArray(gs.vertex)
	gli.EnableVertexAttribArray(gs.texCoord)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(gs.vertex)
	gli.DisableVertexAttribArray(gs.texCoord)

	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)

	cv.disableTextureRenderTarget()

	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	gli.BindBuffer(gl_ARRAY_BUFFER, shadowBuf)
	data = [16]float32{0, 0, 0, float32(cv.h), float32(cv.w), float32(cv.h), float32(cv.w), 0, 0, 0, 0, 1, 1, 1, 1, 0}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, offscr2.tex)

	gli.UseProgram(gs.id)
	gli.Uniform1i(gs.image, 0)
	gli.Uniform2f(gs.canvasSize, float32(cv.fw), float32(cv.fh))
	gli.Uniform2f(gs.kernelScale, 0.0, 1.0/float32(cv.fh))
	gli.Uniform1fv(gs.kernel, int32(len(kernel)), &kernel[0])
	gli.VertexAttribPointer(gs.vertex, 2, gl_FLOAT, false, 0, 0)
	gli.VertexAttribPointer(gs.texCoord, 2, gl_FLOAT, false, 0, 8*4)
	gli.EnableVertexAttribArray(gs.vertex)
	gli.EnableVertexAttribArray(gs.texCoord)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(gs.vertex)
	gli.DisableVertexAttribArray(gs.texCoord)

	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)

	gli.BlendFunc(gl_SRC_ALPHA, gl_ONE_MINUS_SRC_ALPHA)
}

func gaussianKernel(stddev float64, target []float32) {
	stddevSqr := stddev * stddev
	center := float64(len(target) / 2)
	factor := 1.0 / math.Sqrt(2*math.Pi*stddevSqr)
	for i := range target {
		x := float64(i) - center
		target[i] = float32(factor * math.Pow(math.E, -x*x/(2*stddevSqr)))
	}
	// normalizeKernel(target)
}

func normalizeKernel(kernel []float32) {
	var sum float32
	for _, v := range kernel {
		sum += v
	}
	factor := 1.0 / sum
	for i := range kernel {
		kernel[i] *= factor
	}
}
