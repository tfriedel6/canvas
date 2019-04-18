package softwarebackend

import (
	"image"
	"math"

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
		imgx := sx + sw*sx2
		imgy := sy + sh*sy2
		imgxf := math.Floor(imgx)
		imgyf := math.Floor(imgy)
		rx := imgx - imgxf
		ry := imgy - imgyf
		ca := simg.img.At(int(imgxf), int(imgyf))
		cb := simg.img.At(int(imgxf+1), int(imgyf))
		cc := simg.img.At(int(imgxf), int(imgyf+1))
		cd := simg.img.At(int(imgxf+1), int(imgyf+1))
		ctop := lerp(ca, cb, rx)
		cbtm := lerp(cc, cd, rx)
		b.Image.Set(x, y, lerp(ctop, cbtm, ry))
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
