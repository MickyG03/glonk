package engine

import (
	"bytes"
	"io"
	"time"
)

// Terminals report presses and auto-repeats but never releases, so a key
// counts as held until this long passes without another repeat.
const holdTimeout = 150 * time.Millisecond

// Input is the keyboard state handed to Game.Update each frame.
// Letters keep their case: shift+w arrives as 'W'.
type Input struct {
	now      time.Time
	lastSeen map[Key]time.Time
	pressed  map[Key]bool
}

// newInput returns an Input with no keys held.
func newInput() *Input {
	return &Input{
		lastSeen: make(map[Key]time.Time),
		pressed:  make(map[Key]bool),
	}
}

// press records a press or auto-repeat of k at the given time.
func (in *Input) press(k Key, at time.Time) {
	in.lastSeen[k] = at
	in.pressed[k] = true
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

// IsDown reports whether k is held. A new hold reads up briefly before the
// OS starts auto-repeating, and a released key stays down for up to holdTimeout.
func (in *Input) IsDown(k Key) bool {
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
