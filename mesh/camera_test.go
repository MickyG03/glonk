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
