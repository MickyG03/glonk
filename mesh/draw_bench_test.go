package mesh

import (
	"fmt"
	"math"
	"testing"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

// benchScene is what cube3d draws: a floor of lines and six cubes at mixed distances
// and turns, one close enough to fill the view, seen from a camera that looks down a
// little.
func benchScene() (floor Mesh, cubes []Mesh, cam Camera, lights []Light) {
	floor = Grid(40, 20)
	floor.Transform.Position = geometry.Vector3{Y: -1.5}
	for i, z := range []float64{1.5, 3, 5, 7, 9, 12} {
		c := Cube(2)
		c.Transform = Transform{Position: geometry.Vector3{X: float64(i%3-1) * 1.5, Z: z}, Yaw: float64(i) * 0.5, Pitch: float64(i) * -0.3}
		c.Material.Color = engine.Palette(uint8(30 + i*30))
		cubes = append(cubes, c)
	}
	cam = Camera{Position: geometry.Vector3{Y: 0.5, Z: -3}, Pitch: -0.1, FOV: math.Pi / 2, Near: 0.1}
	lights = []Light{Ambient{Level: 0.2}, Directional{Direction: geometry.Vector3{X: 0.4, Y: -1, Z: 0.6}}}
	return floor, cubes, cam, lights
}

// BenchmarkDraw3D measures one frame of the cube3d scene drawn as wireframe, shaded
// into cells and shaded into half-block pixels, at two sizes.
func BenchmarkDraw3D(b *testing.B) {
	floor, cubes, cam, lights := benchScene()
	for _, size := range []struct{ w, h int }{{80, 24}, {200, 60}} {
		fb, err := engine.NewFrameBuffer(size.w, size.h)
		if err != nil {
			b.Fatal(err)
		}
		draws := map[string]func(){
			"wireframe": func() {
				floor.Draw(fb, cam, engine.Cell{Character: '.'})
				for _, c := range cubes {
					c.Draw(fb, cam, engine.Cell{Character: '#'})
				}
			},
			"shaded": func() {
				for _, c := range cubes {
					c.DrawShaded(fb, cam, lights...)
				}
			},
			"pixels": func() {
				for _, c := range cubes {
					c.DrawShadedPixels(fb.Canvas(), cam, lights...)
				}
			},
		}
		for _, name := range []string{"wireframe", "shaded", "pixels"} {
			draw := draws[name]
			b.Run(fmt.Sprintf("%s %dx%d", name, size.w, size.h), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					fb.Clear()
					draw()
				}
			})
		}
	}
}
