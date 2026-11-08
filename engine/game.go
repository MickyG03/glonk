package engine

import "errors"

// ErrQuit, returned from Update, ends Run without an error.
var ErrQuit = errors.New("quit")

// Game is what Run drives: Update advances state, Draw renders it.
// An error returned from Update ends Run with that error, except ErrQuit, which
// ends it cleanly.
type Game interface {
	Update(dt float64, in *Input) error
	Draw(fb *FrameBuffer)
}
