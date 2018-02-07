package canvas

import (
	"bytes"
	"errors"
	"image"
	"io/ioutil"
	"runtime"
	"unsafe"
)

type Image struct {
	w, h    int
	tex     uint32
	deleted bool
}

func LoadImage(src interface{}) (*Image, error) {
	var img *Image
	var err error
	switch v := src.(type) {
	case *image.RGBA:
		img, err = loadImageRGBA(v)
		if err != nil {
			return nil, err
		}
	case *image.Gray:
		img, err = loadImageGray(v)
		if err != nil {
			return nil, err
		}
	case image.Image:
		img, err = loadImageConverted(v)
		if err != nil {
			return nil, err
		}
	case string:
		data, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, err
		}
		srcImg, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return LoadImage(srcImg)
	case []byte:
		srcImg, _, err := image.Decode(bytes.NewReader(v))
		if err != nil {
			return nil, err
		}
		return LoadImage(srcImg)
	default:
		return nil, errors.New("Unsupported source type")
	}
	runtime.SetFinalizer(img, func(img *Image) {
		glChan <- func() {
			gli.DeleteTextures(1, &img.tex)
		}
	})
	return img, nil
}

func loadImageRGBA(src *image.RGBA) (*Image, error) {
	img := &Image{w: src.Bounds().Dx(), h: src.Bounds().Dy()}
	gli.GenTextures(1, &img.tex)
	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, img.tex)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_LINEAR_MIPMAP_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_T, gl_CLAMP_TO_EDGE)
	if err := glError(); err != nil {
		return nil, err
	}
	if src.Stride == img.w*4 {
		gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RGBA, int32(img.w), int32(img.h), 0, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&src.Pix[0]))
	} else {
		data := make([]uint8, 0, img.w*img.h*4)
		for y := 0; y < img.h; y++ {
			start := y * src.Stride
			end := start + img.w*4
			data = append(data, src.Pix[start:end]...)
		}
		gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RGBA, int32(img.w), int32(img.h), 0, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&data[0]))
	}
	if err := glError(); err != nil {
		return nil, err
	}
	gli.GenerateMipmap(gl_TEXTURE_2D)
	if err := glError(); err != nil {
		return nil, err
	}
	return img, nil
}

func loadImageGray(src *image.Gray) (*Image, error) {
	img := &Image{w: src.Bounds().Dx(), h: src.Bounds().Dy()}
	gli.GenTextures(1, &img.tex)
	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, img.tex)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_LINEAR_MIPMAP_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_T, gl_CLAMP_TO_EDGE)
	if err := glError(); err != nil {
		return nil, err
	}
	if src.Stride == img.w {
		gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RED, int32(img.w), int32(img.h), 0, gl_RED, gl_UNSIGNED_BYTE, gli.Ptr(&src.Pix[0]))
	} else {
		data := make([]uint8, 0, img.w*img.h)
		for y := 0; y < img.h; y++ {
			start := y * src.Stride
			end := start + img.w
			data = append(data, src.Pix[start:end]...)
		}
		gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RED, int32(img.w), int32(img.h), 0, gl_RED, gl_UNSIGNED_BYTE, gli.Ptr(&data[0]))
	}
	if err := glError(); err != nil {
		return nil, err
	}
	gli.GenerateMipmap(gl_TEXTURE_2D)
	if err := glError(); err != nil {
		return nil, err
	}
	return img, nil
}

func loadImageConverted(src image.Image) (*Image, error) {
	img := &Image{w: src.Bounds().Dx(), h: src.Bounds().Dy()}
	gli.GenTextures(1, &img.tex)
	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, img.tex)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_LINEAR_MIPMAP_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_T, gl_CLAMP_TO_EDGE)
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
	gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RGBA, int32(img.w), int32(img.h), 0, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&data[0]))
	if err := glError(); err != nil {
		return nil, err
	}
	gli.GenerateMipmap(gl_TEXTURE_2D)
	if err := glError(); err != nil {
		return nil, err
	}
	return img, nil
}

func (img *Image) W() int           { return img.w }
func (img *Image) H() int           { return img.h }
func (img *Image) Size() (int, int) { return img.w, img.h }

func (img *Image) Delete() {
	gli.DeleteTextures(1, &img.tex)
	img.deleted = true
}

func (cv *Canvas) DrawImage(img *Image, coords ...float32) {
	if img.deleted {
		return
	}

	var sx, sy, sw, sh, dx, dy, dw, dh float32
	sw, sh = float32(img.w), float32(img.h)
	dw, dh = float32(img.w), float32(img.h)
	if len(coords) == 2 {
		dx, dy = coords[0], coords[1]
	} else if len(coords) == 4 {
		dx, dy = coords[0], coords[1]
		dw, dh = coords[2], coords[3]
	} else if len(coords) == 8 {
		sx, sy = coords[0], coords[1]
		sw, sh = coords[2], coords[3]
		dx, dy = coords[4], coords[5]
		dw, dh = coords[6], coords[7]
	}

	dx0, dy0 := cv.tfToGL(dx, dy)
	dx1, dy1 := cv.tfToGL(dx, dy+dh)
	dx2, dy2 := cv.tfToGL(dx+dw, dy+dh)
	dx3, dy3 := cv.tfToGL(dx+dw, dy)
	sx /= float32(img.w)
	sy /= float32(img.h)
	sw /= float32(img.w)
	sh /= float32(img.h)

	cv.activate()

	gli.UseProgram(tr.id)

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, img.tex)
	gli.Uniform1i(tr.image, 0)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [16]float32{dx0, dy0, dx1, dy1, dx2, dy2, dx3, dy3,
		sx, sy, sx, sy + sh, sx + sw, sy + sh, sx + sw, sy}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.VertexAttribPointer(tr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.VertexAttribPointer(tr.texCoord, 2, gl_FLOAT, false, 0, gli.PtrOffset(8*4))
	gli.EnableVertexAttribArray(tr.vertex)
	gli.EnableVertexAttribArray(tr.texCoord)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(tr.vertex)
	gli.DisableVertexAttribArray(tr.texCoord)
}
