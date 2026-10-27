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

type SequenceBuilder struct {
	sequence strings.Builder
}

func NewSequenceBuilder() *SequenceBuilder {
	return &SequenceBuilder{}
}

func (sb *SequenceBuilder) Execute() {
	fmt.Print(sb.sequence.String())
}

func (sb *SequenceBuilder) String() string {
	return sb.sequence.String()
}

func (sb *SequenceBuilder) Print(s string) {
	sb.sequence.WriteString(s)
}

func (sb *SequenceBuilder) Up(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dA", n)
}

func (sb *SequenceBuilder) Down(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dB", n)
}

func (sb *SequenceBuilder) Forward(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dC", n)
}

func (sb *SequenceBuilder) Back(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dD", n)
}

func (sb *SequenceBuilder) NextLine(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dE", n)
}

func (sb *SequenceBuilder) PreviousLine(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dF", n)
}

func (sb *SequenceBuilder) MoveToColumn(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dG", n)
}

func (sb *SequenceBuilder) MoveTo(n, m int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%d;%dH", n, m)
}

func (sb *SequenceBuilder) EraseInDisplay(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dJ", n)
}

func (sb *SequenceBuilder) EraseInLine(n int) {
	fmt.Fprintf(&sb.sequence, escapeSeq+"%dK", n)
}

func (sb *SequenceBuilder) SwitchToAlternateScreen() {
	sb.Print(alternateScreenSeq)
}

func (sb *SequenceBuilder) SwitchToMainScreen() {
	sb.Print(mainScreenSeq)
}

func (sb *SequenceBuilder) HideCursor() {
	sb.Print(hideCursorSeq)
}

func (sb *SequenceBuilder) ShowCursor() {
	sb.Print(showCursorSeq)
}
