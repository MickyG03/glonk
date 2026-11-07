package mesh

import (
	"math"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

// shadeRamp runs from dark to bright. Plain ASCII so it renders in any terminal,
// and no space, so a lit face never vanishes into the background.
const shadeRamp = ".:-=+*#%@"

// Light is a directional light with a little ambient fill.
type Light struct {
	// Direction is the way the light travels, e.g. {0, -1, 0} for a lamp straight
	// overhead. It need not be unit length. A zero Direction gives ambient light only.
	Direction geometry.Vector3
	// Ambient is the brightness, from 0 to 1, of a face turned away from the light.
	Ambient float64
}

// intensity returns how bright a face with the given unit normal is, from 0 to 1:
// the ambient level plus the rest scaled by how squarely the face meets the light.
func (l Light) intensity(normal geometry.Vector3) float64 {
	ambient := min(max(l.Ambient, 0), 1)
	facing := max(0, normal.Dot(l.Direction.Normalize().Neg()))
	return ambient + (1-ambient)*facing
}

// shadeChar maps an intensity from 0 (darkest) to 1 (brightest) onto the ramp.
// Out-of-range values clamp; NaN is treated as dark.
func shadeChar(intensity float64) rune {
	if math.IsNaN(intensity) {
		intensity = 0
	}
	i := int(math.Round(min(max(intensity, 0), 1) * float64(len(shadeRamp)-1)))
	return rune(shadeRamp[i])
}

// frontFacing reports whether a face is turned toward the camera. normal points
// out of the face and v0 is one of its corners, both in view space, where the
// camera sits at the origin: the face is visible when the normal opposes the
// direction from the camera to the face.
func frontFacing(normal, v0 geometry.Vector3) bool {
	return normal.Dot(v0) < 0
}

// DrawShaded renders the mesh's faces as solid triangles, flat shaded by light
// and drawn in color. Nearer surfaces hide farther ones, faces turned away from
// the camera are skipped, and faces crossing the near plane are cut where they
// cross it. A mesh without Faces draws nothing; use Draw for wireframes.
func (m Mesh) DrawShaded(fb *engine.FrameBuffer, cam Camera, light Light, color uint8) {
	world := make([]geometry.Vector3, len(m.Vertices))
	view := make([]geometry.Vector3, len(m.Vertices))
	for i, v := range m.Vertices {
		world[i] = m.toWorld(v)
		view[i] = cam.toView(world[i])
	}

	for _, f := range m.Faces {
		v0, v1, v2 := view[f[0]], view[f[1]], view[f[2]]

		if !frontFacing(v1.Sub(v0).Cross(v2.Sub(v0)), v0) {
			continue
		}

		// Lighting needs the normal in world space, where the light lives.
		w0 := world[f[0]]
		normal := world[f[1]].Sub(w0).Cross(world[f[2]].Sub(w0)).Normalize()
		cell := engine.Cell{Character: shadeChar(light.intensity(normal)), Color: color}

		for _, t := range clipTriangleNear(v0, v1, v2, cam.Near) {
			x0, y0 := cam.projectView(t[0], fb.Width(), fb.Height())
			x1, y1 := cam.projectView(t[1], fb.Width(), fb.Height())
			x2, y2 := cam.projectView(t[2], fb.Width(), fb.Height())
			fb.FillTriangle(x0, y0, t[0].Z, x1, y1, t[1].Z, x2, y2, t[2].Z, cell)
		}
	}
}
