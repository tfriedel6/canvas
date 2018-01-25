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

	fill struct {
		r, g, b, a float32
	}
	stroke struct {
		r, g, b, a float32
		lineWidth  float32
	}
	path []pathPoint
	text struct {
		font   *Font
		size   float32
		target *image.RGBA
		tex    uint32
	}
}

type pathPoint struct {
	pos  lm.Vec2
	move bool
}

// New creates a new canvas with the given viewport coordinates.
// While all functions on the canvas use the top left point as
// the origin, since GL uses the bottom left coordinate, the
// coordinates given here also use the bottom left as origin
func New(x, y, w, h int) *Canvas {
	cv := &Canvas{
		x: x, y: y, w: w, h: h,
		fx: float32(x), fy: float32(y),
		fw: float32(w), fh: float32(h),
	}
	cv.stroke.lineWidth = 1
	return cv
}

func (cv *Canvas) xToGL(x float32) float32              { return x*2/cv.fw - 1 }
func (cv *Canvas) yToGL(y float32) float32              { return -y*2/cv.fh + 1 }
func (cv *Canvas) ptToGL(x, y float32) (fx, fy float32) { return x*2/cv.fw - 1, -y*2/cv.fh + 1 }
func (cv *Canvas) vecToGL(v lm.Vec2) (fx, fy float32)   { return v[0]*2/cv.fw - 1, -v[1]*2/cv.fh + 1 }

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
		cv.fill.r, cv.fill.g, cv.fill.b, cv.fill.a = r, g, b, a
	}
}

// SetStrokeColor sets the color for any line drawing calls
func (cv *Canvas) SetStrokeColor(value ...interface{}) {
	r, g, b, a, ok := parseColor(value...)
	if ok {
		cv.stroke.r, cv.stroke.g, cv.stroke.b, cv.stroke.a = r, g, b, a
	}
}

// SetLineWidth sets the line width for any line drawing calls
func (cv *Canvas) SetLineWidth(width float32) {
	cv.stroke.lineWidth = width
}

// SetFont sets the font and font size
func (cv *Canvas) SetFont(font *Font, size float32) {
	cv.text.font = font
	cv.text.size = size
}

// FillRect fills a rectangle with the active color
func (cv *Canvas) FillRect(x, y, w, h float32) {
	cv.activate()

	gli.UseProgram(sr.id)

	x0f, y0f := cv.ptToGL(x, y)
	x1f, y1f := cv.ptToGL(x+w, y+h)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [8]float32{x0f, y0f, x0f, y1f, x1f, y1f, x1f, y0f}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.Uniform4f(sr.color, cv.fill.r, cv.fill.g, cv.fill.b, cv.fill.a)
	gli.EnableVertexAttribArray(sr.vertex)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(sr.vertex)
}
