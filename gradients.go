package canvas

import (
	"runtime"

	"github.com/barnex/fmath"
	"github.com/tfriedel6/lm"
)

type LinearGradient struct {
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
	lg := &LinearGradient{from: lm.Vec2{x0, y0}, to: lm.Vec2{x1, y1}}
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

func (lg *LinearGradient) Delete() {
	gli.DeleteTextures(1, &lg.tex)
}

func (lg *LinearGradient) load() {
	if lg.loaded {
		return
	}

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_1D, lg.tex)
	var pixels [2048 * 4]byte
	pp := 0
	for i := 0; i < 2048; i++ {
		c := lg.colorAt(float32(i) / 2047)
		pixels[pp] = byte(fmath.Floor(c.r*255 + 0.5))
		pixels[pp+1] = byte(fmath.Floor(c.g*255 + 0.5))
		pixels[pp+2] = byte(fmath.Floor(c.b*255 + 0.5))
		pixels[pp+3] = byte(fmath.Floor(c.a*255 + 0.5))
		pp += 4
	}
	gli.TexImage1D(gl_TEXTURE_1D, 0, gl_RGBA, 2048, 0, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&pixels[0]))
	lg.loaded = true
}

func (lg *LinearGradient) colorAt(pos float32) glColor {
	if len(lg.stops) == 0 {
		return glColor{}
	} else if len(lg.stops) == 1 {
		return lg.stops[0].color
	}
	beforeIdx, afterIdx := -1, -1
	for i, stop := range lg.stops {
		if stop.pos > pos {
			afterIdx = i
			break
		}
		beforeIdx = i
	}
	if beforeIdx == -1 {
		return lg.stops[0].color
	} else if afterIdx == -1 {
		return lg.stops[len(lg.stops)-1].color
	}
	before, after := lg.stops[beforeIdx], lg.stops[afterIdx]
	p := (pos - before.pos) / (after.pos - before.pos)
	var c glColor
	c.r = (after.color.r-before.color.r)*p + before.color.r
	c.g = (after.color.g-before.color.g)*p + before.color.g
	c.b = (after.color.b-before.color.b)*p + before.color.b
	c.a = (after.color.a-before.color.a)*p + before.color.a
	return c
}

func (lg *LinearGradient) AddColorStop(pos float32, color ...interface{}) {
	c, _ := parseColor(color...)
	insert := len(lg.stops)
	for i, stop := range lg.stops {
		if stop.pos > pos {
			insert = i
			break
		}
	}
	lg.stops = append(lg.stops, gradientStop{})
	if insert < len(lg.stops)-1 {
		copy(lg.stops[insert+1:], lg.stops[insert:len(lg.stops)-1])
	}
	lg.stops[insert] = gradientStop{pos: pos, color: c}
	lg.loaded = false
}
