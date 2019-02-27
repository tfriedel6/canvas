package xmobilebackend

import (
	"errors"
	"image"
	"runtime"
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"golang.org/x/mobile/gl"
)

// Image represents a loaded image that can be used in various drawing functions
type Image struct {
	b       *XMobileBackend
	w, h    int
	tex     gl.Texture
	deleted bool
	opaque  bool
	flip    bool
}

func (b *XMobileBackend) LoadImage(src image.Image) (backendbase.Image, error) {
	b.activate()

	var tex gl.Texture
	tex = b.glctx.CreateTexture()
	b.glctx.ActiveTexture(gl.TEXTURE0)
	b.glctx.BindTexture(gl.TEXTURE_2D, tex)
	if src == nil {
		return &Image{tex: tex}, nil
	}

	img, err := loadImage(b, src, tex)
	if err != nil {
		return nil, err
	}
	img.b = b

	runtime.SetFinalizer(img, func(img *Image) {
		if !img.deleted {
			b.glChan <- func() {
				b.glctx.DeleteTexture(img.tex)
			}
		}
	})

	return img, nil
}

func loadImage(b *XMobileBackend, src image.Image, tex gl.Texture) (*Image, error) {
	var img *Image
	var err error
	switch v := src.(type) {
	case *image.RGBA:
		img, err = loadImageRGBA(b, v, tex)
		if err != nil {
			return nil, err
		}
	case *image.Gray:
		img, err = loadImageGray(b, v, tex)
		if err != nil {
			return nil, err
		}
	case image.Image:
		img, err = loadImageConverted(b, v, tex)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Unsupported source type")
	}
	return img, nil
}

func loadImageRGBA(b *XMobileBackend, src *image.RGBA, tex gl.Texture) (*Image, error) {
	img := &Image{tex: tex, w: src.Bounds().Dx(), h: src.Bounds().Dy(), opaque: true}

checkOpaque:
	for y := 0; y < img.h; y++ {
		off := src.PixOffset(0, y) + 3
		for x := 0; x < img.w; x++ {
			if src.Pix[off] < 255 {
				img.opaque = false
				break checkOpaque
			}
			off += 4
		}
	}

	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	if err := glError(b); err != nil {
		return nil, err
	}
	if src.Stride == img.w*4 {
		b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, img.w, img.h, gl.RGBA, gl.UNSIGNED_BYTE, src.Pix[0:])
	} else {
		data := make([]uint8, 0, img.w*img.h*4)
		for y := 0; y < img.h; y++ {
			start := y * src.Stride
			end := start + img.w*4
			data = append(data, src.Pix[start:end]...)
		}
		b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, img.w, img.h, gl.RGBA, gl.UNSIGNED_BYTE, data[0:])
	}
	if err := glError(b); err != nil {
		return nil, err
	}
	b.glctx.GenerateMipmap(gl.TEXTURE_2D)
	if err := glError(b); err != nil {
		return nil, err
	}
	return img, nil
}

func loadImageGray(b *XMobileBackend, src *image.Gray, tex gl.Texture) (*Image, error) {
	img := &Image{tex: tex, w: src.Bounds().Dx(), h: src.Bounds().Dy()}
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	if err := glError(b); err != nil {
		return nil, err
	}
	if src.Stride == img.w {
		b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.ALPHA, img.w, img.h, gl.ALPHA, gl.UNSIGNED_BYTE, src.Pix[0:])
	} else {
		data := make([]uint8, 0, img.w*img.h)
		for y := 0; y < img.h; y++ {
			start := y * src.Stride
			end := start + img.w
			data = append(data, src.Pix[start:end]...)
		}
		b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.ALPHA, img.w, img.h, gl.ALPHA, gl.UNSIGNED_BYTE, data[0:])
	}
	if err := glError(b); err != nil {
		return nil, err
	}
	b.glctx.GenerateMipmap(gl.TEXTURE_2D)
	if err := glError(b); err != nil {
		return nil, err
	}
	return img, nil
}

func loadImageConverted(b *XMobileBackend, src image.Image, tex gl.Texture) (*Image, error) {
	img := &Image{tex: tex, w: src.Bounds().Dx(), h: src.Bounds().Dy(), opaque: true}
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	if err := glError(b); err != nil {
		return nil, err
	}
	data := make([]uint8, 0, img.w*img.h*4)
	for y := 0; y < img.h; y++ {
		for x := 0; x < img.w; x++ {
			ir, ig, ib, ia := src.At(x, y).RGBA()
			r, g, b, a := uint8(ir>>8), uint8(ig>>8), uint8(ib>>8), uint8(ia>>8)
			data = append(data, r, g, b, a)
			if a < 255 {
				img.opaque = false
			}
		}
	}
	b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, img.w, img.h, gl.RGBA, gl.UNSIGNED_BYTE, data[0:])
	if err := glError(b); err != nil {
		return nil, err
	}
	b.glctx.GenerateMipmap(gl.TEXTURE_2D)
	if err := glError(b); err != nil {
		return nil, err
	}
	return img, nil
}

// Width returns the width of the image
func (img *Image) Width() int { return img.w }

// Height returns the height of the image
func (img *Image) Height() int { return img.h }

// Size returns the width and height of the image
func (img *Image) Size() (int, int) { return img.w, img.h }

// Delete deletes the image from memory. Any draw calls
// with a deleted image will not do anything
func (img *Image) Delete() {
	b := img.b
	img.b.activate()

	b.glctx.DeleteTexture(img.tex)
	img.deleted = true
}

// IsDeleted returns true if the Delete function has been
// called on this image
func (img *Image) IsDeleted() bool { return img.deleted }

// Replace replaces the image with the new one
func (img *Image) Replace(src image.Image) error {
	b := img.b
	img.b.activate()

	b.glctx.ActiveTexture(gl.TEXTURE0)
	b.glctx.BindTexture(gl.TEXTURE_2D, img.tex)
	newImg, err := loadImage(b, src, img.tex)
	if err != nil {
		return err
	}
	newImg.b = img.b
	*img = *newImg
	return nil
}

// IsOpaque returns true if all pixels in the image
// have a full alpha value
func (img *Image) IsOpaque() bool { return img.opaque }

func (b *XMobileBackend) DrawImage(dimg backendbase.Image, sx, sy, sw, sh float64, pts [4][2]float64, alpha float64) {
	b.activate()

	img := dimg.(*Image)

	sx /= float64(img.w)
	sy /= float64(img.h)
	sw /= float64(img.w)
	sh /= float64(img.h)

	if img.flip {
		sy += sh
		sh = -sh
	}

	var buf [16]float32
	data := buf[:0]
	for _, pt := range pts {
		data = append(data, float32(pt[0]), float32(pt[1]))
	}
	data = append(data,
		float32(sx), float32(sy),
		float32(sx), float32(sy+sh),
		float32(sx+sw), float32(sy+sh),
		float32(sx+sw), float32(sy),
	)

	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)

	b.glctx.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	b.glctx.BufferData(gl.ARRAY_BUFFER, byteSlice(unsafe.Pointer(&data[0]), len(data)*4), gl.STREAM_DRAW)

	b.glctx.ActiveTexture(gl.TEXTURE0)
	b.glctx.BindTexture(gl.TEXTURE_2D, img.tex)

	b.glctx.UseProgram(b.ir.ID)
	b.glctx.Uniform1i(b.ir.Image, 0)
	b.glctx.Uniform2f(b.ir.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.Uniform1f(b.ir.GlobalAlpha, float32(alpha))
	b.glctx.VertexAttribPointer(b.ir.Vertex, 2, gl.FLOAT, false, 0, 0)
	b.glctx.VertexAttribPointer(b.ir.TexCoord, 2, gl.FLOAT, false, 0, 8*4)
	b.glctx.EnableVertexAttribArray(b.ir.Vertex)
	b.glctx.EnableVertexAttribArray(b.ir.TexCoord)
	b.glctx.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	b.glctx.DisableVertexAttribArray(b.ir.Vertex)
	b.glctx.DisableVertexAttribArray(b.ir.TexCoord)

	b.glctx.StencilFunc(gl.ALWAYS, 0, 0xFF)
}
