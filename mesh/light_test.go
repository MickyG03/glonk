package mesh

import (
	"fmt"
	"math"
	"testing"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

func TestShadeChar(t *testing.T) {
	tests := []struct {
		name      string
		intensity float64
		want      rune
	}{
		{"darkest", 0, '.'},
		{"brightest", 1, '@'},
		{"middle", 0.5, '+'},
		{"just above darkest rounds down", 0.05, '.'},
		{"just below brightest rounds up", 0.95, '@'},
		{"below range clamps", -3, '.'},
		{"above range clamps", 3, '@'},
		{"NaN is dark", math.NaN(), '.'},
		{"negative infinity clamps", math.Inf(-1), '.'},
		{"positive infinity clamps", math.Inf(1), '@'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shadeChar(tt.intensity); got != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, got)
			}
		})
	}

	// A lit face must never turn into the blank background.
	for i := 0; i <= 100; i++ {
		if got := shadeChar(float64(i) / 100); got == ' ' {
			t.Fatalf("wanted no blank in the ramp, got one at %d%%", i)
		}
	}
}

func TestLightIntensity(t *testing.T) {
	down := geometry.Vector3{Y: -1}
	up := geometry.Vector3{Y: 1}
	side := geometry.Vector3{X: 1}
	diag := geometry.Vector3{X: math.Sqrt2 / 2, Y: math.Sqrt2 / 2}

	tests := []struct {
		name   string
		light  Light
		normal geometry.Vector3
		want   float64
	}{
		{"facing the light", Light{Direction: down, Ambient: 0.2}, up, 1},
		{"turned away", Light{Direction: down, Ambient: 0.2}, down, 0.2},
		{"edge on", Light{Direction: down, Ambient: 0.2}, side, 0.2},
		{"at 45 degrees", Light{Direction: down, Ambient: 0.2}, diag, 0.2 + 0.8*math.Sqrt2/2},
		{"direction length does not matter", Light{Direction: down.Scale(5), Ambient: 0.2}, up, 1},
		{"direction length does not matter at an angle", Light{Direction: down.Scale(0.01), Ambient: 0.2}, diag, 0.2 + 0.8*math.Sqrt2/2},
		{"no ambient, away", Light{Direction: down}, down, 0},
		{"no ambient, facing", Light{Direction: down}, up, 1},
		{"zero direction is ambient only", Light{Ambient: 0.3}, up, 0.3},
		{"full ambient", Light{Direction: down, Ambient: 1}, down, 1},
		{"ambient above 1 clamps", Light{Direction: down, Ambient: 4}, down, 1},
		{"negative ambient clamps", Light{Direction: down, Ambient: -2}, down, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.light.intensity(tt.normal); !floatEquals(got, tt.want) {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

func TestFrontFacing(t *testing.T) {
	tests := []struct {
		name   string
		normal geometry.Vector3
		v0     geometry.Vector3
		want   bool
	}{
		{"straight at the camera", geometry.Vector3{Z: -1}, geometry.Vector3{Z: 5}, true},
		{"straight away", geometry.Vector3{Z: 1}, geometry.Vector3{Z: 5}, false},
		{"edge on", geometry.Vector3{X: 1}, geometry.Vector3{Z: 5}, false},
		{"off to the side, turned toward the camera", geometry.Vector3{X: -1}, geometry.Vector3{X: 10, Z: 5}, true},
		{"off to the side, turned away", geometry.Vector3{X: 1}, geometry.Vector3{X: 10, Z: 5}, false},
		{"normal length does not matter", geometry.Vector3{Z: -100}, geometry.Vector3{Z: 5}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := frontFacing(tt.normal, tt.v0); got != tt.want {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

// ---- drawing ----

const shadeW, shadeH = 80, 24

func shadeCam() Camera {
	return Camera{FOV: math.Pi / 2, Near: 0.1}
}

func newShadeBuffer(t *testing.T) *engine.FrameBuffer {
	t.Helper()
	fb, err := engine.NewFrameBuffer(shadeW, shadeH)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	return fb
}

// drawn lists the coordinates of every cell that is not blank.
func drawn(fb *engine.FrameBuffer) [][2]int {
	var cells [][2]int
	for y := range fb.Height() {
		for x := range fb.Width() {
			if fb.At(x, y) != engine.BlankCell {
				cells = append(cells, [2]int{x, y})
			}
		}
	}
	return cells
}

func TestDrawShadedCullsBackFaces(t *testing.T) {
	verts := []geometry.Vector3{{X: -1, Y: -1, Z: 5}, {X: 1, Y: -1, Z: 5}, {X: 0, Y: 1, Z: 5}}
	light := Light{Direction: geometry.Vector3{Z: 1}, Ambient: 0.1}

	tests := []struct {
		name string
		face [3]int
		want bool
	}{
		// (b-a)x(c-a) is +z for 0,1,2: pointing away from a camera at the origin.
		{"wound away from the camera", [3]int{0, 1, 2}, false},
		{"wound toward the camera", [3]int{0, 2, 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := newShadeBuffer(t)
			m := Mesh{Vertices: verts, Faces: [][3]int{tt.face}}

			m.DrawShaded(fb, shadeCam(), light, 7)

			if got := len(drawn(fb)) > 0; got != tt.want {
				t.Errorf("wanted drawn=%v, got %v", tt.want, got)
			}
		})
	}
}

func TestDrawShadedUsesColorAndShade(t *testing.T) {
	fb := newShadeBuffer(t)
	cube := Cube(2)
	cube.Position = geometry.Vector3{Z: 5}
	// The light travels along +z, straight into the face that looks back at the camera.
	light := Light{Direction: geometry.Vector3{Z: 1}}

	cube.DrawShaded(fb, shadeCam(), light, 208)

	cells := drawn(fb)
	if len(cells) == 0 {
		t.Fatalf("wanted the cube drawn, got a blank frame")
	}
	for _, c := range cells {
		got := fb.At(c[0], c[1])
		if got.Color != 208 {
			t.Errorf("wanted color 208 at %v, got %v", c, got.Color)
		}
		// Seen head on, only the front face shows, lit at full strength.
		if got.Character != '@' {
			t.Errorf("wanted '@' at %v, got %q", c, got.Character)
		}
	}
}

func TestDrawShadedFacesDiffer(t *testing.T) {
	fb := newShadeBuffer(t)
	cube := Cube(2)
	cube.Position = geometry.Vector3{Z: 5}
	cube.Yaw, cube.Pitch = 0.6, 0.4
	light := Light{Direction: geometry.Vector3{X: 0.3, Y: -1, Z: 0.5}, Ambient: 0.1}

	cube.DrawShaded(fb, shadeCam(), light, 1)

	shades := map[rune]bool{}
	for _, c := range drawn(fb) {
		shades[fb.At(c[0], c[1]).Character] = true
	}
	// Turned to show three faces, each lit differently; never more than three faces show.
	if len(shades) < 2 || len(shades) > 3 {
		t.Errorf("wanted 2 or 3 distinct shades, got %d: %v", len(shades), shades)
	}
}

func TestDrawShadedDepthOrder(t *testing.T) {
	near := Cube(2)
	near.Position = geometry.Vector3{Z: 4}
	far := Cube(2)
	far.Position = geometry.Vector3{Z: 10}
	light := Light{Direction: geometry.Vector3{Z: 1}, Ambient: 0.2}

	for _, nearFirst := range []bool{true, false} {
		fb := newShadeBuffer(t)
		if nearFirst {
			near.DrawShaded(fb, shadeCam(), light, 1)
			far.DrawShaded(fb, shadeCam(), light, 2)
		} else {
			far.DrawShaded(fb, shadeCam(), light, 2)
			near.DrawShaded(fb, shadeCam(), light, 1)
		}

		// The far cube sits entirely behind the near one on screen.
		if got := fb.At(shadeW/2, shadeH/2).Color; got != 1 {
			t.Errorf("nearFirst=%v: wanted the near cube at the centre, got color %d", nearFirst, got)
		}
		for _, c := range drawn(fb) {
			if got := fb.At(c[0], c[1]).Color; got != 1 {
				t.Fatalf("nearFirst=%v: wanted only the near cube, got color %d at %v", nearFirst, got, c)
			}
		}
	}
}

func TestDrawShadedPlacement(t *testing.T) {
	light := Light{Direction: geometry.Vector3{Z: 1}}

	for _, tt := range []struct {
		name  string
		x     float64
		right bool
	}{
		{"to the right", 3, true},
		{"to the left", -3, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fb := newShadeBuffer(t)
			cube := Cube(1)
			cube.Position = geometry.Vector3{X: tt.x, Z: 5}

			cube.DrawShaded(fb, shadeCam(), light, 1)

			cells := drawn(fb)
			if len(cells) == 0 {
				t.Fatalf("wanted the cube drawn, got a blank frame")
			}
			for _, c := range cells {
				if (c[0] > shadeW/2) != tt.right {
					t.Fatalf("wanted every cell on the %v side, got %v", tt.name, c)
				}
			}
		})
	}
}

func TestDrawShadedBehindCamera(t *testing.T) {
	fb := newShadeBuffer(t)
	cube := Cube(2)
	cube.Position = geometry.Vector3{Z: -5}

	cube.DrawShaded(fb, shadeCam(), Light{Direction: geometry.Vector3{Z: 1}}, 1)

	if cells := drawn(fb); len(cells) != 0 {
		t.Errorf("wanted a blank frame, got %d cells", len(cells))
	}
}

// A floor running from behind the camera to far ahead must be cut at the near
// plane, not dropped: it fills the lower screen and never rises above the horizon.
func TestDrawShadedFloorCrossingTheNearPlane(t *testing.T) {
	fb := newShadeBuffer(t)
	floor := Mesh{
		Vertices: []geometry.Vector3{
			{X: -5, Y: -1.5, Z: -5}, {X: -5, Y: -1.5, Z: 10}, {X: 5, Y: -1.5, Z: 10}, {X: 5, Y: -1.5, Z: -5},
		},
		// (b-a)x(c-a) is +y, facing up toward the camera above it.
		Faces: [][3]int{{0, 1, 2}, {0, 2, 3}},
	}

	floor.DrawShaded(fb, shadeCam(), Light{Direction: geometry.Vector3{Y: -1}}, 1)

	// The far edge is at z=10: 1.5/10 below the horizon, so row 13 and down.
	for y := range 13 {
		for x := range shadeW {
			if fb.At(x, y) != engine.BlankCell {
				t.Fatalf("wanted rows above the floor's far edge blank, got a cell at (%d,%d)", x, y)
			}
		}
	}
	if fb.At(shadeW/2, shadeH-1).Character != '@' {
		t.Errorf("wanted the floor to reach the bottom row, got %q", fb.At(shadeW/2, shadeH-1).Character)
	}
	if fb.At(shadeW/2, 14) == engine.BlankCell {
		t.Errorf("wanted the floor to fill the view just below its far edge")
	}
}

func TestDrawShadedWithoutFaces(t *testing.T) {
	fb := newShadeBuffer(t)
	grid := Grid(10, 5)
	grid.Position = geometry.Vector3{Y: -1, Z: 5}

	grid.DrawShaded(fb, shadeCam(), Light{Direction: geometry.Vector3{Y: -1}}, 1)

	if cells := drawn(fb); len(cells) != 0 {
		t.Errorf("wanted a mesh without faces to draw nothing, got %d cells", len(cells))
	}
}

// A floor's depth changes across the screen. Where a wall stands on it, the part of
// the floor behind the wall must be hidden and the part in front must not be.
func TestDrawShadedDepthVariesAcrossAFace(t *testing.T) {
	floor := Mesh{
		Vertices: []geometry.Vector3{
			{X: -5, Y: -1.5, Z: -5}, {X: -5, Y: -1.5, Z: 10}, {X: 5, Y: -1.5, Z: 10}, {X: 5, Y: -1.5, Z: -5},
		},
		Faces: [][3]int{{0, 1, 2}, {0, 2, 3}}, // faces up
	}
	wall := Mesh{
		Vertices: []geometry.Vector3{
			{X: -1, Y: -1.5, Z: 5}, {X: -1, Y: 0.5, Z: 5}, {X: 1, Y: 0.5, Z: 5}, {X: 1, Y: -1.5, Z: 5},
		},
		Faces: [][3]int{{0, 1, 2}, {0, 2, 3}}, // faces the camera
	}
	// Lights the wall harder than the floor, so their characters differ.
	light := Light{Direction: geometry.Vector3{Y: -1, Z: 2}}
	wallChar := shadeChar(light.intensity(geometry.Vector3{Z: -1}))
	floorChar := shadeChar(light.intensity(geometry.Vector3{Y: 1}))
	if wallChar == floorChar {
		t.Fatalf("test setup: wall and floor need different shades, both %q", wallChar)
	}

	for _, floorFirst := range []bool{true, false} {
		fb := newShadeBuffer(t)
		if floorFirst {
			floor.DrawShaded(fb, shadeCam(), light, 1)
			wall.DrawShaded(fb, shadeCam(), light, 2)
		} else {
			wall.DrawShaded(fb, shadeCam(), light, 2)
			floor.DrawShaded(fb, shadeCam(), light, 1)
		}

		// Row 14 on the wall's face: the floor there lies at z of about 7, behind the wall.
		if got := fb.At(shadeW/2, 14).Character; got != wallChar {
			t.Errorf("floorFirst=%v: wanted the wall %q at (40,14), got %q", floorFirst, wallChar, got)
		}
		// Row 18 is below the wall's base: the floor there is closer than the wall.
		if got := fb.At(shadeW/2, 18).Character; got != floorChar {
			t.Errorf("floorFirst=%v: wanted the floor %q at (40,18), got %q", floorFirst, floorChar, got)
		}
	}
}

func BenchmarkDrawShaded(b *testing.B) {
	for _, size := range []struct{ w, h int }{{80, 24}, {200, 60}} {
		b.Run(fmt.Sprintf("%dx%d", size.w, size.h), func(b *testing.B) {
			fb, err := engine.NewFrameBuffer(size.w, size.h)
			if err != nil {
				b.Fatalf("unable to initialize frame buffer: %v", err)
			}
			cam := shadeCam()
			light := Light{Direction: geometry.Vector3{X: 0.4, Y: -1, Z: 0.6}, Ambient: 0.2}

			// Six cubes at mixed distances and turns, one close enough to fill the view.
			var cubes []Mesh
			for i, z := range []float64{1.5, 3, 5, 7, 9, 12} {
				c := Cube(2)
				c.Position = geometry.Vector3{X: float64(i%3-1) * 1.5, Z: z}
				c.Yaw, c.Pitch = float64(i)*0.5, float64(i)*0.3
				cubes = append(cubes, c)
			}

			b.ReportAllocs()
			for b.Loop() {
				fb.Clear()
				for _, c := range cubes {
					c.DrawShaded(fb, cam, light, 1)
				}
			}
		})
	}
}
