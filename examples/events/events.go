package main

import (
	"log"
	"math"
	"time"

	"github.com/tfriedel6/canvas/sdlcanvas"
)

type circle struct {
	x, y  float64
	color string
}

func main() {
	wnd, cv, err := sdlcanvas.CreateWindow(1280, 720, "Canvas Example")
	if err != nil {
		log.Println(err)
		return
	}
	defer wnd.Destroy()

	var mx, my, action float64
	circles := make([]circle, 0, 100)

	wnd.MouseMove = func(x, y int) {
		mx, my = float64(x), float64(y)
	}
	wnd.MouseDown = func(button, x, y int) {
		action = 1
		circles = append(circles, circle{x: mx, y: my, color: "#F00"})
	}
	wnd.KeyDown = func(scancode int, rn rune, name string) {
		switch name {
		case "Escape":
			wnd.Close()
		case "Space":
			action = 1
			circles = append(circles, circle{x: mx, y: my, color: "#0F0"})
		case "Enter":
			action = 1
			circles = append(circles, circle{x: mx, y: my, color: "#00F"})
		}
	}

	lastTime := time.Now()

	wnd.MainLoop(func() {
		now := time.Now()
		diff := now.Sub(lastTime)
		lastTime = now
		action -= diff.Seconds() * 3
		action = math.Max(0, action)

		w, h := float64(cv.Width()), float64(cv.Height())

		// Clear the screen
		cv.SetFillStyle("#000")
		cv.FillRect(0, 0, w, h)

		// Draw a circle around the cursor
		cv.SetStrokeStyle("#F00")
		cv.SetLineWidth(6)
		cv.BeginPath()
		cv.Arc(mx, my, 24+action*24, 0, math.Pi*2, false)
		cv.Stroke()

		// Draw circles where the user has clicked
		for _, circle := range circles {
			cv.SetFillStyle(circle.color)
			cv.BeginPath()
			cv.Arc(circle.x, circle.y, 24, 0, math.Pi*2, false)
			cv.Fill()
		}
	})
}
