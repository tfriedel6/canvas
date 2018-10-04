package canvas

import (
	"errors"
	"image"
	"image/draw"
	"io/ioutil"
	"unsafe"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var fontRenderingContext = newFRContext()

// Font is a loaded font that can be passed to the
// SetFont method
type Font struct {
	font *truetype.Font
}

var fonts = make(map[string]*Font)
var zeroes [alphaTexSize]byte
var textImage *image.Alpha

var defaultFont *Font

// LoadFont loads a font and returns the result. The font
// can be a file name or a byte slice in TTF format
func LoadFont(src interface{}) (*Font, error) {
	if gli == nil {
		panic("LoadGL must be called before fonts can be loaded")
	}
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
	if defaultFont == nil {
		defaultFont = f
	}
	return f, nil
}

// FillText draws the given string at the given coordinates
// using the currently set font and font height
func (cv *Canvas) FillText(str string, x, y float64) {
	cv.activate()

	if cv.state.font == nil {
		return
	}

	frc := fontRenderingContext
	frc.setFont(cv.state.font.font)
	frc.setFontSize(float64(cv.state.fontSize))
	fnt := cv.state.font.font

	curX := x
	var p fixed.Point26_6
	prev, hasPrev := truetype.Index(0), false

	strFrom, strTo := 0, len(str)
	curInside := false

	var textOffset image.Point
	var strWidth, strMaxY int
	for i, rn := range str {
		idx := fnt.Index(rn)
		if idx == 0 {
			idx = fnt.Index(' ')
		}
		advance, bounds, err := frc.glyphMeasure(idx, p)
		if err != nil {
			continue
		}
		if hasPrev {
			kern := fnt.Kern(frc.scale, prev, idx)
			if frc.hinting != font.HintingNone {
				kern = (kern + 32) &^ 63
			}
			curX += float64(kern) / 64
		}

		p0 := cv.tf(vec{float64(bounds.Min.X) + curX, float64(bounds.Min.Y) + y})
		p1 := cv.tf(vec{float64(bounds.Min.X) + curX, float64(bounds.Max.Y) + y})
		p2 := cv.tf(vec{float64(bounds.Max.X) + curX, float64(bounds.Max.Y) + y})
		p3 := cv.tf(vec{float64(bounds.Max.X) + curX, float64(bounds.Min.Y) + y})
		inside := (p0[0] >= 0 || p1[0] >= 0 || p2[0] >= 0 || p3[0] >= 0) &&
			(p0[1] >= 0 || p1[1] >= 0 || p2[1] >= 0 || p3[1] >= 0) &&
			(p0[0] < cv.fw || p1[0] < cv.fw || p2[0] < cv.fw || p3[0] < cv.fw) &&
			(p0[1] < cv.fh || p1[1] < cv.fh || p2[1] < cv.fh || p3[1] < cv.fh)

		if !curInside && inside {
			curInside = true
			strFrom = i
			x = curX
		} else if curInside && !inside {
			strTo = i
			break
		}

		if i == 0 {
			textOffset.X = bounds.Min.X
		}
		if bounds.Min.Y < textOffset.Y {
			textOffset.Y = bounds.Min.Y
		}
		if bounds.Max.Y > strMaxY {
			strMaxY = bounds.Max.Y
		}
		p.X += advance
		curX += float64(advance) / 64
	}
	strWidth = p.X.Ceil() - textOffset.X
	strHeight := strMaxY - textOffset.Y

	if strFrom == strTo || strWidth == 0 || strHeight == 0 {
		return
	}

	if textImage == nil || textImage.Bounds().Dx() < strWidth || textImage.Bounds().Dy() < strHeight {
		var size int
		for size = 2; size < alphaTexSize; size *= 2 {
			if size >= strWidth && size >= strHeight {
				break
			}
		}
		if size > alphaTexSize {
			size = alphaTexSize
		}
		textImage = image.NewAlpha(image.Rect(0, 0, size, size))
	}

	curX = x
	p = fixed.Point26_6{}

	for y := 0; y < strHeight; y++ {
		off := textImage.PixOffset(0, y)
		line := textImage.Pix[off : off+strWidth]
		for i := range line {
			line[i] = 0
		}
	}

	prev, hasPrev = truetype.Index(0), false
	for _, rn := range str[strFrom:strTo] {
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
			curX += float64(kern) / 64
		}
		advance, mask, offset, err := frc.glyph(idx, p)
		if err != nil {
			prev = 0
			hasPrev = false
			continue
		}
		p.X += advance

		draw.Draw(textImage, mask.Bounds().Add(offset).Sub(textOffset), mask, image.ZP, draw.Over)

		curX += float64(advance) / 64
	}

	if cv.state.textAlign == Center {
		x -= float64(strWidth) * 0.5
	} else if cv.state.textAlign == Right || cv.state.textAlign == End {
		x -= float64(strWidth)
	}

	gli.ActiveTexture(gl_TEXTURE1)
	gli.BindTexture(gl_TEXTURE_2D, alphaTex)
	for y := 0; y < strHeight; y++ {
		off := y * textImage.Stride
		gli.TexSubImage2D(gl_TEXTURE_2D, 0, 0, int32(alphaTexSize-1-y), int32(strWidth), 1, gl_ALPHA, gl_UNSIGNED_BYTE, gli.Ptr(&textImage.Pix[off]))
	}

	cv.drawTextShadow(textOffset, strWidth, strHeight, x, y)

	gli.StencilFunc(gl_EQUAL, 0, 0xFF)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	vertex, alphaTexCoord := cv.useAlphaShader(&cv.state.fill, 1)

	gli.EnableVertexAttribArray(vertex)
	gli.EnableVertexAttribArray(alphaTexCoord)

	p0 := cv.tf(vec{float64(textOffset.X) + x, float64(textOffset.Y) + y})
	p1 := cv.tf(vec{float64(textOffset.X) + x, float64(textOffset.Y+strHeight) + y})
	p2 := cv.tf(vec{float64(textOffset.X+strWidth) + x, float64(textOffset.Y+strHeight) + y})
	p3 := cv.tf(vec{float64(textOffset.X+strWidth) + x, float64(textOffset.Y) + y})

	tw := float64(strWidth) / alphaTexSize
	th := float64(strHeight) / alphaTexSize
	data := [16]float32{float32(p0[0]), float32(p0[1]), float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), float32(p3[0]), float32(p3[1]),
		0, 1, 0, float32(1 - th), float32(tw), float32(1 - th), float32(tw), 1}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.VertexAttribPointer(vertex, 2, gl_FLOAT, false, 0, 0)
	gli.VertexAttribPointer(alphaTexCoord, 2, gl_FLOAT, false, 0, 8*4)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)

	gli.DisableVertexAttribArray(vertex)
	gli.DisableVertexAttribArray(alphaTexCoord)

	gli.ActiveTexture(gl_TEXTURE1)
	gli.BindTexture(gl_TEXTURE_2D, alphaTex)

	gli.StencilFunc(gl_ALWAYS, 0, 0xFF)

	for y := 0; y < strHeight; y++ {
		gli.TexSubImage2D(gl_TEXTURE_2D, 0, 0, int32(alphaTexSize-1-y), int32(strWidth), 1, gl_ALPHA, gl_UNSIGNED_BYTE, gli.Ptr(&zeroes[0]))
	}

	gli.ActiveTexture(gl_TEXTURE0)
}

// StrokeText draws the given string at the given coordinates
// using the currently set font and font height and using the
// current stroke style
func (cv *Canvas) StrokeText(str string, x, y float64) {
	cv.activate()

	if cv.state.font == nil {
		return
	}

	frc := fontRenderingContext
	frc.setFont(cv.state.font.font)
	frc.setFontSize(float64(cv.state.fontSize))
	fnt := cv.state.font.font

	for _, rn := range str {
		idx := fnt.Index(rn)
		if idx == 0 {
			idx = fnt.Index(' ')
		}
		advance, _, err := frc.glyphMeasure(idx, fixed.Point26_6{})
		if err != nil {
			continue
		}

		cv.strokeRune(rn, vec{x, y})

		x += float64(advance) / 64
	}
}

func (cv *Canvas) strokeRune(rn rune, pos vec) {
	gb := &truetype.GlyphBuf{}
	gb.Load(cv.state.font.font, fixed.Int26_6(cv.state.fontSize*64), cv.state.font.font.Index(rn), font.HintingNone)

	prevPath := cv.path
	defer func() {
		cv.path = prevPath
	}()

	from := 0
	for _, to := range gb.Ends {
		ps := gb.Points[from:to]

		start := fixed.Point26_6{
			X: ps[0].X,
			Y: ps[0].Y,
		}
		others := []truetype.Point(nil)
		if ps[0].Flags&0x01 != 0 {
			others = ps[1:]
		} else {
			last := fixed.Point26_6{
				X: ps[len(ps)-1].X,
				Y: ps[len(ps)-1].Y,
			}
			if ps[len(ps)-1].Flags&0x01 != 0 {
				start = last
				others = ps[:len(ps)-1]
			} else {
				start = fixed.Point26_6{
					X: (start.X + last.X) / 2,
					Y: (start.Y + last.Y) / 2,
				}
				others = ps
			}
		}

		p0, on0 := gb.Points[from], false
		cv.MoveTo(float64(p0.X)/64+pos[0], pos[1]-float64(p0.Y)/64)
		for _, p := range others {
			on := p.Flags&0x01 != 0
			if on {
				if on0 {
					cv.LineTo(float64(p.X)/64+pos[0], pos[1]-float64(p.Y)/64)
				} else {
					cv.QuadraticCurveTo(float64(p0.X)/64+pos[0], pos[1]-float64(p0.Y)/64, float64(p.X)/64+pos[0], pos[1]-float64(p.Y)/64)
				}
			} else {
				if on0 {
					// No-op.
				} else {
					mid := fixed.Point26_6{
						X: (p0.X + p.X) / 2,
						Y: (p0.Y + p.Y) / 2,
					}
					cv.QuadraticCurveTo(float64(p0.X)/64+pos[0], pos[1]-float64(p0.Y)/64, float64(mid.X)/64+pos[0], pos[1]-float64(mid.Y)/64)
				}
			}
			p0, on0 = p, on
		}

		if on0 {
			cv.LineTo(float64(start.X)/64+pos[0], pos[1]-float64(start.Y)/64)
		} else {
			cv.QuadraticCurveTo(float64(p0.X)/64+pos[0], pos[1]-float64(p0.Y)/64, float64(start.X)/64+pos[0], pos[1]-float64(start.Y)/64)
		}
		cv.ClosePath()
		cv.Stroke()

		from = to
	}
}

// TextMetrics is the result of a MeasureText call
type TextMetrics struct {
	Width                    float64
	ActualBoundingBoxAscent  float64
	ActualBoundingBoxDescent float64
}

// MeasureText measures the given string using the
// current font and font height
func (cv *Canvas) MeasureText(str string) TextMetrics {
	if cv.state.font == nil {
		return TextMetrics{}
	}

	frc := fontRenderingContext
	frc.setFont(cv.state.font.font)
	frc.setFontSize(float64(cv.state.fontSize))
	fnt := cv.state.font.font

	var p fixed.Point26_6
	var x float64
	var minY float64
	var maxY float64
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
			x += float64(kern) / 64
		}

		advance, glyphBounds, err := frc.glyphMeasure(idx, p)
		if err != nil {
			prev = 0
			hasPrev = false
			continue
		}
		if glyphMinY := float64(glyphBounds.Min.Y); glyphMinY < minY {
			minY = glyphMinY
		}
		if glyphMaxY := float64(glyphBounds.Max.Y); glyphMaxY > maxY {
			maxY = glyphMaxY
		}
		x += float64(advance) / 64
		p.X += advance
	}

	return TextMetrics{
		Width: x,
		ActualBoundingBoxAscent:  -minY,
		ActualBoundingBoxDescent: +maxY,
	}
}
