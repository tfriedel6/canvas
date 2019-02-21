package canvas

import (
	"image"
	"math"
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

func (cv *Canvas) drawShadow2(pts [][2]float64) {
	if cv.state.shadowColor.A == 0 {
		return
	}
	if cv.state.shadowOffsetX == 0 && cv.state.shadowOffsetY == 0 {
		return
	}

	if cv.shadowBuf == nil || cap(cv.shadowBuf) < len(pts) {
		cv.shadowBuf = make([][2]float64, 0, len(pts)+1000)
	}
	cv.shadowBuf = cv.shadowBuf[:0]

	for _, pt := range pts {
		cv.shadowBuf = append(cv.shadowBuf, [2]float64{
			pt[0] + cv.state.shadowOffsetX,
			pt[1] + cv.state.shadowOffsetY,
		})
	}

	style := backendbase.FillStyle{Color: cv.state.shadowColor, Blur: cv.state.shadowBlur}
	cv.b.Fill(&style, cv.shadowBuf)
}

func (cv *Canvas) drawTextShadow(offset image.Point, strWidth, strHeight int, x, y float64) {
	if cv.state.shadowColor.A == 0 {
		return
	}

	x += cv.state.shadowOffsetX
	y += cv.state.shadowOffsetY

	if cv.state.shadowBlur > 0 {
		offscr1.alpha = true
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
		gs = gauss127r
		kernel = kernelBuf[:127]
	}

	gaussianKernel(cv.state.shadowBlur, kernel)

	offscr2.alpha = true
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
