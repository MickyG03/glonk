package mesh

import (
	"math"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

// Mesh describes an object that holds
// model-space vertices, edges between them, triangle faces
// and its placement in the world.
//
// Edges drive wireframe drawing and Faces drive shaded drawing; a mesh can have
// either or both. Faces index into Vertices and are wound so the face normal,
// (b-a)x(c-a), points out of the surface.
type Mesh struct {
	Vertices []geometry.Vector3
	Edges    [][2]int
	Faces    [][3]int
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

// clipNear returns where the segment from out to in crosses z = near.
// out must be behind the plane (Z < near) and in on or in front of it,
// so the divisor is never zero.
func clipNear(out, in geometry.Vector3, near float64) geometry.Vector3 {
	t := (near - out.Z) / (in.Z - out.Z)
	p := out.Lerp(in, t)
	// Lerp rounding can leave Z a hair under near, which projectView does not tolerate.
	p.Z = near
	return p
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

	// Vertex i sits at +size/2 on x, y and z when bit 1, 2 and 4 of i is set.
	// Each side is two triangles, listed by the quad's corners in outward order.
	var faces [][3]int
	for _, quad := range [][4]int{
		{0, 4, 6, 2}, // -x
		{1, 3, 7, 5}, // +x
		{0, 1, 5, 4}, // -y
		{2, 6, 7, 3}, // +y
		{0, 2, 3, 1}, // -z
		{4, 5, 7, 6}, // +z
	} {
		faces = append(faces, [3]int{quad[0], quad[1], quad[2]}, [3]int{quad[0], quad[2], quad[3]})
	}

	return Mesh{Vertices: vertices, Edges: edges, Faces: faces}
}

// toWorld places a model-space point in the world: rotate, then translate.
func (m Mesh) toWorld(v geometry.Vector3) geometry.Vector3 {
	return v.RotateX(m.Pitch).RotateY(m.Yaw).Add(m.Position)
}

// Draw renders the mesh's edges as lines.
// Edges crossing the near plane are cut where they cross it;
// edges entirely behind it are skipped.
//
// shortcut: a cut endpoint can project far off screen and DrawLine still steps
// every cell to it, so a tiny Near gets slow; clip lines to the screen in DrawLine to upgrade.
func (m Mesh) Draw(fb *engine.FrameBuffer, cam Camera, cell engine.Cell) {
	view := make([]geometry.Vector3, len(m.Vertices))
	for i, v := range m.Vertices {
		view[i] = cam.toView(m.toWorld(v))
	}

	for _, e := range m.Edges {
		a, b := view[e[0]], view[e[1]]
		aIn := a.Z >= cam.Near
		bIn := b.Z >= cam.Near

		if !aIn && !bIn {
			continue
		}
		if !aIn {
			a = clipNear(a, b, cam.Near)
		}
		if !bIn {
			b = clipNear(b, a, cam.Near)
		}

		ax, ay := cam.projectView(a, fb.Width(), fb.Height())
		bx, by := cam.projectView(b, fb.Width(), fb.Height())
		fb.DrawLine(int(math.Floor(ax)), int(math.Floor(ay)), int(math.Floor(bx)), int(math.Floor(by)), cell)
	}
}
