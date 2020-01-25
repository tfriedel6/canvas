package main

import (
	"log"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/backend/goglbackend"
	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	runtime.LockOSThread()

	// init SDL
	err := sdl.Init(sdl.INIT_VIDEO)
	if err != nil {
		log.Fatalf("Error initializing SDL: %v", err)
	}
	defer sdl.Quit()

	// the stencil size setting is required for the canvas to work
	sdl.GLSetAttribute(sdl.GL_RED_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_GREEN_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_BLUE_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_ALPHA_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_DEPTH_SIZE, 0)
	sdl.GLSetAttribute(sdl.GL_STENCIL_SIZE, 8)
	sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)

	// create window
	const title = "SDL Test"
	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 1280, 720, sdl.WINDOW_RESIZABLE|sdl.WINDOW_OPENGL|sdl.WINDOW_ALLOW_HIGHDPI)
	if err != nil {
		// fallback in case multisample is not available
		sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 0)
		sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 0)
		window, err = sdl.CreateWindow(title, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 1280, 720, sdl.WINDOW_RESIZABLE|sdl.WINDOW_OPENGL|sdl.WINDOW_ALLOW_HIGHDPI)
		if err != nil {
			log.Fatalf("Error creating window: %v", err)
		}
	}
	defer window.Destroy()

	// create GL context
	glContext, err := window.GLCreateContext()
	if err != nil {
		log.Fatalf("Error creating GL context: %v", err)
	}

	// init GL
	err = gl.Init()
	if err != nil {
		log.Fatalf("Error initializing GL: %v", err)
	}

	// enable vsync and multisample (if available)
	sdl.GLSetSwapInterval(1)
	gl.Enable(gl.MULTISAMPLE)

	// load GL backend
	backend, err := goglbackend.New(0, 0, 0, 0, nil)
	if err != nil {
		log.Fatalf("Error loading canvas GL assets: %v", err)
	}

	// initialize canvas with zero size, since size is set in main loop
	cv := canvas.New(backend)

	for running := true; running; {
		err := window.GLMakeCurrent(glContext)
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// find window size and scaling
		ww, wh := window.GetSize()
		fbw, fbh := window.GLGetDrawableSize()
		sx := float64(fbw) / float64(ww)
		sy := float64(fbh) / float64(wh)

		// handle events
		for {
			event := sdl.PollEvent()
			if event == nil {
				break
			}

			switch e := event.(type) {
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN && e.Keysym.Scancode == sdl.SCANCODE_ESCAPE {
					running = false
				}
			case *sdl.MouseMotionEvent:
				mx, my = float64(e.X)*sx, float64(e.Y)*sy
			case *sdl.QuitEvent:
				running = false
			case *sdl.WindowEvent:
				if e.Type == sdl.WINDOWEVENT_CLOSE {
					running = false
				}
			}
		}

		// set canvas size
		backend.SetBounds(0, 0, int(fbw), int(fbh))

		// call the run function to do all the drawing
		run(cv, float64(fbw), float64(fbh))

		// swap back and front buffer
		window.GLSwap()
	}
}

var mx, my float64

func run(cv *canvas.Canvas, w, h float64) {
	cv.SetFillStyle("#000")
	cv.FillRect(0, 0, w, h)
	cv.SetFillStyle("#0F0")
	cv.FillRect(w*0.25, h*0.25, w*0.5, h*0.5)
	cv.SetStrokeStyle("#00F")
	cv.StrokeRect(mx-32, my-32, 64, 64)
}
