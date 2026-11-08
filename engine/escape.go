package engine

import (
	"io"
	"slices"
	"strconv"
	"unicode/utf8"
)

const (
	escapeSeq          = "\x1b["
	hideCursorSeq      = "\x1b[?25l"
	showCursorSeq      = "\x1b[?25h"
	alternateScreenSeq = "\x1b[?1049h"
	mainScreenSeq      = "\x1b[?1049l"
	disableWrapSeq     = "\x1b[?7l"
	enableWrapSeq      = "\x1b[?7h"
	resetSeq           = "\x1b[0m"
	defaultForeSeq     = "\x1b[39m"
	defaultBackSeq     = "\x1b[49m"

	// Kitty keyboard protocol: ask for the current flags (a supporting terminal
	// answers ESC [ ? flags u), ask for device attributes (every terminal answers),
	// and pop the flags we pushed.
	queryKeyboardSeq   = "\x1b[?u"
	queryAttributesSeq = "\x1b[c"
	popKeyboardSeq     = "\x1b[<u"

	// Focus reporting: the terminal sends ESC [ I when the window gains focus and
	// ESC [ O when it loses it.
	enableFocusSeq  = "\x1b[?1004h"
	disableFocusSeq = "\x1b[?1004l"
)

// sequenceBuilder collects the bytes of one write to the terminal. It keeps its
// buffer between uses, so a builder that is reset and filled again each frame
// allocates only until it has grown to the size of a frame.
type sequenceBuilder struct {
	buf []byte
}

func newSequenceBuilder() *sequenceBuilder {
	return &sequenceBuilder{}
}

// reset empties the builder and keeps its buffer.
func (sb *sequenceBuilder) reset() {
	sb.buf = sb.buf[:0]
}

// String returns what has been built so far.
func (sb *sequenceBuilder) String() string {
	return string(sb.buf)
}

// Len returns how many bytes have been built so far.
func (sb *sequenceBuilder) Len() int {
	return len(sb.buf)
}

// Flush writes everything built so far to w in a single Write. The builder is left
// as it is; reset it to build the next one.
func (sb *sequenceBuilder) Flush(w io.Writer) error {
	_, err := w.Write(sb.buf)
	return err
}

func (sb *sequenceBuilder) Print(s string) {
	sb.buf = append(sb.buf, s...)
}

// PrintRune appends r as UTF-8. A value that is not a valid character comes out as
// U+FFFD, the same as string(r).
func (sb *sequenceBuilder) PrintRune(r rune) {
	sb.buf = utf8.AppendRune(sb.buf, r)
}

// appendCSI starts a control sequence with one number and ends it with final.
func (sb *sequenceBuilder) appendCSI(n int, final byte) {
	sb.buf = append(sb.buf, escapeSeq...)
	sb.buf = strconv.AppendInt(sb.buf, int64(n), 10)
	sb.buf = append(sb.buf, final)
}

func (sb *sequenceBuilder) MoveTo(n, m int) {
	sb.buf = append(sb.buf, escapeSeq...)
	sb.buf = strconv.AppendInt(sb.buf, int64(n), 10)
	sb.buf = append(sb.buf, ';')
	sb.buf = strconv.AppendInt(sb.buf, int64(m), 10)
	sb.buf = append(sb.buf, 'H')
}

func (sb *sequenceBuilder) EraseInDisplay(n int) {
	sb.appendCSI(n, 'J')
}

func (sb *sequenceBuilder) SwitchToAlternateScreen() {
	sb.Print(alternateScreenSeq)
}

func (sb *sequenceBuilder) SwitchToMainScreen() {
	sb.Print(mainScreenSeq)
}

// DisableAutowrap makes a character written in the last column stay there, so a row
// that is too long is cut off at the edge instead of continuing on the next line.
func (sb *sequenceBuilder) DisableAutowrap() {
	sb.Print(disableWrapSeq)
}

// EnableAutowrap turns line wrapping back on, as the shell expects it.
func (sb *sequenceBuilder) EnableAutowrap() {
	sb.Print(enableWrapSeq)
}

func (sb *sequenceBuilder) HideCursor() {
	sb.Print(hideCursorSeq)
}

func (sb *sequenceBuilder) ShowCursor() {
	sb.Print(showCursorSeq)
}

// appendPaletteColor writes the sequence that sets the foreground (38) or the
// background (48) to palette color n.
func (sb *sequenceBuilder) appendPaletteColor(base string, n uint8) {
	sb.buf = append(sb.buf, escapeSeq...)
	sb.buf = append(sb.buf, base...)
	sb.buf = strconv.AppendUint(sb.buf, uint64(n), 10)
	sb.buf = append(sb.buf, 'm')
}

func (sb *sequenceBuilder) SetForeground(n uint8) {
	sb.appendPaletteColor("38;5;", n)
}

// SetBackground sets the background to palette color n.
func (sb *sequenceBuilder) SetBackground(n uint8) {
	sb.appendPaletteColor("48;5;", n)
}

// DefaultForeground returns the foreground to the terminal default and leaves the background alone.
func (sb *sequenceBuilder) DefaultForeground() {
	sb.Print(defaultForeSeq)
}

// DefaultBackground returns the background to the terminal default and leaves the foreground alone.
func (sb *sequenceBuilder) DefaultBackground() {
	sb.Print(defaultBackSeq)
}

func (sb *sequenceBuilder) ResetStyle() {
	sb.Print(resetSeq)
}

// Grow reserves room for n more bytes so a frame is built in one allocation
// instead of the builder doubling its way up.
func (sb *sequenceBuilder) Grow(n int) {
	sb.buf = slices.Grow(sb.buf, n)
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
	sb.buf = append(sb.buf, escapeSeq...)
	sb.buf = append(sb.buf, '>')
	sb.buf = strconv.AppendInt(sb.buf, int64(flags), 10)
	sb.buf = append(sb.buf, 'u')
}

// PopKeyboardFlags restores the keyboard mode that was active before the last push.
func (sb *sequenceBuilder) PopKeyboardFlags() {
	sb.Print(popKeyboardSeq)
}

// EnableFocusReporting asks the terminal to report when its window gains or loses focus.
func (sb *sequenceBuilder) EnableFocusReporting() {
	sb.Print(enableFocusSeq)
}

// DisableFocusReporting turns focus reporting back off.
func (sb *sequenceBuilder) DisableFocusReporting() {
	sb.Print(disableFocusSeq)
}
