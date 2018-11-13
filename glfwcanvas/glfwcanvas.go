package glfwcanvas

import (
	"fmt"
	_ "image/gif" // Imported here so that applications based on this package support these formats by default
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/glimpl/gogl"
)

// Window represents the opened window with GL context. The Mouse* and Key*
// functions can be set for callbacks
type Window struct {
	Window     *glfw.Window
	canvas     *canvas.Canvas
	frameTimes [10]time.Time
	frameIndex int
	frameCount int
	fps        float32
	close      bool
	MouseDown  func(button, x, y int)
	MouseMove  func(x, y int)
	MouseUp    func(button, x, y int)
	MouseWheel func(x, y int)
	KeyDown    func(scancode int, rn rune, name string)
	KeyUp      func(scancode int, rn rune, name string)
	KeyChar    func(rn rune)
	SizeChange func(w, h int)
}

// CreateWindow creates a window using SDL and initializes the OpenGL context
func CreateWindow(w, h int, title string) (*Window, *canvas.Canvas, error) {
	runtime.LockOSThread()

	// init GLFW
	err := glfw.Init()
	if err != nil {
		log.Fatalf("Error initializing GLFW: %v", err)
	}

	// the stencil size setting is required for the canvas to work
	glfw.WindowHint(glfw.StencilBits, 8)
	glfw.WindowHint(glfw.DepthBits, 0)

	// create window
	window, err := glfw.CreateWindow(w, h, title, nil, nil)
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

	err = canvas.LoadGL(glimplgogl.GLImpl{})
	if err != nil {
		return nil, nil, fmt.Errorf("Error loading canvas GL assets: %v", err)
	}

	cv := canvas.New(0, 0, w, h)
	wnd := &Window{
		Window: window,
		canvas: cv,
	}

	var mx, my int

	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		if action == glfw.Press && wnd.MouseDown != nil {
			wnd.MouseDown(int(button), mx, my)
		} else if action == glfw.Release && wnd.MouseUp != nil {
			wnd.MouseUp(int(button), mx, my)
		}
	})
	window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
		mx, my = int(math.Round(xpos)), int(math.Round(ypos))
		if wnd.MouseMove != nil {
			wnd.MouseMove(mx, my)
		}
	})
	window.SetScrollCallback(func(w *glfw.Window, xoff, yoff float64) {
		if wnd.MouseWheel != nil {
			wnd.MouseWheel(int(math.Round(xoff)), int(math.Round(yoff)))
		}
	})
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press && wnd.KeyDown != nil {
			wnd.KeyDown(scancode, keyRune(key), keyName(key))
		} else if action == glfw.Release && wnd.KeyUp != nil {
			wnd.KeyUp(scancode, keyRune(key), keyName(key))
		}
	})
	window.SetCharCallback(func(w *glfw.Window, char rune) {
		if wnd.KeyChar != nil {
			wnd.KeyChar(char)
		}
	})
	window.SetSizeCallback(func(w *glfw.Window, width, height int) {
		if wnd.SizeChange != nil {
			wnd.SizeChange(width, height)
		} else {
			cv.SetBounds(0, 0, width, height)
		}
	})
	window.SetCloseCallback(func(w *glfw.Window) {
		wnd.Close()
	})

	return wnd, cv, nil
}

// FPS returns the frames per second (averaged over 10 frames)
func (wnd *Window) FPS() float32 {
	return wnd.fps
}

// Close can be used to end a call to MainLoop
func (wnd *Window) Close() {
	wnd.close = true
}

// StartFrame handles events and gets the window ready for rendering
func (wnd *Window) StartFrame() {
	wnd.Window.MakeContextCurrent()
	glfw.PollEvents()
}

// FinishFrame updates the FPS count and displays the frame
func (wnd *Window) FinishFrame() {
	now := time.Now()
	wnd.frameTimes[wnd.frameIndex] = now
	wnd.frameIndex++
	wnd.frameIndex %= len(wnd.frameTimes)
	if wnd.frameCount < len(wnd.frameTimes) {
		wnd.frameCount++
	} else {
		diff := now.Sub(wnd.frameTimes[wnd.frameIndex]).Seconds()
		wnd.fps = float32(wnd.frameCount-1) / float32(diff)
	}

	wnd.Window.SwapBuffers()
}

// MainLoop runs a main loop and calls run on every frame
func (wnd *Window) MainLoop(run func()) {
	for !wnd.close {
		wnd.StartFrame()
		run()
		wnd.FinishFrame()
	}
}

// Size returns the current width and height of the window
func (wnd *Window) Size() (int, int) {
	return wnd.Window.GetSize()
}
