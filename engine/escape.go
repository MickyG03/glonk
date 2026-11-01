package engine

import (
	"fmt"
	"strings"
)

const (
	escapeSeq          = "\x1b["
	hideCursorSeq      = "\x1b[?25l"
	showCursorSeq      = "\x1b[?25h"
	alternateScreenSeq = "\x1b[?1049h"
	mainScreenSeq      = "\x1b[?1049l"
)

type sequenceBuilder struct {
	sequence strings.Builder
}

func newSequenceBuilder() *sequenceBuilder {
	return &sequenceBuilder{}
}

func (sb *sequenceBuilder) Execute() {
	fmt.Print(sb.sequence.String())
}

func (sb *sequenceBuilder) Print(s string) {
	sb.sequence.WriteString(s)
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
