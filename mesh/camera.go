package mesh

import (
	"math"

	"github.com/kyrokohan/glonk/geometry"
)

// Camera views the world from Position, looking down +z
// when Yaw is 0. Yaw and FOV (vertical) are in radians.
type Camera struct {
	Position geometry.Vector3
	Yaw      float64
	FOV      float64
	Near     float64
}

// cellAspect exists because terminal cells are about twice as tall as they are wide
const cellAspect = 2

// Project returns the 2D projection of 'p'.
// ok=false means that object is behind the camera
// with respect to the Near value
//
// x and y are fractional cell coordinates and the
// caller is expected to floor them.
func (c Camera) Project(p geometry.Vector3, w, h int) (x, y float64, ok bool) {
	v := c.toView(p)

	if v.Z < c.Near {
		return 0, 0, false
	}

	px, py := c.projectView(v, w, h)

	return px, py, true
}

// Forward returns the unit vector the camera faces, flat on the xz plane.
// It is view-space +z rotated by Yaw, the inverse of toView's RotateY(-Yaw).
func (c Camera) Forward() geometry.Vector3 {
	return geometry.Vector3{X: 0, Y: 0, Z: 1}.RotateY(c.Yaw)
}

// Right returns the unit vector to the camera's right, flat on the xz plane.
func (c Camera) Right() geometry.Vector3 {
	return geometry.Vector3{X: 1, Y: 0, Z: 0}.RotateY(c.Yaw)
}

// toView moves a world point into view space: camera at the origin, looking down +z.
func (c Camera) toView(p geometry.Vector3) geometry.Vector3 {
	return p.Sub(c.Position).RotateY(-c.Yaw)
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
