package engine

type terminal struct{}

func newTerminal() *terminal {
	return &terminal{}
}

func (t *terminal) Setup() {
	sb := newSequenceBuilder()
	sb.SwitchToAlternateScreen()
	sb.EraseInDisplay(2)
	sb.HideCursor()
	sb.Execute()
}

func (t *terminal) Teardown() {
	sb := newSequenceBuilder()
	sb.ShowCursor()
	sb.SwitchToMainScreen()
	sb.Execute()
}
