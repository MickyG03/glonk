package engine

import "io"

type terminal struct {
	out io.Writer
}

func newTerminal(out io.Writer) *terminal {
	return &terminal{
		out: out,
	}
}

func (t *terminal) Setup() error {
	sb := newSequenceBuilder()
	sb.SwitchToAlternateScreen()
	sb.EraseInDisplay(2)
	sb.HideCursor()
	return sb.Flush(t.out)
}

func (t *terminal) Teardown() error {
	sb := newSequenceBuilder()
	sb.ShowCursor()
	sb.SwitchToMainScreen()
	return sb.Flush(t.out)
}
