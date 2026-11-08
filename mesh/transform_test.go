package mesh

import (
	"math"
	"testing"

	"github.com/kyrokohan/glonk/geometry"
)

// The doc comments of Transform and Camera say which way their angles turn and in
// what order they apply. This keeps those sentences true.
func TestTransformMatchesItsDocs(t *testing.T) {
	const quarter = math.Pi / 2
	pos := geometry.Vector3{X: 10, Y: 20, Z: 30}

	tests := []struct {
		name string
		t    Transform
		in   geometry.Vector3
		want geometry.Vector3
	}{
		{"positive Pitch turns +z toward +y", Transform{Pitch: quarter}, geometry.Vector3{Z: 1}, geometry.Vector3{Y: 1}},
		{"positive Yaw turns +z toward +x", Transform{Yaw: quarter}, geometry.Vector3{Z: 1}, geometry.Vector3{X: 1}},
		{"Position moves the model", Transform{Position: pos}, geometry.Vector3{X: 1}, geometry.Vector3{X: 11, Y: 20, Z: 30}},
		// Pitch first: (0,0,1) goes to (0,1,0), which Yaw leaves alone. Yaw first would
		// take (0,0,1) to (1,0,0), and Pitch would then leave that alone.
		{"Pitch applies before Yaw", Transform{Pitch: quarter, Yaw: quarter}, geometry.Vector3{Z: 1}, geometry.Vector3{Y: 1}},
		{"Position applies after both", Transform{Pitch: quarter, Yaw: quarter, Position: pos}, geometry.Vector3{Z: 1}, geometry.Vector3{X: 10, Y: 21, Z: 30}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.Apply(tt.in); !vecEquals(got, tt.want) {
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

// A model and a camera given the same Pitch and Yaw tip and turn the same way: the
// model's front points where the camera looks.
func TestTransformAndCameraAgreeOnPitchAndYaw(t *testing.T) {
	for _, pitch := range []float64{-1.2, -0.4, 0, 0.4, 1.2} {
		for _, yaw := range []float64{-2, 0, 0.7, 3} {
			front := Transform{Pitch: pitch, Yaw: yaw}.Apply(geometry.Vector3{Z: 1})
			look := Camera{Pitch: pitch, Yaw: yaw}.Look()
			if !vecEquals(front, look) {
				t.Errorf("pitch %v yaw %v: wanted the model's front %v to match the camera's look %v", pitch, yaw, front, look)
			}
		}
	}
}
