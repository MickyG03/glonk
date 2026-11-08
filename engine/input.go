package engine

import (
	"bytes"
	"io"
	"os"
	"sync"
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
//
// To test a game's Update, build an Input with a [Keyboard].
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
// auto-repeating, and a released key stays down for up to about 150 ms.
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

// inputAction is what Run must do once input has been handled.
type inputAction uint8

const (
	actionNone    inputAction = iota
	actionQuit                // Ctrl-C
	actionSuspend             // Ctrl-Z
)

// inputFeed turns the raw bytes read from the terminal into Input updates. It holds
// a keyParser, and a timer that gives up waiting for the rest of an unfinished
// sequence.
type inputFeed struct {
	parser     keyParser
	in         *Input
	wait       time.Duration
	timer      *time.Timer
	canSuspend bool // whether Ctrl-Z may stop the program
	refresh    bool // the window came back to the front since refreshNeeded was last asked
}

func newInputFeed(in *Input) *inputFeed {
	f := &inputFeed{in: in, wait: escapeTimeout, canSuspend: suspendSupported}
	f.timer = time.NewTimer(f.wait)
	f.timer.Stop()
	return f
}

// feed handles one chunk read from the terminal and says whether it held Ctrl-C
// or Ctrl-Z. What follows that key in the chunk is dropped.
func (f *inputFeed) feed(chunk []byte, now time.Time) inputAction {
	action := f.apply(f.parser.Feed(chunk), now)

	f.stop()
	if f.parser.Pending() {
		f.timer.Reset(f.wait)
	}
	return action
}

// expired is the channel that fires when the wait for the rest of a sequence is over.
func (f *inputFeed) expired() <-chan time.Time {
	return f.timer.C
}

// expire gives up waiting: whatever is held is taken as it stands, so a lone ESC
// becomes the Escape key. It says whether that produced Ctrl-C or Ctrl-Z.
func (f *inputFeed) expire(now time.Time) inputAction {
	return f.apply(f.parser.Flush(), now)
}

// reset forgets every held key and any half-read sequence, for when input picks up
// after a gap in which releases and the rest of a sequence were lost.
func (f *inputFeed) reset() {
	f.stop()
	f.parser = keyParser{}
	f.in.releaseAll()
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

// refreshNeeded reports whether the terminal said its window came back to the front
// since the last time it was asked. What the window showed may have been disturbed
// meanwhile, so the screen is worth drawing again in full.
func (f *inputFeed) refreshNeeded() bool {
	r := f.refresh
	f.refresh = false
	return r
}

func (f *inputFeed) apply(events []keyEvent, now time.Time) inputAction {
	for _, e := range events {
		if e.kind == kindFocusGained {
			f.refresh = true
		}
		switch {
		case e.key == KeyCtrlC && e.kind != kindRelease:
			return actionQuit
		case e.key == keyCtrlZ:
			// Only the press suspends: a repeat would suspend again the moment the
			// program resumes with the key still down.
			if e.kind == kindPress && f.canSuspend {
				return actionSuspend
			}
		default:
			f.in.event(e, now)
		}
	}
	return actionNone
}

// readFrom streams what r yields in chunks. A file gets the platform reader, which
// can be stopped cleanly; anything else can only be read with blocking reads.
func readFrom(r io.Reader) (chunks <-chan []byte, stop func()) {
	if f, ok := r.(*os.File); ok {
		return readInput(f)
	}
	return readBlocking(r)
}

// readBlocking streams chunks read from r until it errors, then closes the channel.
// Its goroutine sits in a blocking Read, which nothing can interrupt, so stop only
// lets go of a chunk nobody is receiving: a goroutine waiting in Read stays behind
// and takes the next input. readInput avoids that where it can.
func readBlocking(r io.Reader) (chunks <-chan []byte, stop func()) {
	ch := make(chan []byte)
	done := make(chan struct{})
	go func() {
		defer close(ch)
		buf := make([]byte, readBuffer)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				select {
				case ch <- bytes.Clone(buf[:n]):
				case <-done:
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	var once sync.Once
	return ch, func() { once.Do(func() { close(done) }) }
}
