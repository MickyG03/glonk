package mesh

import (
	"testing"

	"github.com/kyrokohan/glonk/geometry"
)

func triArea(t [3]geometry.Vector3) float64 {
	return t[1].Sub(t[0]).Cross(t[2].Sub(t[0])).Length() / 2
}

func TestClipTriangleNear(t *testing.T) {
	const near = 0.1
	v := func(x, y, z float64) geometry.Vector3 { return geometry.Vector3{X: x, Y: y, Z: z} }

	tests := []struct {
		name     string
		a, b, c  geometry.Vector3
		wantTris int
		wantArea float64 // total area of the pieces; ignored when wantTris is 0
	}{
		{"all in front", v(0, 0, 2), v(2, 0, 3), v(0, 2, 3), 1, 0},
		{"all behind", v(0, 0, -2), v(2, 0, -1), v(0, 2, -3), 0, 0},
		// a is behind: the cut-off tip is a triangle scaled by t=0.275, so area 6*(1-0.275^2).
		{"one behind", v(0, 0, -1), v(2, 0, 3), v(0, 2, 3), 2, 6 * (1 - 0.275*0.275)},
		{"middle behind", v(2, 0, 3), v(0, 0, -1), v(0, 2, 3), 2, 6 * (1 - 0.275*0.275)},
		{"last behind", v(2, 0, 3), v(0, 2, 3), v(0, 0, -1), 2, 6 * (1 - 0.275*0.275)},
		// only a is in front: what remains is the tip scaled by t=0.725, area 6*0.725^2.
		{"two behind", v(0, 0, 3), v(2, 0, -1), v(0, 2, -1), 1, 6 * 0.725 * 0.725},
		{"two behind, first in front is last", v(2, 0, -1), v(0, 2, -1), v(0, 0, 3), 1, 6 * 0.725 * 0.725},
		{"vertex on the plane, others in front", v(0, 0, near), v(2, 0, 3), v(0, 2, 3), 1, 0},
		{"vertex on the plane, others behind", v(0, 0, near), v(2, 0, -1), v(0, 2, -1), 0, 0},
		{"edge on the plane, third behind", v(0, 0, near), v(2, 0, near), v(0, 2, -1), 0, 0},
		{"edge on the plane, third in front", v(0, 0, near), v(2, 0, near), v(0, 2, 3), 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clipTriangleNear(tt.a, tt.b, tt.c, near)

			if len(got) != tt.wantTris {
				t.Fatalf("wanted %d triangles, got %d: %v", tt.wantTris, len(got), got)
			}

			orig := tt.b.Sub(tt.a).Cross(tt.c.Sub(tt.a))
			var area float64
			for _, tri := range got {
				for _, p := range tri {
					if p.Z < near {
						t.Errorf("wanted every vertex at Z>=%v, got %v", near, p)
					}
				}
				if triArea(tri) == 0 {
					t.Errorf("wanted no degenerate triangles, got %v", tri)
				}
				if n := tri[1].Sub(tri[0]).Cross(tri[2].Sub(tri[0])); n.Dot(orig) <= 0 {
					t.Errorf("wanted winding preserved, normal %v vs original %v", n, orig)
				}
				area += triArea(tri)
			}

			if tt.wantArea != 0 && !floatEquals(area, tt.wantArea) {
				t.Errorf("wanted total area %v, got %v", tt.wantArea, area)
			}
			if tt.wantTris == 1 && tt.wantArea == 0 && !floatEquals(area, triArea([3]geometry.Vector3{tt.a, tt.b, tt.c})) {
				t.Errorf("wanted the triangle unchanged, got area %v", area)
			}
		})
	}
}

// A vertex on the plane with one in front and one behind keeps a triangle: the
// on-plane vertex, the one in front, and the crossing of the remaining edge.
func TestClipTriangleNearVertexOnPlaneSplit(t *testing.T) {
	const near = 0.1
	a := geometry.Vector3{Z: near}
	b := geometry.Vector3{X: 2, Z: 3}
	c := geometry.Vector3{Y: 2, Z: -1}

	got := clipTriangleNear(a, b, c, near)

	if len(got) != 1 {
		t.Fatalf("wanted 1 triangle, got %d: %v", len(got), got)
	}
	for _, p := range got[0] {
		if p.Z < near {
			t.Errorf("wanted every vertex at Z>=%v, got %v", near, p)
		}
	}
	if !slicesContains(got[0][:], a) || !slicesContains(got[0][:], b) {
		t.Errorf("wanted the on-plane and in-front vertices kept, got %v", got[0])
	}
}

func slicesContains(vs []geometry.Vector3, want geometry.Vector3) bool {
	for _, v := range vs {
		if v == want {
			return true
		}
	}
	return false
}
