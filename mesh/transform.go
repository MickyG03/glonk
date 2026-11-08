package mesh

import "github.com/kyrokohan/glonk/geometry"

// Transform says where a model sits in the world. The zero Transform leaves the
// model where it was built, at the origin and unturned. All angles are in radians.
type Transform struct {
	// Position is where the model's origin sits in the world.
	Position geometry.Vector3

	// Yaw turns the model about the y axis. Positive turns +z toward +x. It is
	// applied after Pitch and before Position.
	Yaw float64

	// Pitch turns the model about the x axis. Positive turns +y toward +z. It is
	// applied first.
	Pitch float64
}

// Apply returns where the model-space point p ends up in the world: turned by
// Pitch, then Yaw, then moved to Position.
func (t Transform) Apply(p geometry.Vector3) geometry.Vector3 {
	return p.RotateX(t.Pitch).RotateY(t.Yaw).Add(t.Position)
}
