package mesh

import (
	"math"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

// Mesh describes an object that holds
// model-space vertices, edges between them
// and its placement in the world.
type Mesh struct {
	Vertices []geometry.Vector3
	Edges    [][2]int
	Position geometry.Vector3
	Yaw      float64
	Pitch    float64
}

func corner(i, bit int, size float64) float64 {
	if i&bit == 0 {
		return -size / 2
	}
	return size / 2
}

// Cube returns a cube with edges of length
// size, centered on the origin.
func Cube(size float64) Mesh {
	vertices := make([]geometry.Vector3, 0, 8)

	for i := range 8 {
		x := corner(i, 1, size)
		y := corner(i, 2, size)
		z := corner(i, 4, size)
		v3 := geometry.Vector3{X: x, Y: y, Z: z}
		vertices = append(vertices, v3)
	}

	var edges [][2]int
	for i := range 8 {
		for _, b := range []int{1, 2, 4} {
			j := i | b
			if j != i {
				edges = append(edges, [2]int{i, j})
			}
		}
	}

	return Mesh{Vertices: vertices, Edges: edges}
}

type projection struct {
	x, y float64
	ok   bool
}

// Draw renders the mesh's edges as lines.
// Edges with an endpoint behind the camera
// or inside the near plane are skipped.
func (m Mesh) Draw(fb *engine.FrameBuffer, cam Camera, cell engine.Cell) {
	projected := make([]projection, len(m.Vertices))

	for i, v := range m.Vertices {
		world := v.RotateX(m.Pitch).RotateY(m.Yaw).Add(m.Position)
		x, y, ok := cam.Project(world, fb.Width(), fb.Height())
		projected[i] = projection{x, y, ok}
	}

	for _, e := range m.Edges {
		a := projected[e[0]]
		b := projected[e[1]]
		if !a.ok || !b.ok {
			continue
		}
		fb.DrawLine(int(math.Floor(a.x)), int(math.Floor(a.y)), int(math.Floor(b.x)), int(math.Floor(b.y)), cell)
	}
}
