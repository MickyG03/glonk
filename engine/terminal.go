package engine

import (
	"errors"
	"io"
	"os"

	"golang.org/x/term"
)

type terminal struct {
	fd    int
	state *term.State
	out   io.Writer
}

func newTerminal(in *os.File, out io.Writer) *terminal {
	return &terminal{
		fd:  int(in.Fd()),
		out: out,
	}
}

func (t *terminal) Setup() error {
	var err error
	t.state, err = term.MakeRaw(t.fd)
	if err != nil {
		return err
	}
	sb := newSequenceBuilder()
	sb.SwitchToAlternateScreen()
	sb.EraseInDisplay(2)
	sb.HideCursor()
	err = sb.Flush(t.out)
	if err != nil {
		// ignoring the restore's error because we care about the flush error more
		term.Restore(t.fd, t.state)
		return err
	}
	return nil
}

func (t *terminal) Teardown() error {
	sb := newSequenceBuilder()
	sb.ShowCursor()
	sb.SwitchToMainScreen()
	flushErr := sb.Flush(t.out)
	restoreErr := term.Restore(t.fd, t.state)
	return errors.Join(flushErr, restoreErr)
}
