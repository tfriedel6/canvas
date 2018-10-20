package canvas

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"unsafe"
)

// Image represents a loaded image that can be used in various drawing functions
type Image struct {
	w, h    int
	tex     uint32
	deleted bool
	opaque  bool
}

var images = make(map[interface{}]*Image)

// LoadImage loads an image. The src parameter can be either an image from the
// standard image package, a byte slice that will be loaded, or a file name
// string. If you want the canvas package to load the image, make sure you
// import the required format packages
func LoadImage(src interface{}) (*Image, error) {
	if gli == nil {
		panic("LoadGL must be called before images can be loaded")
	}

	var tex uint32
	gli.GenTextures(1, &tex)
	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, tex)
	if src == nil {
		return &Image{tex: tex}, nil
	}

	img, err := loadImage(src, tex)
	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(img, func(img *Image) {
		if !img.deleted {
			glChan <- func() {
				gli.DeleteTextures(1, &img.tex)
			}
		}
	})

	return img, nil
}

func loadImage(src interface{}, tex uint32) (*Image, error) {
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
	case string:
		data, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, err
		}
		srcImg, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		return loadImage(srcImg, tex)
	case []byte:
		srcImg, _, err := image.Decode(bytes.NewReader(v))
		if err != nil {
			return nil, err
		}
		return loadImage(srcImg, tex)
	default:
		return nil, errors.New("Unsupported source type")
	}
	return img, nil
}

func getImage(src interface{}) *Image {
	if img, ok := images[src]; ok {
		return img
	}
	switch v := src.(type) {
	case *Image:
		return v
	case image.Image:
		img, err := LoadImage(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading image: %v\n", err)
			images[src] = nil
			return nil
		}
		images[v] = img
		return img
	case string:
		img, err := LoadImage(v)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "format") {
				fmt.Fprintf(os.Stderr, "Error loading image %s: %v\nIt may be necessary to import the appropriate decoder, e.g.\nimport _ \"image/jpeg\"\n", v, err)
			} else {
				fmt.Fprintf(os.Stderr, "Error loading image %s: %v\n", v, err)
			}
			images[src] = nil
			return nil
		}
		images[v] = img
		return img
	}
	fmt.Fprintf(os.Stderr, "Unknown image type: %T\n", src)
	images[src] = nil
	return nil
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

func loadImageGray(src *image.Gray, tex uint32) (*Image, error) {
	img := &Image{tex: tex, w: src.Bounds().Dx(), h: src.Bounds().Dy()}
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

func loadImageConverted(src image.Image, tex uint32) (*Image, error) {
	img := &Image{tex: tex, w: src.Bounds().Dx(), h: src.Bounds().Dy(), opaque: true}
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
			if a < 255 {
				img.opaque = false
			}
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

// Width returns the width of the image
func (img *Image) Width() int { return img.w }

// Height returns the height of the image
func (img *Image) Height() int { return img.h }

// Size returns the width and height of the image
func (img *Image) Size() (int, int) { return img.w, img.h }

// Delete deletes the image from memory. Any draw calls with a deleted image
// will not do anything
func (img *Image) Delete() {
	gli.DeleteTextures(1, &img.tex)
	img.deleted = true
}

// Replace replaces the image with the new one
func (img *Image) Replace(src interface{}) {
	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, img.tex)
	newImg, err := loadImage(src, img.tex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error replacing image: %v\n", err)
		return
	}
	*img = *newImg
}

// DrawImage draws the given image to the given coordinates. The image
// parameter can be an Image loaded by LoadImage, a file name string that will
// be loaded and cached, or a name string that corresponds to a previously
// loaded image with the same name parameter.
//
// The coordinates must be one of the following:
//  DrawImage("image", dx, dy)
//  DrawImage("image", dx, dy, dw, dh)
//  DrawImage("image", sx, sy, sw, sh, dx, dy, dw, dh)
// Where dx/dy/dw/dh are the destination coordinates and sx/sy/sw/sh are the
// source coordinates
func (cv *Canvas) DrawImage(image interface{}, coords ...float64) {
	img := getImage(image)

	if img == nil {
		return
	}

	if img.deleted {
		return
	}

	cv.activate()

	var sx, sy, sw, sh, dx, dy, dw, dh float64
	sw, sh = float64(img.w), float64(img.h)
	dw, dh = float64(img.w), float64(img.h)
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

	sx /= float64(img.w)
	sy /= float64(img.h)
	sw /= float64(img.w)
	sh /= float64(img.h)

	p0 := cv.tf(vec{dx, dy})
	p1 := cv.tf(vec{dx, dy + dh})
	p2 := cv.tf(vec{dx + dw, dy + dh})
	p3 := cv.tf(vec{dx + dw, dy})

	if cv.state.shadowColor.a != 0 {
		tris := [24]float32{
			0, 0,
			float32(cv.fw), 0,
			float32(cv.fw), float32(cv.fh),
			0, 0,
			float32(cv.fw), float32(cv.fh),
			0, float32(cv.fh),
			float32(p0[0]), float32(p0[1]),
			float32(p3[0]), float32(p3[1]),
			float32(p2[0]), float32(p2[1]),
			float32(p0[0]), float32(p0[1]),
			float32(p2[0]), float32(p2[1]),
			float32(p1[0]), float32(p1[1]),
		}
		cv.drawShadow(tris[:])
	}

	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [16]float32{float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), float32(p3[0]), float32(p3[1]),
		float32(sx), float32(sy), float32(sx), float32(sy + sh), float32(sx + sw), float32(sy + sh), float32(sx + sw), float32(sy)}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, img.tex)

	gli.UseProgram(ir.id)
	gli.Uniform1i(ir.image, 0)
	gli.Uniform2f(ir.canvasSize, float32(cv.fw), float32(cv.fh))
	gli.Uniform1f(ir.globalAlpha, float32(cv.state.globalAlpha))
	gli.VertexAttribPointer(ir.vertex, 2, gl_FLOAT, false, 0, 0)
	gli.VertexAttribPointer(ir.texCoord, 2, gl_FLOAT, false, 0, 8*4)
	gli.EnableVertexAttribArray(ir.vertex)
	gli.EnableVertexAttribArray(ir.texCoord)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(ir.vertex)
	gli.DisableVertexAttribArray(ir.texCoord)

	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)
}

func (cv *Canvas) drawImageTemp(image interface{}, coords ...float64) {
	img := getImage(image)

	if img == nil {
		return
	}

	if img.deleted {
		return
	}

	cv.activate()

	var sx, sy, sw, sh, dx, dy, dw, dh float64
	sw, sh = float64(img.w), float64(img.h)
	dw, dh = float64(img.w), float64(img.h)
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

	sx /= float64(img.w)
	sy /= float64(img.h)
	sw /= float64(img.w)
	sh /= float64(img.h)

	p0 := vec{dx, dy}
	p1 := vec{dx, dy + dh}
	p2 := vec{dx + dw, dy + dh}
	p3 := vec{dx + dw, dy}

	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	gli.BindBuffer(gl_ARRAY_BUFFER, shadowBuf)
	data := [16]float32{float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), float32(p3[0]), float32(p3[1]),
		float32(sx), float32(sy), float32(sx), float32(sy + sh), float32(sx + sw), float32(sy + sh), float32(sx + sw), float32(sy)}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, img.tex)

	gli.UseProgram(ir.id)
	gli.Uniform1i(ir.image, 0)
	gli.Uniform2f(ir.canvasSize, float32(cv.fw), float32(cv.fh))
	gli.Uniform1f(ir.globalAlpha, float32(cv.state.globalAlpha))
	gli.VertexAttribPointer(ir.vertex, 2, gl_FLOAT, false, 0, 0)
	gli.VertexAttribPointer(ir.texCoord, 2, gl_FLOAT, false, 0, 8*4)
	gli.EnableVertexAttribArray(ir.vertex)
	gli.EnableVertexAttribArray(ir.texCoord)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(ir.vertex)
	gli.DisableVertexAttribArray(ir.texCoord)

	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)
}
