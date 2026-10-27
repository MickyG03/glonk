package geometry

import "math"

type Vector3 struct {
	X, Y, Z float64
}

func (v Vector3) Add(v2 Vector3) Vector3 {
	return Vector3{X: v.X + v2.X, Y: v.Y + v2.Y, Z: v.Z + v2.Z}
}

func (v Vector3) Sub(v2 Vector3) Vector3 {
	return Vector3{X: v.X - v2.X, Y: v.Y - v2.Y, Z: v.Z - v2.Z}
}

func (v Vector3) Scale(s float64) Vector3 {
	return Vector3{X: s * v.X, Y: s * v.Y, Z: s * v.Z}
}

func (v Vector3) Neg() Vector3 {
	return Vector3{X: -v.X, Y: -v.Y, Z: -v.Z}
}

func (v Vector3) Dot(v2 Vector3) float64 {
	return (v.X * v2.X) + (v.Y * v2.Y) + (v.Z * v2.Z)
}

func (v Vector3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vector3) LengthSq() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vector3) Normalize() Vector3 {
	mag := v.Length()
	if mag == 0 {
		return Vector3{}
	}
	return Vector3{X: v.X / mag, Y: v.Y / mag, Z: v.Z / mag}
}

func (v Vector3) Distance(v2 Vector3) float64 {
	return v.Sub(v2).Length()
}

func (v Vector3) Lerp(v2 Vector3, t float64) Vector3 {
	return v.Scale(1 - t).Add(v2.Scale(t))
}

func (v Vector3) Cross(v2 Vector3) Vector3 {
	xp := v.Y*v2.Z - v.Z*v2.Y
	yp := v.Z*v2.X - v.X*v2.Z
	zp := v.X*v2.Y - v.Y*v2.X

	return Vector3{X: xp, Y: yp, Z: zp}
}
