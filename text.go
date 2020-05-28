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
	"unsafe"

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

type fontPathCache struct {
	cache    map[truetype.Index]*Path2D
	lastUsed time.Time
}

func (fpc *fontPathCache) size() int {
	size := 0
	pps := int(unsafe.Sizeof(pathPoint{}))
	for _, p := range fpc.cache {
		size += len(p.p) * pps
	}
	return size
}

type fontTriCache struct {
	cache    map[truetype.Index][]backendbase.Vec
	lastUsed time.Time
}

func (ftc *fontTriCache) size() int {
	size := 0
	for _, p := range ftc.cache {
		size += len(p) * 16
	}
	return size
}

var zeroes [alphaTexSize]byte
var textImage *image.Alpha

var defaultFont *Font

var baseFontSize = fixed.I(42)

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
	frctx.fontSize = size
	frctx.f = font.font
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

	// if the font size is large or rotated or skewed in some way, use the
	// triangulated font rendering
	if fontSize > fixed.I(25) {
		cv.fillText2(str, x, y)
		return
	}
	mat := cv.state.transform
	if mat[1] != 0 || mat[2] != 0 || mat[0] != mat[3] {
		cv.fillText2(str, x, y)
		return
	}

	frc := cv.getFRContext(cv.state.font, fontSize)
	fnt := cv.state.font.font

	strWidth, strHeight, textOffset, str := cv.measureTextRendering(str, &x, &y, frc, scale)
	if strWidth <= 0 || strHeight <= 0 {
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

	// clear the render region in textImage
	for y := 0; y < strHeight; y++ {
		off := textImage.PixOffset(0, y)
		line := textImage.Pix[off : off+strWidth]
		for i := range line {
			line[i] = 0
		}
	}

	// render the string into textImage
	curX := x
	p := fixed.Point26_6{}
	prev, hasPrev := truetype.Index(0), false
	for _, rn := range str {
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
	var pts [4]backendbase.Vec
	pts[0] = cv.tf(backendbase.Vec{float64(textOffset.X)/scale + x, float64(textOffset.Y)/scale + y})
	pts[1] = cv.tf(backendbase.Vec{float64(textOffset.X)/scale + x, float64(textOffset.Y)/scale + float64(strHeight)/scale + y})
	pts[2] = cv.tf(backendbase.Vec{float64(textOffset.X)/scale + float64(strWidth)/scale + x, float64(textOffset.Y)/scale + float64(strHeight)/scale + y})
	pts[3] = cv.tf(backendbase.Vec{float64(textOffset.X)/scale + float64(strWidth)/scale + x, float64(textOffset.Y)/scale + y})

	mask := textImage.SubImage(image.Rect(0, 0, strWidth, strHeight)).(*image.Alpha)

	cv.drawShadow(pts[:], mask, false)

	stl := cv.backendFillStyle(&cv.state.fill, 1)
	cv.b.FillImageMask(&stl, mask, pts)
}

func (cv *Canvas) fillText2(str string, x, y float64) {
	if cv.state.font == nil {
		return
	}

	frc := cv.getFRContext(cv.state.font, cv.state.fontSize)
	fnt := cv.state.font.font

	strWidth, strHeight, _, str := cv.measureTextRendering(str, &x, &y, frc, 1)
	if strWidth <= 0 || strHeight <= 0 {
		return
	}

	scale := float64(cv.state.fontSize) / float64(baseFontSize)
	scaleMat := backendbase.MatScale(backendbase.Vec{scale, scale})

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

		tris := cv.runeTris(rn)
		tf := scaleMat.Mul(backendbase.MatTranslate(backendbase.Vec{x, y})).Mul(cv.state.transform)
		cv.drawShadow(tris, nil, false)
		stl := cv.backendFillStyle(&cv.state.fill, 1)
		cv.b.Fill(&stl, tris, tf, false)

		x += float64(advance) / 64
	}

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

	strWidth, strHeight, _, str := cv.measureTextRendering(str, &x, &y, frc, 1)
	if strWidth <= 0 || strHeight <= 0 {
		return
	}

	scale := float64(cv.state.fontSize) / float64(baseFontSize)
	scaleMat := backendbase.MatScale(backendbase.Vec{scale, scale})

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

		path := cv.runePath(rn)
		tf := scaleMat.Mul(backendbase.MatTranslate(backendbase.Vec{x, y})).Mul(cv.state.transform)
		cv.strokePath(path, tf, backendbase.Mat{}, false)

		x += float64(advance) / 64
	}

}

func (cv *Canvas) measureTextRendering(str string, x, y *float64, frc *frContext, scale float64) (int, int, image.Point, string) {
	// measure rendered text size
	var p fixed.Point26_6
	prev, hasPrev := truetype.Index(0), false
	var textOffset image.Point
	var strWidth, strMaxY int
	strMinY := math.MaxInt32
	for i, rn := range str {
		idx := frc.f.Index(rn)
		if idx == 0 {
			idx = frc.f.Index(' ')
		}
		advance, bounds, err := frc.glyphMeasure(idx, p)
		if err != nil {
			continue
		}
		var kern fixed.Int26_6
		if hasPrev {
			kern = frc.f.Kern(frc.fontSize, prev, idx)
			if frc.hinting != font.HintingNone {
				kern = (kern + 32) &^ 63
			}
		}

		if i == 0 {
			textOffset.X = bounds.Min.X
		}
		if bounds.Max.Y > strMaxY {
			strMaxY = bounds.Max.Y
		}
		if bounds.Min.Y < strMinY {
			strMinY = bounds.Min.Y
		}
		p.X += advance + kern
	}
	textOffset.Y = strMinY
	strWidth = p.X.Ceil() - textOffset.X
	strHeight := strMaxY - textOffset.Y

	if strWidth <= 0 || strHeight <= 0 {
		return 0, 0, image.Point{}, ""
	}

	// calculate offsets
	if cv.state.textAlign == Center {
		*x -= float64(strWidth) / scale * 0.5
	} else if cv.state.textAlign == Right || cv.state.textAlign == End {
		*x -= float64(strWidth) / scale
	}
	metrics := cv.state.fontMetrics
	switch cv.state.textBaseline {
	case Alphabetic:
	case Middle:
		*y += -float64(metrics.Descent)/64 + float64(metrics.Height)*0.5/64
	case Top, Hanging:
		*y += -float64(metrics.Descent)/64 + float64(metrics.Height)/64
	case Bottom, Ideographic:
		*y += -float64(metrics.Descent) / 64
	}

	// find out which characters are inside the visible area
	p = fixed.Point26_6{}
	prev, hasPrev = truetype.Index(0), false
	var insideCount int
	strFrom, strTo := 0, len(str)
	curInside := false
	curX := *x
	for i, rn := range str {
		idx := frc.f.Index(rn)
		if idx == 0 {
			idx = frc.f.Index(' ')
		}
		advance, bounds, err := frc.glyphMeasure(idx, p)
		if err != nil {
			continue
		}
		var kern fixed.Int26_6
		if hasPrev {
			kern = frc.f.Kern(frc.fontSize, prev, idx)
			if frc.hinting != font.HintingNone {
				kern = (kern + 32) &^ 63
			}
			curX += float64(kern) / 64 / scale
		}

		w, h := cv.b.Size()
		fw, fh := float64(w), float64(h)

		p0 := cv.tf(backendbase.Vec{float64(bounds.Min.X)/scale + curX, float64(strMinY)/scale + *y})
		p1 := cv.tf(backendbase.Vec{float64(bounds.Min.X)/scale + curX, float64(strMaxY)/scale + *y})
		p2 := cv.tf(backendbase.Vec{float64(bounds.Max.X)/scale + curX, float64(strMaxY)/scale + *y})
		p3 := cv.tf(backendbase.Vec{float64(bounds.Max.X)/scale + curX, float64(strMinY)/scale + *y})
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
			*x = curX
		} else if curInside && !inside {
			strTo = i
			break
		}

		p.X += advance + kern
		curX += float64(advance) / 64 / scale
	}

	if strFrom == strTo || insideCount == 0 {
		return 0, 0, image.Point{}, ""
	}

	// if necessary, measure rendered text size again with the substring
	if strFrom > 0 || strTo < len(str) {
		str = str[strFrom:strTo]
		p = fixed.Point26_6{}
		prev, hasPrev = truetype.Index(0), false
		textOffset = image.Point{}
		strWidth, strMaxY = 0, 0
		for i, rn := range str {
			idx := frc.f.Index(rn)
			if idx == 0 {
				idx = frc.f.Index(' ')
			}
			advance, bounds, err := frc.glyphMeasure(idx, p)
			if err != nil {
				continue
			}
			var kern fixed.Int26_6
			if hasPrev {
				kern = frc.f.Kern(frc.fontSize, prev, idx)
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
		strHeight = strMaxY - textOffset.Y

		if strWidth <= 0 || strHeight <= 0 {
			return 0, 0, image.Point{}, ""
		}
	}

	return strWidth, strHeight, textOffset, str
}

func (cv *Canvas) runePath(rn rune) *Path2D {
	idx := cv.state.font.font.Index(rn)
	if idx == 0 {
		idx = cv.state.font.font.Index(' ')
	}

	if cache, ok := cv.fontPathCache[cv.state.font]; ok {
		if path, ok := cache.cache[idx]; ok {
			cache.lastUsed = time.Now()
			return path
		}
	}

	path := &Path2D{cv: cv, p: make([]pathPoint, 0, 50), standalone: true, noSelfIntersection: true}

	const scale = 1.0 / 64.0

	var gb truetype.GlyphBuf
	gb.Load(cv.state.font.font, baseFontSize, idx, font.HintingFull)

	from := 0
	for _, to := range gb.Ends {
		ps := gb.Points[from:to]

		start := ps[0]
		others := []truetype.Point(nil)
		if ps[0].Flags&0x01 != 0 {
			others = ps[1:]
		} else {
			last := ps[len(ps)-1]
			if ps[len(ps)-1].Flags&0x01 != 0 {
				start = last
				others = ps[:len(ps)-1]
			} else {
				start = truetype.Point{
					X: (start.X + last.X) / 2,
					Y: (start.Y + last.Y) / 2,
				}
				others = ps
			}
		}

		p0, on0 := start, true
		path.MoveTo(float64(p0.X)*scale, -float64(p0.Y)*scale)
		for _, p := range others {
			on := p.Flags&0x01 != 0
			if on {
				if on0 {
					path.LineTo(float64(p.X)*scale, -float64(p.Y)*scale)
				} else {
					path.QuadraticCurveTo(float64(p0.X)*scale, -float64(p0.Y)*scale, float64(p.X)*scale, -float64(p.Y)*scale)
				}
			} else {
				if on0 {
					// No-op.
				} else {
					mid := fixed.Point26_6{
						X: (p0.X + p.X) / 2,
						Y: (p0.Y + p.Y) / 2,
					}
					path.QuadraticCurveTo(float64(p0.X)*scale, -float64(p0.Y)*scale, float64(mid.X)*scale, -float64(mid.Y)*scale)
				}
			}
			p0, on0 = p, on
		}

		if on0 {
			path.LineTo(float64(start.X)*scale, -float64(start.Y)*scale)
		} else {
			path.QuadraticCurveTo(float64(p0.X)*scale, -float64(p0.Y)*scale, float64(start.X)*scale, -float64(start.Y)*scale)
		}
		path.ClosePath()

		from = to
	}

	cache, ok := cv.fontPathCache[cv.state.font]
	if !ok {
		cache = &fontPathCache{cache: make(map[truetype.Index]*Path2D, 1024)}
		cv.fontPathCache[cv.state.font] = cache
	}
	cache.lastUsed = time.Now()
	cache.cache[idx] = path

	return path
}

func (cv *Canvas) runeTris(rn rune) []backendbase.Vec {
	idx := cv.state.font.font.Index(rn)
	if idx == 0 {
		idx = cv.state.font.font.Index(' ')
	}

	if cache, ok := cv.fontTriCache[cv.state.font]; ok {
		if tris, ok := cache.cache[idx]; ok {
			cache.lastUsed = time.Now()
			return tris
		}
	}

	const scale = 1.0 / 64.0

	var gb truetype.GlyphBuf
	gb.Load(cv.state.font.font, baseFontSize, idx, font.HintingFull)

	contours := make([][]backendbase.Vec, 0, len(gb.Ends))

	from := 0
	for _, to := range gb.Ends {
		ps := gb.Points[from:to]

		start := ps[0]
		others := []truetype.Point(nil)
		if ps[0].Flags&0x01 != 0 {
			others = ps[1:]
		} else {
			last := ps[len(ps)-1]
			if last.Flags&0x01 != 0 {
				start = last
				others = ps[:len(ps)-1]
			} else {
				start = truetype.Point{
					X: (start.X + last.X) / 2,
					Y: (start.Y + last.Y) / 2,
				}
				others = ps
			}
		}

		p0, on0 := start, true
		path := &Path2D{cv: cv, p: make([]pathPoint, 0, 50), standalone: true, noSelfIntersection: true}
		path.MoveTo(float64(p0.X)*scale, -float64(p0.Y)*scale)
		for _, p := range others {
			on := p.Flags&0x01 != 0
			if on {
				if on0 {
					path.LineTo(float64(p.X)*scale, -float64(p.Y)*scale)
				} else {
					path.QuadraticCurveTo(float64(p0.X)*scale, -float64(p0.Y)*scale, float64(p.X)*scale, -float64(p.Y)*scale)
				}
			} else {
				if on0 {
					// No-op.
				} else {
					mid := fixed.Point26_6{
						X: (p0.X + p.X) / 2,
						Y: (p0.Y + p.Y) / 2,
					}
					path.QuadraticCurveTo(float64(p0.X)*scale, -float64(p0.Y)*scale, float64(mid.X)*scale, -float64(mid.Y)*scale)
				}
			}
			p0, on0 = p, on
		}

		if on0 {
			path.LineTo(float64(start.X)*scale, -float64(start.Y)*scale)
		} else {
			path.QuadraticCurveTo(float64(p0.X)*scale, -float64(p0.Y)*scale, float64(start.X)*scale, -float64(start.Y)*scale)
		}
		path.ClosePath()

		contour := make([]backendbase.Vec, len(path.p))
		for i, pt := range path.p {
			contour[i] = pt.pos
		}

		contours = append(contours, contour)

		from = to
	}

	idxs := sortFontContours(contours)
	sortedContours := make([][]backendbase.Vec, 0, len(idxs))
	trisList := make([][]backendbase.Vec, 0, len(contours))

	for i := 0; i < len(idxs); {
		var j int
		for j = i; j < len(idxs); j++ {
			if idxs[j] == -1 {
				break
			}
		}

		sortedContours = sortedContours[:j-i]
		for k, idx := range idxs[i:j] {
			sortedContours[k] = contours[idx]
		}

		var ec earcut
		ec.run(sortedContours)

		tris := make([]backendbase.Vec, len(ec.indices))
		for i, idx := range ec.indices {
			pidx := 0
			poly := sortedContours[pidx]
			for idx >= len(poly) {
				idx -= len(poly)
				pidx++
				poly = sortedContours[pidx]
			}
			tris[i] = poly[idx]
		}
		trisList = append(trisList, tris)

		i = j + 1
	}

	count := 0
	for _, tris := range trisList {
		count += len(tris)
	}

	allTris := make([]backendbase.Vec, count)
	pos := 0
	for _, tris := range trisList {
		copy(allTris[pos:], tris)
		pos += len(tris)
	}

	cache, ok := cv.fontTriCache[cv.state.font]
	if !ok {
		cache = &fontTriCache{cache: make(map[truetype.Index][]backendbase.Vec, 1024)}
		cv.fontTriCache[cv.state.font] = cache
	}
	cache.lastUsed = time.Now()
	cache.cache[idx] = allTris

	return allTris
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
