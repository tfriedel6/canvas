package canvas

import (
	"unsafe"

	"github.com/tfriedel6/lm"
)

// Canvas represents an area on the viewport on which to draw
// using a set of functions very similar to the HTML5 canvas
type Canvas struct {
	x, y, w, h     int
	fx, fy, fw, fh float32

	fill struct {
		r, g, b, a float32
	}
	stroke struct {
		r, g, b, a float32
		lineWidth  float32
	}
	path []pathPoint
}

type pathPoint struct {
	pos  lm.Vec2
	move bool
}

// New creates a new canvas with the given viewport coordinates.
// While all functions on the canvas use the top left point as
// the origin, since GL uses the bottom left coordinate, the
// coordinates given here also use the bottom left as origin
func New(x, y, w, h int) *Canvas {
	cv := &Canvas{
		x: x, y: y, w: w, h: h,
		fx: float32(x), fy: float32(y),
		fw: float32(w), fh: float32(h),
	}
	cv.stroke.lineWidth = 1
	return cv
}

func (cv *Canvas) xToGL(x float32) float32              { return x*2/cv.fw - 1 }
func (cv *Canvas) yToGL(y float32) float32              { return -y*2/cv.fh + 1 }
func (cv *Canvas) ptToGL(x, y float32) (fx, fy float32) { return x*2/cv.fw - 1, -y*2/cv.fh + 1 }
func (cv *Canvas) vecToGL(v lm.Vec2) (fx, fy float32)   { return v[0]*2/cv.fw - 1, -v[1]*2/cv.fh + 1 }

// Activate makes the canvas active and sets the viewport. Only needs
// to be called if any other GL code changes the viewport
func (cv *Canvas) Activate() {
	gli.Viewport(int32(cv.x), int32(cv.y), int32(cv.w), int32(cv.h))
}

var activeCanvas *Canvas

func (cv *Canvas) activate() {
	if activeCanvas != cv {
		activeCanvas = cv
		cv.Activate()
	}
}

// SetFillColor sets the color for any fill calls
func (cv *Canvas) SetFillColor(value ...interface{}) {
	r, g, b, a, ok := parseColor(value...)
	if ok {
		cv.fill.r, cv.fill.g, cv.fill.b, cv.fill.a = r, g, b, a
	}
}

// SetStrokeColor sets the color for any line drawing calls
func (cv *Canvas) SetStrokeColor(value ...interface{}) {
	r, g, b, a, ok := parseColor(value...)
	if ok {
		cv.stroke.r, cv.stroke.g, cv.stroke.b, cv.stroke.a = r, g, b, a
	}
}

// SetLineWidth sets the line width for any line drawing calls
func (cv *Canvas) SetLineWidth(width float32) {
	cv.stroke.lineWidth = width
}

// FillRect fills a rectangle with the active color
func (cv *Canvas) FillRect(x, y, w, h float32) {
	cv.activate()

	gli.UseProgram(sr.id)

	x0f, y0f := cv.ptToGL(x, y)
	x1f, y1f := cv.ptToGL(x+w, y+h)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [8]float32{x0f, y0f, x0f, y1f, x1f, y1f, x1f, y0f}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.Uniform4f(sr.color, cv.fill.r, cv.fill.g, cv.fill.b, cv.fill.a)
	gli.EnableVertexAttribArray(sr.vertex)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(sr.vertex)
}

func (cv *Canvas) BeginPath() {
	if cv.path == nil {
		cv.path = make([]pathPoint, 0, 100)
	}
	cv.path = cv.path[:0]
}

func (cv *Canvas) MoveTo(x, y float32) {
	cv.path = append(cv.path, pathPoint{pos: lm.Vec2{x, y}, move: true})
}

func (cv *Canvas) LineTo(x, y float32) {
	cv.path = append(cv.path, pathPoint{pos: lm.Vec2{x, y}, move: false})
}

func (cv *Canvas) ClosePath() {
	if len(cv.path) == 0 {
		return
	}
	cv.path = append(cv.path, pathPoint{pos: cv.path[0].pos, move: false})
}

func (cv *Canvas) Stroke() {
	if len(cv.path) == 0 {
		return
	}

	cv.activate()

	gli.Enable(gl_STENCIL_TEST)
	gli.ColorMask(false, false, false, false)
	gli.StencilFunc(gl_ALWAYS, 1, 0xFF)
	gli.StencilOp(gl_KEEP, gl_KEEP, gl_REPLACE)
	gli.StencilMask(0xFF)
	gli.Clear(gl_STENCIL_BUFFER_BIT)

	gli.UseProgram(sr.id)
	gli.Uniform4f(sr.color, cv.stroke.r, cv.stroke.g, cv.stroke.b, cv.stroke.a)
	gli.EnableVertexAttribArray(sr.vertex)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	var buf [1000]float32
	tris := buf[:0]
	tris = append(tris, -1, -1, -1, 1, 1, 1, -1, -1, 1, 1, 1, -1)
	p0 := cv.path[0].pos
	for _, p := range cv.path {
		if p.move {
			p0 = p.pos
			continue
		}
		p1 := p.pos

		v1 := p1.Sub(p0).Norm()
		v2 := lm.Vec2{v1[1], -v1[0]}.MulF(cv.stroke.lineWidth * 0.5)
		v1 = v1.MulF(cv.stroke.lineWidth * 0.5)

		x0f, y0f := cv.vecToGL(p0.Sub(v1).Add(v2))
		x1f, y1f := cv.vecToGL(p1.Add(v1).Add(v2))
		x2f, y2f := cv.vecToGL(p1.Add(v1).Sub(v2))
		x3f, y3f := cv.vecToGL(p0.Sub(v1).Sub(v2))

		tris = append(tris, x0f, y0f, x1f, y1f, x2f, y2f, x0f, y0f, x2f, y2f, x3f, y3f)

		p0 = p1
	}

	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)
	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))

	gli.ColorMask(true, true, true, true)
	gli.StencilFunc(gl_EQUAL, 1, 0xFF)
	gli.StencilMask(0)

	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 6)

	gli.DisableVertexAttribArray(sr.vertex)

	gli.Disable(gl_STENCIL_TEST)
}

func (cv *Canvas) Fill() {
	if len(cv.path) < 3 {
		return
	}

	cv.activate()

	gli.UseProgram(sr.id)
	gli.Uniform4f(sr.color, cv.fill.r, cv.fill.g, cv.fill.b, cv.fill.a)
	gli.EnableVertexAttribArray(sr.vertex)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)

	var buf [1000]float32
	tris := buf[:0]
	tris = append(tris, -1, -1, -1, 1, 1, 1, -1, -1, 1, 1, 1, -1)

	tris = triangulatePath(cv.path, tris)
	total := len(tris)
	for i := 12; i < total; i += 2 {
		x, y := tris[i], tris[i+1]
		tris[i], tris[i+1] = cv.ptToGL(x, y)
	}

	gli.BufferData(gl_ARRAY_BUFFER, len(tris)*4, unsafe.Pointer(&tris[0]), gl_STREAM_DRAW)
	gli.VertexAttribPointer(sr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.DrawArrays(gl_TRIANGLES, 6, int32(len(tris)/2-6))

	gli.DisableVertexAttribArray(sr.vertex)
}

func (cv *Canvas) DrawImage(img *Image, coords ...float32) {
	var sx, sy, sw, sh, dx, dy, dw, dh float32
	sw, sh = float32(img.w), float32(img.h)
	dw, dh = float32(img.w), float32(img.h)
	if len(coords) == 2 {
		dx, dy = coords[0], coords[1]
	} else if len(coords) == 4 {
		dx, dy = coords[0], coords[1]
		dw, dh = coords[2], coords[3]
	} else if len(coords) == 8 {
		sx, sy = coords[0], coords[1]
		sw, sh = coords[2], coords[3]
		dx, dy = coords[4], coords[5]
		dw, dh = coords[6], coords[7]
	}

	dx0, dy0 := cv.ptToGL(dx, dy)
	dx1, dy1 := cv.ptToGL(dx+dw, dy+dh)
	sx /= float32(img.w)
	sy /= float32(img.h)
	sw /= float32(img.w)
	sh /= float32(img.h)

	cv.activate()

	gli.UseProgram(tr.id)

	gli.ActiveTexture(gl_TEXTURE0)
	gli.BindTexture(gl_TEXTURE_2D, img.tex)
	gli.Uniform1i(tr.image, 0)

	gli.BindBuffer(gl_ARRAY_BUFFER, buf)
	data := [16]float32{dx0, dy0, dx0, dy1, dx1, dy1, dx1, dy0,
		sx, sy, sx, sy + sh, sx + sw, sy + sh, sx + sw, sy}
	gli.BufferData(gl_ARRAY_BUFFER, len(data)*4, unsafe.Pointer(&data[0]), gl_STREAM_DRAW)

	gli.VertexAttribPointer(tr.vertex, 2, gl_FLOAT, false, 0, nil)
	gli.VertexAttribPointer(tr.texCoord, 2, gl_FLOAT, false, 0, gli.PtrOffset(8*4))
	gli.EnableVertexAttribArray(tr.vertex)
	gli.EnableVertexAttribArray(tr.texCoord)
	gli.DrawArrays(gl_TRIANGLE_FAN, 0, 4)
	gli.DisableVertexAttribArray(tr.vertex)
	gli.DisableVertexAttribArray(tr.texCoord)
}
