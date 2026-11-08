package sprite

// Sprite is one character at a fractional position, floored to a cell when drawn.
type Sprite struct {
	// X is the column and Y the row, in cells from the top left. Fractions are
	// floored when the sprite is drawn.
	X, Y float64

	// Value is the character shown. It is drawn in the terminal's default colors.
	Value rune
}

// New returns a Sprite showing r at (x, y).
func New(x, y float64, r rune) *Sprite {
	return &Sprite{X: x, Y: y, Value: r}
}
