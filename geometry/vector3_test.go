package geometry

import (
	"math"
	"testing"
)

func assertVector3Equals(t *testing.T, want, got Vector3) {
	t.Helper()
	if !floatEquals(want.X, got.X) || !floatEquals(want.Y, got.Y) || !floatEquals(want.Z, got.Z) {
		t.Errorf("wanted %v, got %v", want, got)
	}
}

func TestVector3Dot(t *testing.T) {
	a := Vector3{1, 2, 3}
	b := Vector3{4, 5, 6}
	want := 32.00

	dot := a.Dot(b)
	if !floatEquals(dot, want) {
		t.Errorf("wanted a.Dot(b)=%v, got %v", want, dot)
	}

}

func TestVector3Cross(t *testing.T) {
	a := Vector3{1, 2, 3}
	b := Vector3{4, 5, 6}

	c := a.Cross(b)
	want := Vector3{-3, 6, -3}
	assertVector3Equals(t, want, c)

	cDotA := c.Dot(a)
	if !floatEquals(cDotA, 0) {
		t.Errorf("wanted c.Dot(a)=0.00, got %v", cDotA)
	}

	cDotB := c.Dot(b)
	if !floatEquals(cDotB, 0) {
		t.Errorf("wanted c.Dot(b)=0.00, got %v", cDotB)
	}

	bCrossA := b.Cross(a)
	assertVector3Equals(t, c, bCrossA.Scale(-1))
}

func TestVector3Normalize(t *testing.T) {
	z := Vector3{0, 0, 0}
	a := Vector3{-59, 23, 449}
	aLength := a.Length()

	zNorm := z.Normalize()
	assertVector3Equals(t, z, zNorm)

	norm := a.Normalize()
	assertVector3Equals(t, a, norm.Scale(aLength))
	if !floatEquals(norm.Length(), 1.00) {
		t.Errorf("wanted 1, got %v", norm.Length())
	}
}

func TestVector3Rotate(t *testing.T) {
	a := Vector3{1, 0, 0}
	b := Vector3{0, 1, 0}
	d := Vector3{0, 0, 1}
	wantA := Vector3{0, 0, -1}
	wantB := Vector3{0, 0, 1}
	wantC := Vector3{0, 1, 0}
	wantD := Vector3{1, 0, 0}
	wantE := Vector3{0, -1, 0}

	assertVector3Equals(t, wantA, a.RotateY(math.Pi/2))
	assertVector3Equals(t, wantB, b.RotateX(math.Pi/2))
	assertVector3Equals(t, wantC, b.RotateY(math.Pi/3))
	assertVector3Equals(t, wantD, d.RotateY(math.Pi/2))
	assertVector3Equals(t, wantE, d.RotateX(math.Pi/2))
}

func TestVector3RotateZ(t *testing.T) {
	// Positive turns +x toward +y, and a point on the axis stays put.
	assertVector3Equals(t, Vector3{0, 1, 0}, Vector3{1, 0, 0}.RotateZ(math.Pi/2))
	assertVector3Equals(t, Vector3{-1, 0, 0}, Vector3{0, 1, 0}.RotateZ(math.Pi/2))
	assertVector3Equals(t, Vector3{0, 0, 5}, Vector3{0, 0, 5}.RotateZ(1.3))
	assertVector3Equals(t, Vector3{-1, -2, 3}, Vector3{1, 2, 3}.RotateZ(math.Pi))
	assertVector3Equals(t, Vector3{1, 2, 3}, Vector3{1, 2, 3}.RotateZ(0))
}
