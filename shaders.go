package canvas

import (
	"bytes"
	"fmt"
	"strings"
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
uniform mat3 invmat;
uniform sampler2D gradient;
uniform vec2 from, dir;
uniform float len;
uniform float globalAlpha;
void main() {
	vec3 untf = vec3(v_cp, 1.0) * invmat;
	vec2 v = untf.xy - from;
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
uniform mat3 invmat;
uniform sampler2D gradient;
uniform vec2 from, to, dir;
uniform float radFrom, radTo;
uniform float len;
uniform float globalAlpha;
bool isNaN(float v) {
  return v < 0.0 || 0.0 < v || v == 0.0 ? false : true;
}
void main() {
	vec3 untf = vec3(v_cp, 1.0) * invmat;
	float o_a = 0.5 * sqrt(
		pow(-2.0*from.x*from.x+2.0*from.x*to.x+2.0*from.x*untf.x-2.0*to.x*untf.x-2.0*from.y*from.y+2.0*from.y*to.y+2.0*from.y*untf.y-2.0*to.y*untf.y+2.0*radFrom*radFrom-2.0*radFrom*radTo, 2.0)
		-4.0*(from.x*from.x-2.0*from.x*untf.x+untf.x*untf.x+from.y*from.y-2.0*from.y*untf.y+untf.y*untf.y-radFrom*radFrom)
		*(from.x*from.x-2.0*from.x*to.x+to.x*to.x+from.y*from.y-2.0*from.y*to.y+to.y*to.y-radFrom*radFrom+2.0*radFrom*radTo-radTo*radTo)
	);
	float o_b = (from.x*from.x-from.x*to.x-from.x*untf.x+to.x*untf.x+from.y*from.y-from.y*to.y-from.y*untf.y+to.y*untf.y-radFrom*radFrom+radFrom*radTo);
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
uniform mat3 invmat;
uniform sampler2D image;
uniform float globalAlpha;
void main() {
	vec3 untf = vec3(v_cp, 1.0) * invmat;
	vec4 col = texture2D(image, mod(untf.xy / imageSize, 1.0));
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
uniform mat3 invmat;
uniform sampler2D gradient;
uniform vec2 from, dir;
uniform float len;
uniform sampler2D alphaTex;
uniform float globalAlpha;
void main() {
	vec3 untf = vec3(v_cp, 1.0) * invmat;
	vec2 v = untf.xy - from;
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
uniform mat3 invmat;
uniform sampler2D gradient;
uniform vec2 from, to, dir;
uniform float radFrom, radTo;
uniform float len;
uniform sampler2D alphaTex;
uniform float globalAlpha;
bool isNaN(float v) {
  return v < 0.0 || 0.0 < v || v == 0.0 ? false : true;
}
void main() {
	vec3 untf = vec3(v_cp, 1.0) * invmat;
	float o_a = 0.5 * sqrt(
		pow(-2.0*from.x*from.x+2.0*from.x*to.x+2.0*from.x*untf.x-2.0*to.x*untf.x-2.0*from.y*from.y+2.0*from.y*to.y+2.0*from.y*untf.y-2.0*to.y*untf.y+2.0*radFrom*radFrom-2.0*radFrom*radTo, 2.0)
		-4.0*(from.x*from.x-2.0*from.x*untf.x+untf.x*untf.x+from.y*from.y-2.0*from.y*untf.y+untf.y*untf.y-radFrom*radFrom)
		*(from.x*from.x-2.0*from.x*to.x+to.x*to.x+from.y*from.y-2.0*from.y*to.y+to.y*to.y-radFrom*radFrom+2.0*radFrom*radTo-radTo*radTo)
	);
	float o_b = (from.x*from.x-from.x*to.x-from.x*untf.x+to.x*untf.x+from.y*from.y-from.y*to.y-from.y*untf.y+to.y*untf.y-radFrom*radFrom+radFrom*radTo);
	float o_c = (from.x*from.x-2.0*from.x*to.x+to.x*to.x+from.y*from.y-2.0*from.y*to.y+to.y*to.y-radFrom*radFrom+2.0*radFrom*radTo-radTo*radTo);
	float o1 = (-o_a + o_b) / o_c;
	float o2 = (o_a + o_b) / o_c;
	if (isNaN(o1) && isNaN(o2)) {
		gl_FragColor = vec4(0.0, 0.0, 0.0, 0.0);
		return;
	}
	float o = max(o1, o2);
	float r = radFrom + o * (radTo - radFrom);
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
uniform mat3 invmat;
uniform sampler2D image;
uniform sampler2D alphaTex;
uniform float globalAlpha;
void main() {
	vec3 untf = vec3(v_cp, 1.0) * invmat;
    vec4 col = texture2D(image, mod(untf.xy / imageSize, 1.0));
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

var gaussian255VS = `
attribute vec2 vertex, texCoord;
uniform vec2 canvasSize;
varying vec2 v_texCoord;
void main() {
	v_texCoord = texCoord;
	vec2 glp = vertex * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}`
var gaussian255FS = `
#ifdef GL_ES
precision mediump float;
#endif
varying vec2 v_texCoord;
uniform vec2 kernelScale;
uniform sampler2D image;
uniform float kernel[255];
void main() {
	vec4 color = vec4(0.0, 0.0, 0.0, 0.0);
_SUM_
    gl_FragColor = color;
}`

func init() {
	fstr := "\tcolor += texture2D(image, v_texCoord + vec2(%.1f * kernelScale.x, %.1f * kernelScale.y)) * kernel[%d];\n"
	bb := bytes.Buffer{}
	for i := 0; i < 255; i++ {
		off := float64(i) - 127
		fmt.Fprintf(&bb, fstr, off, off, i)
	}
	gaussian255FS = strings.Replace(gaussian255FS, "_SUM_", bb.String(), -1)
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
