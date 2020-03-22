package goglbackend

import (
	"errors"
	"image"
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"github.com/tfriedel6/canvas/backend/goglbackend/gl"
)

// Image represents a loaded image that can be used in various drawing functions
type Image struct {
	b    *GoGLBackend
	w, h int
	tex  uint32
	flip bool
}

func (b *GoGLBackend) LoadImage(src image.Image) (backendbase.Image, error) {
	b.activate()

	var tex uint32
	gl.GenTextures(1, &tex)
	if tex == 0 {
		return nil, errors.New("glGenTextures failed")
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	if src == nil {
		return &Image{tex: tex}, nil
	}

	img, err := loadImage(src, tex)
	if err != nil {
		return nil, err
	}
	img.b = b

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
	img := &Image{tex: tex, w: src.Bounds().Dx(), h: src.Bounds().Dy()}

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

func loadImageConverted(src image.Image, tex uint32) (*Image, error) {
	img := &Image{tex: tex, w: src.Bounds().Dx(), h: src.Bounds().Dy()}
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
	img.b.activate()

	gl.DeleteTextures(1, &img.tex)
}

// Replace replaces the image with the new one
func (img *Image) Replace(src image.Image) error {
	img.b.activate()

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, img.tex)
	newImg, err := loadImage(src, img.tex)
	if err != nil {
		return err
	}
	newImg.b = img.b
	*img = *newImg
	return nil
}

func (b *GoGLBackend) DrawImage(dimg backendbase.Image, sx, sy, sw, sh float64, pts [4]backendbase.Vec, alpha float64) {
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

	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.buf)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl.STREAM_DRAW)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, img.tex)

	gl.UseProgram(b.shd.ID)
	gl.Uniform1i(b.shd.Image, 0)
	gl.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	gl.UniformMatrix3fv(b.shd.Matrix, 1, false, &mat3identity[0])
	gl.Uniform1f(b.shd.GlobalAlpha, float32(alpha))
	gl.Uniform1i(b.shd.UseAlphaTex, 0)
	gl.Uniform1i(b.shd.Func, shdFuncImage)
	gl.VertexAttribPointer(b.shd.Vertex, 2, gl.FLOAT, false, 0, nil)
	gl.VertexAttribPointer(b.shd.TexCoord, 2, gl.FLOAT, false, 0, gl.PtrOffset(8*4))
	gl.EnableVertexAttribArray(b.shd.Vertex)
	gl.EnableVertexAttribArray(b.shd.TexCoord)
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 4)
	gl.DisableVertexAttribArray(b.shd.Vertex)
	gl.DisableVertexAttribArray(b.shd.TexCoord)

	gl.StencilFunc(gl.ALWAYS, 0, 0xFF)
}

type ImagePattern struct {
	b    *GoGLBackend
	data backendbase.ImagePatternData
}

func (b *GoGLBackend) LoadImagePattern(data backendbase.ImagePatternData) backendbase.ImagePattern {
	return &ImagePattern{
		b:    b,
		data: data,
	}
}

func (ip *ImagePattern) Delete()                                   {}
func (ip *ImagePattern) Replace(data backendbase.ImagePatternData) { ip.data = data }
