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
// x and y are fractional cell coordintes and the
// caller is expected to floor them.
func (c Camera) Project(p geometry.Vector3, w, h int) (x, y float64, ok bool) {
	rel := p.Sub(c.Position)

	cx := rel.X*math.Cos(c.Yaw) - rel.Z*math.Sin(c.Yaw)
	cy := rel.Y
	cz := rel.X*math.Sin(c.Yaw) + rel.Z*math.Cos(c.Yaw)

	if cz < c.Near {
		return 0, 0, false
	}

	f := 1 / math.Tan(c.FOV/2)
	ndcX := f * cx / cz
	ndcY := f * cy / cz

	sx := float64(w)/2 + ndcX*float64(h)/2*cellAspect
	sy := float64(h)/2 - ndcY*float64(h)/2

	return sx, sy, true
}
