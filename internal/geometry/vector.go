package geometry

import "math"

type Vector2 struct {
	X, Y float64
}

func (v Vector2) Add(v2 Vector2) Vector2 {
	return Vector2{X: v.X + v2.X, Y: v.Y + v2.Y}
}

func (v Vector2) Sub(v2 Vector2) Vector2 {
	return Vector2{X: v.X - v2.X, Y: v.Y - v2.Y}
}

func (v Vector2) Scale(s float64) Vector2 {
	return Vector2{X: s * v.X, Y: s * v.Y}
}

func (v Vector2) Neg() Vector2 {
	return Vector2{X: -v.X, Y: -v.Y}
}

func (v Vector2) Dot(v2 Vector2) float64 {
	return (v.X * v2.X) + (v.Y * v2.Y)
}

func (v Vector2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2) LengthSq() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v Vector2) Normalize() Vector2 {
	mag := v.Length()
	if mag == 0 {
		return Vector2{}
	}
	return Vector2{X: v.X / mag, Y: v.Y / mag}
}

func (v Vector2) Distance(v2 Vector2) float64 {
	return v.Sub(v2).Length()
}

func (v Vector2) Lerp(v2 Vector2, t float64) Vector2 {
	return v.Scale(1 - t).Add(v2.Scale(t))
}

func (v Vector2) Rotate(a float64) Vector2 {
	xp := v.X*math.Cos(a) - v.Y*math.Sin(a)
	yp := v.X*math.Sin(a) + v.Y*math.Cos(a)
	return Vector2{X: xp, Y: yp}
}

func (v Vector2) Perp(v2 Vector2) Vector2 {
	dot1 := v2.Dot(v)
	dot2 := v2.Dot(v2)
	if dot2 == 0 {
		return v
	}

	return v.Sub(v2.Scale(dot1 / dot2))
}

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
