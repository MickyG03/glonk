package geometry

import (
	"math"
	"testing"
)

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func assertVector2Equals(t *testing.T, want, got Vector2) {
	t.Helper()
	if !floatEquals(want.X, got.X) || !floatEquals(want.Y, got.Y) {
		t.Errorf("wanted %v, got %v", want, got)
	}
}

func TestVector2Perp(t *testing.T) {
	a := Vector2{1, 1}
	b := Vector2{1, 0}
	want := Vector2{0, 1}

	perp := a.Perp(b)
	assertVector2Equals(t, want, perp)

	zeroRes := perp.Dot(b)
	if !floatEquals(zeroRes, 0) {
		t.Errorf("wanted result.Dot(v2)=0, got %v", zeroRes)
	}

	zeroRes2 := a.Perp(Vector2{0, 0})
	assertVector2Equals(t, a, zeroRes2)
}

func TestVector2Rotate(t *testing.T) {
	a := Vector2{1, 0}
	angle := math.Pi / 2
	want := Vector2{0, 1}

	got := a.Rotate(angle)
	assertVector2Equals(t, want, got)
}

func TestVector2Lerp(t *testing.T) {
	a := Vector2{1, 1}
	b := Vector2{2, 5}
	mp := Vector2{1.5, 3}

	t0 := a.Lerp(b, 0)
	assertVector2Equals(t, a, t0)

	tmp := a.Lerp(b, 0.5)
	assertVector2Equals(t, mp, tmp)

	t1 := a.Lerp(b, 1)
	assertVector2Equals(t, b, t1)
}
