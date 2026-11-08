package mesh

import (
	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

// Light lights the surfaces of a mesh. [Directional] and [Ambient] are the lights
// the package provides; a type with a Brightness method is a Light too, which is how
// a lamp that fades with distance or a spotlight is made.
//
// Several lights can light a mesh at once: DrawShaded and DrawShadedPixels take any
// number. They add up without passing full brightness: each light fills part of what
// the ones before it left dark, so an ambient light of 0.2 and a directional light
// that meets a face squarely at 1 give 1, and two lights of 0.5 give 0.75. The order
// does not matter. With no lights, every face is as dark as can be shown.
type Light interface {
	// Brightness returns how bright this light makes a surface, from 0 (dark) to 1
	// (fully lit). point is the middle of the face and normal is its unit normal,
	// pointing out of the surface toward the viewer, both in world space. A result
	// outside 0 to 1 is clamped, and NaN counts as 0.
	Brightness(point, normal geometry.Vector3) float64
}

// Directional is light that travels one way everywhere, like the sun. A surface is
// brightest facing straight into it and dark when edge-on or turned away.
type Directional struct {
	// Direction is the way the light travels, e.g. {0, -1, 0} for a lamp straight
	// overhead. It need not be unit length. A zero Direction gives no light.
	Direction geometry.Vector3

	// Intensity is the brightness, from 0 to 1, of a surface facing straight into the
	// light. 0 means 1; to switch a light off, leave it out.
	Intensity float64
}

// Brightness is Intensity scaled by how squarely the surface faces the light.
func (d Directional) Brightness(_, normal geometry.Vector3) float64 {
	intensity := d.Intensity
	if intensity == 0 {
		intensity = 1
	}
	return clamp01(intensity) * clamp01(normal.Dot(d.Direction.Normalize().Neg()))
}

// Ambient is light from everywhere at once, which fills in the faces no other light
// reaches so that they do not go dark. It does not depend on which way a face points.
type Ambient struct {
	// Level is the brightness, from 0 to 1, of every surface. The zero Ambient gives
	// no light.
	Level float64
}

// Brightness is Level.
func (a Ambient) Brightness(_, _ geometry.Vector3) float64 {
	return clamp01(a.Level)
}

// clamp01 limits v to 0 to 1. NaN gives 0.
func clamp01(v float64) float64 {
	switch {
	case !(v > 0):
		return 0
	case v > 1:
		return 1
	}
	return v
}

// brightness adds up what lights give a surface; see Light. A nil entry gives
// nothing.
func brightness(lights []Light, point, normal geometry.Vector3) float64 {
	total := 0.0
	for _, l := range lights {
		if l == nil {
			continue
		}
		total += (1 - total) * clamp01(l.Brightness(point, normal))
	}
	return total
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
func (m Mesh) eachShadedTriangle(cam Camera, lights []Light, w, h int, fn func(corners [3]screenPoint, intensity float64)) {
	cam, ok := cam.withDefaults()
	if !ok || !m.Transform.valid() {
		return
	}

	place := m.Transform.placement()
	world := make([]geometry.Vector3, len(m.Vertices))
	view := make([]geometry.Vector3, len(m.Vertices))
	for i, v := range m.Vertices {
		world[i] = place.apply(v)
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
		w0, w1, w2 := world[f[0]], world[f[1]], world[f[2]]
		normal := w1.Sub(w0).Cross(w2.Sub(w0)).Normalize()
		if !front {
			normal = normal.Neg() // light the side the camera is looking at
		}
		center := w0.Add(w1).Add(w2).Scale(1.0 / 3)
		intensity := brightness(lights, center, normal)

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

// DrawShaded renders the mesh's faces as solid triangles, flat shaded by the lights
// and drawn in the mesh's Material. Nearer surfaces hide farther ones, faces turned away
// from the camera are skipped (unless the mesh is DoubleSided), and faces crossing
// the near plane are cut where they cross it. A mesh without Faces draws nothing; use
// Draw for wireframes.
//
// Brightness is shown by the character: a ramp from dim to bright, Material.Ramp.
func (m Mesh) DrawShaded(fb *engine.FrameBuffer, cam Camera, lights ...Light) {
	ramp := m.Material.ramp()
	m.eachShadedTriangle(cam, lights, fb.Width(), fb.Height(), func(c [3]screenPoint, intensity float64) {
		cell := engine.Cell{Character: shadeChar(ramp, intensity), Foreground: m.Material.Color}
		fb.FillTriangle(c[0].x, c[0].y, c[0].z, c[1].x, c[1].y, c[1].z, c[2].x, c[2].y, c[2].z, cell)
	})
}

// DrawShadedPixels is DrawShaded for a pixel Canvas: the same faces, culling,
// clipping and depth rules, but brightness is shown by the color (a darker or
// lighter shade of Material.Color) and there are twice as many rows, so edges are
// finer. Call Canvas.Compose afterwards to put the pixels into the FrameBuffer.
func (m Mesh) DrawShadedPixels(c *engine.Canvas, cam Camera, lights ...Light) {
	// Project onto the cell grid the canvas belongs to, then stretch y to pixel rows.
	m.eachShadedTriangle(cam, lights, c.Width(), c.Height()/2, func(p [3]screenPoint, intensity float64) {
		shade := engine.Shade(m.Material.Color, intensity)
		c.FillTriangle(p[0].x, 2*p[0].y, p[0].z, p[1].x, 2*p[1].y, p[1].z, p[2].x, 2*p[2].y, p[2].z, shade)
	})
}
