package mesh

import (
	"math"
	"testing"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

func TestCube(t *testing.T) {
	size := 2.00
	cube := Cube(size)

	vLen := len(cube.Vertices)
	if vLen != 8 {
		t.Errorf("wanted len(vertices)=8, got %d", vLen)
	}

	eLen := len(cube.Edges)
	if eLen != 12 {
		t.Errorf("wanted len(edges)=12, got %d", eLen)
	}

	seen := map[[2]int]bool{}

	for _, e := range cube.Edges {
		key := e
		if e[0] > e[1] {
			key[0] = e[1]
			key[1] = e[0]
		}
		if seen[key] {
			t.Errorf("duplicate edge %v", e)
		}
		seen[key] = true

		a := cube.Vertices[e[0]]
		b := cube.Vertices[e[1]]
		aDistB := a.Distance(b)
		if !floatEquals(aDistB, size) {
			t.Errorf("wanted edge distance=%f, got %v from edge %v", size, aDistB, e)
		}
	}
}

func TestMeshDrawBehindCamera(t *testing.T) {
	cam := Camera{FOV: math.Pi / 2, Near: 0.1}

	fb, err := engine.NewFrameBuffer(80, 40)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	marker := engine.Cell{Character: '#'}

	cube := Cube(2)
	cube.Position = geometry.Vector3{Z: -5}
	cube.Draw(fb, cam, marker)

	for y := range fb.Height() {
		for x := range fb.Width() {
			if fb.At(x, y) != engine.BlankCell {
				t.Errorf("wanted BlankCell, got %v at (%d, %d)", fb.At(x, y), x, y)
				return
			}
		}
	}
}

func TestMeshDrawInFront(t *testing.T) {
	cam := Camera{FOV: math.Pi / 2, Near: 0.1}

	fb, err := engine.NewFrameBuffer(80, 40)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	marker := engine.Cell{Character: '#'}

	cube := Cube(2)
	cube.Position = geometry.Vector3{Z: 5}
	cube.Draw(fb, cam, marker)

	world := cube.Vertices[7].Add(cube.Position)
	x, y, ok := cam.Project(world, fb.Width(), fb.Height())
	if !ok {
		t.Fatalf("camera setup is incorrect")
	}
	cell := fb.At(int(math.Floor(x)), int(math.Floor(y)))
	if cell != marker {
		t.Errorf("wanted cell=%v, got %v", marker, cell)
	}
}
