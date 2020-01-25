package main

import (
	"log"
	"runtime"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/backend/goglbackend"
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

	// load GL backend
	backend, err := goglbackend.New(0, 0, 0, 0, nil)
	if err != nil {
		log.Fatalf("Error loading canvas GL assets: %v", err)
	}

	var sx, sy float64 = 1, 1
	window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		mx, my = xpos*sx, ypos*sy
	})

	// initialize canvas with zero size, since size is set in main loop
	cv := canvas.New(backend)

	for !window.ShouldClose() {
		window.MakeContextCurrent()

		// find window size and scaling
		ww, wh := window.GetSize()
		fbw, fbh := window.GetFramebufferSize()
		sx = float64(fbw) / float64(ww)
		sy = float64(fbh) / float64(wh)

		glfw.PollEvents()

		// set canvas size
		backend.SetBounds(0, 0, fbw, fbh)

		// call the run function to do all the drawing
		run(cv, float64(fbw), float64(fbh))

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
