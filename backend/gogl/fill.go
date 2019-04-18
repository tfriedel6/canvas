package goglbackend

import (
	"image"
	"math"
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"github.com/tfriedel6/canvas/backend/gogl/gl"
)

func (b *GoGLBackend) Clear(pts [4][2]float64) {
	b.activate()

	// first check if the four points are aligned to form a nice rectangle, which can be more easily
	// cleared using glScissor and glClear
	aligned := pts[0][0] == pts[1][0] && pts[2][0] == pts[3][0] && pts[0][1] == pts[3][1] && pts[1][1] == pts[2][1]
	if !aligned {
		aligned = pts[0][0] == pts[3][0] && pts[1][0] == pts[2][0] && pts[0][1] == pts[1][1] && pts[2][1] == pts[3][1]
	}
	if aligned {
		minX := math.Floor(math.Min(pts[0][0], pts[2][0]))
		maxX := math.Ceil(math.Max(pts[0][0], pts[2][0]))
		minY := math.Floor(math.Min(pts[0][1], pts[2][1]))
		maxY := math.Ceil(math.Max(pts[0][1], pts[2][1]))
		b.clearRect(int(minX), int(minY), int(maxX)-int(minX), int(maxY)-int(minY))
		return
	}

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

func (b *GoGLBackend) clearRect(x, y, w, h int) {
	gl.Enable(gl.SCISSOR_TEST)

	var box [4]int32
	gl.GetIntegerv(gl.SCISSOR_BOX, &box[0])

	gl.Scissor(int32(x), int32(b.h-y-h), int32(w), int32(h))
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Scissor(box[0], box[1], box[2], box[3])

	gl.Disable(gl.SCISSOR_TEST)
}

func (b *GoGLBackend) Fill(style *backendbase.FillStyle, pts [][2]float64) {
	b.activate()

	if style.Blur > 0 {
		b.offscr1.alpha = true
		b.enableTextureRenderTarget(&b.offscr1)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	}

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

	if style.Color.A >= 255 {
		vertex := b.useShader(style)

		gl.StencilFunc(gl.EQUAL, 0, 0xFF)
		gl.EnableVertexAttribArray(vertex)
		gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, nil)
		gl.DrawArrays(mode, 4, int32(len(pts)))
		gl.DisableVertexAttribArray(vertex)
		gl.StencilFunc(gl.ALWAYS, 0, 0xFF)
	} else {
		gl.ColorMask(false, false, false, false)
		gl.StencilFunc(gl.ALWAYS, 1, 0xFF)
		gl.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
		gl.StencilMask(0x01)

		gl.UseProgram(b.sr.ID)
		gl.Uniform4f(b.sr.Color, 0, 0, 0, 0)
		gl.Uniform2f(b.sr.CanvasSize, float32(b.fw), float32(b.fh))
		gl.Uniform1f(b.sr.GlobalAlpha, 1)

		gl.EnableVertexAttribArray(b.sr.Vertex)
		gl.VertexAttribPointer(b.sr.Vertex, 2, gl.FLOAT, false, 0, nil)
		gl.DrawArrays(mode, 4, int32(len(pts)))
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

	if style.Blur > 0 {
		b.drawBlurred(style.Blur)
	}
}

func (b *GoGLBackend) FillImageMask(style *backendbase.FillStyle, mask *image.Alpha, pts [][2]float64) {
	b.activate()

	w, h := mask.Rect.Dx(), mask.Rect.Dy()

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, b.alphaTex)
	for y := 0; y < h; y++ {
		off := y * mask.Stride
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, int32(alphaTexSize-1-y), int32(w), 1, gl.ALPHA, gl.UNSIGNED_BYTE, gl.Ptr(&mask.Pix[off]))
	}

	if style.Blur > 0 {
		b.offscr1.alpha = true
		b.enableTextureRenderTarget(&b.offscr1)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	}

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)

	vertex, alphaTexCoord := b.useAlphaShader(style, 1)

	gl.EnableVertexAttribArray(vertex)
	gl.EnableVertexAttribArray(alphaTexCoord)

	tw := float64(w) / alphaTexSize
	th := float64(h) / alphaTexSize
	var buf [16]float32
	data := buf[:0]
	for _, pt := range pts {
		data = append(data, float32(pt[0]), float32(pt[1]))
	}
	data = append(data, 0, 1, 0, float32(1-th), float32(tw), float32(1-th), float32(tw), 1)

	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(alphaTexCoord, 2, gl.FLOAT, false, 0, gl.PtrOffset(8*4))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	gl.DisableVertexAttribArray(vertex)
	gl.DisableVertexAttribArray(alphaTexCoord)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, b.alphaTex)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

	for y := 0; y < h; y++ {
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, int32(alphaTexSize-1-y), int32(w), 1, gl.ALPHA, gl.UNSIGNED_BYTE, gl.Ptr(&zeroes[0]))
	}

	gl.ActiveTexture(gl.TEXTURE0)

	if style.Blur > 0 {
		b.drawBlurred(style.Blur)
	}
}

func (b *GoGLBackend) drawBlurred(blur float64) {
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

	var kernel []float32
	var kernelBuf [255]float32
	var gs *gaussianShader
	if blur < 3 {
		gs = &b.gauss15r
		kernel = kernelBuf[:15]
	} else if blur < 12 {
		gs = &b.gauss63r
		kernel = kernelBuf[:63]
	} else {
		gs = &b.gauss127r
		kernel = kernelBuf[:127]
	}

	gaussianKernel(blur, kernel)

	b.offscr2.alpha = true
	b.enableTextureRenderTarget(&b.offscr2)
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.shadowBuf)
	data := [16]float32{0, 0, 0, float32(b.h), float32(b.w), float32(b.h), float32(b.w), 0, 0, 0, 0, 1, 1, 1, 1, 0}
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, b.offscr1.tex)

	gl.UseProgram(gs.ID)
	gl.Uniform1i(gs.Image, 0)
	gl.Uniform2f(gs.CanvasSize, float32(b.fw), float32(b.fh))
	gl.Uniform2f(gs.KernelScale, 1.0/float32(b.fw), 0.0)
	gl.Uniform1fv(gs.Kernel, int32(len(kernel)), &kernel[0])
	gl.VertexAttribPointer(gs.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(gs.TexCoord, 2, gl.FLOAT, false, 0, gl.PtrOffset(8*4))
	gl.EnableVertexAttribArray(gs.Vertex)
	gl.EnableVertexAttribArray(gs.TexCoord)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(gs.Vertex)
	gl.DisableVertexAttribArray(gs.TexCoord)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

	b.disableTextureRenderTarget()

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.shadowBuf)
	data = [16]float32{0, 0, 0, float32(b.h), float32(b.w), float32(b.h), float32(b.w), 0, 0, 0, 0, 1, 1, 1, 1, 0}
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, b.offscr2.tex)

	gl.UseProgram(gs.ID)
	gl.Uniform1i(gs.Image, 0)
	gl.Uniform2f(gs.CanvasSize, float32(b.fw), float32(b.fh))
	gl.Uniform2f(gs.KernelScale, 0.0, 1.0/float32(b.fh))
	gl.Uniform1fv(gs.Kernel, int32(len(kernel)), &kernel[0])
	gl.VertexAttribPointer(gs.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(gs.TexCoord, 2, gl.FLOAT, false, 0, gl.PtrOffset(8*4))
	gl.EnableVertexAttribArray(gs.Vertex)
	gl.EnableVertexAttribArray(gs.TexCoord)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(gs.Vertex)
	gl.DisableVertexAttribArray(gs.TexCoord)

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
