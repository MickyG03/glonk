package mesh

import (
	"math"
	"slices"
	"testing"
)

func TestGrid(t *testing.T) {
	tests := []struct {
		name      string
		size      float64
		divisions int
		lines     int // lines per direction after clamping
	}{
		{"even split", 20, 4, 5},
		{"unit cells", 40, 20, 21},
		{"fractional step", 10, 3, 4},
		{"odd size", 5, 2, 3},
		{"zero divisions clamps to one", 8, 0, 2},
		{"negative divisions clamps to one", 8, -3, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := Grid(tt.size, tt.divisions)
			half := tt.size / 2

			if len(g.Vertices) != 4*tt.lines {
				t.Fatalf("wanted %d vertices, got %d", 4*tt.lines, len(g.Vertices))
			}
			if len(g.Edges) != 2*tt.lines {
				t.Fatalf("wanted %d edges, got %d", 2*tt.lines, len(g.Edges))
			}

			for _, v := range g.Vertices {
				if v.Y != 0 {
					t.Errorf("wanted Y=0, got %v", v)
				}
				if math.Abs(v.X) > half+1e-9 || math.Abs(v.Z) > half+1e-9 {
					t.Errorf("wanted vertex within +-%v, got %v", half, v)
				}
			}

			// xs are the x positions of lines running along z, zs the z positions of lines running along x.
			var xs, zs []float64
			for _, e := range g.Edges {
				a, b := g.Vertices[e[0]], g.Vertices[e[1]]

				if !floatEquals(a.Distance(b), tt.size) {
					t.Errorf("wanted edge length %v, got %v for %v-%v", tt.size, a.Distance(b), a, b)
				}

				switch {
				case a.X == b.X && a.Z != b.Z:
					xs = append(xs, a.X)
				case a.Z == b.Z && a.X != b.X:
					zs = append(zs, a.Z)
				default:
					t.Errorf("wanted an edge along exactly one axis, got %v-%v", a, b)
				}
			}

			// Evenly spaced from one side to the other; also catches a step that grows with i.
			for name, got := range map[string][]float64{"x": xs, "z": zs} {
				slices.Sort(got)
				if len(got) != tt.lines {
					t.Fatalf("wanted %d lines along %s, got %d", tt.lines, name, len(got))
				}
				for i, c := range got {
					want := -half + float64(i)*tt.size/float64(tt.lines-1)
					if !floatEquals(c, want) {
						t.Errorf("wanted %s line %d at %v, got %v", name, i, want, c)
					}
				}
			}
		})
	}
}
