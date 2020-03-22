package canvas

import (
	"image"
	"math"

	"github.com/tfriedel6/canvas/backend/backendbase"
)

func (cv *Canvas) drawShadow(pts []backendbase.Vec, mask *image.Alpha, canOverlap bool) {
	if cv.state.shadowColor.A == 0 {
		return
	}
	if cv.state.shadowOffsetX == 0 && cv.state.shadowOffsetY == 0 {
		return
	}

	if cv.shadowBuf == nil || cap(cv.shadowBuf) < len(pts) {
		cv.shadowBuf = make([]backendbase.Vec, 0, len(pts)+1000)
	}
	cv.shadowBuf = cv.shadowBuf[:0]

	for _, pt := range pts {
		cv.shadowBuf = append(cv.shadowBuf, backendbase.Vec{
			pt[0] + cv.state.shadowOffsetX,
			pt[1] + cv.state.shadowOffsetY,
		})
	}

	color := cv.state.shadowColor
	color.A = uint8(math.Round(((float64(color.A) / 255.0) * cv.state.globalAlpha) * 255.0))
	style := backendbase.FillStyle{Color: color, Blur: cv.state.shadowBlur}
	if mask != nil {
		if len(cv.shadowBuf) != 4 {
			panic("invalid number of points to fill with mask, must be 4")
		}
		var quad [4]backendbase.Vec
		copy(quad[:], cv.shadowBuf)
		cv.b.FillImageMask(&style, mask, quad)
	} else {
		cv.b.Fill(&style, cv.shadowBuf, backendbase.MatIdentity, canOverlap)
	}
}
