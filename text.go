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

var fontRenderingContext = newFRContext()

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
		gli.TexImage2D(gl_TEXTURE_2D, 0, gl_RGBA, int32(cv.w), int32(cv.h), 0, gl_RGBA, gl_UNSIGNED_BYTE, nil)
	}

	fontRenderingContext.setFont(cv.state.font.font)
	fontRenderingContext.setFontSize(float64(cv.state.fontSize))
	f := cv.state.fill
	fontRenderingContext.setSrc(image.NewUniform(colorGLToGo(f.r, f.g, f.b, f.a)))
	fontRenderingContext.setDst(cv.text.target)
	fontRenderingContext.setClip(cv.text.target.Bounds())
	_, bounds, _ := fontRenderingContext.drawString(str, fixed.Point26_6{X: fixed.Int26_6(x*64 + 0.5), Y: fixed.Int26_6(y*64 + 0.5)})
	subImg := cv.text.target.SubImage(bounds).(*image.RGBA)

	gli.BlendFunc(gl_ONE, gl_ONE_MINUS_SRC_ALPHA)

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, cv.text.tex)

	for y, w, h := 0, bounds.Dx(), bounds.Dy(); y < h; y++ {
		off := y * subImg.Stride
		pix := subImg.Pix
		gli.TexSubImage2D(gl_TEXTURE_2D, 0, 0, int32(cv.h-1-y), int32(w), 1, gl_RGBA, gl_UNSIGNED_BYTE, gli.Ptr(&pix[off]))
		for b := w * 4; b > 0; b-- {
			pix[off] = 0
			off++
		}
	}

	gli.UseProgram(tr.id)
	gli.Uniform1i(tr.image, 0)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	x0 := float32(bounds.Min.X) / cv.fw
	y0 := float32(bounds.Min.Y) / cv.fh
	x1 := float32(bounds.Max.X) / cv.fw
	y1 := float32(bounds.Max.Y) / cv.fh
	data := [16]float32{x0*2 - 1, -y0*2 + 1, x0*2 - 1, -y1*2 + 1, x1*2 - 1, -y1*2 + 1, x1*2 - 1, -y0*2 + 1,
		0, 1, 0, 1 - (y1 - y0), x1 - x0, 1 - (y1 - y0), x1 - x0, 1}
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
