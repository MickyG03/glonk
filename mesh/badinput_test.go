package mesh

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

// drawEverything draws m with all four ways to draw a mesh into a fresh frame.
func drawEverything(t *testing.T, m Mesh) *engine.FrameBuffer {
	t.Helper()

	fb, err := engine.NewFrameBuffer(40, 12)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	cam := Camera{FOV: math.Pi / 2, Near: 0.1}
	light := Light{Direction: geometry.Vector3{Y: -1}, Ambient: 0.2}

	m.Draw(fb, cam, engine.Cell{Character: '#'})
	m.DoubleSided = true
	m.DrawShaded(fb, cam, light, 1)
	m.DrawShadedPixels(fb.Canvas(), cam, light, 2)
	fb.Canvas().Compose()
	return fb
}

func drawnCells(fb *engine.FrameBuffer) int {
	n := 0
	for y := range fb.Height() {
		for x := range fb.Width() {
			if fb.At(x, y) != engine.BlankCell {
				n++
			}
		}
	}
	return n
}

// triangle is three vertices in front of the camera.
func triangle() []geometry.Vector3 {
	return []geometry.Vector3{{X: -2, Y: -1, Z: 5}, {X: 2, Y: -1, Z: 5}, {X: 0, Y: 2, Z: 5}}
}

func TestDrawSkipsEdgesAndFacesThatNameMissingVertices(t *testing.T) {
	good := Mesh{Vertices: triangle(), Edges: [][2]int{{0, 1}, {1, 2}}, Faces: [][3]int{{0, 1, 2}}}
	want := drawEverything(t, good)
	if drawnCells(want) == 0 {
		t.Fatalf("wanted the good mesh to draw something")
	}

	badEdges := [][2]int{{1, 99}, {-1, 2}, {3, 0}, {0, 3}, {0, math.MaxInt}, {math.MinInt, 1}, {-5, 77}}
	badFaces := [][3]int{{0, 1, 3}, {-1, 1, 2}, {0, 99, 2}, {math.MaxInt, 0, 1}, {0, 1, math.MinInt}, {5, 6, 7}}

	for _, e := range badEdges {
		t.Run(fmt.Sprintf("edge %v", e), func(t *testing.T) {
			m := good
			m.Edges = append(append([][2]int{e}, good.Edges...), e) // first and last
			if got := drawEverything(t, m); !sameFrame(got, want) {
				t.Errorf("wanted the bad edge %v skipped and the rest drawn as before", e)
			}
		})
	}
	for _, f := range badFaces {
		t.Run(fmt.Sprintf("face %v", f), func(t *testing.T) {
			m := good
			m.Faces = append(append([][3]int{f}, good.Faces...), f)
			if got := drawEverything(t, m); !sameFrame(got, want) {
				t.Errorf("wanted the bad face %v skipped and the rest drawn as before", f)
			}
		})
	}
}

func TestDrawWithoutVerticesDrawsNothing(t *testing.T) {
	m := Mesh{Edges: [][2]int{{0, 1}}, Faces: [][3]int{{0, 1, 2}}}
	if n := drawnCells(drawEverything(t, m)); n != 0 {
		t.Errorf("wanted nothing drawn without vertices, got %d cells", n)
	}
}

func TestValidate(t *testing.T) {
	nan := math.NaN()

	tests := []struct {
		name string
		m    Mesh
		want string // "" means valid
	}{
		{"empty mesh", Mesh{}, ""},
		{"cube", Cube(2), ""},
		{"grid", Grid(10, 4), ""},
		{"triangle", Mesh{Vertices: triangle(), Edges: [][2]int{{0, 1}}, Faces: [][3]int{{0, 1, 2}}}, ""},
		{"vertices only", Mesh{Vertices: triangle()}, ""},
		{"last vertex is fine", Mesh{Vertices: triangle(), Edges: [][2]int{{2, 2}}}, ""},

		{"edge past the end", Mesh{Vertices: triangle(), Edges: [][2]int{{0, 1}, {1, 3}}}, "mesh: edge 1 uses vertex 3, but only vertices 0 to 2 exist"},
		{"edge first end", Mesh{Vertices: triangle(), Edges: [][2]int{{9, 1}}}, "mesh: edge 0 uses vertex 9, but only vertices 0 to 2 exist"},
		{"edge negative", Mesh{Vertices: triangle(), Edges: [][2]int{{0, -1}}}, "mesh: edge 0 uses vertex -1, but only vertices 0 to 2 exist"},
		{"face past the end", Mesh{Vertices: triangle(), Faces: [][3]int{{0, 1, 2}, {0, 1, 3}}}, "mesh: face 1 uses vertex 3, but only vertices 0 to 2 exist"},
		{"face first corner", Mesh{Vertices: triangle(), Faces: [][3]int{{3, 1, 2}}}, "mesh: face 0 uses vertex 3, but only vertices 0 to 2 exist"},
		{"face negative", Mesh{Vertices: triangle(), Faces: [][3]int{{0, math.MinInt, 2}}}, "mesh: face 0 uses vertex -9223372036854775808, but only vertices 0 to 2 exist"},
		{"edges without vertices", Mesh{Edges: [][2]int{{0, 1}}}, "mesh: edge 0 uses vertex 0, but the mesh has no vertices"},
		{"faces without vertices", Mesh{Faces: [][3]int{{0, 1, 2}}}, "mesh: face 0 uses vertex 0, but the mesh has no vertices"},
		{"edge reported before face", Mesh{Vertices: triangle(), Edges: [][2]int{{0, 8}}, Faces: [][3]int{{0, 1, 9}}}, "mesh: edge 0 uses vertex 8, but only vertices 0 to 2 exist"},

		{"NaN vertex", Mesh{Vertices: []geometry.Vector3{{}, {Y: nan}}}, "mesh: vertex 1 is not a finite number: {0 NaN 0}"},
		{"infinite vertex", Mesh{Vertices: []geometry.Vector3{{Z: math.Inf(-1)}}}, "mesh: vertex 0 is not a finite number: {0 0 -Inf}"},
		{"vertex reported before edge", Mesh{Vertices: []geometry.Vector3{{X: nan}}, Edges: [][2]int{{0, 5}}}, "mesh: vertex 0 is not a finite number: {NaN 0 0}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.m.Validate()
			switch {
			case tt.want == "" && err != nil:
				t.Errorf("wanted a valid mesh, got %v", err)
			case tt.want != "" && err == nil:
				t.Errorf("wanted error %q, got nil", tt.want)
			case tt.want != "" && err.Error() != tt.want:
				t.Errorf("wanted error %q, got %q", tt.want, err)
			}
		})
	}
}

func TestValidateAcceptsWhatLoadOBJReturns(t *testing.T) {
	obj := "v 0 0 0\nv 1 0 0\nv 1 1 0\nv 0 1 0\nf 1 2 3 4\nf -4 -3 -2\nl 1 3\n"
	m, err := LoadOBJ(strings.NewReader(obj))
	if err != nil {
		t.Fatalf("unable to load: %v", err)
	}
	if err := m.Validate(); err != nil {
		t.Errorf("wanted a loaded model to be valid, got %v", err)
	}
}
