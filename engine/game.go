package engine

import "errors"

// ErrQuit, returned from Update, ends Run without an error.
var ErrQuit = errors.New("quit")

// Game is what Run drives: Update advances state, Draw renders it.
// An error returned from Update ends Run with that error, except ErrQuit, which
// ends it cleanly.
type Game interface {
	// Update advances the game by dt seconds, the time since the previous frame (at
	// most a quarter of a second, so a stall does not fling things across the world).
	// in is the keyboard state for this frame. Return ErrQuit to end Run cleanly.
	Update(dt float64, in *Input) error

	// Draw renders the current state into fb, which is blank at the start of every
	// call.
	Draw(fb *FrameBuffer)
}
