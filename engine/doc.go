// Package engine runs a terminal render loop: a Game drawn into a FrameBuffer of
// cells, with keyboard Input.
//
// A program implements [Game] and gives it to [Run]. Run owns the terminal: it
// switches to the alternate screen, reads the keyboard into an [Input], calls Update
// and then Draw at a fixed frame rate, and puts the terminal back when it returns.
// Given an input and output that are not terminals, Run runs headless, which is how
// a game is tested.
//
// Draw writes [Cell] values into a [FrameBuffer], a grid with its origin at the top left,
// x growing to the right and y growing down. Each cell is one character with a
// foreground and a background [Color] (the terminal's default, one of the 256
// palette colors, or RGB) and text attributes such as [Bold]. Output is 256-color for
// now, so it works in nearly any terminal. A [Canvas] gives the same frame a second
// layer of pixels, two to a cell, drawn with half-block characters.
package engine
