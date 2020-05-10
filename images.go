package canvas

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

// Image is a type holding information on an image loaded with the LoadImage
// function
type Image struct {
	src      interface{}
	cv       *Canvas
	img      backendbase.Image
	deleted  bool
	lastUsed time.Time
}

// LoadImage loads an image. The src parameter can be either an image from the
// standard image package, a byte slice that will be loaded, or a file name
// string. If you want the canvas package to load the image, make sure you
// import the required format packages
func (cv *Canvas) LoadImage(src interface{}) (*Image, error) {
	var reload *Image
	if img, ok := src.(*Image); ok {
		if img.cv != cv {
			panic("image loaded with different canvas")
		}
		if img.deleted {
			reload = img
			src = img.src
		} else {
			img.lastUsed = time.Now()
			return img, nil
		}
	} else if _, ok := src.([]byte); !ok {
		if img, ok := cv.images[src]; ok {
			img.lastUsed = time.Now()
			return img, nil
		}
	}
	cv.reduceCache(Performance.CacheSize, 0)
	var srcImg image.Image
	switch v := src.(type) {
	case image.Image:
		srcImg = v
	case string:
		data, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, err
		}
		srcImg, _, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	case []byte:
		var err error
		srcImg, _, err = image.Decode(bytes.NewReader(v))
		if err != nil {
			return nil, err
		}
	case *Canvas:
		w, h := cv.b.Size()
		src = cv.GetImageData(0, 0, w, h)
	default:
		return nil, errors.New("Unsupported source type")
	}
	backendImg, err := cv.b.LoadImage(srcImg)
	if err != nil {
		return nil, err
	}
	cvimg := &Image{cv: cv, img: backendImg, lastUsed: time.Now(), src: src}
	if reload != nil {
		*reload = *cvimg
		return reload, nil
	}
	if _, ok := src.([]byte); !ok {
		cv.images[src] = cvimg
	}
	return cvimg, nil
}

func (cv *Canvas) getImage(src interface{}) *Image {
	if cv2, ok := src.(*Canvas); ok {
		if !cv.b.CanUseAsImage(cv2.b) {
			w, h := cv2.Size()
			return cv.getImage(cv2.GetImageData(0, 0, w, h))
		}
		bimg := cv2.b.AsImage()
		if bimg == nil {
			w, h := cv2.Size()
			return cv.getImage(cv2.GetImageData(0, 0, w, h))
		}
		return &Image{cv: cv, img: bimg}
	}

	img, err := cv.LoadImage(src)
	if err != nil {
		cv.images[src] = nil
		switch v := src.(type) {
		case image.Image:
			fmt.Fprintf(os.Stderr, "Error loading image: %v\n", err)
		case string:
			if strings.Contains(strings.ToLower(err.Error()), "format") {
				fmt.Fprintf(os.Stderr, "Error loading image %s: %v\nIt may be necessary to import the appropriate decoder, e.g.\nimport _ \"image/jpeg\"\n", v, err)
			} else {
				fmt.Fprintf(os.Stderr, "Error loading image %s: %v\n", v, err)
			}
		default:
			fmt.Fprintf(os.Stderr, "Failed to load image: %v\n", err)
		}
	}
	return img
}

// Width returns the width of the image
func (img *Image) Width() int { return img.img.Width() }

// Height returns the height of the image
func (img *Image) Height() int { return img.img.Height() }

// Size returns the width and height of the image
func (img *Image) Size() (int, int) { return img.img.Size() }

// Delete deletes the image from memory
func (img *Image) Delete() {
	if img == nil || img.deleted {
		return
	}
	img.deleted = true
	img.img.Delete()
	delete(img.cv.images, img.src)
}

// Replace replaces the image with the new one
func (img *Image) Replace(src interface{}) error {
	if img.src == src {
		if origImg, ok := img.src.(image.Image); ok {
			img.img.Replace(origImg)
			return nil
		}
	}
	newImg, err := img.cv.LoadImage(src)
	if err != nil {
		return err
	}
	img.img = newImg.img
	return nil
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
	img := cv.getImage(image)
	if img == nil {
		return
	}

	var sx, sy, sw, sh, dx, dy, dw, dh float64
	sw, sh = float64(img.Width()), float64(img.Height())
	dw, dh = float64(img.Width()), float64(img.Height())
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

	var data [4]backendbase.Vec
	data[0] = cv.tf(backendbase.Vec{dx, dy})
	data[1] = cv.tf(backendbase.Vec{dx, dy + dh})
	data[2] = cv.tf(backendbase.Vec{dx + dw, dy + dh})
	data[3] = cv.tf(backendbase.Vec{dx + dw, dy})

	cv.drawShadow(data[:], nil, false)

	cv.b.DrawImage(img.img, sx, sy, sw, sh, data, cv.state.globalAlpha)
}

// GetImageData returns an RGBA image of the current image
func (cv *Canvas) GetImageData(x, y, w, h int) *image.RGBA {
	return cv.b.GetImageData(x, y, w, h)
}

// PutImageData puts the given image at the given x/y coordinates
func (cv *Canvas) PutImageData(img *image.RGBA, x, y int) {
	cv.b.PutImageData(img, x, y)
}

// ImagePattern is an image pattern that can be used for any
// fill call
type ImagePattern struct {
	cv  *Canvas
	img *Image
	tf  backendbase.Mat
	rep imagePatternRepeat
	ip  backendbase.ImagePattern
}

type imagePatternRepeat uint8

// Image pattern repeat constants
const (
	Repeat   imagePatternRepeat = imagePatternRepeat(backendbase.Repeat)
	RepeatX                     = imagePatternRepeat(backendbase.RepeatX)
	RepeatY                     = imagePatternRepeat(backendbase.RepeatY)
	NoRepeat                    = imagePatternRepeat(backendbase.NoRepeat)
)

func (ip *ImagePattern) data(tf backendbase.Mat) backendbase.ImagePatternData {
	m := tf.Invert().Mul(ip.tf.Invert())
	return backendbase.ImagePatternData{
		Image: ip.img.img,
		Transform: [9]float64{
			m[0], m[2], m[4],
			m[1], m[3], m[5],
			0, 0, 1,
		},
		Repeat: backendbase.ImagePatternRepeat(ip.rep),
	}
}

// SetTransform changes the transformation of the image pattern
// to the given matrix. The matrix is a 3x3 matrix, but three
// of the values are always identity values
func (ip *ImagePattern) SetTransform(tf [6]float64) {
	ip.tf = backendbase.Mat(tf)
}

// CreatePattern creates a new image pattern with the specified
// image and repetition
func (cv *Canvas) CreatePattern(src interface{}, repeat imagePatternRepeat) *ImagePattern {
	ip := &ImagePattern{
		cv:  cv,
		img: cv.getImage(src),
		rep: repeat,
		tf:  backendbase.Mat{1, 0, 0, 1, 0, 0},
	}
	if ip.img != nil {
		ip.ip = cv.b.LoadImagePattern(ip.data(cv.state.transform))
	}
	return ip
}
