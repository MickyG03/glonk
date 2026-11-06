package geometry

import "math"

// Vector2 is a 2D vector.
type Vector2 struct {
	X, Y float64
}

// Add returns v + v2.
func (v Vector2) Add(v2 Vector2) Vector2 {
	return Vector2{X: v.X + v2.X, Y: v.Y + v2.Y}
}

// Sub returns v - v2.
func (v Vector2) Sub(v2 Vector2) Vector2 {
	return Vector2{X: v.X - v2.X, Y: v.Y - v2.Y}
}

// Scale returns v multiplied by s.
func (v Vector2) Scale(s float64) Vector2 {
	return Vector2{X: s * v.X, Y: s * v.Y}
}

// Neg returns -v.
func (v Vector2) Neg() Vector2 {
	return Vector2{X: -v.X, Y: -v.Y}
}

// Dot returns the dot product of v and v2.
func (v Vector2) Dot(v2 Vector2) float64 {
	return (v.X * v2.X) + (v.Y * v2.Y)
}

// Length returns the magnitude of v.
func (v Vector2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// LengthSq returns the squared magnitude of v; cheaper than Length when only comparing.
func (v Vector2) LengthSq() float64 {
	return v.X*v.X + v.Y*v.Y
}

// Normalize returns v scaled to length 1, or the zero vector if v is zero.
func (v Vector2) Normalize() Vector2 {
	mag := v.Length()
	if mag == 0 {
		return Vector2{}
	}
	return Vector2{X: v.X / mag, Y: v.Y / mag}
}

// Distance returns the distance between v and v2.
func (v Vector2) Distance(v2 Vector2) float64 {
	return v.Sub(v2).Length()
}

// Lerp interpolates from v (t=0) to v2 (t=1). t is not clamped.
func (v Vector2) Lerp(v2 Vector2, t float64) Vector2 {
	return v.Scale(1 - t).Add(v2.Scale(t))
}

// Rotate turns v by a radians, counterclockwise when y points up.
func (v Vector2) Rotate(a float64) Vector2 {
	xp := v.X*math.Cos(a) - v.Y*math.Sin(a)
	yp := v.X*math.Sin(a) + v.Y*math.Cos(a)
	return Vector2{X: xp, Y: yp}
}

// Perp returns the part of v perpendicular to v2 (v minus its projection onto v2),
// or v if v2 is zero.
func (v Vector2) Perp(v2 Vector2) Vector2 {
	dot1 := v2.Dot(v)
	dot2 := v2.Dot(v2)
	if dot2 == 0 {
		return v
	}

	return v.Sub(v2.Scale(dot1 / dot2))
}
