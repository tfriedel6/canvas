package goglbackend

import (
	"fmt"
	"image/color"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/tfriedel6/canvas/backend/backendbase"
)

const alphaTexSize = 2048

type GoGLBackend struct {
	x, y, w, h     int
	fx, fy, fw, fh float64

	buf       uint32
	shadowBuf uint32
	alphaTex  uint32
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
	offscr1   offscreenBuffer
	offscr2   offscreenBuffer
	glChan    chan func()

	ptsBuf []float32
}

type offscreenBuffer struct {
	tex              uint32
	w                int
	h                int
	renderStencilBuf uint32
	frameBuf         uint32
	alpha            bool
}

func New(x, y, w, h int) (backendbase.Backend, error) {
	err := gl.Init()
	if err != nil {
		return nil, err
	}

	gl.GetError() // clear error state

	b := &GoGLBackend{
		w:      w,
		h:      h,
		fw:     float64(w),
		fh:     float64(h),
		ptsBuf: make([]float32, 0, 4096),
	}

	err = loadShader(solidVS, solidFS, &b.sr.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.sr.shaderProgram.mustLoadLocations(&b.sr)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(linearGradientVS, linearGradientFS, &b.lgr.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.lgr.shaderProgram.mustLoadLocations(&b.lgr)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(radialGradientVS, radialGradientFS, &b.rgr.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.rgr.shaderProgram.mustLoadLocations(&b.rgr)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(imagePatternVS, imagePatternFS, &b.ipr.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.ipr.shaderProgram.mustLoadLocations(&b.ipr)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(solidAlphaVS, solidAlphaFS, &b.sar.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.sar.shaderProgram.mustLoadLocations(&b.sar)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(linearGradientAlphaVS, linearGradientFS, &b.lgar.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.lgar.shaderProgram.mustLoadLocations(&b.lgar)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(radialGradientAlphaVS, radialGradientAlphaFS, &b.rgar.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.rgar.shaderProgram.mustLoadLocations(&b.rgar)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(imagePatternAlphaVS, imagePatternAlphaFS, &b.ipar.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.ipar.shaderProgram.mustLoadLocations(&b.ipar)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(imageVS, imageFS, &b.ir.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.ir.shaderProgram.mustLoadLocations(&b.ir)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(gaussian15VS, gaussian15FS, &b.gauss15r.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.gauss15r.shaderProgram.mustLoadLocations(&b.gauss15r)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(gaussian63VS, gaussian63FS, &b.gauss63r.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.gauss63r.shaderProgram.mustLoadLocations(&b.gauss63r)
	if err = glError(); err != nil {
		return nil, err
	}

	err = loadShader(gaussian127VS, gaussian127FS, &b.gauss127r.shaderProgram)
	if err != nil {
		return nil, err
	}
	b.gauss127r.shaderProgram.mustLoadLocations(&b.gauss127r)
	if err = glError(); err != nil {
		return nil, err
	}

	gl.GenBuffers(1, &b.buf)
	if err = glError(); err != nil {
		return nil, err
	}

	gl.GenBuffers(1, &b.shadowBuf)
	if err = glError(); err != nil {
		return nil, err
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &b.alphaTex)
	gl.BindTexture(gl.TEXTURE_2D, b.alphaTex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.ALPHA, alphaTexSize, alphaTexSize, 0, gl.ALPHA, gl.UNSIGNED_BYTE, nil)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.STENCIL_TEST)
	gl.StencilMask(0xFF)
	gl.Clear(gl.STENCIL_BUFFER_BIT)
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.KEEP)
	gl.StencilFunc(gl.EQUAL, 0, 0xFF)

	gl.Enable(gl.SCISSOR_TEST)

	return b, nil
}

// SetBounds updates the bounds of the canvas. This would
// usually be called for example when the window is resized
func (b *GoGLBackend) SetBounds(x, y, w, h int) {
	b.x, b.y = x, y
	b.fx, b.fy = float64(x), float64(y)
	b.w, b.h = w, h
	b.fw, b.fh = float64(w), float64(h)
}

func glError() error {
	glErr := gl.GetError()
	if glErr != gl.NO_ERROR {
		return fmt.Errorf("GL Error: %x", glErr)
	}
	return nil
}

// Activate makes this GL backend active and sets the viewport. Only
// needs to be called if any other GL code changes the viewport
func (b *GoGLBackend) Activate() {
	// if b.offscreen {
	// 	gli.Viewport(0, 0, int32(cv.w), int32(cv.h))
	// 	cv.enableTextureRenderTarget(&cv.offscrBuf)
	// 	cv.offscrImg.w = cv.offscrBuf.w
	// 	cv.offscrImg.h = cv.offscrBuf.h
	// 	cv.offscrImg.tex = cv.offscrBuf.tex
	// } else {
	gl.Viewport(int32(b.x), int32(b.y), int32(b.w), int32(b.h))
	b.disableTextureRenderTarget()
	// }
	// b.applyScissor()
	gl.Clear(gl.STENCIL_BUFFER_BIT)
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

func (b *GoGLBackend) useShader(style *backendbase.FillStyle) (vertexLoc uint32) {
	// if lg := style.LinearGradient; lg != nil {
	// 	lg.load()
	// 	gl.ActiveTexture(gl.TEXTURE0)
	// 	gl.BindTexture(gl.TEXTURE_2D, lg.tex)
	// 	gl.UseProgram(lgr.id)
	// 	from := cv.tf(lg.from)
	// 	to := cv.tf(lg.to)
	// 	dir := to.sub(from)
	// 	length := dir.len()
	// 	dir = dir.divf(length)
	// 	gl.Uniform2f(lgr.canvasSize, float32(cv.fw), float32(cv.fh))
	// 	inv := cv.state.transform.invert().f32()
	// 	gl.UniformMatrix3fv(lgr.invmat, 1, false, &inv[0])
	// 	gl.Uniform2f(lgr.from, float32(from[0]), float32(from[1]))
	// 	gl.Uniform2f(lgr.dir, float32(dir[0]), float32(dir[1]))
	// 	gl.Uniform1f(lgr.len, float32(length))
	// 	gl.Uniform1i(lgr.gradient, 0)
	// 	gl.Uniform1f(lgr.globalAlpha, float32(cv.state.globalAlpha))
	// 	return lgr.vertex
	// }
	// if rg := style.RadialGradient; rg != nil {
	// 	rg.load()
	// 	gl.ActiveTexture(gl.TEXTURE0)
	// 	gl.BindTexture(gl.TEXTURE_2D, rg.tex)
	// 	gl.UseProgram(rgr.id)
	// 	from := cv.tf(rg.from)
	// 	to := cv.tf(rg.to)
	// 	dir := to.sub(from)
	// 	length := dir.len()
	// 	dir = dir.divf(length)
	// 	gl.Uniform2f(rgr.canvasSize, float32(cv.fw), float32(cv.fh))
	// 	inv := cv.state.transform.invert().f32()
	// 	gl.UniformMatrix3fv(rgr.invmat, 1, false, &inv[0])
	// 	gl.Uniform2f(rgr.from, float32(from[0]), float32(from[1]))
	// 	gl.Uniform2f(rgr.to, float32(to[0]), float32(to[1]))
	// 	gl.Uniform2f(rgr.dir, float32(dir[0]), float32(dir[1]))
	// 	gl.Uniform1f(rgr.radFrom, float32(rg.radFrom))
	// 	gl.Uniform1f(rgr.radTo, float32(rg.radTo))
	// 	gl.Uniform1f(rgr.len, float32(length))
	// 	gl.Uniform1i(rgr.gradient, 0)
	// 	gl.Uniform1f(rgr.globalAlpha, float32(cv.state.globalAlpha))
	// 	return rgr.vertex
	// }
	// if img := style.Image; img != nil {
	// 	gl.UseProgram(ipr.id)
	// 	gl.ActiveTexture(gl.TEXTURE0)
	// 	gl.BindTexture(gl.TEXTURE_2D, img.tex)
	// 	gl.Uniform2f(ipr.canvasSize, float32(cv.fw), float32(cv.fh))
	// 	inv := cv.state.transform.invert().f32()
	// 	gl.UniformMatrix3fv(ipr.invmat, 1, false, &inv[0])
	// 	gl.Uniform2f(ipr.imageSize, float32(img.w), float32(img.h))
	// 	gl.Uniform1i(ipr.image, 0)
	// 	gl.Uniform1f(ipr.globalAlpha, float32(cv.state.globalAlpha))
	// 	return ipr.vertex
	// }

	gl.UseProgram(b.sr.ID)
	gl.Uniform2f(b.sr.CanvasSize, float32(b.fw), float32(b.fh))
	c := colorGoToGL(style.Color)
	gl.Uniform4f(b.sr.Color, float32(c.r), float32(c.g), float32(c.b), float32(c.a))
	gl.Uniform1f(b.sr.GlobalAlpha, 1)
	return b.sr.Vertex
}

func (b *GoGLBackend) useAlphaShader(style *backendbase.FillStyle, alphaTexSlot int32) (vertexLoc, alphaTexCoordLoc uint32) {
	// if lg := style.LinearGradient; lg != nil {
	// 	lg.load()
	// 	gl.ActiveTexture(gl.TEXTURE0)
	// 	gl.BindTexture(gl.TEXTURE_2D, lg.tex)
	// 	gl.UseProgram(lgar.id)
	// 	from := cv.tf(lg.from)
	// 	to := cv.tf(lg.to)
	// 	dir := to.sub(from)
	// 	length := dir.len()
	// 	dir = dir.divf(length)
	// 	gl.Uniform2f(lgar.canvasSize, float32(cv.fw), float32(cv.fh))
	// 	inv := cv.state.transform.invert().f32()
	// 	gl.UniformMatrix3fv(lgar.invmat, 1, false, &inv[0])
	// 	gl.Uniform2f(lgar.from, float32(from[0]), float32(from[1]))
	// 	gl.Uniform2f(lgar.dir, float32(dir[0]), float32(dir[1]))
	// 	gl.Uniform1f(lgar.len, float32(length))
	// 	gl.Uniform1i(lgar.gradient, 0)
	// 	gl.Uniform1i(lgar.alphaTex, alphaTexSlot)
	// 	gl.Uniform1f(lgar.globalAlpha, float32(cv.state.globalAlpha))
	// 	return lgar.vertex, lgar.alphaTexCoord
	// }
	// if rg := style.RadialGradient; rg != nil {
	// 	rg.load()
	// 	gl.ActiveTexture(gl.TEXTURE0)
	// 	gl.BindTexture(gl.TEXTURE_2D, rg.tex)
	// 	gl.UseProgram(rgar.id)
	// 	from := cv.tf(rg.from)
	// 	to := cv.tf(rg.to)
	// 	dir := to.sub(from)
	// 	length := dir.len()
	// 	dir = dir.divf(length)
	// 	gl.Uniform2f(rgar.canvasSize, float32(cv.fw), float32(cv.fh))
	// 	inv := cv.state.transform.invert().f32()
	// 	gl.UniformMatrix3fv(rgar.invmat, 1, false, &inv[0])
	// 	gl.Uniform2f(rgar.from, float32(from[0]), float32(from[1]))
	// 	gl.Uniform2f(rgar.to, float32(to[0]), float32(to[1]))
	// 	gl.Uniform2f(rgar.dir, float32(dir[0]), float32(dir[1]))
	// 	gl.Uniform1f(rgar.radFrom, float32(rg.radFrom))
	// 	gl.Uniform1f(rgar.radTo, float32(rg.radTo))
	// 	gl.Uniform1f(rgar.len, float32(length))
	// 	gl.Uniform1i(rgar.gradient, 0)
	// 	gl.Uniform1i(rgar.alphaTex, alphaTexSlot)
	// 	gl.Uniform1f(rgar.globalAlpha, float32(cv.state.globalAlpha))
	// 	return rgar.vertex, rgar.alphaTexCoord
	// }
	// if img := style.Image; img != nil {
	// 	gl.UseProgram(ipar.id)
	// 	gl.ActiveTexture(gl.TEXTURE0)
	// 	gl.BindTexture(gl.TEXTURE_2D, img.tex)
	// 	gl.Uniform2f(ipar.canvasSize, float32(cv.fw), float32(cv.fh))
	// 	inv := cv.state.transform.invert().f32()
	// 	gl.UniformMatrix3fv(ipar.invmat, 1, false, &inv[0])
	// 	gl.Uniform2f(ipar.imageSize, float32(img.w), float32(img.h))
	// 	gl.Uniform1i(ipar.image, 0)
	// 	gl.Uniform1i(ipar.alphaTex, alphaTexSlot)
	// 	gl.Uniform1f(ipar.globalAlpha, float32(cv.state.globalAlpha))
	// 	return ipar.vertex, ipar.alphaTexCoord
	// }

	gl.UseProgram(b.sar.ID)
	gl.Uniform2f(b.sar.CanvasSize, float32(b.fw), float32(b.fh))
	c := colorGoToGL(style.Color)
	gl.Uniform4f(b.sar.Color, float32(c.r), float32(c.g), float32(c.b), float32(c.a))
	gl.Uniform1i(b.sar.AlphaTex, alphaTexSlot)
	gl.Uniform1f(b.sar.GlobalAlpha, 1)
	return b.sar.Vertex, b.sar.AlphaTexCoord
}

func (b *GoGLBackend) enableTextureRenderTarget(offscr *offscreenBuffer) {
	if offscr.w != b.w || offscr.h != b.h {
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

		gl.GenFramebuffers(1, &offscr.frameBuf)
		gl.BindFramebuffer(gl.FRAMEBUFFER, offscr.frameBuf)

		gl.GenRenderbuffers(1, &offscr.renderStencilBuf)
		gl.BindRenderbuffer(gl.RENDERBUFFER, offscr.renderStencilBuf)
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, int32(b.w), int32(b.h))
		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, offscr.renderStencilBuf)

		gl.FramebufferTexture(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, offscr.tex, 0)

		if err := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); err != gl.FRAMEBUFFER_COMPLETE {
			// todo this should maybe not panic
			panic(fmt.Sprintf("Failed to set up framebuffer for offscreen texture: %x", err))
		}

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	} else {
		gl.BindFramebuffer(gl.FRAMEBUFFER, offscr.frameBuf)
	}
}

func (b *GoGLBackend) disableTextureRenderTarget() {
	// if b.offscreen {
	// 	b.enableTextureRenderTarget(&b.offscrBuf)
	// } else {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	// }
}
