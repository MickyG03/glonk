package mesh

import (
	"math"

	"github.com/kyrokohan/glonk/geometry"
)

// Camera views the world from Position, looking down +z when Yaw and Pitch
// are 0. Yaw turns it right (positive) or left, Pitch tilts it up (positive) or
// down, and FOV is the vertical field of view. All three are in radians. Pitch
// is not limited; keep it inside +-pi/2 or the view flips over.
//
// Near is how close to the camera something can be and still be seen; anything
// nearer is cut away.
//
// A FOV or Near left at 0 uses a default (a FOV of pi/2, a quarter turn, and a Near
// of 0.1), so Camera{Position: p} is a working camera. A value that makes no sense,
// namely a FOV outside 0 to pi (exclusive), or a Near that is negative, NaN or
// infinite, makes the camera see nothing: Draw, DrawShaded and DrawShadedPixels draw
// nothing and Project reports ok=false. So does a Position, Yaw or Pitch that is not
// a finite number.
type Camera struct {
	Position geometry.Vector3
	Yaw      float64
	Pitch    float64
	FOV      float64
	Near     float64
}

// What a zero FOV and Near stand for.
const (
	defaultFOV  = math.Pi / 2
	defaultNear = 0.1
)

// withDefaults returns the camera as it is used: a zero FOV or Near replaced by its
// default. ok is false when a value makes no sense and nothing can be drawn.
func (c Camera) withDefaults() (cam Camera, ok bool) {
	if c.FOV == 0 {
		c.FOV = defaultFOV
	}
	if c.Near == 0 {
		c.Near = defaultNear
	}
	// Written so that NaN fails every test. An infinite Near passes, and sees nothing.
	ok = c.FOV > 0 && c.FOV < math.Pi && c.Near > 0
	return c, ok
}

// cellAspect exists because terminal cells are about twice as tall as they are wide
const cellAspect = 2

// Project returns the 2D projection of p.
// ok=false means that the point is behind the camera
// with respect to the Near value, has no finite place on the screen, or that the
// camera sees nothing at all (see Camera).
//
// x and y are fractional cell coordinates and the
// caller is expected to floor them.
func (c Camera) Project(p geometry.Vector3, w, h int) (x, y float64, ok bool) {
	c, usable := c.withDefaults()
	if !usable {
		return 0, 0, false
	}
	v := c.toView(p)

	if v.Z < c.Near {
		return 0, 0, false
	}

	px, py := c.projectView(v, w, h)
	if !finite(px, py) { // NaN or infinity: no place on the screen
		return 0, 0, false
	}

	return px, py, true
}

// Forward returns the unit vector the camera faces, flat on the xz plane: Pitch
// is ignored, so walking along it never climbs or dives. Use Look for the way
// the camera actually points. With Yaw 0 it is +z, and a positive Yaw turns it
// toward +x.
func (c Camera) Forward() geometry.Vector3 {
	return geometry.Vector3{X: 0, Y: 0, Z: 1}.RotateY(c.Yaw)
}

// Right returns the unit vector to the camera's right, flat on the xz plane.
// Tilting the camera does not move it.
func (c Camera) Right() geometry.Vector3 {
	return geometry.Vector3{X: 1, Y: 0, Z: 0}.RotateY(c.Yaw)
}

// Look returns the unit vector the camera points along, including Pitch. A point
// at Position plus Look times any distance lands in the middle of the screen.
func (c Camera) Look() geometry.Vector3 {
	return geometry.Vector3{X: 0, Y: 0, Z: 1}.RotateX(-c.Pitch).RotateY(c.Yaw)
}

// toView moves a world point into view space: camera at the origin, looking down +z.
// It undoes the camera's turn first and then its tilt, the reverse of how the
// camera was posed (turn about the vertical axis, then tilt about its own).
func (c Camera) toView(p geometry.Vector3) geometry.Vector3 {
	return p.Sub(c.Position).RotateY(-c.Yaw).RotateX(c.Pitch)
}

// projectView maps a view-space point to fractional cell coordinates.
// The caller must ensure v.Z >= Near; the result is meaningless otherwise.
func (c Camera) projectView(v geometry.Vector3, w, h int) (x, y float64) {
	f := 1 / math.Tan(c.FOV/2)
	ndcX := f * v.X / v.Z
	ndcY := f * v.Y / v.Z

	sx := float64(w)/2 + ndcX*float64(h)/2*cellAspect
	sy := float64(h)/2 - ndcY*float64(h)/2

	return sx, sy
}
