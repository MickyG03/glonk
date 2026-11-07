package engine

import (
	"errors"
	"io"
	"os"

	"golang.org/x/term"
)

// keyboardFlags is what we ask of the kitty keyboard protocol: 1 disambiguate
// escape codes, 2 report press/repeat/release, 8 report all keys as escape codes
// (without it letters send no release events).
const keyboardFlags = 1 | 2 | 8

type terminal struct {
	fd       int
	state    *term.State
	out      io.Writer
	keyboard bool // whether we pushed protocol flags that Teardown must pop
}

func newTerminal(in *os.File, out io.Writer) *terminal {
	return &terminal{
		fd:  int(in.Fd()),
		out: out,
	}
}

func (t *terminal) Setup() error {
	var err error
	t.state, err = term.MakeRaw(t.fd)
	if err != nil {
		return err
	}
	sb := newSequenceBuilder()
	sb.SwitchToAlternateScreen()
	sb.EraseInDisplay(2)
	sb.HideCursor()
	err = sb.Flush(t.out)
	if err != nil {
		// ignoring the restore's error because we care about the flush error more
		term.Restore(t.fd, t.state)
		return err
	}
	return nil
}

// EnableKeyboardProtocol turns on press/repeat/release reporting, plus focus
// reporting: a release that happens while the window is unfocused never arrives,
// so Input needs to know when to forget every held key. Call it only after a
// probe says the terminal supports the protocol.
func (t *terminal) EnableKeyboardProtocol() error {
	// Set first: if the write fails halfway, popping an unpushed flag is harmless
	// (the protocol treats it as a reset) but leaving one pushed would not be.
	t.keyboard = true
	sb := newSequenceBuilder()
	sb.PushKeyboardFlags(keyboardFlags)
	sb.EnableFocusReporting()
	return sb.Flush(t.out)
}

// exitSequence is what Teardown writes: flags are popped first because they were
// pushed on the alternate screen, which has its own stack.
func (t *terminal) exitSequence() *sequenceBuilder {
	sb := newSequenceBuilder()
	if t.keyboard {
		sb.PopKeyboardFlags()
		sb.DisableFocusReporting()
	}
	sb.ShowCursor()
	sb.SwitchToMainScreen()
	return sb
}

func (t *terminal) Teardown() error {
	flushErr := t.exitSequence().Flush(t.out)
	restoreErr := term.Restore(t.fd, t.state)
	return errors.Join(flushErr, restoreErr)
}

// Size returns the terminal's width and height in cells.
func (t *terminal) Size() (int, int, error) {
	return term.GetSize(t.fd)
}

// Clear erases the screen. Needed after a resize because present only writes
// inside the frame, so stale cells outside a smaller frame would stay.
func (t *terminal) Clear() error {
	sb := newSequenceBuilder()
	sb.EraseInDisplay(2)
	return sb.Flush(t.out)
}
