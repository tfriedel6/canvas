package xmobilebackend

import (
	"fmt"
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

	shd unifiedShader

	offscr1 offscreenBuffer
	offscr2 offscreenBuffer

	imageBufTex gl.Texture
	imageBuf    []byte

	ptsBuf []float32
}

// NewGLContext creates all the necessary GL resources,
// like shaders and buffers
func NewGLContext(glctx gl.Context) (*GLContext, error) {
	ctx := &GLContext{
		glctx: glctx,

		ptsBuf: make([]float32, 0, 4096),
	}

	var err error

	b := &XMobileBackend{GLContext: ctx}

	b.glctx.GetError() // clear error state

	err = loadShader(b, unifiedVS, unifiedFS, &ctx.shd.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.shd.shaderProgram.mustLoadLocations(&ctx.shd)
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
	// todo should use gl.RED on OpenGL, gl.ALPHA on OpenGL ES

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
		b.glctx.Viewport(b.x, b.y, b.w, b.h)
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

func (b *XMobileBackend) useShader(style *backendbase.FillStyle, tf [9]float32, useAlpha bool, alphaTexSlot int) (vertexLoc, alphaTexCoordLoc gl.Attrib) {
	b.glctx.UseProgram(b.shd.ID)
	b.glctx.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	b.glctx.UniformMatrix3fv(b.shd.Matrix, tf[:])
	if useAlpha {
		b.glctx.Uniform1i(b.shd.UseAlphaTex, 1)
		b.glctx.Uniform1i(b.shd.AlphaTex, alphaTexSlot)
	} else {
		b.glctx.Uniform1i(b.shd.UseAlphaTex, 0)
	}
	b.glctx.Uniform1f(b.shd.GlobalAlpha, float32(style.Color.A)/255)

	if lg := style.LinearGradient; lg != nil {
		lg := lg.(*LinearGradient)
		b.glctx.ActiveTexture(gl.TEXTURE0)
		b.glctx.BindTexture(gl.TEXTURE_2D, lg.tex)
		from := backendbase.Vec{style.Gradient.X0, style.Gradient.Y0}
		to := backendbase.Vec{style.Gradient.X1, style.Gradient.Y1}
		dir := to.Sub(from)
		length := dir.Len()
		dir = dir.Mulf(1 / length)
		b.glctx.Uniform2f(b.shd.From, float32(from[0]), float32(from[1]))
		b.glctx.Uniform2f(b.shd.Dir, float32(dir[0]), float32(dir[1]))
		b.glctx.Uniform1f(b.shd.Len, float32(length))
		b.glctx.Uniform1i(b.shd.Gradient, 0)
		b.glctx.Uniform1i(b.shd.Func, shdFuncLinearGradient)
		return b.shd.Vertex, b.shd.TexCoord
	}
	if rg := style.RadialGradient; rg != nil {
		rg := rg.(*RadialGradient)
		b.glctx.ActiveTexture(gl.TEXTURE0)
		b.glctx.BindTexture(gl.TEXTURE_2D, rg.tex)
		b.glctx.Uniform2f(b.shd.From, float32(style.Gradient.X0), float32(style.Gradient.Y0))
		b.glctx.Uniform2f(b.shd.To, float32(style.Gradient.X1), float32(style.Gradient.Y1))
		b.glctx.Uniform1f(b.shd.RadFrom, float32(style.Gradient.RadFrom))
		b.glctx.Uniform1f(b.shd.RadTo, float32(style.Gradient.RadTo))
		b.glctx.Uniform1i(b.shd.Gradient, 0)
		b.glctx.Uniform1i(b.shd.Func, shdFuncRadialGradient)
		return b.shd.Vertex, b.shd.TexCoord
	}
	if ip := style.ImagePattern; ip != nil {
		ipd := ip.(*ImagePattern).data
		img := ipd.Image.(*Image)
		b.glctx.ActiveTexture(gl.TEXTURE0)
		b.glctx.BindTexture(gl.TEXTURE_2D, img.tex)
		b.glctx.Uniform2f(b.shd.ImageSize, float32(img.w), float32(img.h))
		b.glctx.Uniform1i(b.shd.Image, 0)
		var f32mat [9]float32
		for i, v := range ipd.Transform {
			f32mat[i] = float32(v)
		}
		b.glctx.UniformMatrix3fv(b.shd.ImageTransform, f32mat[:])
		switch ipd.Repeat {
		case backendbase.Repeat:
			b.glctx.Uniform2f(b.shd.Repeat, 1, 1)
		case backendbase.RepeatX:
			b.glctx.Uniform2f(b.shd.Repeat, 1, 0)
		case backendbase.RepeatY:
			b.glctx.Uniform2f(b.shd.Repeat, 0, 1)
		case backendbase.NoRepeat:
			b.glctx.Uniform2f(b.shd.Repeat, 0, 0)
		}
		b.glctx.Uniform1i(b.shd.Func, shdFuncImagePattern)
		return b.shd.Vertex, b.shd.TexCoord
	}

	cr := float32(style.Color.R) / 255
	cg := float32(style.Color.G) / 255
	cb := float32(style.Color.B) / 255
	ca := float32(style.Color.A) / 255
	b.glctx.Uniform4f(b.shd.Color, cr, cg, cb, ca)
	b.glctx.Uniform1f(b.shd.GlobalAlpha, 1)
	b.glctx.Uniform1i(b.shd.Func, shdFuncSolid)
	return b.shd.Vertex, b.shd.TexCoord
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

func mat3(m backendbase.Mat) (m3 [9]float32) {
	m3[0] = float32(m[0])
	m3[1] = float32(m[1])
	m3[2] = 0
	m3[3] = float32(m[2])
	m3[4] = float32(m[3])
	m3[5] = 0
	m3[6] = float32(m[4])
	m3[7] = float32(m[5])
	m3[8] = 1
	return
}

var mat3identity = [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1}

func byteSlice(ptr unsafe.Pointer, size int) []byte {
	var buf []byte
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Cap = size
	sh.Len = size
	sh.Data = uintptr(ptr)
	return buf
}
