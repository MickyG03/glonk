// Package engine runs a terminal render loop. It owns the raw-mode terminal, a
// grid of cells (FrameBuffer), keyboard Input, and calls a Game's Update and Draw
// at a fixed frame rate. Cells are a rune plus a 256-color index, so output works
// in nearly any terminal.
package engine
