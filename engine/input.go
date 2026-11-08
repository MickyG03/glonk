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
	case kindFocusLost, kindReset:
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

// escapeTimeout is how long an unfinished escape sequence is waited for. The Escape
// key and the start of an arrow key are both a lone ESC byte, so after seeing one
// the input is only taken as the Escape key once this long has passed with nothing
// more. Terminals send a whole key at once, so any value well above the time a read
// takes works; it is the delay a real Escape press waits.
const escapeTimeout = 50 * time.Millisecond

// readBuffer is how many bytes are read from the terminal at once. Keys arrive a few
// bytes at a time, so a large buffer is only filled by pasting or a backlog.
const readBuffer = 4096

// inputFeed turns the raw bytes read from the terminal into Input updates. It holds
// a keyParser, and a timer that gives up waiting for the rest of an unfinished
// sequence.
type inputFeed struct {
	parser keyParser
	in     *Input
	wait   time.Duration
	timer  *time.Timer
}

func newInputFeed(in *Input) *inputFeed {
	f := &inputFeed{in: in, wait: escapeTimeout}
	f.timer = time.NewTimer(f.wait)
	f.timer.Stop()
	return f
}

// feed handles one chunk read from the terminal and reports whether it held Ctrl-C.
func (f *inputFeed) feed(chunk []byte, now time.Time) (quit bool) {
	quit = f.apply(f.parser.Feed(chunk), now)

	f.stop()
	if f.parser.Pending() {
		f.timer.Reset(f.wait)
	}
	return quit
}

// expired is the channel that fires when the wait for the rest of a sequence is over.
func (f *inputFeed) expired() <-chan time.Time {
	return f.timer.C
}

// expire gives up waiting: whatever is held is taken as it stands, so a lone ESC
// becomes the Escape key. It reports whether that produced Ctrl-C.
func (f *inputFeed) expire(now time.Time) (quit bool) {
	return f.apply(f.parser.Flush(), now)
}

// stop cancels the wait, draining the channel so a tick that was already on its way
// cannot be mistaken for a later wait ending (needed before Go 1.23 timer semantics).
func (f *inputFeed) stop() {
	if !f.timer.Stop() {
		select {
		case <-f.timer.C:
		default:
		}
	}
}

func (f *inputFeed) apply(events []keyEvent, now time.Time) (quit bool) {
	for _, e := range events {
		if e.key == KeyCtrlC && e.kind != kindRelease {
			return true
		}
		f.in.event(e, now)
	}
	return false
}

// readInput streams chunks read from r until it errors
// then closes the channel.
func readInput(r io.Reader) <-chan []byte {
	ch := make(chan []byte)
	// the goroutine stays blocked in Read after Run returns;
	// harmless unless Run is called twice.
	go func() {
		defer close(ch)
		buf := make([]byte, readBuffer)
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
