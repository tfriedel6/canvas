package goglbackend

import (
	"runtime"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/tfriedel6/canvas/backend/backendbase"
)

// LinearGradient is a gradient with any number of
// stops and any number of colors. The gradient will
// be drawn such that each point on the gradient
// will correspond to a straight line
type LinearGradient struct {
	gradient
}

// RadialGradient is a gradient with any number of
// stops and any number of colors. The gradient will
// be drawn such that each point on the gradient
// will correspond to a circle
type RadialGradient struct {
	gradient
	radFrom, radTo float64
}

type gradient struct {
	from, to vec
	tex      uint32
	loaded   bool
	deleted  bool
	opaque   bool
}

func (b *GoGLBackend) LoadLinearGradient(data *backendbase.LinearGradientData) backendbase.LinearGradient {
	lg := &LinearGradient{
		gradient: gradient{from: vec{data.X0, data.Y0}, to: vec{data.X1, data.Y1}, opaque: true},
	}
	gl.GenTextures(1, &lg.tex)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, lg.tex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	lg.load(data.Stops)
	runtime.SetFinalizer(lg, func(lg *LinearGradient) {
		b.glChan <- func() {
			gl.DeleteTextures(1, &lg.tex)
		}
	})
	return lg
}

func (b *GoGLBackend) LoadRadialGradient(data *backendbase.RadialGradientData) backendbase.RadialGradient {
	rg := &RadialGradient{
		gradient: gradient{from: vec{data.X0, data.Y0}, to: vec{data.X1, data.Y1}, opaque: true},
		radFrom:  data.RadFrom,
		radTo:    data.RadTo,
	}
	gl.GenTextures(1, &rg.tex)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, rg.tex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	rg.load(data.Stops)
	runtime.SetFinalizer(rg, func(rg *RadialGradient) {
		b.glChan <- func() {
			gl.DeleteTextures(1, &rg.tex)
		}
	})
	return rg
}

// Delete explicitly deletes the gradient
func (g *gradient) Delete() {
	gl.DeleteTextures(1, &g.tex)
	g.deleted = true
}

func (g *gradient) IsDeleted() bool { return g.deleted }
func (g *gradient) IsOpaque() bool  { return g.opaque }

func (lg *LinearGradient) Replace(data *backendbase.LinearGradientData) { lg.load(data.Stops) }
func (rg *RadialGradient) Replace(data *backendbase.RadialGradientData) { rg.load(data.Stops) }

func (g *gradient) load(stops backendbase.Gradient) {
	if g.loaded {
		return
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, g.tex)
	var pixels [2048 * 4]byte
	pp := 0
	for i := 0; i < 2048; i++ {
		c := stops.ColorAt(float64(i) / 2047)
		pixels[pp] = c.R
		pixels[pp+1] = c.G
		pixels[pp+2] = c.B
		pixels[pp+3] = c.A
		pp += 4
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 2048, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(&pixels[0]))
	g.loaded = true
}
