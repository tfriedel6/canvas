// Package canvas provides an API that tries to closely mirror that
// of the HTML5 canvas API, using OpenGL to do the rendering.
package canvas

import (
	"fmt"
	"os"

	"github.com/golang/freetype/truetype"
)

//go:generate go run make_shaders.go
//go:generate go fmt

// Canvas represents an area on the viewport on which to draw
// using a set of functions very similar to the HTML5 canvas
type Canvas struct {
	x, y, w, h     int
	fx, fy, fw, fh float64

	path   []pathPoint
	convex bool
	rect   bool

	state      drawState
	stateStack []drawState
}

type drawState struct {
	transform     mat
	fill          drawStyle
	stroke        drawStyle
	font          *Font
	fontSize      float64
	textAlign     textAlign
	lineAlpha     float64
	lineWidth     float64
	lineJoin      lineJoin
	lineEnd       lineEnd
	miterLimitSqr float64
	globalAlpha   float64

	lineDash       []float64
	lineDashPoint  int
	lineDashOffset float64

	scissor scissor
	clip    []pathPoint

	shadowColor   glColor
	shadowOffsetX float64
	shadowOffsetY float64
	shadowBlur    float64

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

type scissor struct {
	on     bool
	tl, br vec
}

type lineJoin uint8
type lineEnd uint8

// Line join and end constants for SetLineJoin and SetLineEnd
const (
	Miter = iota
	Bevel
	Round
	Square
	Butt
)

type textAlign uint8

// Text alignment constants for SetTextAlign
const (
	Left = iota
	Center
	Right
	Start
	End
)

// New creates a new canvas with the given viewport coordinates.
// While all functions on the canvas use the top left point as
// the origin, since GL uses the bottom left coordinate, the
// coordinates given here also use the bottom left as origin
func New(x, y, w, h int) *Canvas {
	if gli == nil {
		panic("LoadGL must be called before a canvas can be created")
	}
	cv := &Canvas{stateStack: make([]drawState, 0, 20)}
	cv.SetBounds(x, y, w, h)
	cv.state.lineWidth = 1
	cv.state.lineAlpha = 1
	cv.state.miterLimitSqr = 100
	cv.state.globalAlpha = 1
	cv.state.fill.color = glColor{a: 1}
	cv.state.stroke.color = glColor{a: 1}
	cv.state.transform = matIdentity()
	return cv
}

// SetSize changes the internal size of the canvas. This would
// usually be called for example when the window is resized
//
// Deprecated: Use SetBounds instead
func (cv *Canvas) SetSize(w, h int) {
	cv.w, cv.h = w, h
	cv.fw, cv.fh = float64(w), float64(h)
	activeCanvas = nil
}

// SetBounds updates the bounds of the canvas. This would
// usually be called for example when the window is resized
func (cv *Canvas) SetBounds(x, y, w, h int) {
	cv.x, cv.y = x, y
	cv.w, cv.h = w, h
	cv.fx, cv.fy = float64(x), float64(y)
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
	cv.applyScissor()
	gli.Clear(gl_STENCIL_BUFFER_BIT)
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
	gli       GL
	buf       uint32
	shadowBuf uint32
	alphaTex  uint32
	sr        *solidShader
	lgr       *linearGradientShader
	rgr       *radialGradientShader
	ipr       *imagePatternShader
	sar       *solidAlphaShader
	rgar      *radialGradientAlphaShader
	lgar      *linearGradientAlphaShader
	ipar      *imagePatternAlphaShader
	ir        *imageShader
	gauss15r  *gaussianShader
	gauss63r  *gaussianShader
	gauss255r *gaussianShader
	offscr1   offscreenBuffer
	offscr2   offscreenBuffer
	glChan    = make(chan func())
)

type offscreenBuffer struct {
	tex              uint32
	w                int
	h                int
	renderStencilBuf uint32
	frameBuf         uint32
}

type gaussianShader struct {
	id          uint32
	vertex      uint32
	texCoord    uint32
	canvasSize  int32
	kernelScale int32
	image       int32
	kernel      int32
}

// LoadGL needs to be called once per GL context to load the GL assets
// that canvas needs. The parameter is an implementation of the GL interface
// in this package that should make this package neutral to GL implementations.
// The goglimpl subpackage contains an implementation based on Go-GL v3.2
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

	gauss15s, err := loadGaussian15Shader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}
	gauss15r = (*gaussianShader)(gauss15s)

	gauss63s, err := loadGaussian63Shader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}
	gauss63r = (*gaussianShader)(gauss63s)

	gauss255s, err := loadGaussian255Shader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}
	gauss255r = (*gaussianShader)(gauss255s)

	gli.GenBuffers(1, &buf)
	err = glError()
	if err != nil {
		return
	}

	gli.GenBuffers(1, &shadowBuf)
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

	gli.Enable(gl_SCISSOR_TEST)

	return
}

func glError() error {
	glErr := gli.GetError()
	if glErr != gl_NO_ERROR {
		return fmt.Errorf("GL Error: %x", glErr)
	}
	return nil
}

// SetFillStyle sets the color, gradient, or image for any fill calls. To set a
// color, there are several acceptable formats: 3 or 4 int values for RGB(A) in
// the range 0-255, 3 or 4 float values for RGB(A) in the range 0-1, hex strings
// in the format "#AABBCC", "#AABBCCDD", "#ABC", or "#ABCD"
func (cv *Canvas) SetFillStyle(value ...interface{}) {
	cv.state.fill = parseStyle(value...)
}

// SetStrokeStyle sets the color, gradient, or image for any line drawing calls.
// To set a color, there are several acceptable formats: 3 or 4 int values for
// RGB(A) in the range 0-255, 3 or 4 float values for RGB(A) in the range 0-1,
// hex strings in the format "#AABBCC", "#AABBCCDD", "#ABC", or "#ABCD"
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
		}
	}
	c, ok := parseColor(value...)
	if ok {
		style.color = c
	} else if len(value) == 1 {
		switch v := value[0].(type) {
		case *Image, string:
			style.image = getImage(v)
		}
	}
	return style
}

func (s *drawStyle) isOpaque() bool {
	if lg := s.linearGradient; lg != nil {
		return lg.opaque
	}
	if rg := s.radialGradient; rg != nil {
		return rg.opaque
	}
	if img := s.image; img != nil {
		return img.opaque
	}
	return s.color.a >= 1
}

func (cv *Canvas) useShader(style *drawStyle) (vertexLoc uint32) {
	if lg := style.linearGradient; lg != nil {
		lg.load()
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_2D, lg.tex)
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
		gli.Uniform1f(lgr.globalAlpha, float32(cv.state.globalAlpha))
		return lgr.vertex
	}
	if rg := style.radialGradient; rg != nil {
		rg.load()
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_2D, rg.tex)
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
		gli.Uniform1f(rgr.globalAlpha, float32(cv.state.globalAlpha))
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
		gli.Uniform1f(ipr.globalAlpha, float32(cv.state.globalAlpha))
		return ipr.vertex
	}

	gli.UseProgram(sr.id)
	gli.Uniform2f(sr.canvasSize, float32(cv.fw), float32(cv.fh))
	c := style.color
	gli.Uniform4f(sr.color, float32(c.r), float32(c.g), float32(c.b), float32(c.a))
	gli.Uniform1f(sr.globalAlpha, float32(cv.state.globalAlpha))
	return sr.vertex
}

func (cv *Canvas) useAlphaShader(style *drawStyle, alphaTexSlot int32) (vertexLoc, alphaTexCoordLoc uint32) {
	if lg := style.linearGradient; lg != nil {
		lg.load()
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_2D, lg.tex)
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
		gli.Uniform1f(lgar.globalAlpha, float32(cv.state.globalAlpha))
		return lgar.vertex, lgar.alphaTexCoord
	}
	if rg := style.radialGradient; rg != nil {
		rg.load()
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_2D, rg.tex)
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
		gli.Uniform1f(rgar.globalAlpha, float32(cv.state.globalAlpha))
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
		gli.Uniform1f(ipar.globalAlpha, float32(cv.state.globalAlpha))
		return ipar.vertex, ipar.alphaTexCoord
	}

	gli.UseProgram(sar.id)
	gli.Uniform2f(sar.canvasSize, float32(cv.fw), float32(cv.fh))
	c := style.color
	gli.Uniform4f(sar.color, float32(c.r), float32(c.g), float32(c.b), float32(c.a))
	gli.Uniform1i(sar.alphaTex, alphaTexSlot)
	gli.Uniform1f(sar.globalAlpha, float32(cv.state.globalAlpha))
	return sar.vertex, sar.alphaTexCoord
}

func (cv *Canvas) enableTextureRenderTarget(offscr *offscreenBuffer) {
	if offscr.w != cv.w || offscr.h != cv.h {
		if offscr.w != 0 && offscr.h != 0 {
			gli.DeleteTextures(1, &offscr.tex)
			gli.DeleteFramebuffers(1, &offscr.frameBuf)
			gli.DeleteRenderbuffers(1, &offscr.renderStencilBuf)
		}
		offscr.w = cv.w
		offscr.h = cv.h

		gli.ActiveTexture(gl_TEXTURE0)
		gli.GenTextures(1, &offscr.tex)
		gli.BindTexture(gl_TEXTURE_2D, offscr.tex)
		// todo do non-power-of-two textures work everywhere?
		gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RGBA, int32(cv.w), int32(cv.h), 0, gl_RGBA, gl_UNSIGNED_BYTE, nil)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_NEAREST)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_NEAREST)

		gli.GenFramebuffers(1, &offscr.frameBuf)
		gli.BindFramebuffer(gl_FRAMEBUFFER, offscr.frameBuf)

		gli.GenRenderbuffers(1, &offscr.renderStencilBuf)
		gli.BindRenderbuffer(gl_RENDERBUFFER, offscr.renderStencilBuf)
		gli.RenderbufferStorage(gl_RENDERBUFFER, gl_DEPTH24_STENCIL8, int32(cv.w), int32(cv.h))
		gli.FramebufferRenderbuffer(gl_FRAMEBUFFER, gl_DEPTH_STENCIL_ATTACHMENT, gl_RENDERBUFFER, offscr.renderStencilBuf)

		gli.FramebufferTexture(gl_FRAMEBUFFER, gl_COLOR_ATTACHMENT0, offscr.tex, 0)

		if err := gli.CheckFramebufferStatus(gl_FRAMEBUFFER); err != gl_FRAMEBUFFER_COMPLETE {
			// todo this should maybe not panic
			panic(fmt.Sprintf("Failed to set up framebuffer for offscreen texture: %x", err))
		}

		gli.Clear(gl_COLOR_BUFFER_BIT | gl_STENCIL_BUFFER_BIT)
	} else {
		gli.BindFramebuffer(gl_FRAMEBUFFER, offscr.frameBuf)
	}
}

func (cv *Canvas) disableTextureRenderTarget() {
	gli.BindFramebuffer(gl_FRAMEBUFFER, 0)
}

// SetLineWidth sets the line width for any line drawing calls
func (cv *Canvas) SetLineWidth(width float64) {
	if width < 0 {
		cv.state.lineWidth = 1
		cv.state.lineAlpha = 0
	} else if width < 1 {
		cv.state.lineWidth = 1
		cv.state.lineAlpha = width
	} else {
		cv.state.lineWidth = width
		cv.state.lineAlpha = 1
	}
}

// SetFont sets the font and font size. The font parameter can be a font loaded
// with the LoadFont function, a filename for a font to load (which will be
// cached), or nil, in which case the first loaded font will be used
func (cv *Canvas) SetFont(src interface{}, size float64) {
	if src == nil {
		cv.state.font = defaultFont
	} else {
		switch v := src.(type) {
		case *Font:
			cv.state.font = v
		case *truetype.Font:
			cv.state.font = &Font{font: v}
		case string:
			if f, ok := fonts[v]; ok {
				cv.state.font = f
			} else {
				f, err := LoadFont(v)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error loading font %s: %v\n", v, err)
					fonts[v] = nil
				} else {
					fonts[v] = f
					cv.state.font = f
				}
			}
		}
	}
	cv.state.fontSize = size
}

// SetTextAlign sets the text align for any text drawing calls.
// The value can be Left, Center, Right, Start, or End
func (cv *Canvas) SetTextAlign(align textAlign) {
	cv.state.textAlign = align
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

func (cv *Canvas) SetLineDashOffset(offset float64) {
	cv.state.lineDashOffset = offset
}

// GetLineDash gets the line dash style
func (cv *Canvas) GetLineDash() []float64 {
	result := make([]float64, len(cv.state.lineDash))
	copy(result, cv.state.lineDash)
	return result
}

// SetMiterLimit sets the limit for how far a miter line join can be extend.
// The fallback is a bevel join
func (cv *Canvas) SetMiterLimit(limit float64) {
	cv.state.miterLimitSqr = limit * limit
}

// SetGlobalAlpha sets the global alpha value
func (cv *Canvas) SetGlobalAlpha(alpha float64) {
	cv.state.globalAlpha = alpha
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
	gli.StencilMask(0x02)
	gli.Clear(gl_STENCIL_BUFFER_BIT)
	gli.StencilMask(0xFF)
	for _, st := range cv.stateStack {
		if len(st.clip) > 0 {
			cv.clip(st.clip)
		}
	}
	cv.state = cv.stateStack[l-1]
	cv.stateStack = cv.stateStack[:l-1]
	cv.applyScissor()
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

// SetShadowColor sets the color of the shadow. If it is fully transparent (default)
// then no shadow is drawn
func (cv *Canvas) SetShadowColor(color ...interface{}) {
	if c, ok := parseColor(color...); ok {
		cv.state.shadowColor = c
	}
}

// SetShadowOffsetX sets the x offset of the shadow
func (cv *Canvas) SetShadowOffsetX(offset float64) {
	cv.state.shadowOffsetX = offset
}

// SetShadowOffsetY sets the y offset of the shadow
func (cv *Canvas) SetShadowOffsetY(offset float64) {
	cv.state.shadowOffsetY = offset
}

// SetShadowOffset sets the offset of the shadow
func (cv *Canvas) SetShadowOffset(x, y float64) {
	cv.state.shadowOffsetX = x
	cv.state.shadowOffsetY = y
}

// SetShadowBlur sets the gaussian blur radius of the shadow
// (0 for no blur)
func (cv *Canvas) SetShadowBlur(r float64) {
	cv.state.shadowBlur = r
}
