package canvas

import (
	"runtime"

	"github.com/barnex/fmath"
	"github.com/tfriedel6/lm"
)

type LinearGradient struct {
	gradient
}

type RadialGradient struct {
	gradient
	radFrom, radTo float32
}

type gradient struct {
	from, to lm.Vec2
	stops    []gradientStop
	tex      uint32
	loaded   bool
	deleted  bool
}

type gradientStop struct {
	pos   float32
	color glColor
}

func NewLinearGradient(x0, y0, x1, y1 float32) *LinearGradient {
	lg := &LinearGradient{gradient: gradient{from: lm.Vec2{x0, y0}, to: lm.Vec2{x1, y1}}}
	gli.GenTextures(1, &lg.tex)
	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_1D, lg.tex)
	gli.TexParameteri(gl_TEXTURE_1D, gl_TEXTURE_MIN_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_1D, gl_TEXTURE_MAG_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_1D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
	runtime.SetFinalizer(lg, func(lg *LinearGradient) {
		glChan <- func() {
			gli.DeleteTextures(1, &lg.tex)
		}
	})
	return lg
}

func NewRadialGradient(x0, y0, r0, x1, y1, r1 float32) *RadialGradient {
	rg := &RadialGradient{gradient: gradient{from: lm.Vec2{x0, y0}, to: lm.Vec2{x1, y1}}, radFrom: r0, radTo: r1}
	gli.GenTextures(1, &rg.tex)
	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_1D, rg.tex)
	gli.TexParameteri(gl_TEXTURE_1D, gl_TEXTURE_MIN_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_1D, gl_TEXTURE_MAG_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_1D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
	runtime.SetFinalizer(rg, func(rg *RadialGradient) {
		glChan <- func() {
			gli.DeleteTextures(1, &rg.tex)
		}
	})
	return rg
}

func (g *gradient) Delete() {
	gli.DeleteTextures(1, &g.tex)
}

func (g *gradient) load() {
	if g.loaded {
		return
	}

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_1D, g.tex)
	var pixels [2048 * 4]byte
	pp := 0
	for i := 0; i < 2048; i++ {
		c := g.colorAt(float32(i) / 2047)
		pixels[pp] = byte(fmath.Floor(c.r*255 + 0.5))
		pixels[pp+1] = byte(fmath.Floor(c.g*255 + 0.5))
		pixels[pp+2] = byte(fmath.Floor(c.b*255 + 0.5))
		pixels[pp+3] = byte(fmath.Floor(c.a*255 + 0.5))
		pp += 4
	}
	gli.TexImage1D(gl_TEXTURE_1D, 0, gl_RGBA, 2048, 0, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&pixels[0]))
	g.loaded = true
}

func (g *gradient) colorAt(pos float32) glColor {
	if len(g.stops) == 0 {
		return glColor{}
	} else if len(g.stops) == 1 {
		return g.stops[0].color
	}
	beforeIdx, afterIdx := -1, -1
	for i, stop := range g.stops {
		if stop.pos > pos {
			afterIdx = i
			break
		}
		beforeIdx = i
	}
	if beforeIdx == -1 {
		return g.stops[0].color
	} else if afterIdx == -1 {
		return g.stops[len(g.stops)-1].color
	}
	before, after := g.stops[beforeIdx], g.stops[afterIdx]
	p := (pos - before.pos) / (after.pos - before.pos)
	var c glColor
	c.r = (after.color.r-before.color.r)*p + before.color.r
	c.g = (after.color.g-before.color.g)*p + before.color.g
	c.b = (after.color.b-before.color.b)*p + before.color.b
	c.a = (after.color.a-before.color.a)*p + before.color.a
	return c
}

func (g *gradient) AddColorStop(pos float32, color ...interface{}) {
	c, _ := parseColor(color...)
	insert := len(g.stops)
	for i, stop := range g.stops {
		if stop.pos > pos {
			insert = i
			break
		}
	}
	g.stops = append(g.stops, gradientStop{})
	if insert < len(g.stops)-1 {
		copy(g.stops[insert+1:], g.stops[insert:len(g.stops)-1])
	}
	g.stops[insert] = gradientStop{pos: pos, color: c}
	g.loaded = false
}
