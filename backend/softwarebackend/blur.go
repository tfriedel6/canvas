package softwarebackend

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

func (b *SoftwareBackend) activateBlurTarget() {
	b.blurSwap = b.Image
	b.Image = image.NewRGBA(b.Image.Rect)
}

func (b *SoftwareBackend) drawBlurred(size float64) {
	blurred := box3(b.Image, size)
	b.Image = b.blurSwap
	draw.Draw(b.Image, b.Image.Rect, blurred, image.ZP, draw.Over)
}

/*
func gauss(img *image.RGBA, size float64) *image.RGBA {
	const maxDist = 128

	var buf [2*maxDist + 1]float64
	kernel := buf[:]
	gaussianKernel(size, kernel)

	dist := maxDist
	for x := 0; x < maxDist; x++ {
		if kernel[x] < 0.002 {
			continue
		}
		dist = maxDist - x
		kernel = kernel[x : len(kernel)-x]
		break
	}

	fmt.Println(dist)

	img = gaussx(img, kernel, dist)
	return gaussy(img, kernel, dist)
}

func gaussx(img *image.RGBA, kernel []float64, maxDist int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	w, h := bounds.Dx(), bounds.Dy()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			imgFrom, imgTo := x-maxDist, x+maxDist
			kernelFrom := 0
			if imgFrom < 0 {
				kernelFrom -= imgFrom
				imgFrom = 0
			}
			if imgTo >= w {
				imgTo = w - 1
			}
			kernelSum := 0.0
			var r, g, b, a float64
			for xx, j := imgFrom, kernelFrom; xx <= imgTo; xx++ {
				k := kernel[j]
				kernelSum += k
				col := img.RGBAAt(xx, y)
				r += float64(col.R) * k
				g += float64(col.G) * k
				b += float64(col.B) * k
				a += float64(col.A) * k
				j++
			}
			col := color.RGBA{
				R: uint8(math.Round(r / kernelSum)),
				G: uint8(math.Round(g / kernelSum)),
				B: uint8(math.Round(b / kernelSum)),
				A: uint8(math.Round(a / kernelSum)),
			}
			result.SetRGBA(x, y, col)
		}
	}

	return result
}

func gaussy(img *image.RGBA, kernel []float64, maxDist int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	w, h := bounds.Dx(), bounds.Dy()

	for y := 0; y < h; y++ {
		imgFrom, imgTo := y-maxDist, y+maxDist
		kernelFrom := 0
		if imgFrom < 0 {
			kernelFrom -= imgFrom
			imgFrom = 0
		}
		if imgTo >= h {
			imgTo = h - 1
		}
		for x := 0; x < w; x++ {
			kernelSum := 0.0
			var r, g, b, a float64
			for yy, j := imgFrom, kernelFrom; yy <= imgTo; yy++ {
				k := kernel[j]
				kernelSum += k
				col := img.RGBAAt(x, yy)
				r += float64(col.R) * k
				g += float64(col.G) * k
				b += float64(col.B) * k
				a += float64(col.A) * k
				j++
			}
			col := color.RGBA{
				R: uint8(math.Round(r / kernelSum)),
				G: uint8(math.Round(g / kernelSum)),
				B: uint8(math.Round(b / kernelSum)),
				A: uint8(math.Round(a / kernelSum)),
			}
			result.SetRGBA(x, y, col)
		}
	}

	return result
}

func gaussianKernel(stddev float64, target []float64) {
	stddevSqr := stddev * stddev
	center := float64(len(target) / 2)
	factor := 1.0 / math.Sqrt(2*math.Pi*stddevSqr)
	for i := range target {
		x := float64(i) - center
		target[i] = factor * math.Pow(math.E, -x*x/(2*stddevSqr))
	}
}
*/

func box3(img *image.RGBA, size float64) *image.RGBA {
	size *= 1 - 1/(size+1) // this just seems to improve the accuracy

	fsize := math.Floor(size)
	sizea := int(fsize)
	sizeb := sizea
	sizec := sizea
	if size-fsize > 0.333333333 {
		sizeb++
	}
	if size-fsize > 0.666666666 {
		sizec++
	}
	img = box3x(img, sizea)
	img = box3x(img, sizeb)
	img = box3x(img, sizec)
	img = box3y(img, sizea)
	img = box3y(img, sizeb)
	img = box3y(img, sizec)
	return img
}

func box3x(img *image.RGBA, size int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	w, h := bounds.Dx(), bounds.Dy()

	for y := 0; y < h; y++ {
		if size >= w {
			var r, g, b, a float64
			for x := 0; x < w; x++ {
				col := img.RGBAAt(x, y)
				r += float64(col.R)
				g += float64(col.G)
				b += float64(col.B)
				a += float64(col.A)
			}

			factor := 1.0 / float64(w)
			col := color.RGBA{
				R: uint8(math.Round(r * factor)),
				G: uint8(math.Round(g * factor)),
				B: uint8(math.Round(b * factor)),
				A: uint8(math.Round(a * factor)),
			}
			for x := 0; x < w; x++ {
				result.SetRGBA(x, y, col)
			}
			continue
		}

		var r, g, b, a float64
		for x := 0; x <= size; x++ {
			col := img.RGBAAt(x, y)
			r += float64(col.R)
			g += float64(col.G)
			b += float64(col.B)
			a += float64(col.A)
		}

		samples := size + 1
		x := 0
		for {
			factor := 1.0 / float64(samples)
			col := color.RGBA{
				R: uint8(math.Round(r * factor)),
				G: uint8(math.Round(g * factor)),
				B: uint8(math.Round(b * factor)),
				A: uint8(math.Round(a * factor)),
			}
			result.SetRGBA(x, y, col)

			if x >= w-1 {
				break
			}

			if left := x - size; left >= 0 {
				col = img.RGBAAt(left, y)
				r -= float64(col.R)
				g -= float64(col.G)
				b -= float64(col.B)
				a -= float64(col.A)
				samples--
			}

			x++

			if right := x + size; right < w {
				col = img.RGBAAt(right, y)
				r += float64(col.R)
				g += float64(col.G)
				b += float64(col.B)
				a += float64(col.A)
				samples++
			}
		}
	}

	return result
}

func box3y(img *image.RGBA, size int) *image.RGBA {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	w, h := bounds.Dx(), bounds.Dy()

	for x := 0; x < w; x++ {
		if size >= h {
			var r, g, b, a float64
			for y := 0; y < h; y++ {
				col := img.RGBAAt(x, y)
				r += float64(col.R)
				g += float64(col.G)
				b += float64(col.B)
				a += float64(col.A)
			}

			factor := 1.0 / float64(h)
			col := color.RGBA{
				R: uint8(math.Round(r * factor)),
				G: uint8(math.Round(g * factor)),
				B: uint8(math.Round(b * factor)),
				A: uint8(math.Round(a * factor)),
			}
			for y := 0; y < h; y++ {
				result.SetRGBA(x, y, col)
			}
			continue
		}

		var r, g, b, a float64
		for y := 0; y <= size; y++ {
			col := img.RGBAAt(x, y)
			r += float64(col.R)
			g += float64(col.G)
			b += float64(col.B)
			a += float64(col.A)
		}

		samples := size + 1
		y := 0
		for {
			factor := 1.0 / float64(samples)
			col := color.RGBA{
				R: uint8(math.Round(r * factor)),
				G: uint8(math.Round(g * factor)),
				B: uint8(math.Round(b * factor)),
				A: uint8(math.Round(a * factor)),
			}
			result.SetRGBA(x, y, col)

			if y >= h-1 {
				break
			}

			if top := y - size; top >= 0 {
				col = img.RGBAAt(x, top)
				r -= float64(col.R)
				g -= float64(col.G)
				b -= float64(col.B)
				a -= float64(col.A)
				samples--
			}

			y++

			if bottom := y + size; bottom < h {
				col = img.RGBAAt(x, bottom)
				r += float64(col.R)
				g += float64(col.G)
				b += float64(col.B)
				a += float64(col.A)
				samples++
			}
		}
	}

	return result
}
