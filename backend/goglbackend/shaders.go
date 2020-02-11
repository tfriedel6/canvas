package goglbackend

import (
	"bytes"
	"fmt"
	"strings"
)

var unifiedVS = `
attribute vec2 vertex, alphaTexCoord;

uniform vec2 canvasSize;

varying vec2 v_cp, v_atc;

void main() {
    v_atc = alphaTexCoord;
	v_cp = vertex;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}
`
var unifiedFS = `
#ifdef GL_ES
precision mediump float;
#endif

varying vec2 v_cp, v_atc;

uniform vec4 color;
uniform float globalAlpha;

uniform bool useLinearGradient;
uniform bool useRadialGradient;
uniform sampler2D gradient;
uniform vec2 from, dir, to;
uniform float len, radFrom, radTo;

uniform bool useImagePattern;
uniform vec2 imageSize;
uniform sampler2D image;
uniform mat3 imageTransform;
uniform vec2 repeat;

uniform bool useAlphaTex;
uniform sampler2D alphaTex;

bool isNaN(float v) {
  return v < 0.0 || 0.0 < v || v == 0.0 ? false : true;
}

void main() {
	vec4 col = color;

	if (useLinearGradient) {
		vec2 v = v_cp - from;
		float r = dot(v, dir) / len;
		r = clamp(r, 0.0, 1.0);
		col = texture2D(gradient, vec2(r, 0.0));
	} else if (useRadialGradient) {
		float o_a = 0.5 * sqrt(
			pow(-2.0*from.x*from.x+2.0*from.x*to.x+2.0*from.x*v_cp.x-2.0*to.x*v_cp.x-2.0*from.y*from.y+2.0*from.y*to.y+2.0*from.y*v_cp.y-2.0*to.y*v_cp.y+2.0*radFrom*radFrom-2.0*radFrom*radTo, 2.0)
			-4.0*(from.x*from.x-2.0*from.x*v_cp.x+v_cp.x*v_cp.x+from.y*from.y-2.0*from.y*v_cp.y+v_cp.y*v_cp.y-radFrom*radFrom)
			*(from.x*from.x-2.0*from.x*to.x+to.x*to.x+from.y*from.y-2.0*from.y*to.y+to.y*to.y-radFrom*radFrom+2.0*radFrom*radTo-radTo*radTo)
		);
		float o_b = (from.x*from.x-from.x*to.x-from.x*v_cp.x+to.x*v_cp.x+from.y*from.y-from.y*to.y-from.y*v_cp.y+to.y*v_cp.y-radFrom*radFrom+radFrom*radTo);
		float o_c = (from.x*from.x-2.0*from.x*to.x+to.x*to.x+from.y*from.y-2.0*from.y*to.y+to.y*to.y-radFrom*radFrom+2.0*radFrom*radTo-radTo*radTo);
		float o1 = (-o_a + o_b) / o_c;
		float o2 = (o_a + o_b) / o_c;
		if (isNaN(o1) && isNaN(o2)) {
			gl_FragColor = vec4(0.0, 0.0, 0.0, 0.0);
			return;
		}
		float o = max(o1, o2);
		o = clamp(o, 0.0, 1.0);
		col = texture2D(gradient, vec2(o, 0.0));
	} else if (useImagePattern) {
		vec3 tfpt = vec3(v_cp, 1.0) * imageTransform;
		vec2 imgpt = tfpt.xy / imageSize;
		col = texture2D(image, mod(imgpt, 1.0));
		if (imgpt.x < 0.0 || imgpt.x > 1.0) {
			col *= repeat.x;
		}
		if (imgpt.y < 0.0 || imgpt.y > 1.0) {
			col *= repeat.y;
		}
	}

	if (useAlphaTex) {
		col.a *= texture2D(alphaTex, v_atc).a * globalAlpha;
	} else {
		col.a *= globalAlpha;
	}

    gl_FragColor = col;
}
`

var imageVS = `
attribute vec2 vertex, texCoord;
uniform vec2 canvasSize;
varying vec2 v_texCoord;
void main() {
	v_texCoord = texCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var imageFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_texCoord;
uniform sampler2D image;
uniform float globalAlpha;
void main() {
	vec4 col = texture2D(image, v_texCoord);
	col.a *= globalAlpha;
    gl_FragColor = col;
}`

var gaussian15VS = `
attribute vec2 vertex, texCoord;
uniform vec2 canvasSize;
varying vec2 v_texCoord;
void main() {
	v_texCoord = texCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var gaussian15FS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_texCoord;
uniform vec2 kernelScale;
uniform sampler2D image;
uniform float kernel[15];
void main() {
	vec4 color = vec4(0.0, 0.0, 0.0, 0.0);
_SUM_
    gl_FragColor = color;
}`

var gaussian63VS = `
attribute vec2 vertex, texCoord;
uniform vec2 canvasSize;
varying vec2 v_texCoord;
void main() {
	v_texCoord = texCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var gaussian63FS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_texCoord;
uniform vec2 kernelScale;
uniform sampler2D image;
uniform float kernel[63];
void main() {
	vec4 color = vec4(0.0, 0.0, 0.0, 0.0);
_SUM_
    gl_FragColor = color;
}`

var gaussian127VS = `
attribute vec2 vertex, texCoord;
uniform vec2 canvasSize;
varying vec2 v_texCoord;
void main() {
	v_texCoord = texCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var gaussian127FS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_texCoord;
uniform vec2 kernelScale;
uniform sampler2D image;
uniform float kernel[127];
void main() {
	vec4 color = vec4(0.0, 0.0, 0.0, 0.0);
_SUM_
    gl_FragColor = color;
}`

func init() {
	fstr := "\tcolor += texture2D(image, v_texCoord + vec2(%.1f * kernelScale.x, %.1f * kernelScale.y)) * kernel[%d];\n"
	bb := bytes.Buffer{}
	for i := 0; i < 127; i++ {
		off := float64(i) - 63
		fmt.Fprintf(&bb, fstr, off, off, i)
	}
	gaussian127FS = strings.Replace(gaussian127FS, "_SUM_", bb.String(), -1)
	bb.Reset()
	for i := 0; i < 63; i++ {
		off := float64(i) - 31
		fmt.Fprintf(&bb, fstr, off, off, i)
	}
	gaussian63FS = strings.Replace(gaussian63FS, "_SUM_", bb.String(), -1)
	bb.Reset()
	for i := 0; i < 15; i++ {
		off := float64(i) - 7
		fmt.Fprintf(&bb, fstr, off, off, i)
	}
	gaussian15FS = strings.Replace(gaussian15FS, "_SUM_", bb.String(), -1)
}

type unifiedShader struct {
	shaderProgram

	Vertex        uint32
	AlphaTexCoord uint32

	CanvasSize  int32
	Color       int32
	GlobalAlpha int32

	UseAlphaTex int32
	AlphaTex    int32

	UseLinearGradient int32
	UseRadialGradient int32
	Gradient          int32
	From, To, Dir     int32
	Len               int32
	RadFrom, RadTo    int32

	UseImagePattern int32
	ImageSize       int32
	Image           int32
	ImageTransform  int32
	Repeat          int32
}

type imageShader struct {
	shaderProgram
	Vertex      uint32
	TexCoord    uint32
	CanvasSize  int32
	Image       int32
	GlobalAlpha int32
}

type gaussianShader struct {
	shaderProgram
	Vertex      uint32
	TexCoord    uint32
	CanvasSize  int32
	KernelScale int32
	Image       int32
	Kernel      int32
}
