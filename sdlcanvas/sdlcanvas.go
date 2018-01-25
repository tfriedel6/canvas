package sdlcanvas

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/veandco/go-sdl2/sdl"

	"tests/canvas"
	"tests/canvas/goglimpl"
)

type Window struct {
	Window    *sdl.Window
	GLContext sdl.GLContext
	fps       float32
}

func CreateCanvasWindow(w, h int, title string) (*Window, *canvas.Canvas, error) {
	runtime.LockOSThread()

	// init SDL
	err := sdl.Init(sdl.INIT_VIDEO)
	if err != nil {
		return nil, nil, fmt.Errorf("Error initializing SDL: %v", err)
	}

	sdl.GL_SetAttribute(sdl.GL_RED_SIZE, 8)
	sdl.GL_SetAttribute(sdl.GL_GREEN_SIZE, 8)
	sdl.GL_SetAttribute(sdl.GL_BLUE_SIZE, 8)
	sdl.GL_SetAttribute(sdl.GL_ALPHA_SIZE, 8)
	sdl.GL_SetAttribute(sdl.GL_DEPTH_SIZE, 0)
	sdl.GL_SetAttribute(sdl.GL_STENCIL_SIZE, 1)
	sdl.GL_SetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	sdl.GL_SetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)
	sdl.GL_SetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)

	// create window
	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, w, h, sdl.WINDOW_RESIZABLE|sdl.WINDOW_OPENGL)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating window: %v", err)
	}

	// create GL context
	glContext, err := sdl.GL_CreateContext(window)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating GL context: %v", err)
	}

	// init GL
	err = gl.Init()
	if err != nil {
		return nil, nil, fmt.Errorf("Error initializing GL: %v", err)
	}

	sdl.GL_SetSwapInterval(1)
	gl.Enable(gl.MULTISAMPLE)

	err = canvas.LoadGL(goglimpl.GLImpl{})
	if err != nil {
		return nil, nil, fmt.Errorf("Error loading canvas GL assets: %v", err)
	}

	cv := canvas.New(0, 0, w, h)
	wnd := &Window{
		Window:    window,
		GLContext: glContext,
	}

	return wnd, cv, nil
}

func (wnd *Window) Destroy() {
	sdl.GL_DeleteContext(wnd.GLContext)
	wnd.Window.Destroy()
}

func (wnd *Window) FPS() float32 {
	return wnd.fps
}

func (wnd *Window) MainLoop(drawFunc func()) {
	var frameTimes [10]time.Time
	frameIndex := 0
	frameCount := 0

	// main loop
	for running := true; running; {
		for {
			ei := sdl.PollEvent()
			if ei == nil {
				break
			}
			switch e := ei.(type) {
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					running = false
				}
			case *sdl.KeyDownEvent:
				if e.Keysym.Scancode == sdl.SCANCODE_ESCAPE {
					running = false
				}
			}
		}

		err := sdl.GL_MakeCurrent(wnd.Window, wnd.GLContext)
		if err != nil {
			log.Println(err)
			time.Sleep(10 * time.Millisecond)
			continue
		}

		drawFunc()

		now := time.Now()
		frameTimes[frameIndex] = now
		frameIndex++
		frameIndex %= len(frameTimes)
		if frameCount < len(frameTimes) {
			frameCount++
		} else {
			diff := now.Sub(frameTimes[frameIndex]).Seconds()
			wnd.fps = float32(frameCount-1) / float32(diff)
		}

		sdl.GL_SwapWindow(wnd.Window)
	}
}
