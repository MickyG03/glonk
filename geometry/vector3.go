package geometry

import "math"

// Vector3 is a 3D vector.
type Vector3 struct {
	X, Y, Z float64
}

// Add returns v + v2.
func (v Vector3) Add(v2 Vector3) Vector3 {
	return Vector3{X: v.X + v2.X, Y: v.Y + v2.Y, Z: v.Z + v2.Z}
}

// Sub returns v - v2.
func (v Vector3) Sub(v2 Vector3) Vector3 {
	return Vector3{X: v.X - v2.X, Y: v.Y - v2.Y, Z: v.Z - v2.Z}
}

// Scale returns v multiplied by s.
func (v Vector3) Scale(s float64) Vector3 {
	return Vector3{X: s * v.X, Y: s * v.Y, Z: s * v.Z}
}

// Neg returns -v.
func (v Vector3) Neg() Vector3 {
	return Vector3{X: -v.X, Y: -v.Y, Z: -v.Z}
}

// Dot returns the dot product of v and v2.
func (v Vector3) Dot(v2 Vector3) float64 {
	return (v.X * v2.X) + (v.Y * v2.Y) + (v.Z * v2.Z)
}

// Length returns the magnitude of v.
func (v Vector3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// LengthSq returns the squared magnitude of v; cheaper than Length when only comparing.
func (v Vector3) LengthSq() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

// Normalize returns v scaled to length 1, or the zero vector if v is zero.
func (v Vector3) Normalize() Vector3 {
	mag := v.Length()
	if mag == 0 {
		return Vector3{}
	}
	return Vector3{X: v.X / mag, Y: v.Y / mag, Z: v.Z / mag}
}

// Distance returns the distance between v and v2.
func (v Vector3) Distance(v2 Vector3) float64 {
	return v.Sub(v2).Length()
}

// Lerp interpolates from v (t=0) to v2 (t=1). t is not clamped.
func (v Vector3) Lerp(v2 Vector3, t float64) Vector3 {
	return v.Scale(1 - t).Add(v2.Scale(t))
}

// Cross returns the cross product of v and v2, perpendicular to both.
func (v Vector3) Cross(v2 Vector3) Vector3 {
	xp := v.Y*v2.Z - v.Z*v2.Y
	yp := v.Z*v2.X - v.X*v2.Z
	zp := v.X*v2.Y - v.Y*v2.X

	return Vector3{X: xp, Y: yp, Z: zp}
}

// RotateX turns v around the x axis by angle radians; positive turns +y toward +z.
func (v Vector3) RotateX(angle float64) Vector3 {
	xp := v.X
	yp := v.Y*math.Cos(angle) - v.Z*math.Sin(angle)
	zp := v.Y*math.Sin(angle) + v.Z*math.Cos(angle)
	return Vector3{X: xp, Y: yp, Z: zp}
}

// RotateY turns v around the y axis by angle radians; positive turns +z toward +x.
func (v Vector3) RotateY(angle float64) Vector3 {
	xp := v.X*math.Cos(angle) + v.Z*math.Sin(angle)
	yp := v.Y
	zp := -v.X*math.Sin(angle) + v.Z*math.Cos(angle)
	return Vector3{X: xp, Y: yp, Z: zp}
}
