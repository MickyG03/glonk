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

// screenPoint is a mesh corner on screen: x and y in cell coordinates, z the
// distance from the camera.
type screenPoint struct {
	x, y, z float64
}

// eachShadedTriangle does the work shared by every solid drawing mode. For each
// face turned toward the camera it lights the face, cuts it at the near plane,
// projects the pieces onto a w x h grid of cells, and hands each to fn along with
// the face's brightness from 0 to 1. A face that names a vertex the mesh does not
// have is skipped.
func (m Mesh) eachShadedTriangle(cam Camera, light Light, w, h int, fn func(corners [3]screenPoint, intensity float64)) {
	cam, ok := cam.withDefaults()
	if !ok {
		return
	}

	world := make([]geometry.Vector3, len(m.Vertices))
	view := make([]geometry.Vector3, len(m.Vertices))
	for i, v := range m.Vertices {
		world[i] = m.Transform.Apply(v)
		view[i] = cam.toView(world[i])
	}

	for _, f := range m.Faces {
		if !validIndex(f[0], len(view)) || !validIndex(f[1], len(view)) || !validIndex(f[2], len(view)) {
			continue
		}
		v0, v1, v2 := view[f[0]], view[f[1]], view[f[2]]

		front := frontFacing(v1.Sub(v0).Cross(v2.Sub(v0)), v0)
		if !front && !m.DoubleSided {
			continue
		}

		// Lighting needs the normal in world space, where the light lives.
		w0 := world[f[0]]
		normal := world[f[1]].Sub(w0).Cross(world[f[2]].Sub(w0)).Normalize()
		if !front {
			normal = normal.Neg() // light the side the camera is looking at
		}
		intensity := light.intensity(normal)

		for _, t := range clipTriangleNear(v0, v1, v2, cam.Near) {
			var corners [3]screenPoint
			for i, v := range t {
				x, y := cam.projectView(v, w, h)
				corners[i] = screenPoint{x, y, v.Z}
			}
			fn(corners, intensity)
		}
	}
}

// DrawShaded renders the mesh's faces as solid triangles, flat shaded by light
// and drawn in color. Nearer surfaces hide farther ones, faces turned away from
// the camera are skipped (unless the mesh is DoubleSided), and faces crossing the
// near plane are cut where they cross it. A mesh without Faces draws nothing; use
// Draw for wireframes.
//
// Brightness is shown by the character: a ramp from dim to bright.
func (m Mesh) DrawShaded(fb *engine.FrameBuffer, cam Camera, light Light, color engine.Color) {
	m.eachShadedTriangle(cam, light, fb.Width(), fb.Height(), func(c [3]screenPoint, intensity float64) {
		cell := engine.Cell{Character: shadeChar(intensity), Foreground: color}
		fb.FillTriangle(c[0].x, c[0].y, c[0].z, c[1].x, c[1].y, c[1].z, c[2].x, c[2].y, c[2].z, cell)
	})
}

// DrawShadedPixels is DrawShaded for a pixel Canvas: the same faces, culling,
// clipping and depth rules, but brightness is shown by the color (a darker or
// lighter shade of color) and there are twice as many rows, so edges are finer.
// Call Canvas.Compose afterwards to put the pixels into the FrameBuffer.
func (m Mesh) DrawShadedPixels(c *engine.Canvas, cam Camera, light Light, color engine.Color) {
	// Project onto the cell grid the canvas belongs to, then stretch y to pixel rows.
	m.eachShadedTriangle(cam, light, c.Width(), c.Height()/2, func(p [3]screenPoint, intensity float64) {
		shade := engine.Shade(color, intensity)
		c.FillTriangle(p[0].x, 2*p[0].y, p[0].z, p[1].x, 2*p[1].y, p[1].z, p[2].x, 2*p[2].y, p[2].z, shade)
	})
}
