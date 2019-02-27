package softwarebackend

import (
	"image"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

type Image struct {
	img image.Image
}

func (b *SoftwareBackend) LoadImage(img image.Image) (backendbase.Image, error) {
	return &Image{img: img}, nil
}

func (b *SoftwareBackend) DrawImage(dimg backendbase.Image, sx, sy, sw, sh float64, pts [4][2]float64, alpha float64) {
}

func (b *SoftwareBackend) FillImageMask(style *backendbase.FillStyle, mask *image.Alpha, pts [][2]float64) {
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
}

func (img *Image) Replace(src image.Image) error {
	img.img = src
	return nil
}
