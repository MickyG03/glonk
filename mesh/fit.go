package mesh

import (
	"math"

	"github.com/kyrokohan/glonk/geometry"
)

// Bounds returns the smallest axis-aligned box around the mesh's vertices, in
// model space (before Position, Yaw and Pitch are applied). An empty mesh gives
// two zero vectors.
func (m Mesh) Bounds() (min, max geometry.Vector3) {
	if len(m.Vertices) == 0 {
		return geometry.Vector3{}, geometry.Vector3{}
	}

	min, max = m.Vertices[0], m.Vertices[0]
	for _, v := range m.Vertices[1:] {
		min = geometry.Vector3{X: math.Min(min.X, v.X), Y: math.Min(min.Y, v.Y), Z: math.Min(min.Z, v.Z)}
		max = geometry.Vector3{X: math.Max(max.X, v.X), Y: math.Max(max.Y, v.Y), Z: math.Max(max.Z, v.Z)}
	}
	return min, max
}

// Fit returns a copy of the mesh centered on the origin and scaled evenly so that
// its longest side is size, which makes models from different sources comparable
// and puts them where a camera at the origin's side can see them. Position, Yaw
// and Pitch are kept.
//
// A mesh with no size to speak of (an empty mesh, a single point), a size that is
// not a positive finite number, or vertices that are not finite are not scaled; the
// last of these is also not moved. The copy shares nothing with the original.
func (m Mesh) Fit(size float64) Mesh {
	out := m.Clone()
	if len(out.Vertices) == 0 {
		return out
	}

	lo, hi := m.Bounds()
	center := lo.Add(hi).Scale(0.5)
	extent := hi.Sub(lo)
	longest := math.Max(extent.X, math.Max(extent.Y, extent.Z))
	if !finite(center.X, center.Y, center.Z, longest) {
		return out
	}

	scale := 1.0
	if longest > 0 && size > 0 && !math.IsInf(size, 0) {
		scale = size / longest
	}
	for i, v := range out.Vertices {
		out.Vertices[i] = v.Sub(center).Scale(scale)
	}
	return out
}

func finite(vs ...float64) bool {
	for _, v := range vs {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}
