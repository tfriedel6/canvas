package goglbackend

import (
	"github.com/tfriedel6/canvas/backend/backendbase"
	"github.com/tfriedel6/canvas/backend/goglbackend/gl"
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
}

type gradient struct {
	b   *GoGLBackend
	tex uint32
}

func (b *GoGLBackend) LoadLinearGradient(data backendbase.Gradient) backendbase.LinearGradient {
	b.activate()

	lg := &LinearGradient{
		gradient: gradient{b: b},
	}
	gl.GenTextures(1, &lg.tex)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, lg.tex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	lg.load(data)
	return lg
}

func (b *GoGLBackend) LoadRadialGradient(data backendbase.Gradient) backendbase.RadialGradient {
	b.activate()

	rg := &RadialGradient{
		gradient: gradient{b: b},
	}
	gl.GenTextures(1, &rg.tex)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, rg.tex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	rg.load(data)
	return rg
}

// Delete explicitly deletes the gradient
func (g *gradient) Delete() {
	g.b.activate()

	gl.DeleteTextures(1, &g.tex)
}

func (lg *LinearGradient) Replace(data backendbase.Gradient) { lg.load(data) }
func (rg *RadialGradient) Replace(data backendbase.Gradient) { rg.load(data) }

func (g *gradient) load(stops backendbase.Gradient) {
	g.b.activate()

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
}
