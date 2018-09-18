package canvas

import (
	"math"
	"runtime"
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
	stops    []gradientStop
	tex      uint32
	loaded   bool
	deleted  bool
	opaque   bool
}

type gradientStop struct {
	pos   float64
	color glColor
}

// NewLinearGradient creates a new linear gradient with
// the coordinates from where to where the gradient
// will apply on the canvas
func NewLinearGradient(x0, y0, x1, y1 float64) *LinearGradient {
	if gli == nil {
		panic("LoadGL must be called before gradients can be created")
	}
	lg := &LinearGradient{gradient: gradient{from: vec{x0, y0}, to: vec{x1, y1}, opaque: true}}
	gli.GenTextures(1, &lg.tex)
	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, lg.tex)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_T, gl_CLAMP_TO_EDGE)
	runtime.SetFinalizer(lg, func(lg *LinearGradient) {
		glChan <- func() {
			gli.DeleteTextures(1, &lg.tex)
		}
	})
	return lg
}

// NewRadialGradient creates a new linear gradient with
// the coordinates and the radii for two circles. The
// gradient will apply from the first to the second
// circle
func NewRadialGradient(x0, y0, r0, x1, y1, r1 float64) *RadialGradient {
	if gli == nil {
		panic("LoadGL must be called before gradients can be created")
	}
	rg := &RadialGradient{gradient: gradient{from: vec{x0, y0}, to: vec{x1, y1}, opaque: true}, radFrom: r0, radTo: r1}
	gli.GenTextures(1, &rg.tex)
	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, rg.tex)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_LINEAR)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_T, gl_CLAMP_TO_EDGE)
	runtime.SetFinalizer(rg, func(rg *RadialGradient) {
		glChan <- func() {
			gli.DeleteTextures(1, &rg.tex)
		}
	})
	return rg
}

// Delete explicitly deletes the gradient
func (g *gradient) Delete() {
	gli.DeleteTextures(1, &g.tex)
	g.deleted = true
}

func (g *gradient) load() {
	if g.loaded {
		return
	}

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, g.tex)
	var pixels [2048 * 4]byte
	pp := 0
	for i := 0; i < 2048; i++ {
		c := g.colorAt(float64(i) / 2047)
		pixels[pp] = byte(math.Floor(c.r*255 + 0.5))
		pixels[pp+1] = byte(math.Floor(c.g*255 + 0.5))
		pixels[pp+2] = byte(math.Floor(c.b*255 + 0.5))
		pixels[pp+3] = byte(math.Floor(c.a*255 + 0.5))
		pp += 4
	}
	gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RGBA, 2048, 1, 0, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&pixels[0]))
	g.loaded = true
}

func (g *gradient) colorAt(pos float64) glColor {
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

// AddColorStop adds a color stop to the gradient. The stops
// don't have to be added in order, they are sorted into the
// right place
func (g *gradient) AddColorStop(pos float64, color ...interface{}) {
	c, _ := parseColor(color...)
	if c.a < 1 {
		g.opaque = false
	}
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
