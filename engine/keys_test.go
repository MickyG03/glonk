package engine

import (
	"slices"
	"testing"
)

func presses(keys ...Key) []keyEvent {
	events := make([]keyEvent, len(keys))
	for i, k := range keys {
		events[i] = keyEvent{k, kindPress}
	}
	return events
}

func TestParseEvents(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []keyEvent
	}{
		{"printable", "wa", presses('w', 'a')},
		{"arrows", "\x1b[A\x1b[D", presses(KeyUp, KeyLeft)},
		{"ss3 arrow", "\x1bOB", presses(KeyDown)},
		{"modified arrow", "\x1b[1;5C", presses(KeyRight)},
		{"unknown csi skipped", "\x1b[15~x", presses('x')},
		{"lone escape", "\x1b", presses(KeyEscape)},
		{"escape then key", "\x1bx", presses(KeyEscape, 'x')},
		{"truncated csi", "\x1b[", nil},
		{"controls", "\x03\r\x7f\t", presses(KeyCtrlC, KeyEnter, KeyBackspace, KeyTab)},
		{"other control dropped", "\x01", nil},
		{"utf-8", "é", presses('é')},
		{"invalid byte dropped", "\xff", nil},
		{"empty", "", nil},
		{"truncated ss3", "\x1bO", presses(KeyEscape, 'O')},
		{"backspace ctrl-h", "\x08", presses(KeyBackspace)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEvents([]byte(tt.in))

			if !slices.Equal(got, tt.want) {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

// Sequences a terminal sends once the kitty keyboard protocol is on (flags 1|2|8).
func TestParseEventsKitty(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []keyEvent
	}{
		{"letter press", "\x1b[119u", presses('w')},
		{"letter press explicit", "\x1b[119;1:1u", presses('w')},
		{"letter repeat", "\x1b[119;1:2u", []keyEvent{{'w', kindRepeat}}},
		{"letter release", "\x1b[119;1:3u", []keyEvent{{'w', kindRelease}}},
		{"shift stays unshifted", "\x1b[119;2u", presses('w')},
		{"shift release", "\x1b[119;2:3u", []keyEvent{{'w', kindRelease}}},
		{"caps lock ignored", "\x1b[119;65u", presses('w')},
		{"num lock ignored", "\x1b[119;129u", presses('w')},
		{"shift and caps lock", "\x1b[119;66u", presses('w')},
		{"two events in one read", "\x1b[119;1:3u\x1b[101u", []keyEvent{{'w', kindRelease}, {'e', kindPress}}},
		{"escape", "\x1b[27u", presses(KeyEscape)},
		{"enter", "\x1b[13u", presses(KeyEnter)},
		{"tab", "\x1b[9u", presses(KeyTab)},
		{"backspace", "\x1b[127u", presses(KeyBackspace)},
		{"enter release", "\x1b[13;1:3u", []keyEvent{{KeyEnter, kindRelease}}},
		{"ctrl-c", "\x1b[99;5u", presses(KeyCtrlC)},
		{"ctrl-c with caps lock", "\x1b[99;69u", presses(KeyCtrlC)},
		{"ctrl-c release", "\x1b[99;5:3u", []keyEvent{{KeyCtrlC, kindRelease}}},
		{"ctrl-w dropped", "\x1b[119;5u", nil},
		{"alt-w dropped", "\x1b[119;3u", nil},
		{"super-w dropped", "\x1b[119;9u", nil},
		{"ctrl-alt-c dropped", "\x1b[99;7u", nil},
		{"unicode letter", "\x1b[233u", presses('é')},
		{"arrow press", "\x1b[1;1A", presses(KeyUp)},
		{"arrow repeat", "\x1b[1;1:2B", []keyEvent{{KeyDown, kindRepeat}}},
		{"arrow release", "\x1b[1;1:3C", []keyEvent{{KeyRight, kindRelease}}},
		{"arrow release with shift", "\x1b[1;2:3D", []keyEvent{{KeyLeft, kindRelease}}},
		{"bare arrow", "\x1b[A", presses(KeyUp)},
		{"lone modifier key dropped", "\x1b[57441;1u", nil},
		{"keypad key dropped", "\x1b[57399u", nil},
		{"first private use code dropped", "\x1b[57344u", nil},
		{"space key", "\x1b[32u", presses(' ')},
		{"just below space dropped", "\x1b[31u", nil},
		{"private marker lt dropped", "\x1b[<1u", nil},
		{"private marker gt dropped", "\x1b[>11u", nil},
		{"private arrow lookalike dropped", "\x1b[?1A", nil},
		{"private lt arrow lookalike dropped", "\x1b[<1A", nil},
		{"private reply dropped", "\x1b[?11u", nil},
		{"device attributes reply dropped", "\x1b[?62;4c", nil},
		{"missing code dropped", "\x1b[u", nil},
		{"empty code with mods dropped", "\x1b[;5u", nil},
		{"only separators dropped", "\x1b[1;u", nil},
		{"control code dropped", "\x1b[5u", nil},
		{"surrogate dropped", "\x1b[55296u", nil},
		{"past max rune dropped", "\x1b[1114112u", nil},
		{"huge number dropped", "\x1b[99999999999999999999u", nil},
		{"negative code dropped", "\x1b[-5u", nil},
		{"unknown event type is press", "\x1b[119;1:9u", presses('w')},
		{"event type without modifiers value", "\x1b[119;:3u", []keyEvent{{'w', kindRelease}}},
		{"unknown final dropped", "\x1b[119;1:3z", nil},
		{"focus gained", "\x1b[I", []keyEvent{{KeyUnknown, kindFocusGained}}},
		{"focus lost", "\x1b[O", []keyEvent{{KeyUnknown, kindFocusLost}}},
		{"focus lost between keys", "\x1b[119u\x1b[O\x1b[119;1:3u", []keyEvent{{'w', kindPress}, {KeyUnknown, kindFocusLost}, {'w', kindRelease}}},
		{"focus with parameters dropped", "\x1b[1I", nil},
		{"junk then key", "\x1b[?11u\x1b[119u", presses('w')},
		{"key then truncated", "\x1b[119u\x1b[11", presses('w')},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEvents([]byte(tt.in))

			if !slices.Equal(got, tt.want) {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}
