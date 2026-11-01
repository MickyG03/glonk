package engine

import (
	"slices"
	"testing"
)

func TestParseKeys(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []Key
	}{
		{"printable", "wa", []Key{'w', 'a'}},
		{"arrows", "\x1b[A\x1b[D", []Key{KeyUp, KeyLeft}},
		{"ss3 arrow", "\x1bOB", []Key{KeyDown}},
		{"modified arrow", "\x1b[1;5C", []Key{KeyRight}},
		{"unknown csi skipped", "\x1b[15~x", []Key{'x'}},
		{"lone escape", "\x1b", []Key{KeyEscape}},
		{"escape then key", "\x1bx", []Key{KeyEscape, 'x'}},
		{"truncated csi", "\x1b[", nil},
		{"controls", "\x03\r\x7f\t", []Key{KeyCtrlC, KeyEnter, KeyBackspace, KeyTab}},
		{"other control dropped", "\x01", nil},
		{"utf-8", "é", []Key{'é'}},
		{"invalid byte dropped", "\xff", nil},
		{"empty", "", nil},
		{"truncated ss3", "\x1bO", []Key{KeyEscape, 'O'}},
		{"backspace ctrl-h", "\x08", []Key{KeyBackspace}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseKeys([]byte(tt.in))

			if !slices.Equal(got, tt.want) {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}
