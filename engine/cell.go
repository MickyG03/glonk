package engine

// Cell is one terminal character with a color.
type Cell struct {
	Character rune

	// 0 means terminal default
	// 1-255 are the palette indices.
	// Use 16 for black.
	Color uint8
}

// BlankCell is an empty cell: a space in the default color.
var BlankCell = Cell{
	Character: ' ',
	Color:     0,
}
