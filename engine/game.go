package engine

import "errors"

// ErrQuit, returned from Update, ends Run without an error.
var ErrQuit = errors.New("quit")

// Game is what Run drives: Update advances state, Draw renders it.
// An error returned from Update ends Run with that error, except ErrQuit, which
// ends it cleanly. A game that needs to know the size of the frame can also have a
// Resize method; see Resizer.
type Game interface {
	// Update advances the game by dt seconds, the time since the previous frame (at
	// most a quarter of a second, so a stall does not fling things across the world).
	// in is the keyboard state for this frame. Return ErrQuit to end Run cleanly.
	Update(dt float64, in *Input) error

	// Draw renders the current state into fb, which is blank at the start of every
	// call.
	Draw(fb *FrameBuffer)
}

// Resizer is an optional part of a Game. A game that has a Resize method is told
// the size of its frame, which Update and Draw need to know and Update is not
// otherwise given.
type Resizer interface {
	// Resize is called with the frame's size in cells before the first Update, and
	// again whenever that size changes, always before the Update and Draw of the
	// frame that has the new size. The size is that of the FrameBuffer Draw
	// receives: Config.Width and Config.Height where they are set, the terminal's
	// size on an axis left at 0.
	Resize(width, height int)
}
