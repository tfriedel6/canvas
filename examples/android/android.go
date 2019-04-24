package canvasandroidexample

import (
	"math"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/backend/goglbackend"
)

var cv *canvas.Canvas
var mx, my float64

func TouchEvent(typ string, x, y int) {
	mx, my = float64(x), float64(y)
}

func OnSurfaceCreated() {
}

func OnSurfaceChanged(w, h int) {
	backend, err := goglbackend.New(0, 0, w, h, nil)
	if err != nil {
		panic(err)
	}
	cv = canvas.New(backend)
}

func OnDrawFrame() {
	if cv == nil {
		return
	}
	w, h := float64(cv.Width()), float64(cv.Height())

	cv.SetFillStyle("#000")
	cv.FillRect(0, 0, w, h)
	cv.SetFillStyle("#0F0")
	cv.FillRect(w*0.25, h*0.25, w*0.5, h*0.5)
	cv.SetLineWidth(6)
	sqrSize := math.Min(w, h) * 0.1
	cv.SetStrokeStyle("#F00")
	cv.StrokeRect(mx-sqrSize/2, my-sqrSize/2, sqrSize, sqrSize)
}
