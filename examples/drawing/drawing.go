package main

import (
	"image/color"
	"log"
	"math"

	"github.com/tfriedel6/canvas/sdlcanvas"
)

func main() {
	wnd, cv, err := sdlcanvas.CreateWindow(1280, 720, "Canvas Example")
	if err != nil {
		log.Println(err)
		return
	}
	defer wnd.Destroy()

	lg := cv.CreateLinearGradient(320, 200, 480, 520)
	lg.AddColorStop(0, "#ff000040")
	lg.AddColorStop(1, "#00ff0040")
	lg.AddColorStop(0.5, "#0000ff40")

	rg := cv.CreateRadialGradient(540, 300, 80, 740, 300, 100)
	rg.AddColorStop(0, "#ff0000")
	rg.AddColorStop(1, "#00ff00")
	rg.AddColorStop(0.5, "#0000ff")

	wnd.MainLoop(func() {
		w, h := float64(cv.Width()), float64(cv.Height())

		// Clear the screen
		cv.SetFillStyle("#000")
		cv.FillRect(0, 0, w, h)

		// Estimated size used for scaling
		const (
			contentWidth  = 1000
			contentHeight = 350
		)

		// Calculate scaling
		sx := w / contentWidth
		sy := h / contentHeight
		scale := math.Min(sx, sy)
		cv.Save()
		defer cv.Restore()
		cv.Scale(scale, scale)

		// Draw lines with different colors and line thickness
		for x := 1.0; x < 10.5; x += 1.0 {
			cv.SetStrokeStyle(int(x*25), 255, 255)
			cv.SetLineWidth(x)
			cv.BeginPath()
			cv.MoveTo(x*10+20, 20)
			cv.LineTo(x*10+20, 120)
			cv.Stroke()
		}

		// Draw a path
		cv.BeginPath()
		cv.MoveTo(160, 20)
		cv.LineTo(180, 20)
		cv.LineTo(180, 40)
		cv.LineTo(200, 40)
		cv.LineTo(200, 60)
		cv.LineTo(220, 60)
		cv.LineTo(220, 80)
		cv.LineTo(240, 80)
		cv.LineTo(240, 100)
		cv.LineTo(260, 100)
		cv.LineTo(260, 120)
		cv.ArcTo(160, 120, 160, 100, 20)
		cv.ClosePath()
		cv.SetStrokeStyle(color.RGBA{R: 255, G: 128, B: 128, A: 255})
		cv.SetLineWidth(4)
		cv.Stroke()

		// Fill a polygon
		cv.BeginPath()
		cv.MoveTo(300, 20)
		cv.LineTo(340, 20)
		cv.QuadraticCurveTo(370, 20, 370, 50)
		cv.QuadraticCurveTo(370, 80, 400, 80)
		cv.LineTo(400, 80)
		cv.LineTo(400, 120)
		cv.LineTo(360, 120)
		cv.BezierCurveTo(330, 120, 330, 80, 300, 80)
		cv.ClosePath()
		cv.SetFillStyle(color.RGBA{R: 128, G: 255, B: 128, A: 255})
		cv.Fill()

		// Draw with alpha
		cv.SetGlobalAlpha(0.5)
		cv.SetFillStyle("#FF0000")
		cv.BeginPath()
		cv.Arc(100, 275, 60, 0, math.Pi*2, false)
		cv.Fill()
		cv.SetFillStyle("#00FF00")
		cv.BeginPath()
		cv.Arc(140, 210, 60, 0, math.Pi*2, false)
		cv.Fill()
		cv.SetFillStyle("#0000FF")
		cv.BeginPath()
		cv.Arc(180, 275, 60, 0, math.Pi*2, false)
		cv.Fill()
		cv.SetGlobalAlpha(1)

		// Clipped drawing
		cv.Save()
		cv.BeginPath()
		cv.Arc(340, 240, 80, 0, math.Pi*2, true)
		cv.Clip()
		cv.SetStrokeStyle(0, 255, 0)
		for x := 1.0; x < 12.5; x += 1.0 {
			cv.BeginPath()
			cv.MoveTo(260, 140+16*x)
			cv.LineTo(420, 140+16*x)
			cv.Stroke()
		}
		cv.SetFillStyle(0, 0, 255)
		for x := 1.0; x < 12.5; x += 1.0 {
			cv.FillRect(246+x*14, 150, 6, 180)
		}
		cv.Restore()

		// Draw images
		cv.DrawImage("cat.jpg", 480, 40, 320, 265)

		// Draw text
		cv.SetFont("Righteous-Regular.ttf", 40)
		cv.FillText("<-- Cat", 820, 180)
	})
}
