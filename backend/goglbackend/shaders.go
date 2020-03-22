package goglbackend

var unifiedVS = `
attribute vec2 vertex, texCoord;

uniform vec2 canvasSize;
uniform mat3 matrix;

varying vec2 v_cp, v_tc;

void main() {
	v_tc = texCoord;
	vec3 v = matrix * vec3(vertex.xy, 1.0);
	vec2 tf = v.xy / v.z;
	v_cp = tf;
	vec2 glp = tf * 2.0 / canvasSize - 1.0;
    gl_Position = vec4(glp.x, -glp.y, 0.0, 1.0);
}
`
var unifiedFS = `
#ifdef GL_ES
precision mediump float;
#endif

varying vec2 v_cp, v_tc;

uniform int func;

uniform vec4 color;
uniform float globalAlpha;

uniform sampler2D gradient;
uniform vec2 from, dir, to;
uniform float len, radFrom, radTo;

uniform vec2 imageSize;
uniform sampler2D image;
uniform mat3 imageTransform;
uniform vec2 repeat;

uniform bool useAlphaTex;
uniform sampler2D alphaTex;

uniform int boxSize;
uniform bool boxVertical;
uniform float boxScale;
uniform float boxOffset;

bool isNaN(float v) {
  return v < 0.0 || 0.0 < v || v == 0.0 ? false : true;
}

void main() {
	vec4 col = color;

	if (func == 5) {
		vec4 sum = vec4(0.0);
		if (boxVertical) {
			vec2 start = v_tc - vec2(0.0, (float(boxSize) * 0.5 + boxOffset) * boxScale);
			for (int i=0; i <= boxSize; i++) {
				sum += texture2D(image, start + vec2(0.0, float(i) * boxScale));
			}
		} else {
			vec2 start = v_tc - vec2((float(boxSize) * 0.5 + boxOffset) * boxScale, 0.0);
			for (int i=0; i <= boxSize; i++) {
				sum += texture2D(image, start + vec2(float(i) * boxScale, 0.0));
			}
		}
		gl_FragColor = sum / float(boxSize+1);
		return;
	}

	if (func == 1) {
		vec2 v = v_cp - from;
		float r = dot(v, dir) / len;
		r = clamp(r, 0.0, 1.0);
		col = texture2D(gradient, vec2(r, 0.0));
	} else if (func == 2) {
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
	} else if (func == 3) {
		vec3 tfpt = vec3(v_cp, 1.0) * imageTransform;
		vec2 imgpt = tfpt.xy / imageSize;
		col = texture2D(image, mod(imgpt, 1.0));
		if (imgpt.x < 0.0 || imgpt.x > 1.0) {
			col *= repeat.x;
		}
		if (imgpt.y < 0.0 || imgpt.y > 1.0) {
			col *= repeat.y;
		}
	} else if (func == 4) {
		col = texture2D(image, v_tc);
	}

	if (useAlphaTex) {
		col.a *= texture2D(alphaTex, v_tc).a * globalAlpha;
	} else {
		col.a *= globalAlpha;
	}

    gl_FragColor = col;
}
`

const (
	shdFuncSolid int32 = iota
	shdFuncLinearGradient
	shdFuncRadialGradient
	shdFuncImagePattern
	shdFuncImage
	shdFuncBoxBlur
)

type unifiedShader struct {
	shaderProgram

	Vertex   uint32
	TexCoord uint32

	CanvasSize  int32
	Matrix      int32
	Color       int32
	GlobalAlpha int32

	Func int32

	UseAlphaTex int32
	AlphaTex    int32

	Gradient       int32
	From, To, Dir  int32
	Len            int32
	RadFrom, RadTo int32

	ImageSize      int32
	Image          int32
	ImageTransform int32
	Repeat         int32

	BoxSize     int32
	BoxVertical int32
	BoxScale    int32
	BoxOffset   int32
}
