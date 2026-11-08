package sprite

import "github.com/kyrokohan/glonk/geometry"

// Sprite is one character at a fractional position, floored to a cell when drawn.
type Sprite struct {
	// Position is where the sprite is: X is the column and Y the row, in cells from
	// the top left. Fractions are floored when the sprite is drawn.
	Position geometry.Vector2

	// Value is the character shown. It is drawn in the terminal's default colors.
	Value rune
}

// New returns a Sprite showing r at position.
func New(position geometry.Vector2, r rune) *Sprite {
	return &Sprite{Position: position, Value: r}
}
