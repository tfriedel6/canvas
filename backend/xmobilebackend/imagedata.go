package xmobilebackend

import (
	"image"
	"image/color"
	"unsafe"

	"golang.org/x/mobile/gl"
)

// GetImageData returns an RGBA image of the current image
func (b *XMobileBackend) GetImageData(x, y, w, h int) *image.RGBA {
	b.activate()

	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if w > b.w {
		w = b.w
	}
	if h > b.h {
		h = b.h
	}

	var vp [4]int32
	b.glctx.GetIntegerv(vp[:], gl.VIEWPORT)

	size := int(vp[2] * vp[3] * 3)
	if len(b.imageBuf) < size {
		b.imageBuf = make([]byte, size)
	}
	b.glctx.ReadPixels(b.imageBuf[0:], int(vp[0]), int(vp[1]), int(vp[2]), int(vp[3]), gl.RGB, gl.UNSIGNED_BYTE)

	rgba := image.NewRGBA(image.Rect(x, y, x+w, y+h))
	for cy := y; cy < y+h; cy++ {
		bp := (int(vp[3])-h+cy)*int(vp[2])*3 + x*3
		for cx := x; cx < x+w; cx++ {
			rgba.SetRGBA(cx, y+h-1-cy, color.RGBA{R: b.imageBuf[bp], G: b.imageBuf[bp+1], B: b.imageBuf[bp+2], A: 255})
			bp += 3
		}
	}
	return rgba
}

// PutImageData puts the given image at the given x/y coordinates
func (b *XMobileBackend) PutImageData(img *image.RGBA, x, y int) {
	b.activate()

	b.glctx.ActiveTexture(gl.TEXTURE0)
	if b.imageBufTex.Value == 0 {
		b.imageBufTex = b.glctx.CreateTexture()
		b.glctx.BindTexture(gl.TEXTURE_2D, b.imageBufTex)
		b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	} else {
		b.glctx.BindTexture(gl.TEXTURE_2D, b.imageBufTex)
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	if img.Stride == img.Bounds().Dx()*4 {
		b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, w, h, gl.RGBA, gl.UNSIGNED_BYTE, img.Pix[0:])
	} else {
		data := make([]uint8, 0, w*h*4)
		for cy := 0; cy < h; cy++ {
			start := cy * img.Stride
			end := start + w*4
			data = append(data, img.Pix[start:end]...)
		}
		b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, w, h, gl.RGBA, gl.UNSIGNED_BYTE, data[0:])
	}

	dx, dy := float32(x), float32(y)
	dw, dh := float32(w), float32(h)

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	data := [16]float32{dx, dy, dx + dw, dy, dx + dw, dy + dh, dx, dy + dh,
		0, 0, 1, 0, 1, 1, 0, 1}
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&data[0]), len(data)*4), gl.STREAM_DRAW)

	b.glctx.UseProgram(b.shd.ID)
	b.glctx.Uniform1i(b.shd.Image, 0)
	b.glctx.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.UniformMatrix3fv(b.shd.Matrix, mat3identity[:])
	b.glctx.Uniform1f(b.shd.GlobalAlpha, 1)
	b.glctx.Uniform1i(b.shd.UseAlphaTex, 0)
	b.glctx.Uniform1i(b.shd.Func, shdFuncImage)
	b.glctx.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, 8*4)
	b.glctx.EnableVertexAttribArray(b.shd.Vertex)
	b.glctx.EnableVertexAttribArray(b.shd.TexCoord)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	b.glctx.DisableVertexAttribArray(b.shd.Vertex)
	b.glctx.DisableVertexAttribArray(b.shd.TexCoord)
}
