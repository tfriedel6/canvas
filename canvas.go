package canvas

import (
	"fmt"
	"unsafe"

	"github.com/golang/freetype/truetype"
)

//go:generate go run make_shaders.go
//go:generate go fmt

// Canvas represents an area on the viewport on which to draw
// using a set of functions very similar to the HTML5 canvas
type Canvas struct {
	x, y, w, h     int
	fx, fy, fw, fh float64

	polyPath []pathPoint
	linePath []pathPoint

	state      drawState
	stateStack []drawState
}

type pathPoint struct {
	pos    vec
	tf     vec
	move   bool
	next   vec
	attach bool
}

type drawState struct {
	transform mat
	fill      drawStyle
	stroke    drawStyle
	font      *Font
	fontSize  float64
	lineWidth float64
	lineJoin  lineJoin
	lineEnd   lineEnd

	lineDash       []float64
	lineDashPoint  int
	lineDashOffset float64

	clip []pathPoint
	/*
		The current transformation matrix.
		The current clipping region.
		The current dash list.
		The current values of the following attributes: strokeStyle, fillStyle, globalAlpha,
			lineWidth, lineCap, lineJoin, miterLimit, lineDashOffset, shadowOffsetX,
			shadowOffsetY, shadowBlur, shadowColor, globalCompositeOperation, font,
			textAlign, textBaseline, direction, imageSmoothingEnabled
	*/
}

type drawStyle struct {
	color          glColor
	radialGradient *RadialGradient
	linearGradient *LinearGradient
	image          *Image
}

type lineJoin uint8
type lineEnd uint8

const (
	Miter = iota
	Bevel
	Round
	Square
	Butt
)

// New creates a new canvas with the given viewport coordinates.
// While all functions on the canvas use the top left point as
// the origin, since GL uses the bottom left coordinate, the
// coordinates given here also use the bottom left as origin
func New(x, y, w, h int) *Canvas {
	cv := &Canvas{
		x: x, y: y, w: w, h: h,
		fx: float64(x), fy: float64(y),
		fw: float64(w), fh: float64(h),
		stateStack: make([]drawState, 0, 20),
	}
	cv.state.lineWidth = 1
	cv.state.transform = matIdentity()
	return cv
}

// SetSize changes the internal size of the canvas. This would
// usually be called for example when the window is resized
func (cv *Canvas) SetSize(w, h int) {
	cv.w, cv.h = w, h
	cv.fw, cv.fh = float64(w), float64(h)
	activeCanvas = nil
}

// Width returns the internal width of the canvas
func (cv *Canvas) Width() int { return cv.w }

// Height returns the internal height of the canvas
func (cv *Canvas) Height() int { return cv.h }

// Size returns the internal width and height of the canvas
func (cv *Canvas) Size() (int, int) { return cv.w, cv.h }

func (cv *Canvas) tf(v vec) vec {
	v, _ = v.mulMat(cv.state.transform)
	return v
}

// Activate makes the canvas active and sets the viewport. Only needs
// to be called if any other GL code changes the viewport
func (cv *Canvas) Activate() {
	gli.Viewport(int32(cv.x), int32(cv.y), int32(cv.w), int32(cv.h))
}

var activeCanvas *Canvas

func (cv *Canvas) activate() {
	if activeCanvas != cv {
		activeCanvas = cv
		cv.Activate()
	}
loop:
	for {
		select {
		case f := <-glChan:
			f()
		default:
			break loop
		}
	}
}

const alphaTexSize = 2048

var (
	gli      GL
	buf      uint32
	alphaTex uint32
	sr       *solidShader
	lgr      *linearGradientShader
	rgr      *radialGradientShader
	ipr      *imagePatternShader
	sar      *solidAlphaShader
	rgar     *radialGradientAlphaShader
	lgar     *linearGradientAlphaShader
	ipar     *imagePatternAlphaShader
	ir       *imageShader
	glChan   = make(chan func())
)

// LoadGL needs to be called once per GL context to load the GL assets
// that canvas needs. The parameter is an implementation of the GL interface
// in this package that should make this package neutral to GL implementations
func LoadGL(glimpl GL) (err error) {
	gli = glimpl

	gli.GetError() // clear error state

	sr, err = loadSolidShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	lgr, err = loadLinearGradientShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	rgr, err = loadRadialGradientShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	ipr, err = loadImagePatternShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	sar, err = loadSolidAlphaShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	lgar, err = loadLinearGradientAlphaShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	rgar, err = loadRadialGradientAlphaShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	ipar, err = loadImagePatternAlphaShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	ir, err = loadImageShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	gli.GenBuffers(1, &buf)
	err = glError()
	if err != nil {
		return
	}

	gli.ActiveTexture(gl_TEXTURE0)
	gli.GenTextures(1, &alphaTex)
	gli.BindTexture(gl_TEXTURE_2D, alphaTex)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_NEAREST)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_NEAREST)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
	gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_T, gl_CLAMP_TO_EDGE)
	gli.TexImage2D(gl_TEXTURE_2D, 0, gl_ALPHA, alphaTexSize, alphaTexSize, 0, gl_ALPHA, gl_UNSIGNED_BYTE, nil)

	gli.Enable(gl_BLEND)
	gli.BlendFunc(gl_SRC_ALPHA, gl_ONE_MINUS_SRC_ALPHA)
	gli.Enable(gl_STENCIL_TEST)
	gli.StencilMask(0xFF)
	gli.Clear(gl_STENCIL_BUFFER_BIT)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_KEEP)
	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	return
}

func glError() error {
	glErr := gli.GetError()
	if glErr != gl_NO_ERROR {
		return fmt.Errorf("GL Error: %x", glErr)
	}
	return nil
}

// SetFillStyle sets the color, gradient, or image for any fill calls
func (cv *Canvas) SetFillStyle(value ...interface{}) {
	cv.state.fill = parseStyle(value...)
}

// SetStrokeStyle sets the color, gradient, or image for any line drawing calls
func (cv *Canvas) SetStrokeStyle(value ...interface{}) {
	cv.state.stroke = parseStyle(value...)
}

func parseStyle(value ...interface{}) drawStyle {
	var style drawStyle
	if len(value) == 1 {
		switch v := value[0].(type) {
		case *LinearGradient:
			style.linearGradient = v
			return style
		case *RadialGradient:
			style.radialGradient = v
			return style
		case *Image:
			style.image = v
			return style
		case string:
			if img, ok := images[v]; ok {
				style.image = img
			}
		}
	}
	c, ok := parseColor(value...)
	if ok {
		style.color = c
	}
	return style
}

func (cv *Canvas) useShader(style *drawStyle) (vertexLoc uint32) {
	if lg := style.linearGradient; lg != nil {
		lg.load()
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_1D, lg.tex)
		gli.UseProgram(lgr.id)
		from := cv.tf(lg.from)
		to := cv.tf(lg.to)
		dir := to.sub(from)
		length := dir.len()
		dir = dir.divf(length)
		gli.Uniform2f(lgr.canvasSize, float32(cv.fw), float32(cv.fh))
		inv := cv.state.transform.invert().f32()
		gli.UniformMatrix3fv(lgr.invmat, 1, false, &inv[0])
		gli.Uniform2f(lgr.from, float32(from[0]), float32(from[1]))
		gli.Uniform2f(lgr.dir, float32(dir[0]), float32(dir[1]))
		gli.Uniform1f(lgr.len, float32(length))
		gli.Uniform1i(lgr.gradient, 0)
		return lgr.vertex
	}
	if rg := style.radialGradient; rg != nil {
		rg.load()
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_1D, rg.tex)
		gli.UseProgram(rgr.id)
		from := cv.tf(rg.from)
		to := cv.tf(rg.to)
		dir := to.sub(from)
		length := dir.len()
		dir = dir.divf(length)
		gli.Uniform2f(rgr.canvasSize, float32(cv.fw), float32(cv.fh))
		inv := cv.state.transform.invert().f32()
		gli.UniformMatrix3fv(rgr.invmat, 1, false, &inv[0])
		gli.Uniform2f(rgr.from, float32(from[0]), float32(from[1]))
		gli.Uniform2f(rgr.to, float32(to[0]), float32(to[1]))
		gli.Uniform2f(rgr.dir, float32(dir[0]), float32(dir[1]))
		gli.Uniform1f(rgr.radFrom, float32(rg.radFrom))
		gli.Uniform1f(rgr.radTo, float32(rg.radTo))
		gli.Uniform1f(rgr.len, float32(length))
		gli.Uniform1i(rgr.gradient, 0)
		return rgr.vertex
	}
	if img := style.image; img != nil {
		gli.UseProgram(ipr.id)
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_2D, img.tex)
		gli.Uniform2f(ipr.canvasSize, float32(cv.fw), float32(cv.fh))
		inv := cv.state.transform.invert().f32()
		gli.UniformMatrix3fv(ipr.invmat, 1, false, &inv[0])
		gli.Uniform2f(ipr.imageSize, float32(img.w), float32(img.h))
		gli.Uniform1i(ipr.image, 0)
		return ipr.vertex
	}

	gli.UseProgram(sr.id)
	gli.Uniform2f(sr.canvasSize, float32(cv.fw), float32(cv.fh))
	c := style.color
	gli.Uniform4f(sr.color, float32(c.r), float32(c.g), float32(c.b), float32(c.a))
	return sr.vertex
}

func (cv *Canvas) useAlphaShader(style *drawStyle, alphaTexSlot int32) (vertexLoc, alphaTexCoordLoc uint32) {
	if lg := style.linearGradient; lg != nil {
		lg.load()
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_1D, lg.tex)
		gli.UseProgram(lgar.id)
		from := cv.tf(lg.from)
		to := cv.tf(lg.to)
		dir := to.sub(from)
		length := dir.len()
		dir = dir.divf(length)
		gli.Uniform2f(lgar.canvasSize, float32(cv.fw), float32(cv.fh))
		inv := cv.state.transform.invert().f32()
		gli.UniformMatrix3fv(lgar.invmat, 1, false, &inv[0])
		gli.Uniform2f(lgar.from, float32(from[0]), float32(from[1]))
		gli.Uniform2f(lgar.dir, float32(dir[0]), float32(dir[1]))
		gli.Uniform1f(lgar.len, float32(length))
		gli.Uniform1i(lgar.gradient, 0)
		gli.Uniform1i(lgar.alphaTex, alphaTexSlot)
		return lgar.vertex, lgar.alphaTexCoord
	}
	if rg := style.radialGradient; rg != nil {
		rg.load()
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_1D, rg.tex)
		gli.UseProgram(rgar.id)
		from := cv.tf(rg.from)
		to := cv.tf(rg.to)
		dir := to.sub(from)
		length := dir.len()
		dir = dir.divf(length)
		gli.Uniform2f(rgar.canvasSize, float32(cv.fw), float32(cv.fh))
		inv := cv.state.transform.invert().f32()
		gli.UniformMatrix3fv(rgar.invmat, 1, false, &inv[0])
		gli.Uniform2f(rgar.from, float32(from[0]), float32(from[1]))
		gli.Uniform2f(rgar.to, float32(to[0]), float32(to[1]))
		gli.Uniform2f(rgar.dir, float32(dir[0]), float32(dir[1]))
		gli.Uniform1f(rgar.radFrom, float32(rg.radFrom))
		gli.Uniform1f(rgar.radTo, float32(rg.radTo))
		gli.Uniform1f(rgar.len, float32(length))
		gli.Uniform1i(rgar.gradient, 0)
		gli.Uniform1i(rgar.alphaTex, alphaTexSlot)
		return rgar.vertex, rgar.alphaTexCoord
	}
	if img := style.image; img != nil {
		gli.UseProgram(ipar.id)
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_2D, img.tex)
		gli.Uniform2f(ipar.canvasSize, float32(cv.fw), float32(cv.fh))
		inv := cv.state.transform.invert().f32()
		gli.UniformMatrix3fv(ipar.invmat, 1, false, &inv[0])
		gli.Uniform2f(ipar.imageSize, float32(img.w), float32(img.h))
		gli.Uniform1i(ipar.image, 0)
		gli.Uniform1i(ipar.alphaTex, alphaTexSlot)
		return ipar.vertex, ipar.alphaTexCoord
	}

	gli.UseProgram(sar.id)
	gli.Uniform2f(sar.canvasSize, float32(cv.fw), float32(cv.fh))
	c := style.color
	gli.Uniform4f(sar.color, float32(c.r), float32(c.g), float32(c.b), float32(c.a))
	gli.Uniform1i(sar.alphaTex, alphaTexSlot)
	return sar.vertex, sar.alphaTexCoord
}

// SetLineWidth sets the line width for any line drawing calls
func (cv *Canvas) SetLineWidth(width float64) {
	cv.state.lineWidth = width
}

// SetFont sets the font and font size
func (cv *Canvas) SetFont(font interface{}, size float64) {
	switch v := font.(type) {
	case *Font:
		cv.state.font = v
	case *truetype.Font:
		cv.state.font = &Font{font: v}
	case string:
		if f, ok := fonts[v]; ok {
			cv.state.font = f
		} else {
			f, err := LoadFont(v, v)
			if err == nil {
				cv.state.font = f
			}
		}
	}
	cv.state.fontSize = size
}

// SetLineJoin sets the style of line joints for rendering a path with Stroke.
// The value can be Miter, Bevel, or Round
func (cv *Canvas) SetLineJoin(join lineJoin) {
	cv.state.lineJoin = join
}

// SetLineEnd sets the style of line endings for rendering a path with Stroke
// The value can be Butt, Square, or Round
func (cv *Canvas) SetLineEnd(end lineEnd) {
	cv.state.lineEnd = end
}

// SetLineDash sets the line dash style
func (cv *Canvas) SetLineDash(dash []float64) {
	l := len(dash)
	if l%2 == 0 {
		d2 := make([]float64, l)
		copy(d2, dash)
		cv.state.lineDash = d2
	} else {
		d2 := make([]float64, l*2)
		copy(d2[:l], dash)
		copy(d2[l:], dash)
		cv.state.lineDash = d2
	}
	cv.state.lineDashPoint = 0
	cv.state.lineDashOffset = 0
}

// Save saves the current draw state to a stack
func (cv *Canvas) Save() {
	cv.stateStack = append(cv.stateStack, cv.state)
}

// Restore restores the last draw state from the stack if available
func (cv *Canvas) Restore() {
	l := len(cv.stateStack)
	if l <= 0 {
		return
	}
	hadClip := len(cv.state.clip) > 0
	cv.state = cv.stateStack[l-1]
	cv.stateStack = cv.stateStack[:l-1]
	if len(cv.state.clip) > 0 {
		cv.clip(cv.state.clip)
	} else if hadClip {
		gli.StencilMask(0x02)
		gli.Clear(gl_STENCIL_BUFFER_BIT)
		gli.StencilMask(0xFF)
	}
}

// Scale updates the current transformation with a scaling by the given values
func (cv *Canvas) Scale(x, y float64) {
	cv.state.transform = matScale(vec{x, y}).mul(cv.state.transform)
}

// Translate updates the current transformation with a translation by the given values
func (cv *Canvas) Translate(x, y float64) {
	cv.state.transform = matTranslate(vec{x, y}).mul(cv.state.transform)
}

// Rotate updates the current transformation with a rotation by the given angle
func (cv *Canvas) Rotate(angle float64) {
	cv.state.transform = matRotate(angle).mul(cv.state.transform)
}

// Transform updates the current transformation with the given matrix
func (cv *Canvas) Transform(a, b, c, d, e, f float64) {
	cv.state.transform = mat{a, b, 0, c, d, 0, e, f, 1}.mul(cv.state.transform)
}

// SetTransform replaces the current transformation with the given matrix
func (cv *Canvas) SetTransform(a, b, c, d, e, f float64) {
	cv.state.transform = mat{a, b, 0, c, d, 0, e, f, 1}
}

// FillRect fills a rectangle with the active fill style
func (cv *Canvas) FillRect(x, y, w, h float64) {
	cv.activate()

	p0 := cv.tf(vec{x, y})
	p1 := cv.tf(vec{x, y + h})
	p2 := cv.tf(vec{x + w, y + h})
	p3 := cv.tf(vec{x + w, y})

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [8]float32{float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), float32(p3[0]), float32(p3[1])}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	vertex := cv.useShader(&cv.state.fill)
	gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, nil)
	gli.EnableVertexAttribArray(vertex)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(vertex)
}
