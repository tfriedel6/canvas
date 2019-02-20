package goglbackend

/*
import (
	"image"
	"math"
	"unsafe"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/tfriedel6/canvas/backend/backendbase"
)

func (b *GoGLBackend) drawShadow(sh *backendbase.Shadow, tris []float32) {
	if len(tris) == 0 || sh.Color.A == 0 {
		return
	}

	if sh.Blur > 0 {
		b.offscr1.alpha = true
		cv.enableTextureRenderTarget(&b.offscr1)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	}

	ox, oy := float32(sh.OffsetX), float32(sh.OffsetY)

	count := len(tris)
	for i := 12; i < count; i += 2 {
		tris[i] += ox
		tris[i+1] += oy
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, b.shadowBuf)
	gl.BufferData(gl.ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl.STREAM_DRAW)

	gl.ColorMask(false, false, false, false)
	gl.StencilFunc(gl.ALWAYS, 1, 0xFF)
	gl.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
	gl.StencilMask(0x01)

	gl.UseProgram(b.sr.ID)
	gl.Uniform4f(b.sr.Color, 0, 0, 0, 0)
	gl.Uniform2f(b.sr.CanvasSize, float32(b.fw), float32(b.fh))

	gl.EnableVertexAttribArray(b.sr.Vertex)
	gl.VertexAttribPointer(b.sr.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.DrawArrays(gl.TRIANGLES, 6, int32(len(tris)/2-6))
	gl.DisableVertexAttribArray(b.sr.Vertex)

	gl.ColorMask(true, true, true, true)

	gl.StencilFunc(gl.EQUAL, 1, 0xFF)

	var style drawStyle
	style.color = colorGLToGo(sh.Color)

	vertex := b.useShader(&style)
	gl.EnableVertexAttribArray(vertex)
	gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, nil)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.DisableVertexAttribArray(vertex)

	gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

	gl.Clear(gl.STENCIL_BUFFER_BIT)
	gl.StencilMask(0xFF)

	if sh.Blur > 0 {
		b.drawBlurredShadow()
	}
}

func (b *GoGLBackend) drawTextShadow(sh *backendbase.Shadow, offset image.Point, strWidth, strHeight int, x, y float64) {
	x += sh.OffsetX
	y += sh.OffsetY

	if sh.Blur > 0 {
		b.offscr1.alpha = true
		cv.enableTextureRenderTarget(&b.offscr1)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	}

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)

	var style drawStyle
	style.color = colorGLToGo(sh.Color)

	vertex, alphaTexCoord := b.useAlphaShader(&style, 1)

	gl.EnableVertexAttribArray(vertex)
	gl.EnableVertexAttribArray(alphaTexCoord)

	p0 := cv.tf(vec{float64(offset.X) + x, float64(offset.Y) + y})
	p1 := cv.tf(vec{float64(offset.X) + x, float64(offset.Y+strHeight) + y})
	p2 := cv.tf(vec{float64(offset.X+strWidth) + x, float64(offset.Y+strHeight) + y})
	p3 := cv.tf(vec{float64(offset.X+strWidth) + x, float64(offset.Y) + y})

	tw := float64(strWidth) / alphaTexSize
	th := float64(strHeight) / alphaTexSize
	data := [16]float32{float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), float32(p3[0]), float32(p3[1]),
		0, 1, 0, float32(1 - th), float32(tw), float32(1 - th), float32(tw), 1}
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, 0)
	gl.VertexAttribPointer(alphaTexCoord, 2, gl.FLOAT, false, 0, 8*4)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	gl.DisableVertexAttribArray(vertex)
	gl.DisableVertexAttribArray(alphaTexCoord)

	gl.ActiveTexture(gl.TEXTURE0)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

	if cv.state.shadowBlur > 0 {
		cv.drawBlurredShadow()
	}
}

func (b *GoGLBackend) drawBlurredShadow() {
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

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
		gs = gauss127r
		kernel = kernelBuf[:127]
	}

	gaussianKernel(cv.state.shadowBlur, kernel)

	offscr2.alpha = true
	cv.enableTextureRenderTarget(&offscr2)
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.BindBuffer(gl.ARRAY_BUFFER, shadowBuf)
	data := [16]float32{0, 0, 0, float32(cv.h), float32(cv.w), float32(cv.h), float32(cv.w), 0, 0, 0, 0, 1, 1, 1, 1, 0}
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, offscr1.tex)

	gl.UseProgram(gs.id)
	gl.Uniform1i(gs.image, 0)
	gl.Uniform2f(gs.canvasSize, float32(cv.fw), float32(cv.fh))
	gl.Uniform2f(gs.kernelScale, 1.0/float32(cv.fw), 0.0)
	gl.Uniform1fv(gs.kernel, int32(len(kernel)), &kernel[0])
	gl.VertexAttribPointer(gs.vertex, 2, gl.FLOAT, false, 0, 0)
	gl.VertexAttribPointer(gs.texCoord, 2, gl.FLOAT, false, 0, 8*4)
	gl.EnableVertexAttribArray(gs.vertex)
	gl.EnableVertexAttribArray(gs.texCoord)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(gs.vertex)
	gl.DisableVertexAttribArray(gs.texCoord)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

	cv.disableTextureRenderTarget()

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.BindBuffer(gl.ARRAY_BUFFER, shadowBuf)
	data = [16]float32{0, 0, 0, float32(cv.h), float32(cv.w), float32(cv.h), float32(cv.w), 0, 0, 0, 0, 1, 1, 1, 1, 0}
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, offscr2.tex)

	gl.UseProgram(gs.id)
	gl.Uniform1i(gs.image, 0)
	gl.Uniform2f(gs.canvasSize, float32(cv.fw), float32(cv.fh))
	gl.Uniform2f(gs.kernelScale, 0.0, 1.0/float32(cv.fh))
	gl.Uniform1fv(gs.kernel, int32(len(kernel)), &kernel[0])
	gl.VertexAttribPointer(gs.vertex, 2, gl.FLOAT, false, 0, 0)
	gl.VertexAttribPointer(gs.texCoord, 2, gl.FLOAT, false, 0, 8*4)
	gl.EnableVertexAttribArray(gs.vertex)
	gl.EnableVertexAttribArray(gs.texCoord)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(gs.vertex)
	gl.DisableVertexAttribArray(gs.texCoord)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
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
*/
