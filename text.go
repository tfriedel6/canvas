package canvas

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"io/ioutil"
	"math"
	"os"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/tfriedel6/canvas/backend/backendbase"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Font is a loaded font that can be passed to the
// SetFont method
type Font struct {
	font *truetype.Font
}

type fontKey struct {
	font *Font
	size fixed.Int26_6
}

type frCache struct {
	ctx      *frContext
	lastUsed time.Time
}

var zeroes [alphaTexSize]byte
var textImage *image.Alpha

var defaultFont *Font

// LoadFont loads a font and returns the result. The font
// can be a file name or a byte slice in TTF format
func (cv *Canvas) LoadFont(src interface{}) (*Font, error) {
	if f, ok := src.(*Font); ok {
		return f, nil
	} else if _, ok := src.([]byte); !ok {
		if f, ok := cv.fonts[src]; ok {
			return f, nil
		}
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

	if _, ok := src.([]byte); !ok {
		cv.fonts[src] = f
	}
	return f, nil
}

func (cv *Canvas) getFont(src interface{}) *Font {
	f, err := cv.LoadFont(src)
	if err != nil {
		cv.fonts[src] = nil
		fmt.Fprintf(os.Stderr, "Error loading font: %v\n", err)
	} else {
		cv.fonts[src] = f
	}
	return f
}

func (cv *Canvas) getFRContext(font *Font, size fixed.Int26_6) *frContext {
	k := fontKey{font: font, size: size}
	if frctx, ok := cv.fontCtxs[k]; ok {
		frctx.lastUsed = time.Now()
		return frctx.ctx
	}

	cv.reduceCache(Performance.CacheSize, 0)
	frctx := newFRContext()
	frctx.fontSize = k.size
	frctx.f = k.font.font
	frctx.recalc()

	cv.fontCtxs[k] = &frCache{ctx: frctx, lastUsed: time.Now()}

	return frctx
}

// FillText draws the given string at the given coordinates
// using the currently set font and font height
func (cv *Canvas) FillText(str string, x, y float64) {
	if cv.state.font.font == nil {
		return
	}

	scaleX := backendbase.Vec{cv.state.transform[0], cv.state.transform[1]}.Len()
	scaleY := backendbase.Vec{cv.state.transform[2], cv.state.transform[3]}.Len()
	scale := (scaleX + scaleY) * 0.5
	fontSize := fixed.Int26_6(math.Round(float64(cv.state.fontSize) * scale))

	frc := cv.getFRContext(cv.state.font, fontSize)
	fnt := cv.state.font.font

	// measure rendered text size
	var p fixed.Point26_6
	prev, hasPrev := truetype.Index(0), false
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
		var kern fixed.Int26_6
		if hasPrev {
			kern = fnt.Kern(fontSize, prev, idx)
			if frc.hinting != font.HintingNone {
				kern = (kern + 32) &^ 63
			}
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
		p.X += advance + kern
	}
	strWidth = p.X.Ceil() - textOffset.X
	strHeight := strMaxY - textOffset.Y

	if strWidth <= 0 || strHeight <= 0 {
		return
	}

	fstrWidth := float64(strWidth) / scale
	fstrHeight := float64(strHeight) / scale

	// calculate offsets
	if cv.state.textAlign == Center {
		x -= float64(fstrWidth) * 0.5
	} else if cv.state.textAlign == Right || cv.state.textAlign == End {
		x -= float64(fstrWidth)
	}
	metrics := cv.state.fontMetrics
	switch cv.state.textBaseline {
	case Alphabetic:
	case Middle:
		y += (-float64(metrics.Descent)/64 + float64(metrics.Height)*0.5/64) / scale
	case Top, Hanging:
		y += (-float64(metrics.Descent)/64 + float64(metrics.Height)/64) / scale
	case Bottom, Ideographic:
		y += -float64(metrics.Descent) / 64 / scale
	}

	// find out which characters are inside the visible area
	p = fixed.Point26_6{}
	prev, hasPrev = truetype.Index(0), false
	var insideCount int
	strFrom, strTo := 0, len(str)
	curInside := false
	curX := x
	for i, rn := range str {
		idx := fnt.Index(rn)
		if idx == 0 {
			idx = fnt.Index(' ')
		}
		advance, bounds, err := frc.glyphMeasure(idx, p)
		if err != nil {
			continue
		}
		var kern fixed.Int26_6
		if hasPrev {
			kern = fnt.Kern(fontSize, prev, idx)
			if frc.hinting != font.HintingNone {
				kern = (kern + 32) &^ 63
			}
			curX += float64(kern) / 64 / scale
		}

		w, h := cv.b.Size()
		fw, fh := float64(w), float64(h)

		p0 := cv.tf(backendbase.Vec{float64(bounds.Min.X)/scale + curX, float64(bounds.Min.Y)/scale + y})
		p1 := cv.tf(backendbase.Vec{float64(bounds.Min.X)/scale + curX, float64(bounds.Max.Y)/scale + y})
		p2 := cv.tf(backendbase.Vec{float64(bounds.Max.X)/scale + curX, float64(bounds.Max.Y)/scale + y})
		p3 := cv.tf(backendbase.Vec{float64(bounds.Max.X)/scale + curX, float64(bounds.Min.Y)/scale + y})
		inside := (p0[0] >= 0 || p1[0] >= 0 || p2[0] >= 0 || p3[0] >= 0) &&
			(p0[1] >= 0 || p1[1] >= 0 || p2[1] >= 0 || p3[1] >= 0) &&
			(p0[0] < fw || p1[0] < fw || p2[0] < fw || p3[0] < fw) &&
			(p0[1] < fh || p1[1] < fh || p2[1] < fh || p3[1] < fh)

		if inside {
			insideCount++
		}
		if !curInside && inside {
			curInside = true
			strFrom = i
			x = curX
		} else if curInside && !inside {
			strTo = i
			break
		}

		p.X += advance + kern
		curX += float64(advance) / 64 / scale
	}

	if strFrom == strTo || insideCount == 0 {
		return
	}

	// make sure textImage is large enough for the rendered string
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

	// clear the render region in textImage
	for y := 0; y < strHeight; y++ {
		off := textImage.PixOffset(0, y)
		line := textImage.Pix[off : off+strWidth]
		for i := range line {
			line[i] = 0
		}
	}

	// render the string into textImage
	prev, hasPrev = truetype.Index(0), false
	for _, rn := range str[strFrom:strTo] {
		idx := fnt.Index(rn)
		if idx == 0 {
			prev = 0
			hasPrev = false
			continue
		}
		if hasPrev {
			kern := fnt.Kern(fontSize, prev, idx)
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

	// render textImage to the screen
	var pts [4][2]float64
	pts[0] = cv.tf(backendbase.Vec{float64(textOffset.X)/scale + x, float64(textOffset.Y)/scale + y})
	pts[1] = cv.tf(backendbase.Vec{float64(textOffset.X)/scale + x, float64(textOffset.Y)/scale + fstrHeight + y})
	pts[2] = cv.tf(backendbase.Vec{float64(textOffset.X)/scale + fstrWidth + x, float64(textOffset.Y)/scale + fstrHeight + y})
	pts[3] = cv.tf(backendbase.Vec{float64(textOffset.X)/scale + fstrWidth + x, float64(textOffset.Y)/scale + y})

	mask := textImage.SubImage(image.Rect(0, 0, strWidth, strHeight)).(*image.Alpha)

	cv.drawShadow(pts[:], mask, false)

	stl := cv.backendFillStyle(&cv.state.fill, 1)
	cv.b.FillImageMask(&stl, mask, pts)
}

// StrokeText draws the given string at the given coordinates
// using the currently set font and font height and using the
// current stroke style
func (cv *Canvas) StrokeText(str string, x, y float64) {
	if cv.state.font == nil {
		return
	}

	frc := cv.getFRContext(cv.state.font, cv.state.fontSize)
	fnt := cv.state.font.font

	prevPath := cv.path
	cv.BeginPath()

	prev, hasPrev := truetype.Index(0), false
	for _, rn := range str {
		idx := fnt.Index(rn)
		if idx == 0 {
			idx = fnt.Index(' ')
		}

		if hasPrev {
			kern := fnt.Kern(cv.state.fontSize, prev, idx)
			if frc.hinting != font.HintingNone {
				kern = (kern + 32) &^ 63
			}
			x += float64(kern) / 64
		}
		advance, _, err := frc.glyphMeasure(idx, fixed.Point26_6{})
		if err != nil {
			continue
		}

		cv.runePath(rn, backendbase.Vec{x, y})

		x += float64(advance) / 64
	}

	cv.Stroke()
	cv.path = prevPath
}

func (cv *Canvas) runePath(rn rune, pos backendbase.Vec) {
	gb := &truetype.GlyphBuf{}
	gb.Load(cv.state.font.font, cv.state.fontSize, cv.state.font.font.Index(rn), font.HintingNone)

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

	frc := cv.getFRContext(cv.state.font, cv.state.fontSize)
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
		var kern fixed.Int26_6
		if hasPrev {
			kern = fnt.Kern(frc.fontSize, prev, idx)
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
		p.X += advance + kern
	}

	return TextMetrics{
		Width:                    x,
		ActualBoundingBoxAscent:  -minY,
		ActualBoundingBoxDescent: +maxY,
	}
}
