// Package engine runs a terminal render loop.
//
// A program implements Game and gives it to Run. Run owns the terminal: it switches
// to the alternate screen, reads the keyboard into an Input, calls Update and then
// Draw at a fixed frame rate, and puts the terminal back when it returns.
//
// Draw writes Cells into a FrameBuffer, a grid with its origin at the top left,
// x growing to the right and y growing down. Each cell is one character with a
// color and a background color from the 256-color palette, so output works in
// nearly any terminal. A Canvas gives the same frame a second layer of pixels, two
// to a cell, drawn with half-block characters.
package engine
