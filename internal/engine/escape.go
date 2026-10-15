package engine

import "fmt"

const (
	ESCAPE      = "\x1b["
	HIDE_CURSOR = "\x1b[?25l"
)

type SequenceBuilder struct {
	sequence string
}

func NewSequenceBuilder() *SequenceBuilder {
	return &SequenceBuilder{}
}

func (sb *SequenceBuilder) Execute() {
	fmt.Print(sb.sequence)
}

func (sb *SequenceBuilder) String() string {
	return fmt.Sprintf("%q\n", sb.sequence)
}

func (sb *SequenceBuilder) Print(s string) {
	sb.sequence += s
}

func (sb *SequenceBuilder) Up(n int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%dA", n)
}

func (sb *SequenceBuilder) Down(n int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%dB", n)
}

func (sb *SequenceBuilder) Forward(n int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%dC", n)
}

func (sb *SequenceBuilder) Back(n int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%dD", n)
}

func (sb *SequenceBuilder) NextLine(n int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%dE", n)
}

func (sb *SequenceBuilder) PreviousLine(n int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%dF", n)
}

func (sb *SequenceBuilder) MoveToColumn(n int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%dG", n)
}

func (sb *SequenceBuilder) MoveTo(n, m int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%d;%dH", n, m)
}

func (sb *SequenceBuilder) EraseInDisplay(n int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%dJ", n)
}

func (sb *SequenceBuilder) EraseInLine(n int) {
	sb.sequence += ESCAPE + fmt.Sprintf("%dK", n)
}
