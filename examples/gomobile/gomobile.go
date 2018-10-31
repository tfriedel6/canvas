package main

import (
	"math"
	"time"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/glimpl/xmobile"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/gl"
)

func main() {
	app.Main(func(a app.App) {
		var cv, painter *canvas.Canvas
		var w, h int

		var glctx gl.Context
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					canvas.LoadGL(glimplxmobile.New(glctx))
					cv = canvas.NewOffscreen(0, 0)
					painter = canvas.New(0, 0, 0, 0)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					glctx = nil
				}
			case size.Event:
				w, h = e.WidthPx, e.HeightPx
			case paint.Event:
				if glctx != nil {
					glctx.ClearColor(0, 0, 0, 0)
					glctx.Clear(gl.COLOR_BUFFER_BIT)

					cv.SetBounds(0, 0, w, h)
					painter.SetBounds(0, 0, w, h)

					fw, fh := float64(w), float64(h)
					color := math.Sin(float64(time.Now().UnixNano())*0.000000002)*0.3 + 0.7

					cv.SetFillStyle(color*0.2, color*0.2, color*0.8)
					cv.FillRect(fw*0.25, fh*0.25, fw*0.5, fh*0.5)

					painter.DrawImage(cv)

					a.Publish()
					a.Send(paint.Event{})
				}
			}
		}
	})
}
