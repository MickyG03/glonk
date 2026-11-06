package sprite

// Sprite is one character at a fractional position, floored to a cell when drawn.
type Sprite struct {
	X     float64
	Y     float64
	Value rune
}

// New returns a Sprite showing r at (x, y).
func New(x, y float64, r rune) *Sprite {
	return &Sprite{X: x, Y: y, Value: r}
}
