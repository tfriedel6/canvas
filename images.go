package canvas

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

type Image struct {
	cv  *Canvas
	img backendbase.Image
}

// LoadImage loads an image. The src parameter can be either an image from the
// standard image package, a byte slice that will be loaded, or a file name
// string. If you want the canvas package to load the image, make sure you
// import the required format packages
func (cv *Canvas) LoadImage(src interface{}) (*Image, error) {
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
		src = cv.GetImageData(0, 0, cv.Width(), cv.Height())
	default:
		return nil, errors.New("Unsupported source type")
	}
	backendImg, err := cv.b.LoadImage(srcImg)
	if err != nil {
		return nil, err
	}
	return &Image{cv: cv, img: backendImg}, nil
}

func (cv *Canvas) getImage(src interface{}) *Image {
	if img, ok := cv.images[src]; ok {
		return img
	}
	switch v := src.(type) {
	case *Image:
		return v
	case image.Image:
		img, err := cv.LoadImage(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading image: %v\n", err)
			cv.images[src] = nil
			return nil
		}
		cv.images[v] = img
		return img
	case string:
		img, err := cv.LoadImage(v)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "format") {
				fmt.Fprintf(os.Stderr, "Error loading image %s: %v\nIt may be necessary to import the appropriate decoder, e.g.\nimport _ \"image/jpeg\"\n", v, err)
			} else {
				fmt.Fprintf(os.Stderr, "Error loading image %s: %v\n", v, err)
			}
			cv.images[src] = nil
			return nil
		}
		cv.images[v] = img
		return img
	}
	fmt.Fprintf(os.Stderr, "Unknown image type: %T\n", src)
	cv.images[src] = nil
	return nil
}

// Width returns the width of the image
func (img *Image) Width() int { return img.img.Width() }

// Height returns the height of the image
func (img *Image) Height() int { return img.img.Height() }

// Size returns the width and height of the image
func (img *Image) Size() (int, int) { return img.img.Size() }

// Delete deletes the image from memory. Any draw calls with a deleted image
// will not do anything
func (img *Image) Delete() { img.img.Delete() }

// Replace replaces the image with the new one
func (img *Image) Replace(src interface{}) error {
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
	var img *Image
	// var flip bool
	// if cv2, ok := image.(*Canvas); ok && cv2.offscreen {
	// 	img = &cv2.offscrImg
	// 	flip = true
	// } else {
	img = cv.getImage(image)
	// }

	if img == nil {
		return
	}

	if img.img.IsDeleted() {
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

	// if flip {
	// 	dy += dh
	// 	dh = -dh
	// }

	var data [4][2]float64
	data[0] = cv.tf(vec{dx, dy})
	data[1] = cv.tf(vec{dx, dy + dh})
	data[2] = cv.tf(vec{dx + dw, dy + dh})
	data[3] = cv.tf(vec{dx + dw, dy})

	cv.drawShadow2(data[:], nil)

	cv.b.DrawImage(img.img, sx, sy, sw, sh, dx, dy, dw, dh, cv.state.globalAlpha)
}

// GetImageData returns an RGBA image of the current image
func (cv *Canvas) GetImageData(x, y, w, h int) *image.RGBA {
	return cv.b.GetImageData(x, y, w, h)
}

// PutImageData puts the given image at the given x/y coordinates
func (cv *Canvas) PutImageData(img *image.RGBA, x, y int) {
	cv.b.PutImageData(img, x, y)
}
