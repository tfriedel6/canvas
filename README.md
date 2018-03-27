# Go canvas

Canvas is a Go library based on OpenGL that tries to provide the HTML5 canvas API as closely as possible.

Many of the basic functions are supported, but it is still a work in progress. The library aims to accept a lot of different parameters on each function in a similar way as the Javascript API does.

Since the library uses OpenGL directly in many places, the performance is likely to be decent, but not great. Browser implementation are likely to be much more optimized, but this is certainly useable.

[GoDoc is available here](https://godoc.org/github.com/tfriedel6/canvas)

# sdlcanvas

The sdlcanvas subpackage provides a very simple way to get started with just a few lines of code. As the name implies it is based on the SDL library. It creates a window for you and gives you a canvas to draw with. It also serves as a useful example for more complex programs.

