package main

import (
	"log"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/backend/xmobilebackend"
	"golang.org/x/exp/shiny/driver/gldriver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/exp/shiny/widget"
	"golang.org/x/exp/shiny/widget/glwidget"
	"golang.org/x/exp/shiny/widget/node"
)

var cv *canvas.Canvas
var sheet *widget.Sheet

func main() {
	gldriver.Main(func(s screen.Screen) {
		glw := glwidget.NewGL(draw)
		sheet = widget.NewSheet(glw)
		ctx, err := xmobilebackend.NewGLContext(glw.Ctx)
		if err != nil {
			log.Fatal(err)
		}
		backend, err := xmobilebackend.New(0, 0, 600, 600, ctx)
		if err != nil {
			log.Fatal(err)
		}
		cv = canvas.New(backend)

		err = widget.RunWindow(s, sheet, &widget.RunWindowOptions{
			NewWindowOptions: screen.NewWindowOptions{
				Title:  "Shiny Canvas Example",
				Width:  600,
				Height: 600,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	})
}

func draw(w *glwidget.GL) {
	cv.Save()
	defer cv.Restore()

	cv.Translate(0, 600)
	cv.Scale(1, -1)

	cv.ClearRect(0, 0, 600, 600)
	cv.SetFillStyle("#FF00FF")
	cv.FillRect(100, 100, 200, 200)

	w.Publish()
	w.Mark(node.MarkNeedsPaintBase)
}
