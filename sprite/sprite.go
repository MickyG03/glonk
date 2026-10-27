package sprite

type Sprite struct {
	X     float64
	Y     float64
	Value rune
}

func New(x, y float64, r rune) *Sprite {
	return &Sprite{X: x, Y: y, Value: r}
}
