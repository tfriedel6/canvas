package backendbase

import (
	"fmt"
	"math"
)

type Vec [2]float64

func (v Vec) String() string {
	return fmt.Sprintf("[%f,%f]", v[0], v[1])
}

func (v Vec) Add(v2 Vec) Vec {
	return Vec{v[0] + v2[0], v[1] + v2[1]}
}

func (v Vec) Sub(v2 Vec) Vec {
	return Vec{v[0] - v2[0], v[1] - v2[1]}
}

func (v Vec) Mul(v2 Vec) Vec {
	return Vec{v[0] * v2[0], v[1] * v2[1]}
}

func (v Vec) Mulf(f float64) Vec {
	return Vec{v[0] * f, v[1] * f}
}

func (v Vec) MulMat(m Mat) Vec {
	return Vec{
		m[0]*v[0] + m[2]*v[1] + m[4],
		m[1]*v[0] + m[3]*v[1] + m[5]}
}

func (v Vec) MulMat2(m Mat2) Vec {
	return Vec{m[0]*v[0] + m[2]*v[1], m[1]*v[0] + m[3]*v[1]}
}

func (v Vec) Div(v2 Vec) Vec {
	return Vec{v[0] / v2[0], v[1] / v2[1]}
}

func (v Vec) Divf(f float64) Vec {
	return Vec{v[0] / f, v[1] / f}
}

func (v Vec) Dot(v2 Vec) float64 {
	return v[0]*v2[0] + v[1]*v2[1]
}

func (v Vec) Len() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1])
}

func (v Vec) LenSqr() float64 {
	return v[0]*v[0] + v[1]*v[1]
}

func (v Vec) Norm() Vec {
	return v.Mulf(1.0 / v.Len())
}

func (v Vec) Atan2() float64 {
	return math.Atan2(v[1], v[0])
}

func (v Vec) Angle() float64 {
	return math.Pi*0.5 - math.Atan2(v[1], v[0])
}

func (v Vec) AngleTo(v2 Vec) float64 {
	return math.Acos(v.Norm().Dot(v2.Norm()))
}

type Mat [6]float64

func (m *Mat) String() string {
	return fmt.Sprintf("[%f,%f,0,\n %f,%f,0,\n %f,%f,1,]", m[0], m[2], m[4], m[1], m[3], m[5])
}

var MatIdentity = Mat{
	1, 0,
	0, 1,
	0, 0}

func MatTranslate(v Vec) Mat {
	return Mat{
		1, 0,
		0, 1,
		v[0], v[1]}
}

func MatScale(v Vec) Mat {
	return Mat{
		v[0], 0,
		0, v[1],
		0, 0}
}

func MatRotate(radians float64) Mat {
	s, c := math.Sincos(radians)
	return Mat{
		c, s,
		-s, c,
		0, 0}
}

func (m Mat) Mul(m2 Mat) Mat {
	return Mat{
		m[0]*m2[0] + m[1]*m2[2],
		m[0]*m2[1] + m[1]*m2[3],
		m[2]*m2[0] + m[3]*m2[2],
		m[2]*m2[1] + m[3]*m2[3],
		m[4]*m2[0] + m[5]*m2[2] + m2[4],
		m[4]*m2[1] + m[5]*m2[3] + m2[5]}
}

func (m Mat) Invert() Mat {
	identity := 1.0 / (m[0]*m[3] - m[2]*m[1])

	return Mat{
		m[3] * identity,
		-m[1] * identity,
		-m[2] * identity,
		m[0] * identity,
		(m[2]*m[5] - m[3]*m[4]) * identity,
		(m[1]*m[4] - m[0]*m[5]) * identity,
	}
}

type Mat2 [4]float64

func (m Mat) Mat2() Mat2 {
	return Mat2{m[0], m[1], m[2], m[3]}
}

func (m *Mat2) String() string {
	return fmt.Sprintf("[%f,%f,\n %f,%f]", m[0], m[2], m[1], m[3])
}
