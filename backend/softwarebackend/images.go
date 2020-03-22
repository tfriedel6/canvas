package softwarebackend

import (
	"image"
	"image/color"
	"math"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

type Image struct {
	mips    []image.Image
	deleted bool
}

func (b *SoftwareBackend) LoadImage(img image.Image) (backendbase.Image, error) {
	bimg := &Image{mips: make([]image.Image, 1, 10)}
	bimg.Replace(img)
	return bimg, nil
}

func halveImage(img image.Image) (*image.RGBA, int, int) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	w = w / 2
	h = h / 2
	rimg := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		sy := y * 2
		for x := 0; x < w; x++ {
			sx := x * 2
			r1, g1, b1, a1 := img.At(sx, sy).RGBA()
			r2, g2, b2, a2 := img.At(sx+1, sy).RGBA()
			r3, g3, b3, a3 := img.At(sx, sy+1).RGBA()
			r4, g4, b4, a4 := img.At(sx+1, sy+1).RGBA()
			mixr := uint8((int(r1) + int(r2) + int(r3) + int(r4)) / 1024)
			mixg := uint8((int(g1) + int(g2) + int(g3) + int(g4)) / 1024)
			mixb := uint8((int(b1) + int(b2) + int(b3) + int(b4)) / 1024)
			mixa := uint8((int(a1) + int(a2) + int(a3) + int(a4)) / 1024)
			rimg.Set(x, y, color.RGBA{R: mixr, G: mixg, B: mixb, A: mixa})
		}
	}
	return rimg, w, h
}

func (b *SoftwareBackend) DrawImage(dimg backendbase.Image, sx, sy, sw, sh float64, pts [4]backendbase.Vec, alpha float64) {
	simg := dimg.(*Image)
	if simg.deleted {
		return
	}

	bounds := simg.mips[0].Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	factor := float64(w*h) / (sw * sh)
	area := quadArea(pts) * factor
	mip := simg.mips[0]
	closest := math.MaxFloat64
	mipW, mipH := w, h
	for _, img := range simg.mips {
		bounds := img.Bounds()
		w, h := bounds.Dx(), bounds.Dy()
		dist := math.Abs(float64(w*h) - area)
		if dist < closest {
			closest = dist
			mip = img
			mipW = w
			mipH = h
		}
	}

	mipScaleX := float64(mipW) / float64(w)
	mipScaleY := float64(mipH) / float64(h)
	sx *= mipScaleX
	sy *= mipScaleY
	sw *= mipScaleX
	sh *= mipScaleY

	b.fillQuad(pts, func(x, y, tx, ty float64) color.RGBA {
		imgx := sx + sw*tx
		imgy := sy + sh*ty
		imgxf := math.Floor(imgx)
		imgyf := math.Floor(imgy)
		return toRGBA(mip.At(int(imgxf), int(imgyf)))

		// rx := imgx - imgxf
		// ry := imgy - imgyf
		// ca := mip.At(int(imgxf), int(imgyf))
		// cb := mip.At(int(imgxf+1), int(imgyf))
		// cc := mip.At(int(imgxf), int(imgyf+1))
		// cd := mip.At(int(imgxf+1), int(imgyf+1))
		// ctop := lerp(ca, cb, rx)
		// cbtm := lerp(cc, cd, rx)
		// b.Image.Set(x, y, lerp(ctop, cbtm, ry))
	})
}

func (img *Image) Width() int {
	return img.mips[0].Bounds().Dx()
}

func (img *Image) Height() int {
	return img.mips[0].Bounds().Dy()
}

func (img *Image) Size() (w, h int) {
	b := img.mips[0].Bounds()
	return b.Dx(), b.Dy()
}

func (img *Image) Delete() {
	img.deleted = true
}

func (img *Image) Replace(src image.Image) error {
	img.mips = img.mips[:1]
	img.mips[0] = src

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	for w > 1 && h > 1 {
		src, w, h = halveImage(src)
		img.mips = append(img.mips, src)
	}

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
