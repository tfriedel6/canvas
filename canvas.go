package canvas

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/tfriedel6/lm"
)

// Canvas represents an area on the viewport on which to draw
// using a set of functions very similar to the HTML5 canvas
type Canvas struct {
	x, y, w, h     int
	fx, fy, fw, fh float32

	polyPath []pathPoint
	linePath []pathPoint
	text     struct {
		target *image.RGBA
		tex    uint32
	}

	state      drawState
	stateStack []drawState
}

type pathPoint struct {
	pos    lm.Vec2
	tf     lm.Vec2
	move   bool
	next   lm.Vec2
	attach bool
}

type drawState struct {
	transform lm.Mat3x3
	fill      drawStyle
	stroke    drawStyle
	font      *Font
	fontSize  float32
	lineWidth float32
	lineJoin  lineJoin
	lineEnd   lineEnd

	lineDash       []float32
	lineDashPoint  int
	lineDashOffset float32

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
		fx: float32(x), fy: float32(y),
		fw: float32(w), fh: float32(h),
		stateStack: make([]drawState, 0, 20),
	}
	cv.state.lineWidth = 1
	cv.state.transform = lm.Mat3x3Identity()
	return cv
}

func (cv *Canvas) tf(v lm.Vec2) lm.Vec2 {
	v, _ = v.MulMat3x3(cv.state.transform)
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

var (
	gli    GL
	buf    uint32
	sr     *solidShader
	tr     *textureShader
	lgr    *linearGradientShader
	rgr    *radialGradientShader
	ipr    *imagePatternShader
	glChan = make(chan func())
)

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

	tr, err = loadTextureShader()
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

	gli.GenBuffers(1, &buf)
	err = glError()
	if err != nil {
		return
	}

	gli.Enable(gl_BLEND)
	gli.BlendFunc(gl_SRC_ALPHA, gl_ONE_MINUS_SRC_ALPHA)
	gli.Enable(gl_STENCIL_TEST)
	gli.StencilFunc(gl_EQUAL, 1, 0x00)

	return
}

//go:generate go run make_shaders.go
//go:generate go fmt

var solidVS = `
attribute vec2 vertex;
uniform vec2 canvasSize;
void main() {
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var solidFS = `
#ifdef GL_ES
precision mediump float;
#endif
uniform vec4 color;
void main() {
    gl_FragColor = color;
}`

var textureVS = `
attribute vec2 vertex, texCoord;
uniform vec2 canvasSize;
varying vec2 v_texCoord;
void main() {
	v_texCoord = texCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var textureFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_texCoord;
uniform sampler2D image;
void main() {
    gl_FragColor = texture2D(image, v_texCoord);
}`
var linearGradientVS = `
attribute vec2 vertex;
uniform vec2 canvasSize;
varying vec2 v_cp;
void main() {
	v_cp = vertex;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var linearGradientFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_cp;
uniform sampler1D gradient;
uniform vec2 from, dir;
uniform float len;
void main() {
	vec2 v = v_cp - from;
	float r = dot(v, dir) / len;
	r = clamp(r, 0.0, 1.0);
    gl_FragColor = texture1D(gradient, r);
}`
var radialGradientVS = `
attribute vec2 vertex;
uniform vec2 canvasSize;
varying vec2 v_cp;
void main() {
	v_cp = vertex;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var radialGradientFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_cp;
uniform sampler1D gradient;
uniform vec2 from, to, dir;
uniform float radFrom, radTo;
uniform float len;
bool isNaN(float v) {
  return v < 0.0 || 0.0 < v || v == 0.0 ? false : true;
}
void main() {
	float o_a = 0.5 * sqrt(
		pow(-2.0*from.x*from.x+2.0*from.x*to.x+2.0*from.x*v_cp.x-2.0*to.x*v_cp.x-2.0*from.y*from.y+2.0*from.y*to.y+2.0*from.y*v_cp.y-2.0*to.y*v_cp.y+2.0*radFrom*radFrom-2.0*radFrom*radTo, 2.0)
		-4.0*(from.x*from.x-2.0*from.x*v_cp.x+v_cp.x*v_cp.x+from.y*from.y-2.0*from.y*v_cp.y+v_cp.y*v_cp.y-radFrom*radFrom)
		*(from.x*from.x-2.0*from.x*to.x+to.x*to.x+from.y*from.y-2.0*from.y*to.y+to.y*to.y-radFrom*radFrom+2.0*radFrom*radTo-radTo*radTo)
	);
	float o_b = (from.x*from.x-from.x*to.x-from.x*v_cp.x+to.x*v_cp.x+from.y*from.y-from.y*to.y-from.y*v_cp.y+to.y*v_cp.y-radFrom*radFrom+radFrom*radTo);
	float o_c = (from.x*from.x-2.0*from.x*to.x+to.x*to.x+from.y*from.y-2.0*from.y*to.y+to.y*to.y-radFrom*radFrom+2.0*radFrom*radTo-radTo*radTo);
	float o1 = (-o_a + o_b) / o_c;
	float o2 = (o_a + o_b) / o_c;
	if (isNaN(o1) && isNaN(o2)) {
		gl_FragColor = vec4(0.0, 0.0, 0.0, 0.0);
		return;
	}
	float o = max(o1, o2);
	float r = radFrom + o * (radTo - radFrom);
	gl_FragColor = texture1D(gradient, o);
}`
var imagePatternVS = `
attribute vec2 vertex;
uniform vec2 canvasSize;
varying vec2 v_cp;
void main() {
	v_cp = vertex;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var imagePatternFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_cp;
uniform vec2 imageSize;
uniform sampler2D image;
void main() {
    gl_FragColor = texture2D(image, mod(v_cp / imageSize, 1.0));
    //gl_FragColor = vec4(v_cp * 0.1, 0.0, 1.0);
}`

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
		dir := to.Sub(from)
		length := dir.Len()
		dir = dir.DivF(length)
		gli.Uniform2f(lgr.canvasSize, cv.fw, cv.fh)
		gli.Uniform2f(lgr.from, from[0], from[1])
		gli.Uniform2f(lgr.dir, dir[0], dir[1])
		gli.Uniform1f(lgr.len, length)
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
		dir := to.Sub(from)
		length := dir.Len()
		dir = dir.DivF(length)
		gli.Uniform2f(rgr.canvasSize, cv.fw, cv.fh)
		gli.Uniform2f(rgr.from, from[0], from[1])
		gli.Uniform2f(rgr.to, to[0], to[1])
		gli.Uniform2f(rgr.dir, dir[0], dir[1])
		gli.Uniform1f(rgr.radFrom, rg.radFrom)
		gli.Uniform1f(rgr.radTo, rg.radTo)
		gli.Uniform1f(rgr.len, length)
		gli.Uniform1i(rgr.gradient, 0)
		return rgr.vertex
	}
	if img := style.image; img != nil {
		gli.UseProgram(ipr.id)
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_2D, img.tex)
		gli.Uniform2f(ipr.canvasSize, cv.fw, cv.fh)
		gli.Uniform2f(ipr.imageSize, float32(img.w), float32(img.h))
		gli.Uniform1i(ipr.image, 0)
		return ipr.vertex
	}

	gli.UseProgram(sr.id)
	gli.Uniform2f(sr.canvasSize, cv.fw, cv.fh)
	c := cv.state.fill.color
	gli.Uniform4f(sr.color, c.r, c.g, c.b, c.a)
	return sr.vertex
}

// SetLineWidth sets the line width for any line drawing calls
func (cv *Canvas) SetLineWidth(width float32) {
	cv.state.lineWidth = width
}

// SetFont sets the font and font size
func (cv *Canvas) SetFont(font *Font, size float32) {
	cv.state.font = font
	cv.state.fontSize = size
}

// SetLineJoin sets the style of line joints for rendering a path with Stroke
func (cv *Canvas) SetLineJoin(join lineJoin) {
	cv.state.lineJoin = join
}

// SetLineEnd sets the style of line endings for rendering a path with Stroke
func (cv *Canvas) SetLineEnd(end lineEnd) {
	cv.state.lineEnd = end
}

// SetLineDash sets the line dash style
func (cv *Canvas) SetLineDash(dash []float32) {
	l := len(dash)
	if l%2 == 0 {
		d2 := make([]float32, l)
		copy(d2, dash)
		cv.state.lineDash = d2
	} else {
		d2 := make([]float32, l*2)
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

func (cv *Canvas) Scale(x, y float32) {
	cv.state.transform = cv.state.transform.Mul(lm.Mat3x3Scale(lm.Vec2{x, y}))
}

func (cv *Canvas) Translate(x, y float32) {
	cv.state.transform = cv.state.transform.Mul(lm.Mat3x3Translate(lm.Vec2{x, y}))
}

func (cv *Canvas) Rotate(angle float32) {
	cv.state.transform = cv.state.transform.Mul(lm.Mat3x3Rotate(angle))
}

func (cv *Canvas) Transform(a, b, c, d, e, f float32) {
	cv.state.transform = cv.state.transform.Mul(lm.Mat3x3{a, b, 0, c, d, 0, e, f, 1})
}

func (cv *Canvas) SetTransform(a, b, c, d, e, f float32) {
	cv.state.transform = lm.Mat3x3{a, b, 0, c, d, 0, e, f, 1}
}

// FillRect fills a rectangle with the active fill style
func (cv *Canvas) FillRect(x, y, w, h float32) {
	cv.activate()

	p0 := cv.tf(lm.Vec2{x, y})
	p1 := cv.tf(lm.Vec2{x, y + h})
	p2 := cv.tf(lm.Vec2{x + w, y + h})
	p3 := cv.tf(lm.Vec2{x + w, y})

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [8]float32{p0[0], p0[1], p1[0], p1[1], p2[0], p2[1], p3[0], p3[1]}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	vertex := cv.useShader(&cv.state.fill)
	gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, nil)
	gli.EnableVertexAttribArray(vertex)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(vertex)
}
