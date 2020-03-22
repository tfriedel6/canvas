package goglbackend

import (
	"fmt"

	"github.com/tfriedel6/canvas/backend/backendbase"
	"github.com/tfriedel6/canvas/backend/goglbackend/gl"
)

const alphaTexSize = 2048

var zeroes [alphaTexSize]byte

// GLContext is a context that contains all the
// shaders and buffers necessary for rendering
type GLContext struct {
	buf       uint32
	shadowBuf uint32
	alphaTex  uint32

	shd unifiedShader

	offscr1 offscreenBuffer
	offscr2 offscreenBuffer

	imageBufTex uint32
	imageBuf    []byte

	ptsBuf []float32
}

// NewGLContext creates all the necessary GL resources,
// like shaders and buffers
func NewGLContext() (*GLContext, error) {
	ctx := &GLContext{
		ptsBuf: make([]float32, 0, 4096),
	}

	err := gl.Init()
	if err != nil {
		return nil, err
	}

	gl.GetError() // clear error state

	err = loadShader(unifiedVS, unifiedFS, &ctx.shd.shaderProgram)
	if err != nil {
		return nil, err
	}
	ctx.shd.shaderProgram.mustLoadLocations(&ctx.shd)
	if err = glError(); err != nil {
		return nil, err
	}

	gl.GenBuffers(1, &ctx.buf)
	if err = glError(); err != nil {
		return nil, err
	}

	gl.GenBuffers(1, &ctx.shadowBuf)
	if err = glError(); err != nil {
		return nil, err
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &ctx.alphaTex)
	gl.BindTexture(gl.TEXTURE_2D, ctx.alphaTex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.ALPHA, alphaTexSize, alphaTexSize, 0, gl.ALPHA, gl.UNSIGNED_BYTE, nil)
	// todo should use gl.RED on OpenGL, gl.ALPHA on OpenGL ES

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.STENCIL_TEST)
	gl.StencilMask(0xFF)
	gl.Clear(gl.STENCIL_BUFFER_BIT)
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.Disable(gl.SCISSOR_TEST)

	return ctx, nil
}

// GoGLBackend is a canvas backend using Go-GL
type GoGLBackend struct {
	x, y, w, h     int
	fx, fy, fw, fh float64

	*GLContext

	activateFn                 func()
	disableTextureRenderTarget func()
}

type offscreenBuffer struct {
	tex              uint32
	w                int
	h                int
	renderStencilBuf uint32
	frameBuf         uint32
	alpha            bool
}

// New returns a new canvas backend. x, y, w, h define the target
// rectangle in the window. ctx is a GLContext created with
// NewGLContext, but can be nil for a default one. It makes sense
// to pass one in when using for example an onscreen and an
// offscreen backend using the same GL context.
func New(x, y, w, h int, ctx *GLContext) (*GoGLBackend, error) {
	if ctx == nil {
		var err error
		ctx, err = NewGLContext()
		if err != nil {
			return nil, err
		}
	}

	b := &GoGLBackend{
		w:         w,
		h:         h,
		fw:        float64(w),
		fh:        float64(h),
		GLContext: ctx,
	}

	b.activateFn = func() {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.Viewport(int32(b.x), int32(b.y), int32(b.w), int32(b.h))
		// todo reapply clipping since another application may have used the stencil buffer
	}
	b.disableTextureRenderTarget = func() {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.Viewport(int32(b.x), int32(b.y), int32(b.w), int32(b.h))
	}

	return b, nil
}

// GoGLBackendOffscreen is a canvas backend using an offscreen
// texture
type GoGLBackendOffscreen struct {
	GoGLBackend

	TextureID uint32

	offscrBuf offscreenBuffer
	offscrImg Image
}

// NewOffscreen returns a new offscreen canvas backend. w, h define
// the size of the offscreen texture. ctx is a GLContext created
// with NewGLContext, but can be nil for a default one. It makes
// sense to pass one in when using for example an onscreen and an
// offscreen backend using the same GL context.
func NewOffscreen(w, h int, alpha bool, ctx *GLContext) (*GoGLBackendOffscreen, error) {
	b, err := New(0, 0, w, h, ctx)
	if err != nil {
		return nil, err
	}
	bo := &GoGLBackendOffscreen{GoGLBackend: *b}
	bo.offscrBuf.alpha = alpha
	bo.offscrImg.flip = true

	bo.activateFn = func() {
		bo.enableTextureRenderTarget(&bo.offscrBuf)
		gl.Viewport(0, 0, int32(bo.w), int32(bo.h))
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
func (b *GoGLBackend) SetBounds(x, y, w, h int) {
	b.x, b.y = x, y
	b.fx, b.fy = float64(x), float64(y)
	b.w, b.h = w, h
	b.fw, b.fh = float64(w), float64(h)
	if b == activeContext {
		gl.Viewport(0, 0, int32(b.w), int32(b.h))
		gl.Clear(gl.STENCIL_BUFFER_BIT)
	}
}

// SetSize updates the size of the offscreen texture
func (b *GoGLBackendOffscreen) SetSize(w, h int) {
	b.GoGLBackend.SetBounds(0, 0, w, h)
	b.offscrImg.w = b.offscrBuf.w
	b.offscrImg.h = b.offscrBuf.h
}

// Size returns the size of the window or offscreen
// texture
func (b *GoGLBackend) Size() (int, int) {
	return b.w, b.h
}

func glError() error {
	glErr := gl.GetError()
	if glErr != gl.NO_ERROR {
		return fmt.Errorf("GL Error: %x", glErr)
	}
	return nil
}

// Activate only needs to be called if there is other
// code also using the GL state
func (b *GoGLBackend) Activate() {
	b.activate()
}

var activeContext *GoGLBackend

func (b *GoGLBackend) activate() {
	if activeContext != b {
		activeContext = b
		b.activateFn()
	}
}

// Delete deletes the offscreen texture. After calling this
// the backend can no longer be used
func (b *GoGLBackendOffscreen) Delete() {
	gl.DeleteTextures(1, &b.offscrBuf.tex)
	gl.DeleteFramebuffers(1, &b.offscrBuf.frameBuf)
	gl.DeleteRenderbuffers(1, &b.offscrBuf.renderStencilBuf)
}

// CanUseAsImage returns true if the given backend can be
// directly used by this backend to avoid a conversion.
// Used internally
func (b *GoGLBackend) CanUseAsImage(b2 backendbase.Backend) bool {
	_, ok := b2.(*GoGLBackendOffscreen)
	return ok
}

// AsImage returns nil, since this backend cannot be directly
// used as an image. Used internally
func (b *GoGLBackend) AsImage() backendbase.Image {
	return nil
}

// AsImage returns an implementation of the Image interface
// that can be used to render this offscreen texture
// directly. Used internally
func (b *GoGLBackendOffscreen) AsImage() backendbase.Image {
	return &b.offscrImg
}

func (b *GoGLBackend) useShader(style *backendbase.FillStyle, tf [9]float32, useAlpha bool, alphaTexSlot int32) (vertexLoc, alphaTexCoordLoc uint32) {
	gl.UseProgram(b.shd.ID)
	gl.Uniform2f(b.shd.CanvasSize, float32(b.fw), float32(b.fh))
	gl.UniformMatrix3fv(b.shd.Matrix, 1, false, &tf[0])
	if useAlpha {
		gl.Uniform1i(b.shd.UseAlphaTex, 1)
		gl.Uniform1i(b.shd.AlphaTex, alphaTexSlot)
	} else {
		gl.Uniform1i(b.shd.UseAlphaTex, 0)
	}
	gl.Uniform1f(b.shd.GlobalAlpha, float32(style.Color.A)/255)

	if lg := style.LinearGradient; lg != nil {
		lg := lg.(*LinearGradient)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, lg.tex)
		from := backendbase.Vec{style.Gradient.X0, style.Gradient.Y0}
		to := backendbase.Vec{style.Gradient.X1, style.Gradient.Y1}
		dir := to.Sub(from)
		length := dir.Len()
		dir = dir.Mulf(1 / length)
		gl.Uniform2f(b.shd.From, float32(from[0]), float32(from[1]))
		gl.Uniform2f(b.shd.Dir, float32(dir[0]), float32(dir[1]))
		gl.Uniform1f(b.shd.Len, float32(length))
		gl.Uniform1i(b.shd.Gradient, 0)
		gl.Uniform1i(b.shd.Func, shdFuncLinearGradient)
		return b.shd.Vertex, b.shd.TexCoord
	}
	if rg := style.RadialGradient; rg != nil {
		rg := rg.(*RadialGradient)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, rg.tex)
		gl.Uniform2f(b.shd.From, float32(style.Gradient.X0), float32(style.Gradient.Y0))
		gl.Uniform2f(b.shd.To, float32(style.Gradient.X1), float32(style.Gradient.Y1))
		gl.Uniform1f(b.shd.RadFrom, float32(style.Gradient.RadFrom))
		gl.Uniform1f(b.shd.RadTo, float32(style.Gradient.RadTo))
		gl.Uniform1i(b.shd.Gradient, 0)
		gl.Uniform1i(b.shd.Func, shdFuncRadialGradient)
		return b.shd.Vertex, b.shd.TexCoord
	}
	if ip := style.ImagePattern; ip != nil {
		ipd := ip.(*ImagePattern).data
		img := ipd.Image.(*Image)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, img.tex)
		gl.Uniform2f(b.shd.ImageSize, float32(img.w), float32(img.h))
		gl.Uniform1i(b.shd.Image, 0)
		var f32mat [9]float32
		for i, v := range ipd.Transform {
			f32mat[i] = float32(v)
		}
		gl.UniformMatrix3fv(b.shd.ImageTransform, 1, false, &f32mat[0])
		switch ipd.Repeat {
		case backendbase.Repeat:
			gl.Uniform2f(b.shd.Repeat, 1, 1)
		case backendbase.RepeatX:
			gl.Uniform2f(b.shd.Repeat, 1, 0)
		case backendbase.RepeatY:
			gl.Uniform2f(b.shd.Repeat, 0, 1)
		case backendbase.NoRepeat:
			gl.Uniform2f(b.shd.Repeat, 0, 0)
		}
		gl.Uniform1i(b.shd.Func, shdFuncImagePattern)
		return b.shd.Vertex, b.shd.TexCoord
	}

	cr := float32(style.Color.R) / 255
	cg := float32(style.Color.G) / 255
	cb := float32(style.Color.B) / 255
	ca := float32(style.Color.A) / 255
	gl.Uniform4f(b.shd.Color, cr, cg, cb, ca)
	gl.Uniform1f(b.shd.GlobalAlpha, 1)
	gl.Uniform1i(b.shd.Func, shdFuncSolid)
	return b.shd.Vertex, b.shd.TexCoord
}

func (b *GoGLBackend) enableTextureRenderTarget(offscr *offscreenBuffer) {
	if offscr.w == b.w && offscr.h == b.h {
		gl.BindFramebuffer(gl.FRAMEBUFFER, offscr.frameBuf)
		return
	}

	if b.w == 0 || b.h == 0 {
		return
	}

	if offscr.w != 0 && offscr.h != 0 {
		gl.DeleteTextures(1, &offscr.tex)
		gl.DeleteFramebuffers(1, &offscr.frameBuf)
		gl.DeleteRenderbuffers(1, &offscr.renderStencilBuf)
	}
	offscr.w = b.w
	offscr.h = b.h

	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &offscr.tex)
	gl.BindTexture(gl.TEXTURE_2D, offscr.tex)
	// todo do non-power-of-two textures work everywhere?
	if offscr.alpha {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(b.w), int32(b.h), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	} else {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB, int32(b.w), int32(b.h), 0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	}
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.GenFramebuffers(1, &offscr.frameBuf)
	gl.BindFramebuffer(gl.FRAMEBUFFER, offscr.frameBuf)

	gl.GenRenderbuffers(1, &offscr.renderStencilBuf)
	gl.BindRenderbuffer(gl.RENDERBUFFER, offscr.renderStencilBuf)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.STENCIL_INDEX8, int32(b.w), int32(b.h))
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.STENCIL_ATTACHMENT, gl.RENDERBUFFER, offscr.renderStencilBuf)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, offscr.tex, 0)

	if err := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); err != gl.FRAMEBUFFER_COMPLETE {
		// todo this should maybe not panic
		panic(fmt.Sprintf("Failed to set up framebuffer for offscreen texture: %x", err))
	}

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
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
