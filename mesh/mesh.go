package mesh

import (
	"math"
	"slices"

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
	// Vertices are the points of the model, in model space: the space before
	// Transform is applied.
	Vertices []geometry.Vector3

	// Edges are pairs of indices into Vertices. Draw draws each as a line.
	Edges [][2]int

	// Faces are triples of indices into Vertices. DrawShaded fills each as a
	// triangle.
	Faces [][3]int

	// Transform places the model in the world.
	Transform Transform

	// DoubleSided draws faces from behind as well as from the front, lit as the
	// side the camera sees. Use it for models wound the wrong way round (they
	// look inside out or vanish), and for open surfaces such as a flat leaf.
	// It only affects shaded drawing, not wireframes.
	DoubleSided bool
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

// Clone returns a copy of the mesh that shares nothing with it, so changing the
// copy's Vertices, Edges or Faces leaves the original alone. Assigning a Mesh to a
// new variable copies the struct but not these lists, which both would still share.
func (m Mesh) Clone() Mesh {
	out := m
	out.Vertices = slices.Clone(m.Vertices)
	out.Edges = slices.Clone(m.Edges)
	out.Faces = slices.Clone(m.Faces)
	return out
}

// validIndex reports whether i names one of n items.
func validIndex(i, n int) bool {
	return i >= 0 && i < n
}

// Draw renders the mesh's edges as lines, all in cell. A mesh whose Transform is
// not valid is not drawn.
// Edges crossing the near plane are cut where they cross it;
// edges entirely behind it are skipped, and so are edges that name a vertex the
// mesh does not have (Validate finds them). An edge that ends far off screen
// costs no more than one that stays on screen.
func (m Mesh) Draw(fb *engine.FrameBuffer, cam Camera, cell engine.Cell) {
	cam, ok := cam.withDefaults()
	if !ok || !m.Transform.valid() {
		return
	}

	place := m.Transform.placement()
	view := make([]geometry.Vector3, len(m.Vertices))
	for i, v := range m.Vertices {
		view[i] = cam.toView(place.apply(v))
	}

	for _, e := range m.Edges {
		if !validIndex(e[0], len(view)) || !validIndex(e[1], len(view)) {
			continue
		}
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
		drawEdge(fb, ax, ay, bx, by, cell)
	}
}

// farEdge is how far from the origin, in cells, a projected point can be before
// drawEdge trims the edge to the screen. It is well inside the range DrawLine
// handles exactly.
const farEdge = 1e8

// drawEdge draws the line between two projected points. A point almost on the
// camera's plane projects absurdly far away, or to infinity; such an edge is cut
// back to the screen first so that it keeps its direction instead of being bent by
// DrawLine's coordinate limit. An edge with a non-finite end cannot be drawn.
func drawEdge(fb *engine.FrameBuffer, ax, ay, bx, by float64, cell engine.Cell) {
	for _, v := range [...]float64{ax, ay, bx, by} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return
		}
	}

	if math.Abs(ax) > farEdge || math.Abs(ay) > farEdge || math.Abs(bx) > farEdge || math.Abs(by) > farEdge {
		var ok bool
		// A cell of margin keeps the cut ends outside the visible area.
		ax, ay, bx, by, ok = clipSegment(ax, ay, bx, by, -1, -1, float64(fb.Width()), float64(fb.Height()))
		if !ok {
			return
		}
	}

	fb.DrawLine(int(math.Floor(ax)), int(math.Floor(ay)), int(math.Floor(bx)), int(math.Floor(by)), cell)
}
