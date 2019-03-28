package canvas

import (
	"fmt"
	"math"
)

type vec [2]float64

func (v vec) String() string {
	return fmt.Sprintf("[%f,%f]", v[0], v[1])
}

func (v vec) add(v2 vec) vec {
	return vec{v[0] + v2[0], v[1] + v2[1]}
}

func (v vec) sub(v2 vec) vec {
	return vec{v[0] - v2[0], v[1] - v2[1]}
}

func (v vec) mul(v2 vec) vec {
	return vec{v[0] * v2[0], v[1] * v2[1]}
}

func (v vec) mulf(f float64) vec {
	return vec{v[0] * f, v[1] * f}
}

func (v vec) mulMat(m mat) vec {
	return vec{
		m[0]*v[0] + m[2]*v[1] + m[4],
		m[1]*v[0] + m[3]*v[1] + m[5]}
}

func (v vec) mulMat2(m mat2) vec {
	return vec{m[0]*v[0] + m[2]*v[1], m[1]*v[0] + m[3]*v[1]}
}

func (v vec) div(v2 vec) vec {
	return vec{v[0] / v2[0], v[1] / v2[1]}
}

func (v vec) divf(f float64) vec {
	return vec{v[0] / f, v[1] / f}
}

func (v vec) dot(v2 vec) float64 {
	return v[0]*v2[0] + v[1]*v2[1]
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

type mat [6]float64

func (m *mat) String() string {
	return fmt.Sprintf("[%f,%f,0,\n %f,%f,0,\n %f,%f,1,]", m[0], m[2], m[4], m[1], m[3], m[5])
}

func matIdentity() mat {
	return mat{
		1, 0,
		0, 1,
		0, 0}
}

func matTranslate(v vec) mat {
	return mat{
		1, 0,
		0, 1,
		v[0], v[1]}
}

func matScale(v vec) mat {
	return mat{
		v[0], 0,
		0, v[1],
		0, 0}
}

func matRotate(radians float64) mat {
	s, c := math.Sincos(radians)
	return mat{
		c, s,
		-s, c,
		0, 0}
}

func (m mat) mul(m2 mat) mat {
	return mat{
		m[0]*m2[0] + m[1]*m2[2],
		m[0]*m2[1] + m[1]*m2[3],
		m[2]*m2[0] + m[3]*m2[2],
		m[2]*m2[1] + m[3]*m2[3],
		m[4]*m2[0] + m[5]*m2[2] + m2[4],
		m[4]*m2[1] + m[5]*m2[3] + m2[5]}
}

func (m mat) invert() mat {
	identity := 1.0 / (m[0]*m[3] - m[2]*m[1])

	return mat{
		m[3] * identity,
		-m[1] * identity,
		-m[2] * identity,
		m[0] * identity,
		(m[2]*m[5] - m[3]*m[4]) * identity,
		(m[1]*m[4] - m[0]*m[5]) * identity,
	}
}

type mat2 [4]float64

func (m mat) mat2() mat2 {
	return mat2{m[0], m[1], m[2], m[3]}
}

func (m *mat2) String() string {
	return fmt.Sprintf("[%f,%f,\n %f,%f]", m[0], m[2], m[1], m[3])
}
