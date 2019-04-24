package xmobilebackend

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/mobile/gl"
)

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

var solidVS = `
attribute vec2 vertex;
uniform vec2 canvasSize;
void main() {
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var solidFS = `
#ifdef GL_ES
precision mediump float;
#endif
uniform vec4 color;
uniform float globalAlpha;
void main() {
	vec4 col = color;
	col.a *= globalAlpha;
    gl_FragColor = col;
}`

var linearGradientVS = `
attribute vec2 vertex;
uniform vec2 canvasSize;
varying vec2 v_cp;
void main() {
	v_cp = vertex;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var linearGradientFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_cp;
uniform sampler2D gradient;
uniform vec2 from, dir;
uniform float len;
uniform float globalAlpha;
void main() {
	vec2 v = v_cp - from;
	float r = dot(v, dir) / len;
	r = clamp(r, 0.0, 1.0);
	vec4 col = texture2D(gradient, vec2(r, 0.0));
	col.a *= globalAlpha;
    gl_FragColor = col;
}`

var radialGradientVS = `
attribute vec2 vertex;
uniform vec2 canvasSize;
varying vec2 v_cp;
void main() {
	v_cp = vertex;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var radialGradientFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_cp;
uniform sampler2D gradient;
uniform vec2 from, to;
uniform float radFrom, radTo;
uniform float globalAlpha;
bool isNaN(float v) {
  return v < 0.0 || 0.0 < v || v == 0.0 ? false : true;
}
void main() {
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
	//float r = radFrom + o * (radTo - radFrom);
	vec4 col = texture2D(gradient, vec2(o, 0.0));
	col.a *= globalAlpha;
    gl_FragColor = col;
}`

var imagePatternVS = `
attribute vec2 vertex;
uniform vec2 canvasSize;
varying vec2 v_cp;
void main() {
	v_cp = vertex;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var imagePatternFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_cp;
uniform vec2 imageSize;
uniform sampler2D image;
uniform mat3 imageTransform;
uniform vec2 repeat;
uniform float globalAlpha;
void main() {
	vec3 tfpt = vec3(v_cp, 1.0) * imageTransform;
	vec2 imgpt = tfpt.xy / imageSize;
	vec4 col = texture2D(image, mod(imgpt, 1.0));
	if (imgpt.x < 0.0 || imgpt.x > 1.0) {
		col *= repeat.x;
	}
	if (imgpt.y < 0.0 || imgpt.y > 1.0) {
		col *= repeat.y;
	}
	col.a *= globalAlpha;
    gl_FragColor = col;
}`

var solidAlphaVS = `
attribute vec2 vertex, alphaTexCoord;
uniform vec2 canvasSize;
varying vec2 v_atc;
void main() {
    v_atc = alphaTexCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var solidAlphaFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_atc;
uniform vec4 color;
uniform sampler2D alphaTex;
uniform float globalAlpha;
void main() {
    vec4 col = color;
    col.a *= texture2D(alphaTex, v_atc).a * globalAlpha;
    gl_FragColor = col;
}`

var linearGradientAlphaVS = `
attribute vec2 vertex, alphaTexCoord;
uniform vec2 canvasSize;
varying vec2 v_cp;
varying vec2 v_atc;
void main() {
	v_cp = vertex;
    v_atc = alphaTexCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var linearGradientAlphaFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_cp;
varying vec2 v_atc;
varying vec2 v_texCoord;
uniform sampler2D gradient;
uniform vec2 from, dir;
uniform float len;
uniform sampler2D alphaTex;
uniform float globalAlpha;
void main() {
	vec2 v = v_cp - from;
	float r = dot(v, dir) / len;
	r = clamp(r, 0.0, 1.0);
    vec4 col = texture2D(gradient, vec2(r, 0.0));
    col.a *= texture2D(alphaTex, v_atc).a * globalAlpha;
    gl_FragColor = col;
}`

var radialGradientAlphaVS = `
attribute vec2 vertex, alphaTexCoord;
uniform vec2 canvasSize;
varying vec2 v_cp;
varying vec2 v_atc;
void main() {
	v_cp = vertex;
    v_atc = alphaTexCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var radialGradientAlphaFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_cp;
varying vec2 v_atc;
uniform sampler2D gradient;
uniform vec2 from, to;
uniform float radFrom, radTo;
uniform sampler2D alphaTex;
uniform float globalAlpha;
bool isNaN(float v) {
  return v < 0.0 || 0.0 < v || v == 0.0 ? false : true;
}
void main() {
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
	//float r = radFrom + o * (radTo - radFrom);
    vec4 col = texture2D(gradient, vec2(o, 0.0));
    col.a *= texture2D(alphaTex, v_atc).a * globalAlpha;
	gl_FragColor = col;
}`

var imagePatternAlphaVS = `
attribute vec2 vertex, alphaTexCoord;
uniform vec2 canvasSize;
varying vec2 v_cp;
varying vec2 v_atc;
void main() {
	v_cp = vertex;
    v_atc = alphaTexCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var imagePatternAlphaFS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_cp;
varying vec2 v_atc;
uniform vec2 imageSize;
uniform sampler2D image;
uniform mat3 imageTransform;
uniform vec2 repeat;
uniform sampler2D alphaTex;
uniform float globalAlpha;
void main() {
	vec3 tfpt = vec3(v_cp, 1.0) * imageTransform;
	vec2 imgpt = tfpt.xy / imageSize;
	vec4 col = texture2D(image, mod(imgpt, 1.0));
	if (imgpt.x < 0.0 || imgpt.x > 1.0) {
		col *= repeat.x;
	}
	if (imgpt.y < 0.0 || imgpt.y > 1.0) {
		col *= repeat.y;
	}
    col.a *= texture2D(alphaTex, v_atc).a * globalAlpha;
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

type solidShader struct {
	shaderProgram
	Vertex      gl.Attrib
	CanvasSize  gl.Uniform
	Color       gl.Uniform
	GlobalAlpha gl.Uniform
}

type imageShader struct {
	shaderProgram
	Vertex      gl.Attrib
	TexCoord    gl.Attrib
	CanvasSize  gl.Uniform
	Image       gl.Uniform
	GlobalAlpha gl.Uniform
}

type linearGradientShader struct {
	shaderProgram
	Vertex      gl.Attrib
	CanvasSize  gl.Uniform
	Gradient    gl.Uniform
	From        gl.Uniform
	Dir         gl.Uniform
	Len         gl.Uniform
	GlobalAlpha gl.Uniform
}

type radialGradientShader struct {
	shaderProgram
	Vertex      gl.Attrib
	CanvasSize  gl.Uniform
	Gradient    gl.Uniform
	From        gl.Uniform
	To          gl.Uniform
	RadFrom     gl.Uniform
	RadTo       gl.Uniform
	GlobalAlpha gl.Uniform
}

type imagePatternShader struct {
	shaderProgram
	Vertex         gl.Attrib
	CanvasSize     gl.Uniform
	ImageSize      gl.Uniform
	Image          gl.Uniform
	ImageTransform gl.Uniform
	Repeat         gl.Uniform
	GlobalAlpha    gl.Uniform
}

type solidAlphaShader struct {
	shaderProgram
	Vertex        gl.Attrib
	AlphaTexCoord gl.Attrib
	CanvasSize    gl.Uniform
	Color         gl.Uniform
	AlphaTex      gl.Uniform
	GlobalAlpha   gl.Uniform
}

type linearGradientAlphaShader struct {
	shaderProgram
	Vertex        gl.Attrib
	AlphaTexCoord gl.Attrib
	CanvasSize    gl.Uniform
	Gradient      gl.Uniform
	From          gl.Uniform
	Dir           gl.Uniform
	Len           gl.Uniform
	AlphaTex      gl.Uniform
	GlobalAlpha   gl.Uniform
}

type radialGradientAlphaShader struct {
	shaderProgram
	Vertex        gl.Attrib
	AlphaTexCoord gl.Attrib
	CanvasSize    gl.Uniform
	Gradient      gl.Uniform
	From          gl.Uniform
	To            gl.Uniform
	RadFrom       gl.Uniform
	RadTo         gl.Uniform
	AlphaTex      gl.Uniform
	GlobalAlpha   gl.Uniform
}

type imagePatternAlphaShader struct {
	shaderProgram
	Vertex         gl.Attrib
	AlphaTexCoord  gl.Attrib
	CanvasSize     gl.Uniform
	ImageSize      gl.Uniform
	Image          gl.Uniform
	ImageTransform gl.Uniform
	Repeat         gl.Uniform
	AlphaTex       gl.Uniform
	GlobalAlpha    gl.Uniform
}

type gaussianShader struct {
	shaderProgram
	Vertex      gl.Attrib
	TexCoord    gl.Attrib
	CanvasSize  gl.Uniform
	KernelScale gl.Uniform
	Image       gl.Uniform
	Kernel      gl.Uniform
}
