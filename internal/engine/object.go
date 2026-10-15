package engine

type SceneObject struct {
	X     int
	Y     int
	Value byte
}

func NewSceneObject(x, y int, b byte) *SceneObject {
	return &SceneObject{X: x, Y: y, Value: b}
}
