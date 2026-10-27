package engine

import (
	"fmt"
	"strings"
)

const (
	ESCAPE_SEQ           = "\x1b["
	HIDE_CURSOR_SEQ      = "\x1b[?25l"
	SHOW_CURSOR_SEQ      = "\x1b[?25h"
	ALTERNATE_SCREEN_SEQ = "\x1b[?1049h"
	MAIN_SCREEN_SEQ      = "\x1b[?1049l"
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
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%dA", n)
}

func (sb *SequenceBuilder) Down(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%dB", n)
}

func (sb *SequenceBuilder) Forward(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%dC", n)
}

func (sb *SequenceBuilder) Back(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%dD", n)
}

func (sb *SequenceBuilder) NextLine(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%dE", n)
}

func (sb *SequenceBuilder) PreviousLine(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%dF", n)
}

func (sb *SequenceBuilder) MoveToColumn(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%dG", n)
}

func (sb *SequenceBuilder) MoveTo(n, m int) {
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%d;%dH", n, m)
}

func (sb *SequenceBuilder) EraseInDisplay(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%dJ", n)
}

func (sb *SequenceBuilder) EraseInLine(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE_SEQ+"%dK", n)
}

func (sb *SequenceBuilder) SwitchToAlternateScreen() {
	sb.Print(ALTERNATE_SCREEN_SEQ)
}

func (sb *SequenceBuilder) SwitchToMainScreen() {
	sb.Print(MAIN_SCREEN_SEQ)
}

func (sb *SequenceBuilder) HideCursor() {
	sb.Print(HIDE_CURSOR_SEQ)
}

func (sb *SequenceBuilder) ShowCursor() {
	sb.Print(SHOW_CURSOR_SEQ)
}
