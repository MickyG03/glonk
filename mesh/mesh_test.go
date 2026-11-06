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

func TestClipNear(t *testing.T) {
	const near = 0.1

	tests := []struct {
		name    string
		out, in geometry.Vector3
		want    geometry.Vector3
	}{
		{
			"straight along z",
			geometry.Vector3{Z: -1}, geometry.Vector3{Z: 3},
			geometry.Vector3{Z: near},
		},
		{
			"diagonal",
			geometry.Vector3{X: -2, Y: 4, Z: -1}, geometry.Vector3{X: 2, Y: 0, Z: 3},
			geometry.Vector3{X: -0.9, Y: 2.9, Z: near},
		},
		{
			"in end exactly on the plane",
			geometry.Vector3{X: 1, Y: 1, Z: -1}, geometry.Vector3{X: 3, Y: 3, Z: near},
			geometry.Vector3{X: 3, Y: 3, Z: near},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clipNear(tt.out, tt.in, near)

			// Exact, not epsilon: projectView would misread a Z a hair under near.
			if got.Z != near {
				t.Errorf("wanted Z=%v exactly, got %v", near, got.Z)
			}
			if !vecEquals(got, tt.want) {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

func TestMeshDrawNearPlane(t *testing.T) {
	cam := Camera{FOV: math.Pi / 2, Near: 0.1}
	marker := engine.Cell{Character: '#'}

	// All edges sit on the row at the screen's center (y=12) unless noted.
	tests := []struct {
		name  string
		verts []geometry.Vector3
		lit   func(x, y int) bool
	}{
		{
			// Far end lands at x=46; the cut end is far off the right edge.
			"first end behind",
			[]geometry.Vector3{{X: 1, Z: -1}, {X: 1, Z: 4}},
			func(x, y int) bool { return y == 12 && x >= 46 },
		},
		{
			"second end behind",
			[]geometry.Vector3{{X: 1, Z: 4}, {X: 1, Z: -1}},
			func(x, y int) bool { return y == 12 && x >= 46 },
		},
		{
			// Both ends exactly on the plane: dividing to cut them would give NaN.
			"edge lying on the near plane",
			[]geometry.Vector3{{X: -1, Z: cam.Near}, {X: 1, Z: cam.Near}},
			func(x, y int) bool { return y == 12 },
		},
		{
			// A cut here would extrapolate to the screen center and draw a stray cell.
			"edge entirely behind",
			[]geometry.Vector3{{Z: -1}, {Z: -2}},
			func(x, y int) bool { return false },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := engine.NewFrameBuffer(80, 24)
			if err != nil {
				t.Fatalf("unable to initialize frame buffer: %v", err)
			}

			m := Mesh{Vertices: tt.verts, Edges: [][2]int{{0, 1}}}
			m.Draw(fb, cam, marker)

			for y := range fb.Height() {
				for x := range fb.Width() {
					want := engine.BlankCell
					if tt.lit(x, y) {
						want = marker
					}
					if got := fb.At(x, y); got != want {
						t.Fatalf("wanted %v at (%d,%d), got %v", want, x, y, got)
					}
				}
			}
		})
	}
}
