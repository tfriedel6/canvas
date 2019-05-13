# Go canvas [![GoDoc](https://godoc.org/github.com/tfriedel6/canvas?status.svg)](https://godoc.org/github.com/tfriedel6/canvas)

Canvas is a pure Go library that provides drawing functionality as similar as possible to the HTML5 canvas API. It has nothing to do with HTML or Javascript, the functions are just made to be approximately the same.

Most of the functions are supported, but it is still a work in progress. The library aims to accept a lot of different parameters on each function in a similar way as the Javascript API does.

Whereas the Javascript API uses a context that all draw calls go to, here all draw calls are directly on the canvas type. The other difference is that here setters are used instead of properties for things like fonts and line width. 

## OpenGL backend

The OpenGL backend is intended to provide decent performance. Obviously it will not be able to rival hand coded OpenGL for a given purpose, but for many purposes it will be enough. It can also be combined with hand coded OpenGL.

## Software backend

The software backend can also be used if no OpenGL context is available. It will render into a standard Go RGBA image. 

There is experimental MSAA anti-aliasing, but it doesn't fully work properly yet. The best option for anti-aliasing currently is to render to a larger image and then scale it down.

## SDL/GLFW convenience packages

The sdlcanvas and glfwcanvas subpackages provide a very simple way to get started with just a few lines of code. As the names imply they are based on the SDL library and the GLFW library respectively. They create a window for you and give you a canvas to draw with.

# OS support

Both the OpenGL and software backends work on the following operating systems:

- Linux
- Windows
- macOS
- Android
- iOS

Using gomobile to build a full Go app using ```gomobile build``` now works by using an offscreen texture to render to and then rendering that to the screen. See the example in examples/gomobile. The offscreen texture is necessary since gomobile automatically creates a GL context without a stencil buffer, which this library requires.

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
- textBaseline
- fillStyle
- strokeText
- strokeStyle
- linear gradients
- radial gradients
- image patterns with repeat and transform
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
- isPointInPath
- isPointInStroke
- self intersecting polygons

# Missing features

- globalCompositeOperation
- imageSmoothingEnabled
- textBaseline hanging and ideographic (currently work just like top and bottom)
