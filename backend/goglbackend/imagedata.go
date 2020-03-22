package goglbackend

import (
	"image"
	"image/color"
	"unsafe"

	"github.com/tfriedel6/canvas/backend/goglbackend/gl"
)

// GetImageData returns an RGBA image of the current image
func (b *GoGLBackend) GetImageData(x, y, w, h int) *image.RGBA {
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
	gl.GetIntegerv(gl.VIEWPORT, &vp[0])

	size := int(vp[2] * vp[3] * 3)
	if len(b.imageBuf) < size {
		b.imageBuf = make([]byte, size)
	}
	gl.ReadPixels(vp[0], vp[1], vp[2], vp[3], gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(&b.imageBuf[0]))

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
func (b *GoGLBackend) PutImageData(img *image.RGBA, x, y int) {
	b.activate()

	gl.ActiveTexture(gl.TEXTURE0)
	if b.imageBufTex == 0 {
		gl.GenTextures(1, &b.imageBufTex)
		gl.BindTexture(gl.TEXTURE_2D, b.imageBufTex)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	} else {
		gl.BindTexture(gl.TEXTURE_2D, b.imageBufTex)
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	if img.Stride == img.Bounds().Dx()*4 {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&img.Pix[0]))
	} else {
		data := make([]uint8, 0, w*h*4)
		for cy := 0; cy < h; cy++ {
			start := cy * img.Stride
			end := start + w*4
			data = append(data, img.Pix[start:end]...)
		}
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(w), int32(h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&data[0]))
	}

	dx, dy := float32(x), float32(y)
	dw, dh := float32(w), float32(h)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	data := [16]float32{dx, dy, dx + dw, dy, dx + dw, dy + dh, dx, dy + dh,
		0, 0, 1, 0, 1, 1, 0, 1}
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.UseProgram(b.shd.ID)
	gl.Uniform1i(b.shd.Image, 0)
	gl.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	gl.UniformMatrix3fv(b.shd.Matrix, 1, false, &mat3identity[0])
	gl.Uniform1f(b.shd.GlobalAlpha, 1)
	gl.Uniform1i(b.shd.UseAlphaTex, 0)
	gl.Uniform1i(b.shd.Func, shdFuncImage)
	gl.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, gl.PtrOffset(8*4))
	gl.EnableVertexAttribArray(b.shd.Vertex)
	gl.EnableVertexAttribArray(b.shd.TexCoord)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(b.shd.Vertex)
	gl.DisableVertexAttribArray(b.shd.TexCoord)
}
