package xmobilebackend

import (
	"image"
	"math"
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"golang.org/x/mobile/gl"
)

func (b *XMobileBackend) Clear(pts [4]backendbase.Vec) {
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

	b.glctx.UseProgram(b.shd.ID)
	b.glctx.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.UniformMatrix3fv(b.shd.Matrix, mat3identity[:])
	b.glctx.Uniform4f(b.shd.Color, 0, 0, 0, 0)
	b.glctx.Uniform1f(b.shd.GlobalAlpha, 1)
	b.glctx.Uniform1i(b.shd.UseAlphaTex, 0)
	b.glctx.Uniform1i(b.shd.Func, shdFuncSolid)

	b.glctx.Disable(gl.BLEND)

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&data[0]), len(data)*4), gl.STREAM_DRAW)

	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)

	b.glctx.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, 0)
	b.glctx.EnableVertexAttribArray(b.shd.Vertex)
	b.glctx.EnableVertexAttribArray(b.shd.TexCoord)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	b.glctx.DisableVertexAttribArray(b.shd.Vertex)
	b.glctx.DisableVertexAttribArray(b.shd.TexCoord)

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

func extent(pts []backendbase.Vec) (min, max backendbase.Vec) {
	max[0] = -math.MaxFloat64
	max[1] = -math.MaxFloat64
	min[0] = math.MaxFloat64
	min[1] = math.MaxFloat64
	for _, v := range pts {
		min[0] = math.Min(min[0], v[0])
		min[1] = math.Min(min[1], v[1])
		max[0] = math.Max(max[0], v[0])
		max[1] = math.Max(max[1], v[1])
	}
	return
}

func (b *XMobileBackend) Fill(style *backendbase.FillStyle, pts []backendbase.Vec, tf backendbase.Mat, canOverlap bool) {
	b.activate()

	if style.Blur > 0 {
		b.offscr1.alpha = true
		b.enableTextureRenderTarget(&b.offscr1)
		b.glctx.ClearColor(0, 0, 0, 0)
		b.glctx.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	}

	b.ptsBuf = b.ptsBuf[:0]
	min, max := extent(pts)
	b.ptsBuf = append(b.ptsBuf,
		float32(min[0]), float32(min[1]),
		float32(min[0]), float32(max[1]),
		float32(max[0]), float32(max[1]),
		float32(max[0]), float32(min[1]))
	for _, pt := range pts {
		b.ptsBuf = append(b.ptsBuf, float32(pt[0]), float32(pt[1]))
	}

	mode := gl.Enum(gl.TRIANGLES)
	if len(pts) == 4 {
		mode = gl.TRIANGLE_FAN
	}

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&b.ptsBuf[0]), len(b.ptsBuf)*4), gl.STREAM_DRAW)

	if !canOverlap || style.Color.A >= 255 {
		vertex, _ := b.useShader(style, mat3(tf), false, 0)

		b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)
		b.glctx.EnableVertexAttribArray(vertex)
		b.glctx.EnableVertexAttribArray(b.shd.TexCoord)
		b.glctx.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, 0)
		b.glctx.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, 0)
		b.glctx.DrawArrays(mode, 4, len(pts))
		b.glctx.DisableVertexAttribArray(vertex)
		b.glctx.DisableVertexAttribArray(b.shd.TexCoord)
		b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)
	} else {
		b.glctx.ColorMask(false, false, false, false)
		b.glctx.StencilFunc(gl.ALWAYS, 1, 0xFF)
		b.glctx.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
		b.glctx.StencilMask(0x01)

		b.glctx.UseProgram(b.shd.ID)
		b.glctx.Uniform4f(b.shd.Color, 0, 0, 0, 0)
		b.glctx.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
		m3 := mat3(tf)
		b.glctx.UniformMatrix3fv(b.shd.Matrix, m3[:])
		b.glctx.Uniform1f(b.shd.GlobalAlpha, 1)
		b.glctx.Uniform1i(b.shd.UseAlphaTex, 0)
		b.glctx.Uniform1i(b.shd.Func, shdFuncSolid)

		b.glctx.EnableVertexAttribArray(b.shd.Vertex)
		b.glctx.EnableVertexAttribArray(b.shd.TexCoord)
		b.glctx.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, 0)
		b.glctx.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, 0)
		b.glctx.DrawArrays(mode, 4, len(pts))
		b.glctx.DisableVertexAttribArray(b.shd.Vertex)
		b.glctx.DisableVertexAttribArray(b.shd.TexCoord)

		b.glctx.ColorMask(true, true, true, true)

		b.glctx.StencilFunc(gl.EQUAL, 1, 0xFF)

		vertex, _ := b.useShader(style, mat3identity, false, 0)
		b.glctx.EnableVertexAttribArray(vertex)
		b.glctx.EnableVertexAttribArray(b.shd.TexCoord)
		b.glctx.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, 0)
		b.glctx.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, 0)

		b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
		b.glctx.DisableVertexAttribArray(vertex)
		b.glctx.DisableVertexAttribArray(b.shd.TexCoord)

		b.glctx.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
		b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)

		b.glctx.Clear(gl.STENCIL_BUFFER_BIT)
		b.glctx.StencilMask(0xFF)
	}

	if style.Blur > 0 {
		b.drawBlurred(style.Blur, min, max)
	}
}

func (b *XMobileBackend) FillImageMask(style *backendbase.FillStyle, mask *image.Alpha, pts [4]backendbase.Vec) {
	b.activate()

	w, h := mask.Rect.Dx(), mask.Rect.Dy()

	b.glctx.ActiveTexture(gl.TEXTURE1)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.alphaTex)
	b.glctx.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, mask.Stride, h, gl.ALPHA, gl.UNSIGNED_BYTE, mask.Pix[0:])
	if w < alphaTexSize {
		b.glctx.TexSubImage2D(gl.TEXTURE_2D, 0, w, 0, 1, h, gl.ALPHA, gl.UNSIGNED_BYTE, zeroes[0:])
	}
	if h < alphaTexSize {
		b.glctx.TexSubImage2D(gl.TEXTURE_2D, 0, 0, h, w, 1, gl.ALPHA, gl.UNSIGNED_BYTE, zeroes[0:])
	}

	if style.Blur > 0 {
		b.offscr1.alpha = true
		b.enableTextureRenderTarget(&b.offscr1)
		b.glctx.ClearColor(0, 0, 0, 0)
		b.glctx.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	}

	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buf)

	vertex, alphaTexCoord := b.useShader(style, mat3identity, true, 1)

	b.glctx.EnableVertexAttribArray(vertex)
	b.glctx.EnableVertexAttribArray(alphaTexCoord)

	tw := float64(w) / alphaTexSize
	th := float64(h) / alphaTexSize
	var buf [16]float32
	data := buf[:0]
	for _, pt := range pts {
		data = append(data, float32(math.Round(pt[0])), float32(math.Round(pt[1])))
	}
	data = append(data, 0, 0, 0, float32(th), float32(tw), float32(th), float32(tw), 0)

	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&data[0]), len(data)*4), gl.STREAM_DRAW)

	b.glctx.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.VertexAttribPointer(alphaTexCoord, 2, gl.FLOAT, false, 0, 8*4)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	b.glctx.DisableVertexAttribArray(vertex)
	b.glctx.DisableVertexAttribArray(alphaTexCoord)

	b.glctx.ActiveTexture(gl.TEXTURE1)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.alphaTex)

	b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)

	b.glctx.ActiveTexture(gl.TEXTURE0)

	if style.Blur > 0 {
		min, max := extent(pts[:])
		b.drawBlurred(style.Blur, min, max)
	}
}

func (b *XMobileBackend) drawBlurred(size float64, min, max backendbase.Vec) {
	b.offscr1.alpha = true
	b.offscr2.alpha = true

	// calculate box blur size
	fsize := math.Max(1, math.Floor(size))
	sizea := int(fsize)
	sizeb := sizea
	sizec := sizea
	if size-fsize > 0.333333333 {
		sizeb++
	}
	if size-fsize > 0.666666666 {
		sizec++
	}

	min[0] -= fsize * 3
	min[1] -= fsize * 3
	max[0] += fsize * 3
	max[1] += fsize * 3
	min[0] = math.Max(0.0, math.Min(b.fw, min[0]))
	min[1] = math.Max(0.0, math.Min(b.fh, min[1]))
	max[0] = math.Max(0.0, math.Min(b.fw, max[0]))
	max[1] = math.Max(0.0, math.Min(b.fh, max[1]))

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.shadowBuf)
	data := [16]float32{
		float32(min[0]), float32(min[1]),
		float32(min[0]), float32(max[1]),
		float32(max[0]), float32(max[1]),
		float32(max[0]), float32(min[1]),
		float32(min[0] / b.fw), 1 - float32(min[1]/b.fh),
		float32(min[0] / b.fw), 1 - float32(max[1]/b.fh),
		float32(max[0] / b.fw), 1 - float32(max[1]/b.fh),
		float32(max[0] / b.fw), 1 - float32(min[1]/b.fh),
	}
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&data[0]), len(data)*4), gl.STREAM_DRAW)

	b.glctx.UseProgram(b.shd.ID)
	b.glctx.Uniform1i(b.shd.Image, 0)
	b.glctx.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.UniformMatrix3fv(b.shd.Matrix, mat3identity[:])
	b.glctx.Uniform1i(b.shd.UseAlphaTex, 0)
	b.glctx.Uniform1i(b.shd.Func, shdFuncBoxBlur)

	b.glctx.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, 8*4)
	b.glctx.EnableVertexAttribArray(b.shd.Vertex)
	b.glctx.EnableVertexAttribArray(b.shd.TexCoord)

	b.glctx.Disable(gl.BLEND)

	b.glctx.ActiveTexture(gl.TEXTURE0)

	b.glctx.ClearColor(0, 0, 0, 0)

	b.enableTextureRenderTarget(&b.offscr2)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.offscr1.tex)
	b.glctx.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	b.box3(sizea, 0, false)
	b.enableTextureRenderTarget(&b.offscr1)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.offscr2.tex)
	b.box3(sizeb, -0.5, false)
	b.enableTextureRenderTarget(&b.offscr2)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.offscr1.tex)
	b.box3(sizec, 0, false)
	b.enableTextureRenderTarget(&b.offscr1)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.offscr2.tex)
	b.box3(sizea, 0, true)
	b.enableTextureRenderTarget(&b.offscr2)
	b.glctx.BindTexture(gl.TEXTURE_2D, b.offscr1.tex)
	b.box3(sizeb, -0.5, true)
	b.glctx.Enable(gl.BLEND)
	b.glctx.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	b.disableTextureRenderTarget()
	b.glctx.BindTexture(gl.TEXTURE_2D, b.offscr2.tex)
	b.box3(sizec, 0, true)

	b.glctx.DisableVertexAttribArray(b.shd.Vertex)
	b.glctx.DisableVertexAttribArray(b.shd.TexCoord)

	b.glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (b *XMobileBackend) box3(size int, offset float32, vertical bool) {
	b.glctx.Uniform1i(b.shd.BoxSize, size)
	if vertical {
		b.glctx.Uniform1i(b.shd.BoxVertical, 1)
		b.glctx.Uniform1f(b.shd.BoxScale, 1/float32(b.fh))
	} else {
		b.glctx.Uniform1i(b.shd.BoxVertical, 0)
		b.glctx.Uniform1f(b.shd.BoxScale, 1/float32(b.fw))
	}
	b.glctx.Uniform1f(b.shd.BoxOffset, offset)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
}
