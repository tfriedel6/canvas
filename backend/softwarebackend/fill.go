package softwarebackend

import (
	"image"
	"image/color"
	"math"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

func (b *SoftwareBackend) Clear(pts [4]backendbase.Vec) {
	iterateTriangles(pts[:], func(tri []backendbase.Vec) {
		b.fillTriangleNoAA(tri, func(x, y int) {
			if b.clip.AlphaAt(x, y).A == 0 {
				return
			}
			b.Image.SetRGBA(x, y, color.RGBA{})
		})
	})
}

func (b *SoftwareBackend) Fill(style *backendbase.FillStyle, pts []backendbase.Vec, tf backendbase.Mat, canOverlap bool) {
	ffn := fillFunc(style)

	var triBuf [500]backendbase.Vec
	if tf != backendbase.MatIdentity {
		ptsOld := pts
		if len(pts) < len(triBuf) {
			pts = triBuf[:len(pts)]
		} else {
			pts = make([]backendbase.Vec, len(pts))
		}
		for i, pt := range ptsOld {
			pts[i] = pt.MulMat(tf)
		}
	}

	if style.Blur > 0 {
		b.activateBlurTarget()
		b.fillTriangles(pts, ffn)
		b.drawBlurred(style.Blur)
	} else {
		b.fillTriangles(pts, ffn)
	}
}

func (b *SoftwareBackend) FillImageMask(style *backendbase.FillStyle, mask *image.Alpha, pts [4]backendbase.Vec) {
	ffn := fillFunc(style)

	mw := float64(mask.Bounds().Dx())
	mh := float64(mask.Bounds().Dy())
	b.fillQuad(pts, func(x, y, sx2, sy2 float64) color.RGBA {
		sxi := int(mw * sx2)
		syi := int(mh * sy2)
		a := mask.AlphaAt(sxi, syi)
		if a.A == 0 {
			return color.RGBA{}
		}
		col := ffn(x, y)
		return alphaColor(col, a)
	})
}

func fillFunc(style *backendbase.FillStyle) func(x, y float64) color.RGBA {
	if lg := style.LinearGradient; lg != nil {
		lg := lg.(*LinearGradient)
		from := backendbase.Vec{style.Gradient.X0, style.Gradient.Y0}
		dir := backendbase.Vec{style.Gradient.X1 - style.Gradient.X0, style.Gradient.Y1 - style.Gradient.Y0}
		dirlen := math.Sqrt(dir[0]*dir[0] + dir[1]*dir[1])
		dir[0] /= dirlen
		dir[1] /= dirlen
		return func(x, y float64) color.RGBA {
			pos := backendbase.Vec{x - from[0], y - from[1]}
			r := (pos[0]*dir[0] + pos[1]*dir[1]) / dirlen
			return lg.data.ColorAt(r)
		}
	} else if rg := style.RadialGradient; rg != nil {
		rg := rg.(*RadialGradient)
		from := backendbase.Vec{style.Gradient.X0, style.Gradient.Y0}
		to := backendbase.Vec{style.Gradient.X1, style.Gradient.Y1}
		radFrom := style.Gradient.RadFrom
		radTo := style.Gradient.RadTo
		return func(x, y float64) color.RGBA {
			pos := backendbase.Vec{x, y}
			oa := 0.5 * math.Sqrt(
				math.Pow(-2.0*from[0]*from[0]+2.0*from[0]*to[0]+2.0*from[0]*pos[0]-2.0*to[0]*pos[0]-2.0*from[1]*from[1]+2.0*from[1]*to[1]+2.0*from[1]*pos[1]-2.0*to[1]*pos[1]+2.0*radFrom*radFrom-2.0*radFrom*radTo, 2.0)-
					4.0*(from[0]*from[0]-2.0*from[0]*pos[0]+pos[0]*pos[0]+from[1]*from[1]-2.0*from[1]*pos[1]+pos[1]*pos[1]-radFrom*radFrom)*
						(from[0]*from[0]-2.0*from[0]*to[0]+to[0]*to[0]+from[1]*from[1]-2.0*from[1]*to[1]+to[1]*to[1]-radFrom*radFrom+2.0*radFrom*radTo-radTo*radTo))
			ob := (from[0]*from[0] - from[0]*to[0] - from[0]*pos[0] + to[0]*pos[0] + from[1]*from[1] - from[1]*to[1] - from[1]*pos[1] + to[1]*pos[1] - radFrom*radFrom + radFrom*radTo)
			oc := (from[0]*from[0] - 2.0*from[0]*to[0] + to[0]*to[0] + from[1]*from[1] - 2.0*from[1]*to[1] + to[1]*to[1] - radFrom*radFrom + 2.0*radFrom*radTo - radTo*radTo)
			o1 := (-oa + ob) / oc
			o2 := (oa + ob) / oc
			if math.IsNaN(o1) && math.IsNaN(o2) {
				return color.RGBA{}
			}
			o := math.Max(o1, o2)
			return rg.data.ColorAt(o)
		}
	} else if ip := style.ImagePattern; ip != nil {
		ip := ip.(*ImagePattern)
		img := ip.data.Image.(*Image)
		mip := img.mips[0] // todo select the right mip size
		w, h := img.Size()
		fw, fh := float64(w), float64(h)
		rx := ip.data.Repeat == backendbase.Repeat || ip.data.Repeat == backendbase.RepeatX
		ry := ip.data.Repeat == backendbase.Repeat || ip.data.Repeat == backendbase.RepeatY
		return func(x, y float64) color.RGBA {
			pos := backendbase.Vec{x, y}
			tfptx := pos[0]*ip.data.Transform[0] + pos[1]*ip.data.Transform[1] + ip.data.Transform[2]
			tfpty := pos[0]*ip.data.Transform[3] + pos[1]*ip.data.Transform[4] + ip.data.Transform[5]

			if !rx && (tfptx < 0 || tfptx >= fw) {
				return color.RGBA{}
			}
			if !ry && (tfpty < 0 || tfpty >= fh) {
				return color.RGBA{}
			}

			mx := int(math.Floor(tfptx)) % w
			if mx < 0 {
				mx += w
			}
			my := int(math.Floor(tfpty)) % h
			if my < 0 {
				my += h
			}

			return toRGBA(mip.At(mx, my))
		}
	}
	return func(x, y float64) color.RGBA {
		return style.Color
	}
}

func (b *SoftwareBackend) clearStencil() {
	p := b.stencil.Pix
	for i := range p {
		p[i] = 0
	}
}

func (b *SoftwareBackend) ClearClip() {
	p := b.clip.Pix
	for i := range p {
		p[i] = 255
	}
}

func (b *SoftwareBackend) Clip(pts []backendbase.Vec) {
	b.clearStencil()

	iterateTriangles(pts[:], func(tri []backendbase.Vec) {
		b.fillTriangleNoAA(tri, func(x, y int) {
			b.stencil.SetAlpha(x, y, color.Alpha{A: 255})
		})
	})

	p := b.clip.Pix
	p2 := b.stencil.Pix
	for i := range p {
		if p2[i] == 0 {
			p[i] = 0
		}
	}
}
