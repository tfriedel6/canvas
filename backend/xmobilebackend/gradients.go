package xmobilebackend

import (
	"github.com/tfriedel6/canvas/backend/backendbase"
	"golang.org/x/mobile/gl"
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
	b   *XMobileBackend
	tex gl.Texture
}

func (b *XMobileBackend) LoadLinearGradient(data backendbase.Gradient) backendbase.LinearGradient {
	b.activate()

	lg := &LinearGradient{
		gradient: gradient{b: b},
	}
	lg.tex = b.glctx.CreateTexture()
	b.glctx.ActiveTexture(gl.TEXTURE0)
	b.glctx.BindTexture(gl.TEXTURE_2D, lg.tex)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	lg.load(data)
	return lg
}

func (b *XMobileBackend) LoadRadialGradient(data backendbase.Gradient) backendbase.RadialGradient {
	b.activate()

	rg := &RadialGradient{
		gradient: gradient{b: b},
	}
	rg.tex = b.glctx.CreateTexture()
	b.glctx.ActiveTexture(gl.TEXTURE0)
	b.glctx.BindTexture(gl.TEXTURE_2D, rg.tex)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	rg.load(data)
	return rg
}

// Delete explicitly deletes the gradient
func (g *gradient) Delete() {
	b := g.b
	g.b.activate()

	b.glctx.DeleteTexture(g.tex)
}

func (lg *LinearGradient) Replace(data backendbase.Gradient) { lg.load(data) }
func (rg *RadialGradient) Replace(data backendbase.Gradient) { rg.load(data) }

func (g *gradient) load(stops backendbase.Gradient) {
	b := g.b
	g.b.activate()

	b.glctx.ActiveTexture(gl.TEXTURE0)
	b.glctx.BindTexture(gl.TEXTURE_2D, g.tex)
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

	b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 2048, 1, gl.RGBA, gl.UNSIGNED_BYTE, pixels[0:])
}
