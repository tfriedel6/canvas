package canvas

import (
	"image"
	"image/color"
	"unsafe"
)

var imageBufTex uint32
var imageBuf []byte

// GetImageData returns an RGBA image of the currently displayed image. The
// alpha channel is always opaque
func (cv *Canvas) GetImageData(x, y, w, h int) *image.RGBA {
	cv.activate()

	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if w > cv.w {
		w = cv.w
	}
	if h > cv.h {
		h = cv.h
	}
	if len(imageBuf) < w*h*3 {
		imageBuf = make([]byte, w*h*3)
	}

	gli.ReadPixels(int32(x), int32(y), int32(w), int32(h), gl_RGB, gl_UNSIGNED_BYTE, gli.Ptr(&imageBuf[0]))
	rgba := image.NewRGBA(image.Rect(x, y, x+w, y+h))
	bp := 0
	for cy := y; cy < y+h; cy++ {
		for cx := x; cx < x+w; cx++ {
			rgba.SetRGBA(cx, y+h-1-cy, color.RGBA{R: imageBuf[bp], G: imageBuf[bp+1], B: imageBuf[bp+2], A: 255})
			bp += 3
		}
	}
	return rgba
}

// PutImageData puts the given image at the given x/y coordinates
func (cv *Canvas) PutImageData(img *image.RGBA, x, y int) {
	cv.activate()

	gli.ActiveTexture(gl_TEXTURE0)
	if imageBufTex == 0 {
		gli.GenTextures(1, &imageBufTex)
		gli.BindTexture(gl_TEXTURE_2D, imageBufTex)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_LINEAR)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_LINEAR)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_T, gl_CLAMP_TO_EDGE)
	} else {
		gli.BindTexture(gl_TEXTURE_2D, imageBufTex)
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	if img.Stride == img.Bounds().Dx()*4 {
		gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RGBA, int32(w), int32(h), 0, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&img.Pix[0]))
	} else {
		data := make([]uint8, 0, w*h*4)
		for cy := 0; cy < h; cy++ {
			start := cy * img.Stride
			end := start + w*4
			data = append(data, img.Pix[start:end]...)
		}
		gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RGBA, int32(w), int32(h), 0, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&data[0]))
	}

	dx, dy := float32(x), float32(y)
	dw, dh := float32(w), float32(h)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [16]float32{dx, dy, dx + dw, dy, dx + dw, dy + dh, dx, dy + dh,
		0, 0, 1, 0, 1, 1, 0, 1}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.UseProgram(ir.id)
	gli.Uniform1i(ir.image, 0)
	gli.Uniform2f(ir.canvasSize, float32(cv.fw), float32(cv.fh))
	gli.Uniform1f(ir.globalAlpha, 1)
	gli.VertexAttribPointer(ir.vertex, 2, gl_FLOAT, false, 0, 0)
	gli.VertexAttribPointer(ir.texCoord, 2, gl_FLOAT, false, 0, 8*4)
	gli.EnableVertexAttribArray(ir.vertex)
	gli.EnableVertexAttribArray(ir.texCoord)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(ir.vertex)
	gli.DisableVertexAttribArray(ir.texCoord)
}
