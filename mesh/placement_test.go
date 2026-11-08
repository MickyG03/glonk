package mesh

import (
	"math"
	"testing"

	"github.com/kyrokohan/glonk/geometry"
)

// The doc comments of Mesh and Camera say which way their angles turn and in what
// order they apply. This keeps those sentences true.
func TestMeshPlacementMatchesItsDocs(t *testing.T) {
	const quarter = math.Pi / 2
	pos := geometry.Vector3{X: 10, Y: 20, Z: 30}

	tests := []struct {
		name string
		m    Mesh
		in   geometry.Vector3
		want geometry.Vector3
	}{
		{"positive Pitch turns +y toward +z", Mesh{Pitch: quarter}, geometry.Vector3{Y: 1}, geometry.Vector3{Z: 1}},
		{"positive Yaw turns +z toward +x", Mesh{Yaw: quarter}, geometry.Vector3{Z: 1}, geometry.Vector3{X: 1}},
		{"Position moves the model", Mesh{Position: pos}, geometry.Vector3{X: 1}, geometry.Vector3{X: 11, Y: 20, Z: 30}},
		// Pitch first: (0,1,0) goes to (0,0,1), then Yaw takes that to (1,0,0). Yaw first
		// would leave (0,1,0) alone and Pitch would then end at (0,0,1).
		{"Pitch applies before Yaw", Mesh{Pitch: quarter, Yaw: quarter}, geometry.Vector3{Y: 1}, geometry.Vector3{X: 1}},
		{"Position applies after both", Mesh{Pitch: quarter, Yaw: quarter, Position: pos}, geometry.Vector3{Y: 1}, geometry.Vector3{X: 11, Y: 20, Z: 30}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.toWorld(tt.in); !vecEquals(got, tt.want) {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCameraAnglesMatchTheirDocs(t *testing.T) {
	const eighth = math.Pi / 4

	if f := (Camera{Yaw: eighth}).Forward(); !(f.X > 0.5 && f.Z > 0.5) {
		t.Errorf("wanted a positive Yaw to turn the camera right, toward +x, got %v", f)
	}
	if l := (Camera{Pitch: eighth}).Look(); !(l.Y > 0.5 && l.Z > 0.5) {
		t.Errorf("wanted a positive Pitch to tilt the camera up, got %v", l)
	}
	if f := (Camera{}).Forward(); !vecEquals(f, geometry.Vector3{Z: 1}) {
		t.Errorf("wanted a camera at Yaw 0 to face +z, got %v", f)
	}
}
