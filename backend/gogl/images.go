package goglbackend

import (
	"errors"
	"image"
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/tfriedel6/canvas/backend/backendbase"
)

// Image represents a loaded image that can be used in various drawing functions
type Image struct {
	w, h    int
	tex     uint32
	deleted bool
	opaque  bool
}

func (b *GoGLBackend) LoadImage(src image.Image) (backendbase.Image, error) {
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	if src == nil {
		return &Image{tex: tex}, nil
	}

	img, err := loadImage(src, tex)
	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(img, func(img *Image) {
		if !img.deleted {
			b.glChan <- func() {
				gl.DeleteTextures(1, &img.tex)
			}
		}
	})

	return img, nil
}

func loadImage(src image.Image, tex uint32) (*Image, error) {
	var img *Image
	var err error
	switch v := src.(type) {
	case *image.RGBA:
		img, err = loadImageRGBA(v, tex)
		if err != nil {
			return nil, err
		}
	case *image.Gray:
		img, err = loadImageGray(v, tex)
		if err != nil {
			return nil, err
		}
	case image.Image:
		img, err = loadImageConverted(v, tex)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Unsupported source type")
	}
	return img, nil
}

func loadImageRGBA(src *image.RGBA, tex uint32) (*Image, error) {
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

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	if err := glError(); err != nil {
		return nil, err
	}
	if src.Stride == img.w*4 {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(img.w), int32(img.h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&src.Pix[0]))
	} else {
		data := make([]uint8, 0, img.w*img.h*4)
		for y := 0; y < img.h; y++ {
			start := y * src.Stride
			end := start + img.w*4
			data = append(data, src.Pix[start:end]...)
		}
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(img.w), int32(img.h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&data[0]))
	}
	if err := glError(); err != nil {
		return nil, err
	}
	gl.GenerateMipmap(gl.TEXTURE_2D)
	if err := glError(); err != nil {
		return nil, err
	}
	return img, nil
}

func loadImageGray(src *image.Gray, tex uint32) (*Image, error) {
	img := &Image{tex: tex, w: src.Bounds().Dx(), h: src.Bounds().Dy()}
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	if err := glError(); err != nil {
		return nil, err
	}
	if src.Stride == img.w {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, int32(img.w), int32(img.h), 0, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(&src.Pix[0]))
	} else {
		data := make([]uint8, 0, img.w*img.h)
		for y := 0; y < img.h; y++ {
			start := y * src.Stride
			end := start + img.w
			data = append(data, src.Pix[start:end]...)
		}
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, int32(img.w), int32(img.h), 0, gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(&data[0]))
	}
	if err := glError(); err != nil {
		return nil, err
	}
	gl.GenerateMipmap(gl.TEXTURE_2D)
	if err := glError(); err != nil {
		return nil, err
	}
	return img, nil
}

func loadImageConverted(src image.Image, tex uint32) (*Image, error) {
	img := &Image{tex: tex, w: src.Bounds().Dx(), h: src.Bounds().Dy(), opaque: true}
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	if err := glError(); err != nil {
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
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(img.w), int32(img.h), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&data[0]))
	if err := glError(); err != nil {
		return nil, err
	}
	gl.GenerateMipmap(gl.TEXTURE_2D)
	if err := glError(); err != nil {
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
	gl.DeleteTextures(1, &img.tex)
	img.deleted = true
}

// IsDeleted returns true if the Delete function has been
// called on this image
func (img *Image) IsDeleted() bool { return img.deleted }

// Replace replaces the image with the new one
func (img *Image) Replace(src image.Image) error {
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, img.tex)
	newImg, err := loadImage(src, img.tex)
	if err != nil {
		return err
	}
	*img = *newImg
	return nil
}

// IsOpaque returns true if all pixels in the image
// have a full alpha value
func (img *Image) IsOpaque() bool { return img.opaque }

func (b *GoGLBackend) DrawImage(dimg backendbase.Image, sx, sy, sw, sh, dx, dy, dw, dh float64, alpha float64) {
	img := dimg.(*Image)

	sx /= float64(img.w)
	sy /= float64(img.h)
	sw /= float64(img.w)
	sh /= float64(img.h)

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	data := [16]float32{
		float32(dx), float32(dy),
		float32(dx), float32(dy + dh),
		float32(dx + dw), float32(dy + dh),
		float32(dx + dw), float32(dy),
		float32(sx), float32(sy),
		float32(sx), float32(sy + sh),
		float32(sx + sw), float32(sy + sh),
		float32(sx + sw), float32(sy),
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, img.tex)

	gl.UseProgram(b.ir.ID)
	gl.Uniform1i(b.ir.Image, 0)
	gl.Uniform2f(b.ir.CanvasSize, float32(b.fw), float32(b.fh))
	gl.Uniform1f(b.ir.GlobalAlpha, float32(alpha))
	gl.VertexAttribPointer(b.ir.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(b.ir.TexCoord, 2, gl.FLOAT, false, 0, gl.PtrOffset(8*4))
	gl.EnableVertexAttribArray(b.ir.Vertex)
	gl.EnableVertexAttribArray(b.ir.TexCoord)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(b.ir.Vertex)
	gl.DisableVertexAttribArray(b.ir.TexCoord)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)
}
