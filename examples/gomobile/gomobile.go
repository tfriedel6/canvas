package main

import (
	"log"
	"math"
	"time"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/backend/xmobilebackend"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/gl"
)

func main() {
	app.Main(func(a app.App) {
		var cv, painter *canvas.Canvas
		var cvb *xmobilebackend.XMobileBackendOffscreen
		var painterb *xmobilebackend.XMobileBackend
		var w, h int

		var glctx gl.Context
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					var err error
					glctx = e.DrawContext.(gl.Context)
					ctx, err := xmobilebackend.NewGLContext(glctx)
					if err != nil {
						log.Fatal(err)
					}
					cvb, err = xmobilebackend.NewOffscreen(0, 0, false, ctx)
					if err != nil {
						log.Fatalln(err)
					}
					painterb, err = xmobilebackend.New(0, 0, 0, 0, ctx)
					if err != nil {
						log.Fatalln(err)
					}
					cv = canvas.New(cvb)
					painter = canvas.New(painterb)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					cvb.Delete()
					glctx = nil
				}
			case size.Event:
				w, h = e.WidthPx, e.HeightPx
			case paint.Event:
				if glctx != nil {
					fw, fh := float64(w), float64(h)
					color := math.Sin(float64(time.Now().UnixNano())*0.000000002)*0.3 + 0.7

					cvb.SetSize(w, h)
					cv.SetFillStyle(color*0.2, color*0.2, color*0.8)
					cv.FillRect(fw*0.25, fh*0.25, fw*0.5, fh*0.5)

					painterb.SetBounds(0, 0, w, h)
					painter.DrawImage(cv)

					a.Publish()
					a.Send(paint.Event{})
				}
			}
		}
	})
}
