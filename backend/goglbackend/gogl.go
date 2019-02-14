package goglbackend

import (
	"fmt"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/tfriedel6/canvas"
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
}

type offscreenBuffer struct {
	tex              uint32
	w                int
	h                int
	renderStencilBuf uint32
	frameBuf         uint32
	alpha            bool
}

func New(x, y, w, h int) (canvas.Backend, error) {
	err := gl.Init()
	if err != nil {
		return nil, err
	}

	gl.GetError() // clear error state

	b := &GoGLBackend{
		w:  w,
		h:  h,
		fw: float64(w),
		fh: float64(h),
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
