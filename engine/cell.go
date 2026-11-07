package engine

// Cell is one terminal character with a color and a background color.
type Cell struct {
	Character rune

	// Color is the character's color.
	// 0 means terminal default
	// 1-255 are the palette indices.
	// Use 16 for black.
	Color uint8

	// Background is the color behind the character, with the same values as Color.
	Background uint8
}

// BlankCell is an empty cell: a space in the default colors.
var BlankCell = Cell{
	Character: ' ',
}
