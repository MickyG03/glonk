package engine

// Cell is one terminal character with colors and text attributes.
//
// A cell is one column wide, and only a printable character that takes one column
// is shown as it is. What FrameBuffer.At returns is what was stored; the frame is
// checked when it is sent, so nothing a cell holds can send commands to the terminal.
type Cell struct {
	// Character is the character to draw. NUL, which is what an unset field holds,
	// is shown as a space, so Cell{Background: engine.Palette(4)} is a colored blank.
	// Control characters (tab, newline, escape), wide characters (CJK, most emoji) and
	// combining marks are shown as '?'. Real wide-character support needs a cell that
	// can span two columns and is not there yet.
	Character rune

	// Foreground is the color of the character. The zero Color is the terminal's
	// default.
	Foreground Color

	// Background is the color behind the character. The zero Color is the terminal's
	// default.
	Background Color

	// Attrs are the text attributes, such as Bold. The zero Attr is plain text.
	Attrs Attr
}

// BlankCell is an empty cell: a space in the default colors.
var BlankCell = Cell{
	Character: ' ',
}

// style is the part of the cell that is not its character.
func (c Cell) style() style {
	return style{fg: c.Foreground, bg: c.Background, attrs: c.Attrs}
}
