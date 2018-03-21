package canvas

import (
	"fmt"
	"math"
)

type vec [2]float64

func (v vec) String() string {
	return fmt.Sprintf("[%f,%f]", v[0], v[1])
}

func (v1 vec) add(v2 vec) vec {
	return vec{v1[0] + v2[0], v1[1] + v2[1]}
}

func (v1 vec) sub(v2 vec) vec {
	return vec{v1[0] - v2[0], v1[1] - v2[1]}
}

func (v1 vec) mul(v2 vec) vec {
	return vec{v1[0] * v2[0], v1[1] * v2[1]}
}

func (v vec) mulf(f float64) vec {
	return vec{v[0] * f, v[1] * f}
}

func (v vec) mulMat(m mat) (vec, float64) {
	return vec{
			m[0]*v[0] + m[3]*v[1] + m[6],
			m[1]*v[0] + m[4]*v[1] + m[7]},
		m[2]*v[0] + m[5]*v[1] + m[8]
}

func (v1 vec) div(v2 vec) vec {
	return vec{v1[0] / v2[0], v1[1] / v2[1]}
}

func (v vec) divf(f float64) vec {
	return vec{v[0] / f, v[1] / f}
}

func (v1 vec) dot(v2 vec) float64 {
	return v1[0]*v2[0] + v1[1]*v2[1]
}

func (v vec) len() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1])
}

func (v vec) lenSqr() float64 {
	return v[0]*v[0] + v[1]*v[1]
}

func (v vec) norm() vec {
	return v.mulf(1.0 / v.len())
}

func (v vec) atan2() float64 {
	return math.Atan2(v[1], v[0])
}

func (v vec) angle() float64 {
	return math.Pi*0.5 - math.Atan2(v[1], v[0])
}

func (v vec) angleTo(v2 vec) float64 {
	return math.Acos(v.norm().dot(v2.norm()))
}

type mat [9]float64

func (m *mat) String() string {
	return fmt.Sprintf("[%f,%f,%f,\n %f,%f,%f,\n %f,%f,%f,]", m[0], m[3], m[6], m[1], m[4], m[7], m[2], m[5], m[8])
}

func matIdentity() mat {
	return mat{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1}
}

func matTranslate(v vec) mat {
	return mat{
		1, 0, 0,
		0, 1, 0,
		v[0], v[1], 1}
}

func matScale(v vec) mat {
	return mat{
		v[0], 0, 0,
		0, v[1], 0,
		0, 0, 1}
}

func matRotate(radians float64) mat {
	s, c := math.Sincos(radians)
	return mat{
		c, s, 0,
		-s, c, 0,
		0, 0, 1}
}

func (m1 mat) mul(m2 mat) mat {
	return mat{
		m1[0]*m2[0] + m1[1]*m2[3] + m1[2]*m2[6],
		m1[0]*m2[1] + m1[1]*m2[4] + m1[2]*m2[7],
		m1[0]*m2[2] + m1[1]*m2[5] + m1[2]*m2[8],
		m1[3]*m2[0] + m1[4]*m2[3] + m1[5]*m2[6],
		m1[3]*m2[1] + m1[4]*m2[4] + m1[5]*m2[7],
		m1[3]*m2[2] + m1[4]*m2[5] + m1[5]*m2[8],
		m1[6]*m2[0] + m1[7]*m2[3] + m1[8]*m2[6],
		m1[6]*m2[1] + m1[7]*m2[4] + m1[8]*m2[7],
		m1[6]*m2[2] + m1[7]*m2[5] + m1[8]*m2[8]}
}

func (m mat) invert() mat {
	var identity float64 = 1.0 / (m[0]*m[4]*m[8] + m[3]*m[7]*m[2] + m[6]*m[1]*m[5] - m[6]*m[4]*m[2] - m[3]*m[1]*m[8] - m[0]*m[7]*m[5])

	return mat{
		(m[4]*m[8] - m[5]*m[7]) * identity,
		(m[2]*m[7] - m[1]*m[8]) * identity,
		(m[1]*m[5] - m[2]*m[4]) * identity,
		(m[5]*m[6] - m[3]*m[8]) * identity,
		(m[0]*m[8] - m[2]*m[6]) * identity,
		(m[2]*m[3] - m[0]*m[5]) * identity,
		(m[3]*m[7] - m[4]*m[6]) * identity,
		(m[1]*m[6] - m[0]*m[7]) * identity,
		(m[0]*m[4] - m[1]*m[3]) * identity}
}

func (m mat) f32() [9]float32 {
	return [9]float32{
		float32(m[0]), float32(m[1]), float32(m[2]),
		float32(m[3]), float32(m[4]), float32(m[5]),
		float32(m[6]), float32(m[7]), float32(m[8])}
}
