package canvas

import "github.com/tfriedel6/canvas/backend/backendbase"

type Backend interface {
	ClearRect(x, y, w, h int)
	Clear(pts [4][2]float64)
	Fill(style *backendbase.Style, pts [][2]float64)
	// BlurredShadow(shadow *backendbase.Shadow, pts [][2]float64)
}
