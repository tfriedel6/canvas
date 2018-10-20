[GoDoc is available here](https://godoc.org/github.com/tfriedel6/canvas)

# Go canvas

Canvas is a Go library based on OpenGL that tries to provide the HTML5 canvas API as closely as possible.

Many of the basic functions are supported, but it is still a work in progress. The library aims to accept a lot of different parameters on each function in a similar way as the Javascript API does.

Whereas the Javascript API uses a context that all draw calls go to, here all draw calls are directly on the canvas type. The other difference is that here setters are used instead of properties for things like fonts and line width. 

The library is intended to provide decent performance. Obviously it will not be able to rival hand coded OpenGL for a given purpose, but for many purposes it will be enough. It can also be combined with hand coded OpenGL.

# SDL/GLFW convenience packages

The sdlcanvas and glfwcanvas subpackages provide a very simple way to get started with just a few lines of code. As the names imply they are based on the SDL library and the GLFW library respectively. They create a window for you and give you a canvas to draw with.

# OS support

- Linux
- Windows
- macOS
- Android
- iOS

Unfortunately using full Go apps using gomobile doesn't work since gomobile does not seem to create a GL view with a stencil buffer, and the canvas package makes heavy use of the stencil buffer. Therefore the ```gomobile bind``` command has to be used together with platform specific projects.

# Example

Look at the example/drawing package for some drawing examples. 

Here is a simple example for how to get started:

```go
package main

import (
	"math"

	"github.com/tfriedel6/canvas/sdlcanvas"
)

func main() {
	wnd, cv, err := sdlcanvas.CreateWindow(1280, 720, "Hello")
	if err != nil {
		panic(err)
	}
	defer wnd.Destroy()

	wnd.MainLoop(func() {
		w, h := float64(cv.Width()), float64(cv.Height())
		cv.SetFillStyle("#000")
		cv.FillRect(0, 0, w, h)

		for r := 0.0; r < math.Pi*2; r += math.Pi * 0.1 {
			cv.SetFillStyle(int(r*10), int(r*20), int(r*40))
			cv.BeginPath()
			cv.MoveTo(w*0.5, h*0.5)
			cv.Arc(w*0.5, h*0.5, math.Min(w, h)*0.4, r, r+0.1*math.Pi, false)
			cv.ClosePath()
			cv.Fill()
		}

		cv.SetStrokeStyle("#FFF")
		cv.SetLineWidth(10)
		cv.BeginPath()
		cv.Arc(w*0.5, h*0.5, math.Min(w, h)*0.4, 0, math.Pi*2, false)
		cv.Stroke()
	})
}
```

The result:

<img src="https://i.imgur.com/Nz8cT4M.png" width="320">

# Implemented features

These features *should* work just like their HTML5 counterparts, but there are likely to be a lot of edge cases where they don't work exactly the same way.

- beginPath
- closePath
- moveTo
- lineTo
- rect
- arc
- arcTo
- quadraticCurveTo
- bezierCurveTo
- stroke
- fill
- clip
- save
- restore
- scale
- translate
- rotate
- transform
- setTransform
- fillText
- measureText
- textAlign
- fillStyle
- strokeText
- strokeStyle
- linear gradients
- radial gradients
- image patterns
- lineWidth
- lineEnd (square, butt, round)
- lineJoin (bevel, miter, round)
- miterLimit
- lineDash
- getLineDash
- lineDashOffset
- global alpha
- drawImage
- getImageData
- putImageData
- clearRect
- shadowColor
- shadowOffset(X/Y)
- shadowBlur

# Missing features

- globalCompositeOperation
- textBaseline
- isPointInPath
- isPointInStroke
