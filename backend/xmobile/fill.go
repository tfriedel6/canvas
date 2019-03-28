package xmobilebackend

import (
	"image"
	"math"
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"golang.org/x/mobile/gl"
)

func (b *XMobileBackend) Clear(pts [4][2]float64) {
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

	b.glctx.UseProgram(b.sr.ID)
	b.glctx.Uniform2f(b.sr.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.Uniform4f(b.sr.Color, 0, 0, 0, 0)
	b.glctx.Uniform1f(b.sr.GlobalAlpha, 1)

	b.glctx.Disable(gl.BLEND)

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&data[0]), len(data)*4), gl.STREAM_DRAW)

	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)

	b.glctx.VertexAttribPointer(b.sr.Vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.EnableVertexAttribArray(b.sr.Vertex)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	b.glctx.DisableVertexAttribArray(b.sr.Vertex)

	b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)

	b.glctx.Enable(gl.BLEND)
}

func (b *XMobileBackend) clearRect(x, y, w, h int) {
	b.glctx.Enable(gl.SCISSOR_TEST)

	var box [4]int32
	b.glctx.GetIntegerv(box[:], gl.SCISSOR_BOX)

	b.glctx.Scissor(int32(x), int32(b.h-y-h), int32(w), int32(h))
	b.glctx.ClearColor(0, 0, 0, 0)
	b.glctx.Clear(gl.COLOR_BUFFER_BIT)
	b.glctx.Scissor(box[0], box[1], box[2], box[3])

	b.glctx.Disable(gl.SCISSOR_TEST)
}

func (b *XMobileBackend) Fill(style *backendbase.FillStyle, pts [][2]float64) {
	b.activate()

	if style.Blur > 0 {
		b.offscr1.alpha = true
		b.enableTextureRenderTarget(&b.offscr1)
		b.glctx.ClearColor(0, 0, 0, 0)
		b.glctx.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
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

	mode := gl.Enum(gl.TRIANGLES)
	if len(pts) == 4 {
		mode = gl.TRIANGLE_FAN
	}

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&b.ptsBuf[0]), len(b.ptsBuf)*4), gl.STREAM_DRAW)

	if style.Color.A >= 255 {
		vertex := b.useShader(style)

		b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)
		b.glctx.EnableVertexAttribArray(vertex)
		b.glctx.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, 0)
		b.glctx.DrawArrays(mode, 4, len(pts))
		b.glctx.DisableVertexAttribArray(vertex)
		b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)
	} else {
		b.glctx.ColorMask(false, false, false, false)
		b.glctx.StencilFunc(gl.ALWAYS, 1, 0xFF)
		b.glctx.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
		b.glctx.StencilMask(0x01)

		b.glctx.UseProgram(b.sr.ID)
		b.glctx.Uniform4f(b.sr.Color, 0, 0, 0, 0)
		b.glctx.Uniform2f(b.sr.CanvasSize, float32(b.fw), float32(b.fh))

		b.glctx.EnableVertexAttribArray(b.sr.Vertex)
		b.glctx.VertexAttribPointer(b.sr.Vertex, 2, gl.FLOAT, false, 0, 0)
		b.glctx.DrawArrays(mode, 4, len(pts))
		b.glctx.DisableVertexAttribArray(b.sr.Vertex)

		b.glctx.ColorMask(true, true, true, true)

		b.glctx.StencilFunc(gl.EQUAL, 1, 0xFF)

		vertex := b.useShader(style)
		b.glctx.EnableVertexAttribArray(vertex)
		b.glctx.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, 0)

		b.ptsBuf = append(b.ptsBuf[:0], 0, 0, float32(b.fw), 0, float32(b.fw), float32(b.fh), 0, float32(b.fh))
		b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
		b.glctx.DisableVertexAttribArray(vertex)

		b.glctx.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
		b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)

		b.glctx.Clear(gl.STENCIL_BUFFER_BIT)
		b.glctx.StencilMask(0xFF)
	}

	if style.Blur > 0 {
		b.drawBlurred(style.Blur)
	}
}

func (b *XMobileBackend) FillImageMask(style *backendbase.FillStyle, mask *image.Alpha, pts [][2]float64) {
	b.activate()

	w, h := mask.Rect.Dx(), mask.Rect.Dy()

	b.glctx.ActiveTexture(gl.TEXTURE1)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.alphaTex)
	for y := 0; y < h; y++ {
		off := y * mask.Stride
		b.glctx.TexSubImage2D(gl.TEXTURE_2D, 0, 0, alphaTexSize-1-y, w, 1, gl.ALPHA, gl.UNSIGNED_BYTE, mask.Pix[off:])
	}

	if style.Blur > 0 {
		b.offscr1.alpha = true
		b.enableTextureRenderTarget(&b.offscr1)
		b.glctx.ClearColor(0, 0, 0, 0)
		b.glctx.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	}

	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buf)

	vertex, alphaTexCoord := b.useAlphaShader(style, 1)

	b.glctx.EnableVertexAttribArray(vertex)
	b.glctx.EnableVertexAttribArray(alphaTexCoord)

	tw := float64(w) / alphaTexSize
	th := float64(h) / alphaTexSize
	var buf [16]float32
	data := buf[:0]
	for _, pt := range pts {
		data = append(data, float32(pt[0]), float32(pt[1]))
	}
	data = append(data, 0, 1, 0, float32(1-th), float32(tw), float32(1-th), float32(tw), 1)

	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&data[0]), len(data)*4), gl.STREAM_DRAW)

	b.glctx.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.VertexAttribPointer(alphaTexCoord, 2, gl.FLOAT, false, 0, 8*4)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	b.glctx.DisableVertexAttribArray(vertex)
	b.glctx.DisableVertexAttribArray(alphaTexCoord)

	b.glctx.ActiveTexture(gl.TEXTURE1)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.alphaTex)

	b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)

	for y := 0; y < h; y++ {
		b.glctx.TexSubImage2D(gl.TEXTURE_2D, 0, 0, alphaTexSize-1-y, w, 1, gl.ALPHA, gl.UNSIGNED_BYTE, zeroes[0:])
	}

	b.glctx.ActiveTexture(gl.TEXTURE0)

	if style.Blur > 0 {
		b.drawBlurred(style.Blur)
	}
}

func (b *XMobileBackend) drawBlurred(blur float64) {
	b.glctx.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

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
	b.glctx.ClearColor(0, 0, 0, 0)
	b.glctx.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)

	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.shadowBuf)
	data := [16]float32{0, 0, 0, float32(b.h), float32(b.w), float32(b.h), float32(b.w), 0, 0, 0, 0, 1, 1, 1, 1, 0}
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&data[0]), len(data)*4), gl.STREAM_DRAW)

	b.glctx.ActiveTexture(gl.TEXTURE0)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.offscr1.tex)

	b.glctx.UseProgram(gs.ID)
	b.glctx.Uniform1i(gs.Image, 0)
	b.glctx.Uniform2f(gs.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.Uniform2f(gs.KernelScale, 1.0/float32(b.fw), 0.0)
	b.glctx.Uniform1fv(gs.Kernel, kernel)
	b.glctx.VertexAttribPointer(gs.Vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.VertexAttribPointer(gs.TexCoord, 2, gl.FLOAT, false, 0, 8*4)
	b.glctx.EnableVertexAttribArray(gs.Vertex)
	b.glctx.EnableVertexAttribArray(gs.TexCoord)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	b.glctx.DisableVertexAttribArray(gs.Vertex)
	b.glctx.DisableVertexAttribArray(gs.TexCoord)

	b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)

	b.disableTextureRenderTarget()

	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.shadowBuf)
	data = [16]float32{0, 0, 0, float32(b.h), float32(b.w), float32(b.h), float32(b.w), 0, 0, 0, 0, 1, 1, 1, 1, 0}
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&data[0]), len(data)*4), gl.STREAM_DRAW)

	b.glctx.ActiveTexture(gl.TEXTURE0)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.offscr2.tex)

	b.glctx.UseProgram(gs.ID)
	b.glctx.Uniform1i(gs.Image, 0)
	b.glctx.Uniform2f(gs.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.Uniform2f(gs.KernelScale, 0.0, 1.0/float32(b.fh))
	b.glctx.Uniform1fv(gs.Kernel, kernel)
	b.glctx.VertexAttribPointer(gs.Vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.VertexAttribPointer(gs.TexCoord, 2, gl.FLOAT, false, 0, 8*4)
	b.glctx.EnableVertexAttribArray(gs.Vertex)
	b.glctx.EnableVertexAttribArray(gs.TexCoord)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	b.glctx.DisableVertexAttribArray(gs.Vertex)
	b.glctx.DisableVertexAttribArray(gs.TexCoord)

	b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)

	b.glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
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
