package engine

import (
	"fmt"
	"io"
	"strings"
)

const (
	escapeSeq          = "\x1b["
	hideCursorSeq      = "\x1b[?25l"
	showCursorSeq      = "\x1b[?25h"
	alternateScreenSeq = "\x1b[?1049h"
	mainScreenSeq      = "\x1b[?1049l"
	resetSeq           = "\x1b[0m"

	// Kitty keyboard protocol: ask for the current flags (a supporting terminal
	// answers ESC [ ? flags u), ask for device attributes (every terminal answers),
	// and pop the flags we pushed.
	queryKeyboardSeq   = "\x1b[?u"
	queryAttributesSeq = "\x1b[c"
	popKeyboardSeq     = "\x1b[<u"
)

type sequenceBuilder struct {
	sequence strings.Builder
}

func newSequenceBuilder() *sequenceBuilder {
	return &sequenceBuilder{}
}

func (sb *sequenceBuilder) Flush(w io.Writer) error {
	_, err := io.WriteString(w, sb.sequence.String())
	return err
}

func (sb *sequenceBuilder) Print(s string) {
	sb.sequence.WriteString(s)
}

// PrintRune appends r as UTF-8 without building a string first.
// Invalid runes come out as U+FFFD, the same as string(r).
func (sb *sequenceBuilder) PrintRune(r rune) {
	sb.sequence.WriteRune(r)
}

func (sb *sequenceBuilder) MoveTo(n, m int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%d;%dH", n, m)
}

func (sb *sequenceBuilder) EraseInDisplay(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dJ", n)
}

func (sb *sequenceBuilder) SwitchToAlternateScreen() {
	sb.Print(alternateScreenSeq)
}

func (sb *sequenceBuilder) SwitchToMainScreen() {
	sb.Print(mainScreenSeq)
}

func (sb *sequenceBuilder) HideCursor() {
	sb.Print(hideCursorSeq)
}

func (sb *sequenceBuilder) ShowCursor() {
	sb.Print(showCursorSeq)
}

func (sb *sequenceBuilder) SetForeground(n uint8) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"38;5;%dm", n)
}

func (sb *sequenceBuilder) ResetStyle() {
	sb.Print(resetSeq)
}

// Grow reserves room for n more bytes so a frame is built in one allocation
// instead of the builder doubling its way up.
func (sb *sequenceBuilder) Grow(n int) {
	sb.sequence.Grow(n)
}

// QueryKeyboardProtocol asks whether the terminal speaks the kitty keyboard protocol.
func (sb *sequenceBuilder) QueryKeyboardProtocol() {
	sb.Print(queryKeyboardSeq)
}

// QueryDeviceAttributes asks for the terminal's device attributes. Sent right after
// the protocol query: every terminal answers it, so its reply marks the end of the wait.
func (sb *sequenceBuilder) QueryDeviceAttributes() {
	sb.Print(queryAttributesSeq)
}

// PushKeyboardFlags turns on the given kitty keyboard protocol flags until popped.
func (sb *sequenceBuilder) PushKeyboardFlags(flags int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+">%du", flags)
}

// PopKeyboardFlags restores the keyboard mode that was active before the last push.
func (sb *sequenceBuilder) PopKeyboardFlags() {
	sb.Print(popKeyboardSeq)
}
