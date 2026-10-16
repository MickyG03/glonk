package engine

import (
	"fmt"
	"strings"
)

const (
	ESCAPE      = "\x1b["
	HIDE_CURSOR = "\x1b[?25l"
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
	fmt.Fprintf(&sb.sequence, ESCAPE+"%dA", n)
}

func (sb *SequenceBuilder) Down(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE+"%dB", n)
}

func (sb *SequenceBuilder) Forward(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE+"%dC", n)
}

func (sb *SequenceBuilder) Back(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE+"%dD", n)
}

func (sb *SequenceBuilder) NextLine(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE+"%dE", n)
}

func (sb *SequenceBuilder) PreviousLine(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE+"%dF", n)
}

func (sb *SequenceBuilder) MoveToColumn(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE+"%dG", n)
}

func (sb *SequenceBuilder) MoveTo(n, m int) {
	fmt.Fprintf(&sb.sequence, ESCAPE+"%d;%dH", n, m)
}

func (sb *SequenceBuilder) EraseInDisplay(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE+"%dJ", n)
}

func (sb *SequenceBuilder) EraseInLine(n int) {
	fmt.Fprintf(&sb.sequence, ESCAPE+"%dK", n)
}
