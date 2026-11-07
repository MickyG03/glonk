package engine

import (
	"bytes"
	"os"
	"testing"
)

func TestKeyboardSequences(t *testing.T) {
	tests := []struct {
		name  string
		build func(sb *sequenceBuilder)
		want  string
	}{
		{"query protocol", func(sb *sequenceBuilder) { sb.QueryKeyboardProtocol() }, "\x1b[?u"},
		{"query attributes", func(sb *sequenceBuilder) { sb.QueryDeviceAttributes() }, "\x1b[c"},
		{"push flags", func(sb *sequenceBuilder) { sb.PushKeyboardFlags(11) }, "\x1b[>11u"},
		{"push other flags", func(sb *sequenceBuilder) { sb.PushKeyboardFlags(1) }, "\x1b[>1u"},
		{"pop flags", func(sb *sequenceBuilder) { sb.PopKeyboardFlags() }, "\x1b[<u"},
		{"enable focus reporting", func(sb *sequenceBuilder) { sb.EnableFocusReporting() }, "\x1b[?1004h"},
		{"disable focus reporting", func(sb *sequenceBuilder) { sb.DisableFocusReporting() }, "\x1b[?1004l"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := newSequenceBuilder()
			tt.build(sb)

			if got := sb.sequence.String(); got != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, got)
			}
		})
	}
}

func TestEnableKeyboardProtocol(t *testing.T) {
	var out bytes.Buffer
	tty := newTerminal(os.Stdin, &out)

	if err := tty.EnableKeyboardProtocol(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := "\x1b[>11u\x1b[?1004h"; out.String() != want {
		t.Errorf("wanted %q, got %q", want, out.String())
	}
	if !tty.keyboard {
		t.Errorf("wanted the terminal to remember it must pop the flags, got false")
	}
}

// Even a failed write must be remembered, or the flags could be left pushed.
func TestEnableKeyboardProtocolFailedWrite(t *testing.T) {
	tty := newTerminal(os.Stdin, failingWriter{})

	if err := tty.EnableKeyboardProtocol(); err == nil {
		t.Errorf("wanted the write error, got nil")
	}
	if !tty.keyboard {
		t.Errorf("wanted the terminal to remember it must pop the flags, got false")
	}
}

func TestExitSequence(t *testing.T) {
	tests := []struct {
		name     string
		keyboard bool
		want     string
	}{
		{"without the protocol", false, "\x1b[?25h\x1b[?1049l"},
		// Popped before leaving the alternate screen, where the flags were pushed.
		{"with the protocol", true, "\x1b[<u\x1b[?1004l\x1b[?25h\x1b[?1049l"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tty := newTerminal(os.Stdin, &bytes.Buffer{})
			tty.keyboard = tt.keyboard

			if got := tty.exitSequence().sequence.String(); got != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, got)
			}
		})
	}
}
