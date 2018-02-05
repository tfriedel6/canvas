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
	move   bool
	next   lm.Vec2
	attach bool
}

type drawState struct {
	transform lm.Mat3x3
	fill      struct {
		r, g, b, a float32
	}
	stroke struct {
		r, g, b, a float32
		lineWidth  float32
	}
	font     *Font
	fontSize float32
	lineJoin lineJoin
	lineEnd  lineEnd

	lineDash       []float32
	lineDashPoint  int
	lineDashOffset float32
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
	cv.state.stroke.lineWidth = 1
	cv.state.transform = lm.Mat3x3Identity()
	return cv
}

func (cv *Canvas) ptToGL(x, y float32) (fx, fy float32) {
	return cv.vecToGL(lm.Vec2{x, y})
}

func (cv *Canvas) vecToGL(v lm.Vec2) (fx, fy float32) {
	v, _ = v.MulMat3x3(cv.state.transform)
	return v[0]*2/cv.fw - 1, -v[1]*2/cv.fh + 1
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
}

var (
	gli GL
	buf uint32
	sr  *solidShader
	tr  *textureShader
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
void main() {
    gl_Position = vec4(vertex, 0.0, 1.0);
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
varying vec2 v_texCoord;
void main() {
	v_texCoord = texCoord;
    gl_Position = vec4(vertex, 0.0, 1.0);
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

func glError() error {
	glErr := gli.GetError()
	if glErr != gl_NO_ERROR {
		return fmt.Errorf("GL Error: %x", glErr)
	}
	return nil
}

// SetFillColor sets the color for any fill calls
func (cv *Canvas) SetFillColor(value ...interface{}) {
	r, g, b, a, ok := parseColor(value...)
	if ok {
		f := &cv.state.fill
		f.r, f.g, f.b, f.a = r, g, b, a
	}
}

// SetStrokeColor sets the color for any line drawing calls
func (cv *Canvas) SetStrokeColor(value ...interface{}) {
	r, g, b, a, ok := parseColor(value...)
	if ok {
		s := &cv.state.stroke
		s.r, s.g, s.b, s.a = r, g, b, a
	}
}

// SetLineWidth sets the line width for any line drawing calls
func (cv *Canvas) SetLineWidth(width float32) {
	cv.state.stroke.lineWidth = width
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
	cv.state = cv.stateStack[l-1]
	cv.stateStack = cv.stateStack[:l-1]
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

// FillRect fills a rectangle with the active color
func (cv *Canvas) FillRect(x, y, w, h float32) {
	cv.activate()

	gli.UseProgram(sr.id)

	x0f, y0f := cv.ptToGL(x, y)
	x1f, y1f := cv.ptToGL(x, y+h)
	x2f, y2f := cv.ptToGL(x+w, y+h)
	x3f, y3f := cv.ptToGL(x+w, y)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [8]float32{x0f, y0f, x1f, y1f, x2f, y2f, x3f, y3f}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)
	f := cv.state.fill
	gli.Uniform4f(sr.color, f.r, f.g, f.b, f.a)
	gli.EnableVertexAttribArray(sr.vertex)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(sr.vertex)
}
