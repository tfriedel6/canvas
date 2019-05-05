package xmobilebackend

import (
	"fmt"
	"image/color"
	"math"
	"reflect"
	"unsafe"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"golang.org/x/mobile/gl"
)

const alphaTexSize = 2048

var zeroes [alphaTexSize]byte

// GLContext is a context that contains all the
// shaders and buffers necessary for rendering
type GLContext struct {
	glctx gl.Context

	buf       gl.Buffer
	shadowBuf gl.Buffer
	alphaTex  gl.Texture

	sr        solidShader
	lgr       linearGradientShader
	rgr       radialGradientShader
	ipr       imagePatternShader
	sar       solidAlphaShader
	rgar      radialGradientAlphaShader
	lgar      linearGradientAlphaShader
	ipar      imagePatternAlphaShader
	ir        imageShader
	gauss15r  gaussianShader
	gauss63r  gaussianShader
	gauss127r gaussianShader

	offscr1 offscreenBuffer
	offscr2 offscreenBuffer

	imageBufTex gl.Texture
	imageBuf    []byte

	ptsBuf []float32

	glChan chan func()
}

// NewGLContext creates all the necessary GL resources,
// like shaders and buffers
func NewGLContext(glctx gl.Context) (*GLContext, error) {
	ctx := &GLContext{
		glctx: glctx,

		ptsBuf: make([]float32, 0, 4096),
		glChan: make(chan func()),
	}

	var err error

	b := &XMobileBackend{GLContext: ctx}

	b.glctx.GetError() // clear error state

	err = loadShader(b, solidVS, solidFS, &ctx.sr.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.sr.shaderProgram.mustLoadLocations(&ctx.sr)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, linearGradientVS, linearGradientFS, &ctx.lgr.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.lgr.shaderProgram.mustLoadLocations(&ctx.lgr)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, radialGradientVS, radialGradientFS, &ctx.rgr.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.rgr.shaderProgram.mustLoadLocations(&ctx.rgr)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, imagePatternVS, imagePatternFS, &ctx.ipr.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.ipr.shaderProgram.mustLoadLocations(&ctx.ipr)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, solidAlphaVS, solidAlphaFS, &ctx.sar.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.sar.shaderProgram.mustLoadLocations(&ctx.sar)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, linearGradientAlphaVS, linearGradientFS, &ctx.lgar.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.lgar.shaderProgram.mustLoadLocations(&ctx.lgar)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, radialGradientAlphaVS, radialGradientAlphaFS, &ctx.rgar.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.rgar.shaderProgram.mustLoadLocations(&ctx.rgar)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, imagePatternAlphaVS, imagePatternAlphaFS, &ctx.ipar.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.ipar.shaderProgram.mustLoadLocations(&ctx.ipar)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, imageVS, imageFS, &ctx.ir.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.ir.shaderProgram.mustLoadLocations(&ctx.ir)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, gaussian15VS, gaussian15FS, &ctx.gauss15r.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.gauss15r.shaderProgram.mustLoadLocations(&ctx.gauss15r)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, gaussian63VS, gaussian63FS, &ctx.gauss63r.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.gauss63r.shaderProgram.mustLoadLocations(&ctx.gauss63r)
	if err = glError(b); err != nil {
		return nil, err
	}

	err = loadShader(b, gaussian127VS, gaussian127FS, &ctx.gauss127r.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.gauss127r.shaderProgram.mustLoadLocations(&ctx.gauss127r)
	if err = glError(b); err != nil {
		return nil, err
	}

	ctx.buf = b.glctx.CreateBuffer()
	if err = glError(b); err != nil {
		return nil, err
	}

	ctx.shadowBuf = b.glctx.CreateBuffer()
	if err = glError(b); err != nil {
		return nil, err
	}

	b.glctx.ActiveTexture(gl.TEXTURE0)
	ctx.alphaTex = b.glctx.CreateTexture()
	b.glctx.BindTexture(gl.TEXTURE_2D, ctx.alphaTex)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.ALPHA, alphaTexSize, alphaTexSize, gl.ALPHA, gl.UNSIGNED_BYTE, nil)

	b.glctx.Enable(gl.BLEND)
	b.glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	b.glctx.Enable(gl.STENCIL_TEST)
	b.glctx.StencilMask(0xFF)
	b.glctx.Clear(gl.STENCIL_BUFFER_BIT)
	b.glctx.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
	b.glctx.StencilFunc(gl.EQUAL, 0, 0xFF)

	b.glctx.Disable(gl.SCISSOR_TEST)

	return ctx, nil
}

// XMobileBackend is a canvas backend using Go-GL
type XMobileBackend struct {
	x, y, w, h     int
	fx, fy, fw, fh float64

	*GLContext

	activateFn                 func()
	disableTextureRenderTarget func()
}

type offscreenBuffer struct {
	tex              gl.Texture
	w                int
	h                int
	renderStencilBuf gl.Renderbuffer
	frameBuf         gl.Framebuffer
	alpha            bool
}

// New returns a new canvas backend. x, y, w, h define the target
// rectangle in the window. ctx is a GLContext created with
// NewGLContext
func New(x, y, w, h int, ctx *GLContext) (*XMobileBackend, error) {
	b := &XMobileBackend{
		w:         w,
		h:         h,
		fw:        float64(w),
		fh:        float64(h),
		GLContext: ctx,
	}

	b.activateFn = func() {
		b.glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{Value: 0})
		b.glctx.Viewport(b.x, b.y, b.w, b.h)
		// todo reapply clipping since another application may have used the stencil buffer
	}
	b.disableTextureRenderTarget = func() {
		b.glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{Value: 0})
	}

	return b, nil
}

// XMobileBackendOffscreen is a canvas backend using an offscreen
// texture
type XMobileBackendOffscreen struct {
	XMobileBackend

	TextureID gl.Texture

	offscrBuf offscreenBuffer
	offscrImg Image
}

// NewOffscreen returns a new offscreen canvas backend. w, h define
// the size of the offscreen texture. ctx is a GLContext created
// with NewGLContext
func NewOffscreen(w, h int, alpha bool, ctx *GLContext) (*XMobileBackendOffscreen, error) {
	b, err := New(0, 0, w, h, ctx)
	if err != nil {
		return nil, err
	}
	bo := &XMobileBackendOffscreen{XMobileBackend: *b}
	bo.offscrBuf.alpha = alpha
	bo.offscrImg.flip = true

	bo.activateFn = func() {
		bo.enableTextureRenderTarget(&bo.offscrBuf)
		b.glctx.Viewport(0, 0, bo.w, bo.h)
		bo.offscrImg.w = bo.offscrBuf.w
		bo.offscrImg.h = bo.offscrBuf.h
		bo.offscrImg.tex = bo.offscrBuf.tex
		bo.TextureID = bo.offscrBuf.tex
	}
	bo.disableTextureRenderTarget = func() {
		bo.enableTextureRenderTarget(&bo.offscrBuf)
	}

	return bo, nil
}

// SetBounds updates the bounds of the canvas. This would
// usually be called for example when the window is resized
func (b *XMobileBackend) SetBounds(x, y, w, h int) {
	b.x, b.y = x, y
	b.fx, b.fy = float64(x), float64(y)
	b.w, b.h = w, h
	b.fw, b.fh = float64(w), float64(h)
	if b == activeContext {
		b.glctx.Viewport(0, 0, b.w, b.h)
		b.glctx.Clear(gl.STENCIL_BUFFER_BIT)
	}
}

// SetSize updates the size of the offscreen texture
func (b *XMobileBackendOffscreen) SetSize(w, h int) {
	b.XMobileBackend.SetBounds(0, 0, w, h)
	b.offscrImg.w = b.offscrBuf.w
	b.offscrImg.h = b.offscrBuf.h
}

// Size returns the size of the window or offscreen
// texture
func (b *XMobileBackend) Size() (int, int) {
	return b.w, b.h
}

func glError(b *XMobileBackend) error {
	glErr := b.glctx.GetError()
	if glErr != gl.NO_ERROR {
		return fmt.Errorf("GL Error: %x", glErr)
	}
	return nil
}

// Activate only needs to be called if there is other
// code also using the GL state
func (b *XMobileBackend) Activate() {
	b.activate()
}

var activeContext *XMobileBackend

func (b *XMobileBackend) activate() {
	if activeContext != b {
		activeContext = b
		b.activateFn()
	}
	b.runGLQueue()
}

func (b *XMobileBackend) runGLQueue() {
	for {
		select {
		case f := <-b.glChan:
			f()
		default:
			return
		}
	}
}

// Delete deletes the offscreen texture. After calling this
// the backend can no longer be used
func (b *XMobileBackendOffscreen) Delete() {
	b.glctx.DeleteTexture(b.offscrBuf.tex)
	b.glctx.DeleteFramebuffer(b.offscrBuf.frameBuf)
	b.glctx.DeleteRenderbuffer(b.offscrBuf.renderStencilBuf)
}

// CanUseAsImage returns true if the given backend can be
// directly used by this backend to avoid a conversion.
// Used internally
func (b *XMobileBackend) CanUseAsImage(b2 backendbase.Backend) bool {
	_, ok := b2.(*XMobileBackendOffscreen)
	return ok
}

// AsImage returns nil, since this backend cannot be directly
// used as an image. Used internally
func (b *XMobileBackend) AsImage() backendbase.Image {
	return nil
}

// AsImage returns an implementation of the Image interface
// that can be used to render this offscreen texture
// directly. Used internally
func (b *XMobileBackendOffscreen) AsImage() backendbase.Image {
	return &b.offscrImg
}

type glColor struct {
	r, g, b, a float64
}

func colorGoToGL(c color.RGBA) glColor {
	var glc glColor
	glc.r = float64(c.R) / 255
	glc.g = float64(c.G) / 255
	glc.b = float64(c.B) / 255
	glc.a = float64(c.A) / 255
	return glc
}

func (b *XMobileBackend) useShader(style *backendbase.FillStyle) (vertexLoc gl.Attrib) {
	if lg := style.LinearGradient; lg != nil {
		lg := lg.(*LinearGradient)
		b.glctx.ActiveTexture(gl.TEXTURE0)
		b.glctx.BindTexture(gl.TEXTURE_2D, lg.tex)
		b.glctx.UseProgram(b.lgr.ID)
		from := vec{style.Gradient.X0, style.Gradient.Y0}
		to := vec{style.Gradient.X1, style.Gradient.Y1}
		dir := to.sub(from)
		length := dir.len()
		dir = dir.scale(1 / length)
		b.glctx.Uniform2f(b.lgr.CanvasSize, float32(b.fw), float32(b.fh))
		b.glctx.Uniform2f(b.lgr.From, float32(from[0]), float32(from[1]))
		b.glctx.Uniform2f(b.lgr.Dir, float32(dir[0]), float32(dir[1]))
		b.glctx.Uniform1f(b.lgr.Len, float32(length))
		b.glctx.Uniform1i(b.lgr.Gradient, 0)
		b.glctx.Uniform1f(b.lgr.GlobalAlpha, float32(style.Color.A)/255)
		return b.lgr.Vertex
	}
	if rg := style.RadialGradient; rg != nil {
		rg := rg.(*RadialGradient)
		b.glctx.ActiveTexture(gl.TEXTURE0)
		b.glctx.BindTexture(gl.TEXTURE_2D, rg.tex)
		b.glctx.UseProgram(b.rgr.ID)
		from := vec{style.Gradient.X0, style.Gradient.Y0}
		to := vec{style.Gradient.X1, style.Gradient.Y1}
		b.glctx.Uniform2f(b.rgr.CanvasSize, float32(b.fw), float32(b.fh))
		b.glctx.Uniform2f(b.rgr.From, float32(from[0]), float32(from[1]))
		b.glctx.Uniform2f(b.rgr.To, float32(to[0]), float32(to[1]))
		b.glctx.Uniform1f(b.rgr.RadFrom, float32(style.Gradient.RadFrom))
		b.glctx.Uniform1f(b.rgr.RadTo, float32(style.Gradient.RadTo))
		b.glctx.Uniform1i(b.rgr.Gradient, 0)
		b.glctx.Uniform1f(b.rgr.GlobalAlpha, float32(style.Color.A)/255)
		return b.rgr.Vertex
	}
	if ip := style.ImagePattern; ip != nil {
		ipd := ip.(*ImagePattern).data
		img := ipd.Image.(*Image)
		b.glctx.UseProgram(b.ipr.ID)
		b.glctx.ActiveTexture(gl.TEXTURE0)
		b.glctx.BindTexture(gl.TEXTURE_2D, img.tex)
		b.glctx.Uniform2f(b.ipr.CanvasSize, float32(b.fw), float32(b.fh))
		b.glctx.Uniform2f(b.ipr.ImageSize, float32(img.w), float32(img.h))
		b.glctx.Uniform1i(b.ipr.Image, 0)
		var f32mat [9]float32
		for i, v := range ipd.Transform {
			f32mat[i] = float32(v)
		}
		b.glctx.UniformMatrix3fv(b.ipr.ImageTransform, f32mat[:])
		switch ipd.Repeat {
		case backendbase.Repeat:
			b.glctx.Uniform2f(b.ipr.Repeat, 1, 1)
		case backendbase.RepeatX:
			b.glctx.Uniform2f(b.ipr.Repeat, 1, 0)
		case backendbase.RepeatY:
			b.glctx.Uniform2f(b.ipr.Repeat, 0, 1)
		case backendbase.NoRepeat:
			b.glctx.Uniform2f(b.ipr.Repeat, 0, 0)
		}
		b.glctx.Uniform1f(b.ipr.GlobalAlpha, float32(style.Color.A)/255)
		return b.ipr.Vertex
	}

	b.glctx.UseProgram(b.sr.ID)
	b.glctx.Uniform2f(b.sr.CanvasSize, float32(b.fw), float32(b.fh))
	c := colorGoToGL(style.Color)
	b.glctx.Uniform4f(b.sr.Color, float32(c.r), float32(c.g), float32(c.b), float32(c.a))
	b.glctx.Uniform1f(b.sr.GlobalAlpha, 1)
	return b.sr.Vertex
}

func (b *XMobileBackend) useAlphaShader(style *backendbase.FillStyle, alphaTexSlot int) (vertexLoc, alphaTexCoordLoc gl.Attrib) {
	if lg := style.LinearGradient; lg != nil {
		lg := lg.(*LinearGradient)
		b.glctx.ActiveTexture(gl.TEXTURE0)
		b.glctx.BindTexture(gl.TEXTURE_2D, lg.tex)
		b.glctx.UseProgram(b.lgar.ID)
		from := vec{style.Gradient.X0, style.Gradient.Y0}
		to := vec{style.Gradient.X1, style.Gradient.Y1}
		dir := to.sub(from)
		length := dir.len()
		dir = dir.scale(1 / length)
		b.glctx.Uniform2f(b.lgar.CanvasSize, float32(b.fw), float32(b.fh))
		b.glctx.Uniform2f(b.lgar.From, float32(from[0]), float32(from[1]))
		b.glctx.Uniform2f(b.lgar.Dir, float32(dir[0]), float32(dir[1]))
		b.glctx.Uniform1f(b.lgar.Len, float32(length))
		b.glctx.Uniform1i(b.lgar.Gradient, 0)
		b.glctx.Uniform1i(b.lgar.AlphaTex, alphaTexSlot)
		b.glctx.Uniform1f(b.lgar.GlobalAlpha, float32(style.Color.A)/255)
		return b.lgar.Vertex, b.lgar.AlphaTexCoord
	}
	if rg := style.RadialGradient; rg != nil {
		rg := rg.(*RadialGradient)
		b.glctx.ActiveTexture(gl.TEXTURE0)
		b.glctx.BindTexture(gl.TEXTURE_2D, rg.tex)
		b.glctx.UseProgram(b.rgar.ID)
		from := vec{style.Gradient.X0, style.Gradient.Y0}
		to := vec{style.Gradient.X1, style.Gradient.Y1}
		b.glctx.Uniform2f(b.rgar.CanvasSize, float32(b.fw), float32(b.fh))
		b.glctx.Uniform2f(b.rgar.From, float32(from[0]), float32(from[1]))
		b.glctx.Uniform2f(b.rgar.To, float32(to[0]), float32(to[1]))
		b.glctx.Uniform1f(b.rgar.RadFrom, float32(style.Gradient.RadFrom))
		b.glctx.Uniform1f(b.rgar.RadTo, float32(style.Gradient.RadTo))
		b.glctx.Uniform1i(b.rgar.Gradient, 0)
		b.glctx.Uniform1i(b.rgar.AlphaTex, alphaTexSlot)
		b.glctx.Uniform1f(b.rgar.GlobalAlpha, float32(style.Color.A)/255)
		return b.rgar.Vertex, b.rgar.AlphaTexCoord
	}
	if ip := style.ImagePattern; ip != nil {
		ipd := ip.(*ImagePattern).data
		img := ipd.Image.(*Image)
		b.glctx.UseProgram(b.ipar.ID)
		b.glctx.ActiveTexture(gl.TEXTURE0)
		b.glctx.BindTexture(gl.TEXTURE_2D, img.tex)
		b.glctx.Uniform2f(b.ipar.CanvasSize, float32(b.fw), float32(b.fh))
		b.glctx.Uniform2f(b.ipar.ImageSize, float32(img.w), float32(img.h))
		b.glctx.Uniform1i(b.ipar.Image, 0)
		var f32mat [9]float32
		for i, v := range ipd.Transform {
			f32mat[i] = float32(v)
		}
		b.glctx.UniformMatrix3fv(b.ipr.ImageTransform, f32mat[:])
		switch ipd.Repeat {
		case backendbase.Repeat:
			b.glctx.Uniform2f(b.ipr.Repeat, 1, 1)
		case backendbase.RepeatX:
			b.glctx.Uniform2f(b.ipr.Repeat, 1, 0)
		case backendbase.RepeatY:
			b.glctx.Uniform2f(b.ipr.Repeat, 0, 1)
		case backendbase.NoRepeat:
			b.glctx.Uniform2f(b.ipr.Repeat, 0, 0)
		}
		b.glctx.Uniform1i(b.ipar.AlphaTex, alphaTexSlot)
		b.glctx.Uniform1f(b.ipar.GlobalAlpha, float32(style.Color.A)/255)
		return b.ipar.Vertex, b.ipar.AlphaTexCoord
	}

	b.glctx.UseProgram(b.sar.ID)
	b.glctx.Uniform2f(b.sar.CanvasSize, float32(b.fw), float32(b.fh))
	c := colorGoToGL(style.Color)
	b.glctx.Uniform4f(b.sar.Color, float32(c.r), float32(c.g), float32(c.b), float32(c.a))
	b.glctx.Uniform1i(b.sar.AlphaTex, alphaTexSlot)
	b.glctx.Uniform1f(b.sar.GlobalAlpha, 1)
	return b.sar.Vertex, b.sar.AlphaTexCoord
}

func (b *XMobileBackend) enableTextureRenderTarget(offscr *offscreenBuffer) {
	if offscr.w == b.w && offscr.h == b.h {
		b.glctx.BindFramebuffer(gl.FRAMEBUFFER, offscr.frameBuf)
		return
	}

	if b.w == 0 || b.h == 0 {
		return
	}

	if offscr.w != 0 && offscr.h != 0 {
		b.glctx.DeleteTexture(offscr.tex)
		b.glctx.DeleteFramebuffer(offscr.frameBuf)
		b.glctx.DeleteRenderbuffer(offscr.renderStencilBuf)
	}
	offscr.w = b.w
	offscr.h = b.h

	b.glctx.ActiveTexture(gl.TEXTURE0)
	offscr.tex = b.glctx.CreateTexture()
	b.glctx.BindTexture(gl.TEXTURE_2D, offscr.tex)
	// todo do non-power-of-two textures work everywhere?
	if offscr.alpha {
		b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, b.w, b.h, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	} else {
		b.glctx.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, b.w, b.h, gl.RGB, gl.UNSIGNED_BYTE, nil)
	}
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	b.glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	offscr.frameBuf = b.glctx.CreateFramebuffer()
	b.glctx.BindFramebuffer(gl.FRAMEBUFFER, offscr.frameBuf)

	offscr.renderStencilBuf = b.glctx.CreateRenderbuffer()
	b.glctx.BindRenderbuffer(gl.RENDERBUFFER, offscr.renderStencilBuf)
	b.glctx.RenderbufferStorage(gl.RENDERBUFFER, gl.STENCIL_INDEX8, b.w, b.h)
	b.glctx.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.STENCIL_ATTACHMENT, gl.RENDERBUFFER, offscr.renderStencilBuf)

	b.glctx.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, offscr.tex, 0)

	if err := b.glctx.CheckFramebufferStatus(gl.FRAMEBUFFER); err != gl.FRAMEBUFFER_COMPLETE {
		// todo this should maybe not panic
		panic(fmt.Sprintf("Failed to set up framebuffer for offscreen texture: %x", err))
	}

	b.glctx.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
}

type vec [2]float64

func (v vec) sub(v2 vec) vec {
	return vec{v[0] - v2[0], v[1] - v2[1]}
}

func (v vec) len() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1])
}

func (v vec) scale(f float64) vec {
	return vec{v[0] * f, v[1] * f}
}

func byteSlice(ptr unsafe.Pointer, size int) []byte {
	var buf []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = size
	sh.Len = size
	sh.Data = uintptr(ptr)
	return buf
}
