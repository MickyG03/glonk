package mesh

import "fmt"

// Validate checks a mesh that was built by hand or by other code, and returns an
// error describing the first problem: a vertex that is not a finite number, or an
// edge or face that names a vertex the mesh does not have.
//
// Drawing skips edges and faces with a bad index instead of failing, so Validate is
// how to find out that a mesh is broken. Meshes from LoadOBJ, Cube and Grid are
// valid.
func (m Mesh) Validate() error {
	n := len(m.Vertices)

	for i, v := range m.Vertices {
		if !finite(v.X, v.Y, v.Z) {
			return fmt.Errorf("mesh: vertex %d is not a finite number: %v", i, v)
		}
	}
	for i, e := range m.Edges {
		for _, v := range e {
			if !validIndex(v, n) {
				return missingVertex("edge", i, v, n)
			}
		}
	}
	for i, f := range m.Faces {
		for _, v := range f {
			if !validIndex(v, n) {
				return missingVertex("face", i, v, n)
			}
		}
	}
	return nil
}

// missingVertex describes edge or face number i naming vertex v in a mesh of n vertices.
func missingVertex(part string, i, v, n int) error {
	if n == 0 {
		return fmt.Errorf("mesh: %s %d uses vertex %d, but the mesh has no vertices", part, i, v)
	}
	return fmt.Errorf("mesh: %s %d uses vertex %d, but only vertices 0 to %d exist", part, i, v, n-1)
}
