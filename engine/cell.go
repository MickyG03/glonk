package engine

// Cell is one terminal character with a color and a background color.
//
// A cell is one column wide, and only a printable character that takes one column
// is shown as it is. What FrameBuffer.At returns is what was stored; the frame is
// checked when it is sent, so nothing a cell holds can send commands to the terminal.
type Cell struct {
	// Character is the character to draw. NUL, which is what an unset field holds,
	// is shown as a space, so Cell{Background: 4} is a colored blank. Control
	// characters (tab, newline, escape), wide characters (CJK, most emoji) and
	// combining marks are shown as '?'. Real wide-character support needs a cell that
	// can span two columns and is not there yet.
	Character rune

	// Color is the character's color. 0 means the terminal's default; 1 to 255 are
	// palette indices. Use 16 for black.
	Color uint8

	// Background is the color behind the character, with the same values as Color.
	Background uint8
}

// BlankCell is an empty cell: a space in the default colors.
var BlankCell = Cell{
	Character: ' ',
}
