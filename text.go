package canvas

import (
	"errors"
	"image"
	"io/ioutil"
	"unsafe"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
)

var fontRenderingContext = freetype.NewContext()

type Font struct {
	font *truetype.Font
}

func LoadFont(src interface{}) (*Font, error) {
	switch v := src.(type) {
	case *truetype.Font:
		return &Font{font: v}, nil
	case string:
		data, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, err
		}
		font, err := freetype.ParseFont(data)
		if err != nil {
			return nil, err
		}
		return &Font{font: font}, nil
	case []byte:
		font, err := freetype.ParseFont(v)
		if err != nil {
			return nil, err
		}
		return &Font{font: font}, nil
	}
	return nil, errors.New("Unsupported source type")
}

func (cv *Canvas) FillText(str string, x, y float32) {
	cv.activate()

	if cv.text.target == nil || cv.text.target.Bounds().Dx() != cv.w || cv.text.target.Bounds().Dy() != cv.h {
		if cv.text.tex != 0 {
			gli.DeleteTextures(1, &cv.text.tex)
		}
		cv.text.target = image.NewRGBA(image.Rect(0, 0, cv.w, cv.h))
		gli.GenTextures(1, &cv.text.tex)
		gli.ActiveTexture(gl_TEXTURE0)
		gli.BindTexture(gl_TEXTURE_2D, cv.text.tex)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MIN_FILTER, gl_NEAREST)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_MAG_FILTER, gl_NEAREST)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_S, gl_CLAMP_TO_EDGE)
		gli.TexParameteri(gl_TEXTURE_2D, gl_TEXTURE_WRAP_T, gl_CLAMP_TO_EDGE)
	}

	for i := range cv.text.target.Pix {
		cv.text.target.Pix[i] = 0
	}

	fontRenderingContext.SetFont(cv.text.font.font)
	fontRenderingContext.SetFontSize(float64(cv.text.size))
	fontRenderingContext.SetSrc(image.NewUniform(colorGLToGo(cv.fill.r, cv.fill.g, cv.fill.b, cv.fill.a)))
	fontRenderingContext.SetDst(cv.text.target)
	fontRenderingContext.SetClip(cv.text.target.Bounds())
	fontRenderingContext.DrawString(str, fixed.Point26_6{X: fixed.Int26_6(x*64 + 0.5), Y: fixed.Int26_6(y*64 + 0.5)})

	gli.BlendFunc(gl_ONE, gl_ONE_MINUS_SRC_ALPHA)

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, cv.text.tex)
	gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RGBA, int32(cv.w), int32(cv.h), 0, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&cv.text.target.Pix[0]))

	gli.UseProgram(tr.id)
	gli.Uniform1i(tr.image, 0)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [16]float32{-1, -1, -1, 1, 1, 1, 1, -1, 0, 1, 0, 0, 1, 0, 1, 1}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.VertexAttribPointer(tr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.VertexAttribPointer(tr.texCoord, 2, gl_FLOAT, false, 0, gli.PtrOffset(8*4))
	gli.EnableVertexAttribArray(tr.vertex)
	gli.EnableVertexAttribArray(tr.texCoord)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(tr.vertex)
	gli.DisableVertexAttribArray(tr.texCoord)

	gli.BlendFunc(gl_SRC_ALPHA, gl_ONE_MINUS_SRC_ALPHA)
}
