package canvas

import (
	"fmt"
)

var gli GL

//go:generate go run make_shaders.go
//go:generate go fmt

func glError() error {
	glErr := gli.GetError()
	if glErr != gl_NO_ERROR {
		return fmt.Errorf("GL Error: %x", glErr)
	}
	return nil
}

var (
	buf uint32
	sr  *solidShader
)

func LoadGL(glimpl GL) (err error) {
	gli = glimpl

	gli.GetError() // clear error state

	sr, err = loadSolidShader()
	if err != nil {
		return
	}
	err = glError()
	if err != nil {
		return
	}

	gli.GenBuffers(1, &buf)
	err = glError()
	if err != nil {
		return
	}

	gli.Enable(gl_BLEND)
	gli.BlendFunc(gl_SRC_ALPHA, gl_ONE_MINUS_SRC_ALPHA)

	return
}

var solidVS = `
attribute vec2 vertex;
void main() {
    gl_Position = vec4(vertex, 0.0, 1.0);
}`
var solidFS = `
#ifdef GL_ES
precision mediump float;
#endif
uniform vec4 color;
void main() {
    gl_FragColor = color;
}`
