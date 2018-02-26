package canvas

import (
	"errors"
	"io/ioutil"
	"unsafe"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/tfriedel6/lm"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var fontRenderingContext = newFRContext()

type Font struct {
	font *truetype.Font
}

var fonts = make(map[string]*Font)

func LoadFont(src interface{}, name string) (*Font, error) {
	var f *Font
	switch v := src.(type) {
	case *truetype.Font:
		f = &Font{font: v}
	case string:
		data, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, err
		}
		font, err := freetype.ParseFont(data)
		if err != nil {
			return nil, err
		}
		f = &Font{font: font}
	case []byte:
		font, err := freetype.ParseFont(v)
		if err != nil {
			return nil, err
		}
		f = &Font{font: font}
	default:
		return nil, errors.New("Unsupported source type")
	}
	if name != "" {
		fonts[name] = f
	}
	return f, nil
}

func (cv *Canvas) FillText(str string, x, y float32) {
	cv.activate()

	frc := fontRenderingContext
	frc.setFont(cv.state.font.font)
	frc.setFontSize(float64(cv.state.fontSize))

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	vertex, alphaTexCoord := cv.useAlphaShader(&cv.state.fill, 1)

	gli.ActiveTexture(gl_TEXTURE1)
	gli.BindTexture(gl_TEXTURE_2D, alphaTex)

	gli.EnableVertexAttribArray(vertex)
	gli.EnableVertexAttribArray(alphaTexCoord)

	fnt := cv.state.font.font

	prev, hasPrev := truetype.Index(0), false
	for _, rn := range str {
		idx := fnt.Index(rn)
		if idx == 0 {
			prev = 0
			hasPrev = false
			continue
		}
		if hasPrev {
			kern := fnt.Kern(frc.scale, prev, idx)
			if frc.hinting != font.HintingNone {
				kern = (kern + 32) &^ 63
			}
			x += float32(kern) / 64
		}
		advance, mask, offset, err := frc.glyph(idx, fixed.Point26_6{})
		if err != nil {
			prev = 0
			hasPrev = false
			continue
		}
		bounds := mask.Bounds().Add(offset)

		for y, w, h := 0, bounds.Dx(), bounds.Dy(); y < h; y++ {
			off := y * mask.Stride
			gli.TexSubImage2D(gl_TEXTURE_2D, 0, 0, int32(bufferTextureSize-1-y), int32(w), 1, gl_ALPHA, gl_UNSIGNED_BYTE, gli.Ptr(&mask.Pix[off]))
		}

		p0 := cv.tf(lm.Vec2{float32(bounds.Min.X) + x, float32(bounds.Min.Y) + y})
		p1 := cv.tf(lm.Vec2{float32(bounds.Min.X) + x, float32(bounds.Max.Y) + y})
		p2 := cv.tf(lm.Vec2{float32(bounds.Max.X) + x, float32(bounds.Max.Y) + y})
		p3 := cv.tf(lm.Vec2{float32(bounds.Max.X) + x, float32(bounds.Min.Y) + y})
		tw := float32(bounds.Dx()) / bufferTextureSize
		th := float32(bounds.Dy()) / bufferTextureSize
		data := [16]float32{p0[0], p0[1], p1[0], p1[1], p2[0], p2[1], p3[0], p3[1],
			0, 1, 0, 1 - th, tw, 1 - th, tw, 1}
		gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

		gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, nil)
		gli.VertexAttribPointer(alphaTexCoord, 2, gl_FLOAT, false, 0, gli.PtrOffset(8*4))
		gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)

		x += float32(advance) / 64
	}

	gli.DisableVertexAttribArray(vertex)
	gli.DisableVertexAttribArray(alphaTexCoord)

	gli.ActiveTexture(gl_TEXTURE0)
}
