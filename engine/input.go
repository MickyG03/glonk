package engine

import (
	"bytes"
	"io"
	"time"
)

// Without release events a key counts as held until this long passes
// without another repeat.
const holdTimeout = 150 * time.Millisecond

// Input is the keyboard state handed to Game.Update each frame.
//
// Most terminals report presses and auto-repeats but never releases, and the OS
// repeats only the last key pressed, so holding two keys at once does not work
// there. Terminals that implement the kitty keyboard protocol report releases;
// there Input tracks exactly which keys are down. Letters keep their case in
// the first kind of terminal (shift+w arrives as 'W'), not in the second.
type Input struct {
	now      time.Time
	exact    bool // true when the terminal reports key releases
	lastSeen map[Key]time.Time
	held     map[Key]bool
	pressed  map[Key]bool
}

// newInput returns an Input with no keys held. exact says whether the terminal
// reports key releases.
func newInput(exact bool) *Input {
	return &Input{
		exact:    exact,
		lastSeen: make(map[Key]time.Time),
		held:     make(map[Key]bool),
		pressed:  make(map[Key]bool),
	}
}

// event records a key event at the given time.
func (in *Input) event(e keyEvent, at time.Time) {
	switch e.kind {
	case kindRelease:
		delete(in.held, e.key)
	case kindFocusLost:
		in.releaseAll()
	case kindFocusGained:
		// Nothing to do: keys held across a focus change send fresh repeats.
	default:
		in.lastSeen[e.key] = at
		in.held[e.key] = true
		in.pressed[e.key] = true
	}
}

// releaseAll marks every key as released, for when the terminal loses focus
// and the real releases will never arrive.
func (in *Input) releaseAll() {
	clear(in.held)
}

// beginFrame sets the time IsDown measures against.
// Passed in rather than read from the clock so tests can use exact times.
func (in *Input) beginFrame(now time.Time) {
	in.now = now
}

// endFrame forgets this frame's presses so Pressed fires once per press.
func (in *Input) endFrame() {
	clear(in.pressed)
}

// IsDown reports whether k is held. Where the terminal reports releases this is
// exact. Elsewhere a new hold reads up briefly before the OS starts
// auto-repeating, and a released key stays down for up to holdTimeout.
func (in *Input) IsDown(k Key) bool {
	if in.exact {
		// pressed keeps a tap shorter than one frame visible.
		return in.held[k] || in.pressed[k]
	}
	t, ok := in.lastSeen[k]
	return ok && in.now.Sub(t) < holdTimeout
}

// Pressed reports whether k was pressed or repeated since the previous frame.
func (in *Input) Pressed(k Key) bool {
	return in.pressed[k]
}

// readInput streams chunks read from r until it errors
// then closes the channel.
func readInput(r io.Reader) <-chan []byte {
	ch := make(chan []byte)
	// the goroutine stays blocked in Read after Run returns;
	// harmless unless Run is called twice.
	go func() {
		defer close(ch)
		buf := make([]byte, 64)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				ch <- bytes.Clone(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()
	return ch
}
