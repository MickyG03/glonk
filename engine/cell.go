package engine

type Cell struct {
	Character rune

	// 0 means terminal default
	// 1-255 are the palette indices.
	// Use 16 for black.
	Color uint8
}

var BlankCell = Cell{
	Character: ' ',
	Color:     0,
}
