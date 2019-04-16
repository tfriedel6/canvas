package softwarebackend

import (
	"image"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

type Image struct {
	img     image.Image
	deleted bool
}

func (b *SoftwareBackend) LoadImage(img image.Image) (backendbase.Image, error) {
	return &Image{img: img}, nil
}

func (b *SoftwareBackend) DrawImage(dimg backendbase.Image, sx, sy, sw, sh float64, pts [4][2]float64, alpha float64) {
	simg := dimg.(*Image)
	if simg.deleted {
		return
	}
	b.fillQuad(pts, func(x, y int, sx2, sy2 float64) {
		sxi := int(sx + sw*sx2 + 0.5)
		syi := int(sy + sh*sy2 + 0.5)
		b.Image.Set(x, y, simg.img.At(sxi, syi))
	})
}

func (img *Image) Width() int {
	return img.img.Bounds().Dx()
}

func (img *Image) Height() int {
	return img.img.Bounds().Dy()
}

func (img *Image) Size() (w, h int) {
	b := img.img.Bounds()
	return b.Dx(), b.Dy()
}

func (img *Image) Delete() {
	img.deleted = true
}

func (img *Image) Replace(src image.Image) error {
	img.img = src
	return nil
}

type ImagePattern struct {
	data backendbase.ImagePatternData
}

func (b *SoftwareBackend) LoadImagePattern(data backendbase.ImagePatternData) backendbase.ImagePattern {
	return &ImagePattern{
		data: data,
	}
}

func (ip *ImagePattern) Delete()                                   {}
func (ip *ImagePattern) Replace(data backendbase.ImagePatternData) { ip.data = data }
