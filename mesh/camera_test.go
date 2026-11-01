package mesh

import (
	"math"
	"testing"

	"github.com/kyrokohan/glonk/geometry"
)

type cameraOutput struct {
	x, y float64
	ok   bool
}

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestCameraProject(t *testing.T) {
	var tests = []struct {
		name  string
		yaw   float64
		point geometry.Vector3
		want  cameraOutput
	}{
		{"straight ahead", 0.00, geometry.Vector3{X: 0, Y: 0, Z: 5}, cameraOutput{40, 12, true}},
		{"behind", 0.00, geometry.Vector3{X: 0, Y: 0, Z: -5}, cameraOutput{0, 0, false}},
		{"right", 0.00, geometry.Vector3{X: 1, Y: 0, Z: 2}, cameraOutput{52, 12, true}},
		{"up", 0.00, geometry.Vector3{X: 0, Y: 1, Z: 2}, cameraOutput{40, 6, true}},
		{"turned", math.Pi / 2, geometry.Vector3{X: 5, Y: 0, Z: 0}, cameraOutput{40, 12, true}},
		{"inside near plane", 0.00, geometry.Vector3{X: 0, Y: 0, Z: 0.05}, cameraOutput{0, 0, false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Camera{
				Position: geometry.Vector3{X: 0, Y: 0, Z: 0},
				FOV:      math.Pi / 2,
				Yaw:      tt.yaw,
				Near:     0.1,
			}

			x, y, ok := c.Project(tt.point, 80, 24)
			if ok != tt.want.ok {
				t.Fatalf("wanted ok=%v, got %v", tt.want.ok, ok)
			}
			if !ok {
				return
			}

			if !floatEquals(x, tt.want.x) {
				t.Errorf("wanted x=%v, got %v", tt.want.x, x)
			}

			if !floatEquals(y, tt.want.y) {
				t.Errorf("wanted y=%v, got %v", tt.want.y, y)
			}
		})
	}
}

func vecEquals(a, b geometry.Vector3) bool {
	return floatEquals(a.X, b.X) && floatEquals(a.Y, b.Y) && floatEquals(a.Z, b.Z)
}

func TestCameraForwardRight(t *testing.T) {
	tests := []struct {
		name           string
		yaw            float64
		forward, right geometry.Vector3
	}{
		{"yaw 0", 0, geometry.Vector3{Z: 1}, geometry.Vector3{X: 1}},
		{"yaw quarter turn", math.Pi / 2, geometry.Vector3{X: 1}, geometry.Vector3{Z: -1}},
		{"yaw half turn", math.Pi, geometry.Vector3{Z: -1}, geometry.Vector3{X: -1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cam := Camera{Yaw: tt.yaw}

			if got := cam.Forward(); !vecEquals(got, tt.forward) {
				t.Errorf("wanted Forward()=%v, got %v", tt.forward, got)
			}
			if got := cam.Right(); !vecEquals(got, tt.right) {
				t.Errorf("wanted Right()=%v, got %v", tt.right, got)
			}
		})
	}
}

// Forward and Right must agree with Project, or WASD moves the wrong way.
func TestCameraForwardRightMatchProject(t *testing.T) {
	const w, h = 80, 24

	for _, yaw := range []float64{0, 0.7, math.Pi / 2, 2.5, -1.2} {
		cam := Camera{
			Position: geometry.Vector3{X: 3, Y: 1, Z: -2},
			Yaw:      yaw,
			FOV:      math.Pi / 2,
			Near:     0.1,
		}
		ahead := cam.Position.Add(cam.Forward().Scale(5))

		x, y, ok := cam.Project(ahead, w, h)
		if !ok || !floatEquals(x, w/2) || !floatEquals(y, h/2) {
			t.Errorf("yaw %v: wanted point ahead at center (%v,%v,true), got (%v,%v,%v)", yaw, w/2, h/2, x, y, ok)
		}

		x, _, ok = cam.Project(ahead.Add(cam.Right()), w, h)
		if !ok || x <= w/2 {
			t.Errorf("yaw %v: wanted point to the right at x>%v, got x=%v ok=%v", yaw, w/2, x, ok)
		}
	}
}
