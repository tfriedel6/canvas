package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/glimpl/gogl"
)

func main() {
	runtime.LockOSThread()

	// init GLFW
	err := glfw.Init()
	if err != nil {
		log.Fatalf("Error initializing GLFW: %v", err)
	}
	defer glfw.Terminate()

	// the stencil size setting is required for the canvas to work
	glfw.WindowHint(glfw.StencilBits, 8)
	glfw.WindowHint(glfw.DepthBits, 0)

	// create window
	window, err := glfw.CreateWindow(1280, 720, "GLFW Test", nil, nil)
	if err != nil {
		log.Fatalf("Error creating window: %v", err)
	}
	window.MakeContextCurrent()

	// init GL
	err = gl.Init()
	if err != nil {
		log.Fatalf("Error initializing GL: %v", err)
	}

	// set vsync on, enable multisample (if available)
	glfw.SwapInterval(1)
	gl.Enable(gl.MULTISAMPLE)

	// load canvas GL assets
	err = canvas.LoadGL(glimplgogl.GLImpl{})
	if err != nil {
		log.Fatalf("Error loading canvas GL assets: %v", err)
	}

	window.SetCursorPosCallback(func(w *glfw.Window, xpos float64, ypos float64) {
		mx, my = xpos, ypos
	})

	// initialize canvas with zero size, since size is set in main loop
	cv := canvas.New(0, 0, 0, 0)

	for !window.ShouldClose() {
		window.MakeContextCurrent()
		glfw.PollEvents()

		// set canvas size
		ww, wh := window.GetSize()
		cv.SetBounds(0, 0, ww, wh)

		// call the run function to do all the drawing
		run(cv, float64(ww), float64(wh))

		// swap back and front buffer
		window.SwapBuffers()
	}
}

var mx, my float64

func run(cv *canvas.Canvas, w, h float64) {
	cv.SetFillStyle("#000")
	cv.FillRect(0, 0, w, h)
	cv.SetFillStyle("#00F")
	cv.FillRect(w*0.25, h*0.25, w*0.5, h*0.5)
	cv.SetStrokeStyle("#0F0")
	cv.StrokeRect(mx-32, my-32, 64, 64)
}
