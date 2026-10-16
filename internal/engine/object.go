package engine

type SceneObject struct {
	X     int
	Y     int
	Value rune
}

func NewSceneObject(x, y int, r rune) *SceneObject {
	return &SceneObject{X: x, Y: y, Value: r}
}
