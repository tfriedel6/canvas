package canvas

import (
	"image/color"
	"runtime"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

// LinearGradient is a gradient with any number of
// stops and any number of colors. The gradient will
// be drawn such that each point on the gradient
// will correspond to a straight line
type LinearGradient struct {
	cv       *Canvas
	from, to backendbase.Vec
	created  bool
	loaded   bool
	opaque   bool
	grad     backendbase.LinearGradient
	data     backendbase.Gradient
}

// RadialGradient is a gradient with any number of
// stops and any number of colors. The gradient will
// be drawn such that each point on the gradient
// will correspond to a circle
type RadialGradient struct {
	cv       *Canvas
	from, to backendbase.Vec
	radFrom  float64
	radTo    float64
	created  bool
	loaded   bool
	opaque   bool
	grad     backendbase.RadialGradient
	data     backendbase.Gradient
}

// CreateLinearGradient creates a new linear gradient with
// the coordinates from where to where the gradient
// will apply on the canvas
func (cv *Canvas) CreateLinearGradient(x0, y0, x1, y1 float64) *LinearGradient {
	lg := &LinearGradient{
		cv:     cv,
		opaque: true,
		from:   backendbase.Vec{x0, y0},
		to:     backendbase.Vec{x1, y1},
		data:   make(backendbase.Gradient, 0, 20),
	}
	runtime.SetFinalizer(lg, func(*LinearGradient) {
		lg.grad.Delete()
	})
	return lg
}

// CreateRadialGradient creates a new radial gradient with
// the coordinates and the radii for two circles. The
// gradient will apply from the first to the second
// circle
func (cv *Canvas) CreateRadialGradient(x0, y0, r0, x1, y1, r1 float64) *RadialGradient {
	rg := &RadialGradient{
		cv:      cv,
		opaque:  true,
		from:    backendbase.Vec{x0, y0},
		to:      backendbase.Vec{x1, y1},
		radFrom: r0,
		radTo:   r1,
		data:    make(backendbase.Gradient, 0, 20),
	}
	runtime.SetFinalizer(rg, func(*RadialGradient) {
		rg.grad.Delete()
	})
	return rg
}

func (lg *LinearGradient) load() {
	if lg.loaded || len(lg.data) < 1 {
		return
	}

	if !lg.created {
		lg.grad = lg.cv.b.LoadLinearGradient(lg.data)
	} else {
		lg.grad.Replace(lg.data)
	}
	lg.created = true
	lg.loaded = true
}

func (rg *RadialGradient) load() {
	if rg.loaded || len(rg.data) < 1 {
		return
	}

	if !rg.created {
		rg.grad = rg.cv.b.LoadRadialGradient(rg.data)
	} else {
		rg.grad.Replace(rg.data)
	}
	rg.created = true
	rg.loaded = true
}

// AddColorStop adds a color stop to the gradient. The stops
// don't have to be added in order, they are sorted into the
// right place
func (lg *LinearGradient) AddColorStop(pos float64, stopColor ...interface{}) {
	var c color.RGBA
	lg.data, c = addColorStop(lg.data, pos, stopColor...)
	if c.A < 255 {
		lg.opaque = false
	}
	lg.loaded = false
}

// AddColorStop adds a color stop to the gradient. The stops
// don't have to be added in order, they are sorted into the
// right place
func (rg *RadialGradient) AddColorStop(pos float64, stopColor ...interface{}) {
	var c color.RGBA
	rg.data, c = addColorStop(rg.data, pos, stopColor...)
	if c.A < 255 {
		rg.opaque = false
	}
	rg.loaded = false
}

func addColorStop(stops backendbase.Gradient, pos float64, stopColor ...interface{}) (backendbase.Gradient, color.RGBA) {
	c, _ := parseColor(stopColor...)
	insert := len(stops)
	for i, stop := range stops {
		if stop.Pos > pos {
			insert = i
			break
		}
	}
	stops = append(stops, backendbase.GradientStop{})
	if insert < len(stops)-1 {
		copy(stops[insert+1:], stops[insert:len(stops)-1])
	}
	stops[insert] = backendbase.GradientStop{Pos: pos, Color: c}
	return stops, c
}
