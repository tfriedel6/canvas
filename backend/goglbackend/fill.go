package goglbackend

import (
	"image"
	"math"
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"github.com/tfriedel6/canvas/backend/goglbackend/gl"
)

func (b *GoGLBackend) Clear(pts [4]backendbase.Vec) {
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

	gl.UseProgram(b.shd.ID)
	gl.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	gl.UniformMatrix3fv(b.shd.Matrix, 1, false, &mat3identity[0])
	gl.Uniform4f(b.shd.Color, 0, 0, 0, 0)
	gl.Uniform1f(b.shd.GlobalAlpha, 1)
	gl.Uniform1i(b.shd.UseAlphaTex, 0)
	gl.Uniform1i(b.shd.Func, shdFuncSolid)

	gl.Disable(gl.BLEND)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, nil)
	gl.EnableVertexAttribArray(b.shd.Vertex)
	gl.EnableVertexAttribArray(b.shd.TexCoord)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(b.shd.Vertex)
	gl.DisableVertexAttribArray(b.shd.TexCoord)

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

func (b *GoGLBackend) Fill(style *backendbase.FillStyle, pts []backendbase.Vec, tf backendbase.Mat, canOverlap bool) {
	b.activate()

	if style.Blur > 0 {
		b.offscr1.alpha = true
		b.enableTextureRenderTarget(&b.offscr1)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
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

	mode := uint32(gl.TRIANGLES)
	if len(pts) == 4 {
		mode = gl.TRIANGLE_FAN
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	gl.BufferData(gl.ARRAY_BUFFER, len(b.ptsBuf)*4, unsafe.Pointer(&b.ptsBuf[0]), gl.STREAM_DRAW)

	if !canOverlap || style.Color.A >= 255 {
		vertex, _ := b.useShader(style, mat3(tf), false, 0)

		gl.StencilFunc(gl.EQUAL, 0, 0xFF)
		gl.EnableVertexAttribArray(vertex)
		gl.EnableVertexAttribArray(b.shd.TexCoord)
		gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, nil)
		gl.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, nil)
		gl.DrawArrays(mode, 4, int32(len(pts)))
		gl.DisableVertexAttribArray(vertex)
		gl.DisableVertexAttribArray(b.shd.TexCoord)
		gl.StencilFunc(gl.ALWAYS, 0, 0xFF)
	} else {
		gl.ColorMask(false, false, false, false)
		gl.StencilFunc(gl.ALWAYS, 1, 0xFF)
		gl.StencilOp(gl.REPLACE, gl.REPLACE, gl.REPLACE)
		gl.StencilMask(0x01)

		gl.UseProgram(b.shd.ID)
		gl.Uniform4f(b.shd.Color, 0, 0, 0, 0)
		gl.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
		m3 := mat3(tf)
		gl.UniformMatrix3fv(b.shd.Matrix, 1, false, &m3[0])
		gl.Uniform1f(b.shd.GlobalAlpha, 1)
		gl.Uniform1i(b.shd.UseAlphaTex, 0)
		gl.Uniform1i(b.shd.Func, shdFuncSolid)

		gl.EnableVertexAttribArray(b.shd.Vertex)
		gl.EnableVertexAttribArray(b.shd.TexCoord)
		gl.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, nil)
		gl.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, nil)
		gl.DrawArrays(mode, 4, int32(len(pts)))
		gl.DisableVertexAttribArray(b.shd.Vertex)
		gl.DisableVertexAttribArray(b.shd.TexCoord)

		gl.ColorMask(true, true, true, true)

		gl.StencilFunc(gl.EQUAL, 1, 0xFF)

		vertex, _ := b.useShader(style, mat3identity, false, 0)
		gl.EnableVertexAttribArray(vertex)
		gl.EnableVertexAttribArray(b.shd.TexCoord)
		gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, nil)
		gl.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, nil)

		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
		gl.DisableVertexAttribArray(vertex)
		gl.DisableVertexAttribArray(b.shd.TexCoord)

		gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
		gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

		gl.Clear(gl.STENCIL_BUFFER_BIT)
		gl.StencilMask(0xFF)
	}

	if style.Blur > 0 {
		b.drawBlurred(style.Blur, min, max)
	}
}

func (b *GoGLBackend) FillImageMask(style *backendbase.FillStyle, mask *image.Alpha, pts [4]backendbase.Vec) {
	b.activate()

	w, h := mask.Rect.Dx(), mask.Rect.Dy()

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, b.alphaTex)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(mask.Stride), int32(h), gl.ALPHA, gl.UNSIGNED_BYTE, gl.Ptr(&mask.Pix[0]))
	if w < alphaTexSize {
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, int32(w), 0, 1, int32(h), gl.ALPHA, gl.UNSIGNED_BYTE, gl.Ptr(&zeroes[0]))
	}
	if h < alphaTexSize {
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, int32(h), int32(w), 1, gl.ALPHA, gl.UNSIGNED_BYTE, gl.Ptr(&zeroes[0]))
	}

	if style.Blur > 0 {
		b.offscr1.alpha = true
		b.enableTextureRenderTarget(&b.offscr1)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	}

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)

	vertex, alphaTexCoord := b.useShader(style, mat3identity, true, 1)

	gl.EnableVertexAttribArray(vertex)
	gl.EnableVertexAttribArray(alphaTexCoord)

	tw := float64(w) / alphaTexSize
	th := float64(h) / alphaTexSize
	var buf [16]float32
	data := buf[:0]
	for _, pt := range pts {
		data = append(data, float32(math.Round(pt[0])), float32(math.Round(pt[1])))
	}
	data = append(data, 0, 0, 0, float32(th), float32(tw), float32(th), float32(tw), 0)

	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.VertexAttribPointer(vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(alphaTexCoord, 2, gl.FLOAT, false, 0, gl.PtrOffset(8*4))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)

	gl.DisableVertexAttribArray(vertex)
	gl.DisableVertexAttribArray(alphaTexCoord)

	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, b.alphaTex)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)

	gl.ActiveTexture(gl.TEXTURE0)

	if style.Blur > 0 {
		min, max := extent(pts[:])
		b.drawBlurred(style.Blur, min, max)
	}
}

func (b *GoGLBackend) drawBlurred(size float64, min, max backendbase.Vec) {
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

	gl.BindBuffer(gl.ARRAY_BUFFER, b.shadowBuf)
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
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.UseProgram(b.shd.ID)
	gl.Uniform1i(b.shd.Image, 0)
	gl.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	gl.UniformMatrix3fv(b.shd.Matrix, 1, false, &mat3identity[0])
	gl.Uniform1i(b.shd.UseAlphaTex, 0)
	gl.Uniform1i(b.shd.Func, shdFuncBoxBlur)

	gl.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, gl.PtrOffset(8*4))
	gl.EnableVertexAttribArray(b.shd.Vertex)
	gl.EnableVertexAttribArray(b.shd.TexCoord)

	gl.Disable(gl.BLEND)

	gl.ActiveTexture(gl.TEXTURE0)

	gl.ClearColor(0, 0, 0, 0)

	b.enableTextureRenderTarget(&b.offscr2)
	gl.BindTexture(gl.TEXTURE_2D, b.offscr1.tex)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	b.box3(sizea, 0, false)
	b.enableTextureRenderTarget(&b.offscr1)
	gl.BindTexture(gl.TEXTURE_2D, b.offscr2.tex)
	b.box3(sizeb, -0.5, false)
	b.enableTextureRenderTarget(&b.offscr2)
	gl.BindTexture(gl.TEXTURE_2D, b.offscr1.tex)
	b.box3(sizec, 0, false)
	b.enableTextureRenderTarget(&b.offscr1)
	gl.BindTexture(gl.TEXTURE_2D, b.offscr2.tex)
	b.box3(sizea, 0, true)
	b.enableTextureRenderTarget(&b.offscr2)
	gl.BindTexture(gl.TEXTURE_2D, b.offscr1.tex)
	b.box3(sizeb, -0.5, true)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	b.disableTextureRenderTarget()
	gl.BindTexture(gl.TEXTURE_2D, b.offscr2.tex)
	b.box3(sizec, 0, true)

	gl.DisableVertexAttribArray(b.shd.Vertex)
	gl.DisableVertexAttribArray(b.shd.TexCoord)

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (b *GoGLBackend) box3(size int, offset float32, vertical bool) {
	gl.Uniform1i(b.shd.BoxSize, int32(size))
	if vertical {
		gl.Uniform1i(b.shd.BoxVertical, 1)
		gl.Uniform1f(b.shd.BoxScale, 1/float32(b.fh))
	} else {
		gl.Uniform1i(b.shd.BoxVertical, 0)
		gl.Uniform1f(b.shd.BoxScale, 1/float32(b.fw))
	}
	gl.Uniform1f(b.shd.BoxOffset, offset)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
}
