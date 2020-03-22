package softwarebackend

import (
	"image/color"
	"math"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

func triangleLR(tri []backendbase.Vec, y float64) (l, r float64, outside bool) {
	a, b, c := tri[0], tri[1], tri[2]

	// sort by y
	if a[1] > b[1] {
		a, b = b, a
	}
	if b[1] > c[1] {
		b, c = c, b
		if a[1] > b[1] {
			a, b = b, a
		}
	}

	// check general bounds
	if y <= a[1] {
		return a[0], a[0], true
	}
	if y > c[1] {
		return c[0], c[0], true
	}

	// find left and right x at y
	if y >= a[1] && y <= b[1] && a[1] < b[1] {
		r0 := (y - a[1]) / (b[1] - a[1])
		l = (b[0]-a[0])*r0 + a[0]
		r1 := (y - a[1]) / (c[1] - a[1])
		r = (c[0]-a[0])*r1 + a[0]
	} else {
		r0 := (y - b[1]) / (c[1] - b[1])
		l = (c[0]-b[0])*r0 + b[0]
		r1 := (y - a[1]) / (c[1] - a[1])
		r = (c[0]-a[0])*r1 + a[0]
	}
	if l > r {
		l, r = r, l
	}

	return
}

func (b *SoftwareBackend) fillTriangleNoAA(tri []backendbase.Vec, fn func(x, y int)) {
	minY := int(math.Floor(math.Min(math.Min(tri[0][1], tri[1][1]), tri[2][1])))
	maxY := int(math.Ceil(math.Max(math.Max(tri[0][1], tri[1][1]), tri[2][1])))
	if minY < 0 {
		minY = 0
	} else if minY >= b.h {
		return
	}
	if maxY < 0 {
		return
	} else if maxY >= b.h {
		maxY = b.h - 1
	}
	for y := minY; y <= maxY; y++ {
		l, r, out := triangleLR(tri, float64(y)+0.5)
		if out {
			continue
		}
		if l < 0 {
			l = 0
		} else if l > float64(b.w) {
			continue
		}
		if r < 0 {
			continue
		} else if r > float64(b.w) {
			r = float64(b.w)
		}
		if l >= r {
			continue
		}
		fl, cr := int(math.Floor(l)), int(math.Ceil(r))
		for x := fl; x <= cr; x++ {
			fx := float64(x) + 0.5
			if fx < l || fx >= r {
				continue
			}
			fn(x, y)
		}
	}
}

type msaaPixel struct {
	ix, iy int
	fx, fy float64
	tx, ty float64
}

func (b *SoftwareBackend) fillTriangleMSAA(tri []backendbase.Vec, msaaLevel int, msaaPixels []msaaPixel, fn func(x, y int)) []msaaPixel {
	msaaStep := 1.0 / float64(msaaLevel+1)

	minY := int(math.Floor(math.Min(math.Min(tri[0][1], tri[1][1]), tri[2][1])))
	maxY := int(math.Ceil(math.Max(math.Max(tri[0][1], tri[1][1]), tri[2][1])))
	if minY < 0 {
		minY = 0
	} else if minY >= b.h {
		return msaaPixels
	}
	if maxY < 0 {
		return msaaPixels
	} else if maxY >= b.h {
		maxY = b.h - 1
	}

	for y := minY; y <= maxY; y++ {
		var l, r [5]float64
		allOut := true
		minL, maxR := math.MaxFloat64, 0.0

		sy := float64(y) + msaaStep*0.5
		for step := 0; step <= msaaLevel; step++ {
			var out bool
			l[step], r[step], out = triangleLR(tri, sy)
			if l[step] < 0 {
				l[step] = 0
			} else if l[step] > float64(b.w) {
				l[step] = float64(b.w)
				out = true
			}
			if r[step] < 0 {
				r[step] = 0
				out = true
			} else if r[step] > float64(b.w) {
				r[step] = float64(b.w)
			}
			if r[step] <= l[step] {
				out = true
			}
			if !out {
				allOut = false
				minL = math.Min(minL, l[step])
				maxR = math.Max(maxR, r[step])
			}
			sy += msaaStep
		}

		if allOut {
			continue
		}

		fl, cr := int(math.Floor(minL)), int(math.Ceil(maxR))
		for x := fl; x <= cr; x++ {
			sy = float64(y) + msaaStep*0.5
			allIn := true
		check:
			for stepy := 0; stepy <= msaaLevel; stepy++ {
				sx := float64(x) + msaaStep*0.5
				for stepx := 0; stepx <= msaaLevel; stepx++ {
					if sx < l[stepy] || sx >= r[stepy] {
						allIn = false
						break check
					}
					sx += msaaStep
				}
				sy += msaaStep
			}

			if allIn {
				fn(x, y)
				continue
			}

			sy = float64(y) + msaaStep*0.5
			for stepy := 0; stepy <= msaaLevel; stepy++ {
				sx := float64(x) + msaaStep*0.5
				for stepx := 0; stepx <= msaaLevel; stepx++ {
					if sx >= l[stepy] && sx < r[stepy] {
						msaaPixels = addMSAAPixel(msaaPixels, msaaPixel{ix: x, iy: y, fx: sx, fy: sy})
					}
					sx += msaaStep
				}
				sy += msaaStep
			}
		}
	}

	return msaaPixels
}

func addMSAAPixel(msaaPixels []msaaPixel, px msaaPixel) []msaaPixel {
	for _, px2 := range msaaPixels {
		if px == px2 {
			return msaaPixels
		}
	}
	return append(msaaPixels, px)
}

func quadArea(quad [4]backendbase.Vec) float64 {
	leftv := backendbase.Vec{quad[1][0] - quad[0][0], quad[1][1] - quad[0][1]}
	topv := backendbase.Vec{quad[3][0] - quad[0][0], quad[3][1] - quad[0][1]}
	return math.Abs(leftv[0]*topv[1] - leftv[1]*topv[0])
}

func (b *SoftwareBackend) fillQuadNoAA(quad [4]backendbase.Vec, fn func(x, y int, tx, ty float64)) {
	minY := int(math.Floor(math.Min(math.Min(quad[0][1], quad[1][1]), math.Min(quad[2][1], quad[3][1]))))
	maxY := int(math.Ceil(math.Max(math.Max(quad[0][1], quad[1][1]), math.Max(quad[2][1], quad[3][1]))))
	if minY < 0 {
		minY = 0
	} else if minY >= b.h {
		return
	}
	if maxY < 0 {
		return
	} else if maxY >= b.h {
		maxY = b.h - 1
	}

	leftv := backendbase.Vec{quad[1][0] - quad[0][0], quad[1][1] - quad[0][1]}
	leftLen := math.Sqrt(leftv[0]*leftv[0] + leftv[1]*leftv[1])
	leftv[0] /= leftLen
	leftv[1] /= leftLen
	topv := backendbase.Vec{quad[3][0] - quad[0][0], quad[3][1] - quad[0][1]}
	topLen := math.Sqrt(topv[0]*topv[0] + topv[1]*topv[1])
	topv[0] /= topLen
	topv[1] /= topLen

	tri1 := [3]backendbase.Vec{quad[0], quad[1], quad[2]}
	tri2 := [3]backendbase.Vec{quad[0], quad[2], quad[3]}
	for y := minY; y <= maxY; y++ {
		lf1, rf1, out1 := triangleLR(tri1[:], float64(y)+0.5)
		lf2, rf2, out2 := triangleLR(tri2[:], float64(y)+0.5)
		if out1 && out2 {
			continue
		}
		l := math.Min(lf1, lf2)
		r := math.Max(rf1, rf2)
		if l < 0 {
			l = 0
		} else if l > float64(b.w) {
			continue
		}
		if r < 0 {
			continue
		} else if r > float64(b.w) {
			r = float64(b.w)
		}
		if l >= r {
			continue
		}

		tfy := float64(y) + 0.5 - quad[0][1]
		fl, cr := int(math.Floor(l)), int(math.Ceil(r))
		for x := fl; x <= cr; x++ {
			fx := float64(x) + 0.5
			if fx < l || fx >= r {
				continue
			}
			tfx := fx - quad[0][0]

			var tx, ty float64
			if math.Abs(leftv[0]) > math.Abs(leftv[1]) {
				tx = (tfy - tfx*(leftv[1]/leftv[0])) / (topv[1] - topv[0]*(leftv[1]/leftv[0]))
				ty = (tfx - topv[0]*tx) / leftv[0]
			} else {
				tx = (tfx - tfy*(leftv[0]/leftv[1])) / (topv[0] - topv[1]*(leftv[0]/leftv[1]))
				ty = (tfy - topv[1]*tx) / leftv[1]
			}

			fn(x, y, tx/topLen, ty/leftLen)
		}
	}
}

func (b *SoftwareBackend) fillQuadMSAA(quad [4]backendbase.Vec, msaaLevel int, msaaPixels []msaaPixel, fn func(x, y int, tx, ty float64)) []msaaPixel {
	msaaStep := 1.0 / float64(msaaLevel+1)

	minY := int(math.Floor(math.Min(math.Min(quad[0][1], quad[1][1]), math.Min(quad[2][1], quad[3][1]))))
	maxY := int(math.Ceil(math.Max(math.Max(quad[0][1], quad[1][1]), math.Max(quad[2][1], quad[3][1]))))
	if minY < 0 {
		minY = 0
	} else if minY >= b.h {
		return msaaPixels
	}
	if maxY < 0 {
		return msaaPixels
	} else if maxY >= b.h {
		maxY = b.h - 1
	}

	leftv := backendbase.Vec{quad[1][0] - quad[0][0], quad[1][1] - quad[0][1]}
	leftLen := math.Sqrt(leftv[0]*leftv[0] + leftv[1]*leftv[1])
	leftv[0] /= leftLen
	leftv[1] /= leftLen
	topv := backendbase.Vec{quad[3][0] - quad[0][0], quad[3][1] - quad[0][1]}
	topLen := math.Sqrt(topv[0]*topv[0] + topv[1]*topv[1])
	topv[0] /= topLen
	topv[1] /= topLen

	tri1 := [3]backendbase.Vec{quad[0], quad[1], quad[2]}
	tri2 := [3]backendbase.Vec{quad[0], quad[2], quad[3]}
	for y := minY; y <= maxY; y++ {
		var l, r [5]float64
		allOut := true
		minL, maxR := math.MaxFloat64, 0.0

		sy := float64(y) + msaaStep*0.5
		for step := 0; step <= msaaLevel; step++ {
			lf1, rf1, out1 := triangleLR(tri1[:], sy)
			lf2, rf2, out2 := triangleLR(tri2[:], sy)
			l[step] = math.Min(lf1, lf2)
			r[step] = math.Max(rf1, rf2)
			out := out1 || out2

			if l[step] < 0 {
				l[step] = 0
			} else if l[step] > float64(b.w) {
				l[step] = float64(b.w)
				out = true
			}
			if r[step] < 0 {
				r[step] = 0
				out = true
			} else if r[step] > float64(b.w) {
				r[step] = float64(b.w)
			}
			if r[step] <= l[step] {
				out = true
			}
			if !out {
				allOut = false
				minL = math.Min(minL, l[step])
				maxR = math.Max(maxR, r[step])
			}
			sy += msaaStep
		}

		if allOut {
			continue
		}

		fl, cr := int(math.Floor(minL)), int(math.Ceil(maxR))
		for x := fl; x <= cr; x++ {
			sy = float64(y) + msaaStep*0.5
			allIn := true
		check:
			for stepy := 0; stepy <= msaaLevel; stepy++ {
				sx := float64(x) + msaaStep*0.5
				for stepx := 0; stepx <= msaaLevel; stepx++ {
					if sx < l[stepy] || sx >= r[stepy] {
						allIn = false
						break check
					}
					sx += msaaStep
				}
				sy += msaaStep
			}

			if allIn {
				tfx := float64(x) + 0.5 - quad[0][0]
				tfy := float64(y) + 0.5 - quad[0][1]

				var tx, ty float64
				if math.Abs(leftv[0]) > math.Abs(leftv[1]) {
					tx = (tfy - tfx*(leftv[1]/leftv[0])) / (topv[1] - topv[0]*(leftv[1]/leftv[0]))
					ty = (tfx - topv[0]*tx) / leftv[0]
				} else {
					tx = (tfx - tfy*(leftv[0]/leftv[1])) / (topv[0] - topv[1]*(leftv[0]/leftv[1]))
					ty = (tfy - topv[1]*tx) / leftv[1]
				}

				fn(x, y, tx/topLen, ty/leftLen)
				continue
			}

			sy = float64(y) + msaaStep*0.5
			for stepy := 0; stepy <= msaaLevel; stepy++ {
				sx := float64(x) + msaaStep*0.5
				for stepx := 0; stepx <= msaaLevel; stepx++ {
					if sx >= l[stepy] && sx < r[stepy] {
						tfx := sx - quad[0][0]
						tfy := sy - quad[0][1]

						var tx, ty float64
						if math.Abs(leftv[0]) > math.Abs(leftv[1]) {
							tx = (tfy - tfx*(leftv[1]/leftv[0])) / (topv[1] - topv[0]*(leftv[1]/leftv[0]))
							ty = (tfx - topv[0]*tx) / leftv[0]
						} else {
							tx = (tfx - tfy*(leftv[0]/leftv[1])) / (topv[0] - topv[1]*(leftv[0]/leftv[1]))
							ty = (tfy - topv[1]*tx) / leftv[1]
						}

						msaaPixels = addMSAAPixel(msaaPixels, msaaPixel{ix: x, iy: y, fx: sx, fy: sy, tx: tx / topLen, ty: ty / leftLen})
					}
					sx += msaaStep
				}
				sy += msaaStep
			}
		}
	}

	return msaaPixels
}

func (b *SoftwareBackend) fillQuad(pts [4]backendbase.Vec, fn func(x, y, tx, ty float64) color.RGBA) {
	b.clearStencil()

	if b.MSAA > 0 {
		var msaaPixelBuf [500]msaaPixel
		msaaPixels := msaaPixelBuf[:0]

		msaaPixels = b.fillQuadMSAA(pts, b.MSAA, msaaPixels, func(x, y int, tx, ty float64) {
			if b.clip.AlphaAt(x, y).A == 0 {
				return
			}
			if b.stencil.AlphaAt(x, y).A > 0 {
				return
			}
			b.stencil.SetAlpha(x, y, color.Alpha{A: 255})
			col := fn(float64(x)+0.5, float64(y)+0.5, tx, ty)
			if col.A > 0 {
				b.Image.SetRGBA(x, y, mix(col, b.Image.RGBAAt(x, y)))
			}
		})

		samples := (b.MSAA + 1) * (b.MSAA + 1)

		for i, px := range msaaPixels {
			if px.ix < 0 || b.clip.AlphaAt(px.ix, px.iy).A == 0 || b.stencil.AlphaAt(px.ix, px.iy).A > 0 {
				continue
			}
			b.stencil.SetAlpha(px.ix, px.iy, color.Alpha{A: 255})

			var mr, mg, mb, ma int
			for j, px2 := range msaaPixels[i:] {
				if px2.ix != px.ix || px2.iy != px.iy {
					continue
				}

				col := fn(px2.fx, px2.fy, px2.tx, px2.ty)
				mr += int(col.R)
				mg += int(col.G)
				mb += int(col.B)
				ma += int(col.A)

				msaaPixels[i+j].ix = -1
			}

			combined := color.RGBA{
				R: uint8(mr / samples),
				G: uint8(mg / samples),
				B: uint8(mb / samples),
				A: uint8(ma / samples),
			}
			b.Image.SetRGBA(px.ix, px.iy, mix(combined, b.Image.RGBAAt(px.ix, px.iy)))
		}

	} else {
		b.fillQuadNoAA(pts, func(x, y int, tx, ty float64) {
			if b.clip.AlphaAt(x, y).A == 0 {
				return
			}
			if b.stencil.AlphaAt(x, y).A > 0 {
				return
			}
			b.stencil.SetAlpha(x, y, color.Alpha{A: 255})
			col := fn(float64(x)+0.5, float64(y)+0.5, tx, ty)
			if col.A > 0 {
				b.Image.SetRGBA(x, y, mix(col, b.Image.RGBAAt(x, y)))
			}
		})
	}
}

func iterateTriangles(pts []backendbase.Vec, fn func(tri []backendbase.Vec)) {
	if len(pts) == 4 {
		var buf [3]backendbase.Vec
		buf[0] = pts[0]
		buf[1] = pts[1]
		buf[2] = pts[2]
		fn(buf[:])
		buf[1] = pts[2]
		buf[2] = pts[3]
		fn(buf[:])
		return
	}
	for i := 3; i <= len(pts); i += 3 {
		fn(pts[i-3 : i])
	}
}

func (b *SoftwareBackend) fillTrianglesNoAA(pts []backendbase.Vec, fn func(x, y float64) color.RGBA) {
	iterateTriangles(pts[:], func(tri []backendbase.Vec) {
		b.fillTriangleNoAA(tri, func(x, y int) {
			if b.clip.AlphaAt(x, y).A == 0 {
				return
			}
			if b.stencil.AlphaAt(x, y).A > 0 {
				return
			}
			b.stencil.SetAlpha(x, y, color.Alpha{A: 255})
			col := fn(float64(x), float64(y))
			if col.A > 0 {
				b.Image.SetRGBA(x, y, mix(col, b.Image.RGBAAt(x, y)))
			}
		})
	})
}

func (b *SoftwareBackend) fillTrianglesMSAA(pts []backendbase.Vec, msaaLevel int, fn func(x, y float64) color.RGBA) {
	var msaaPixelBuf [500]msaaPixel
	msaaPixels := msaaPixelBuf[:0]

	iterateTriangles(pts[:], func(tri []backendbase.Vec) {
		msaaPixels = b.fillTriangleMSAA(tri, msaaLevel, msaaPixels, func(x, y int) {
			if b.clip.AlphaAt(x, y).A == 0 {
				return
			}
			if b.stencil.AlphaAt(x, y).A > 0 {
				return
			}
			b.stencil.SetAlpha(x, y, color.Alpha{A: 255})
			col := fn(float64(x), float64(y))
			if col.A > 0 {
				b.Image.SetRGBA(x, y, mix(col, b.Image.RGBAAt(x, y)))
			}
		})
	})

	samples := (msaaLevel + 1) * (msaaLevel + 1)

	for i, px := range msaaPixels {
		if px.ix < 0 || b.clip.AlphaAt(px.ix, px.iy).A == 0 || b.stencil.AlphaAt(px.ix, px.iy).A > 0 {
			continue
		}
		b.stencil.SetAlpha(px.ix, px.iy, color.Alpha{A: 255})

		var mr, mg, mb, ma int
		for j, px2 := range msaaPixels[i:] {
			if px2.ix != px.ix || px2.iy != px.iy {
				continue
			}

			col := fn(px2.fx, px2.fy)
			mr += int(col.R)
			mg += int(col.G)
			mb += int(col.B)
			ma += int(col.A)

			msaaPixels[i+j].ix = -1
		}

		combined := color.RGBA{
			R: uint8(mr / samples),
			G: uint8(mg / samples),
			B: uint8(mb / samples),
			A: uint8(ma / samples),
		}
		b.Image.SetRGBA(px.ix, px.iy, mix(combined, b.Image.RGBAAt(px.ix, px.iy)))
	}
}

func (b *SoftwareBackend) fillTriangles(pts []backendbase.Vec, fn func(x, y float64) color.RGBA) {
	b.clearStencil()

	if b.MSAA > 0 {
		b.fillTrianglesMSAA(pts, b.MSAA, fn)
	} else {
		b.fillTrianglesNoAA(pts, fn)
	}
}
