package engine

import "fmt"

type Terminal struct{}

func NewTerminal() *Terminal {
	return &Terminal{}
}

func (t *Terminal) Setup() {
	sb := NewSequenceBuilder()
	sb.SwitchToAlternateScreen()
	sb.EraseInDisplay(2)
	sb.HideCursor()
	sb.Execute()
}

func (t *Terminal) Teardown() {
	sb := NewSequenceBuilder()
	sb.ShowCursor()
	sb.SwitchToMainScreen()
	sb.Execute()

	fmt.Println("\nGOODBYE!\n")
}
