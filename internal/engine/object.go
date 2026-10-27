package engine

type SceneObject struct {
	X     float64
	Y     float64
	Value rune
}

func NewSceneObject(x, y float64, r rune) *SceneObject {
	return &SceneObject{X: x, Y: y, Value: r}
}
