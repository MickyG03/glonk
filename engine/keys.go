package engine

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Key is a keypress: printable keys are their own rune, special keys are negative.
type Key rune

// Special keys. Negative values so they can never collide with a real rune.
const (
	KeyUnknown Key = -1 - iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEscape
	KeyEnter
	KeyBackspace
	KeyTab
	// Raw mode turns off the terminal's signal handling, so Ctrl-C arrives
	// as a byte instead of SIGINT.
	KeyCtrlC
	// keyCtrlZ asks Run to suspend. Like Ctrl-C it is Run's, never the game's.
	keyCtrlZ
)

// keyKind says what happened to a key. Only terminals speaking the kitty
// keyboard protocol report repeats and releases; everything else sends presses.
// The focus kinds are not about a key: they carry KeyUnknown and tell Input the
// window gained or lost focus, which we only ask for alongside the protocol.
// kindReset also carries KeyUnknown: part of the input was thrown away, so any
// release in it may have been lost and Input must forget which keys are held.
type keyKind uint8

const (
	kindPress keyKind = iota
	kindRepeat
	kindRelease
	kindFocusGained
	kindFocusLost
	kindReset
)

// keyEvent is one decoded key event.
type keyEvent struct {
	key  Key
	kind keyKind
}

// Kitty keyboard protocol modifier bits, after subtracting the protocol's +1 offset.
// Shift and the lock keys are deliberately absent: they do not make a shortcut.
const (
	modAlt   = 1 << 1
	modCtrl  = 1 << 2
	modSuper = 1 << 3
	modHyper = 1 << 4
	modMeta  = 1 << 5

	shortcutMods = modAlt | modCtrl | modSuper | modHyper | modMeta
)

// privateUseStart is where the protocol puts keys with no character, such as a
// lone Shift or a keypad key. We have no use for them.
const privateUseStart = 57344

// arrow maps the final byte of an arrow-key escape sequence to its Key.
func arrow(c byte) Key {
	switch c {
	case 'A':
		return KeyUp
	case 'B':
		return KeyDown
	case 'C':
		return KeyRight
	case 'D':
		return KeyLeft
	default:
		return KeyUnknown
	}
}

// maxPendingInput bounds how many bytes of an unfinished sequence keyParser holds.
// Real key sequences are a few bytes long; more than this is garbage.
const maxPendingInput = 128

// parseEvents decodes input that is known to be complete into key events. For
// input that arrives in pieces use a keyParser.
func parseEvents(b []byte) []keyEvent {
	events, _ := parse(b, true)
	return events
}

// parse decodes raw terminal input into key events and reports how many bytes it
// used.
//
// A sequence cut off at the end of b is not used when final is false: the caller
// keeps those bytes and tries again once more have arrived. When final is true
// nothing more is coming, so a lone ESC is the Escape key, an unfinished UTF-8
// rune is dropped, and an unfinished CSI sequence is dropped with a reset event,
// since the release it may have carried is lost.
//
// ESC followed by a key is an Alt chord (Alt+x arrives as ESC x). Games do not get
// chords, and reading them as Escape then x would make Alt+anything press Escape,
// so they are dropped, as the kitty protocol path already does.
func parse(b []byte, final bool) (events []keyEvent, consumed int) {
	press := func(k Key) { events = append(events, keyEvent{k, kindPress}) }

	i := 0
	for i < len(b) {
		c := b[i]
		switch {
		case c == 0x1b:
			if i+1 >= len(b) {
				if !final {
					return events, i
				}
				press(KeyEscape)
				i++
				continue
			}

			switch next := b[i+1]; {
			case next == '[':
				j := i + 2
				// Skipping parameters lets modified arrows (ESC [1;5C) read as
				// plain arrows and unknown keys (ESC [15~) drop cleanly.
				for j < len(b) && b[j] >= 0x20 && b[j] <= 0x3f {
					j++
				}
				if j >= len(b) {
					if !final {
						return events, i
					}
					events = append(events, keyEvent{KeyUnknown, kindReset})
					i = len(b)
					continue
				}
				events = appendCSI(events, b[i+2:j], b[j])
				i = j + 1

			case next == 'O':
				// Some terminals send arrows as SS3 (ESC O A) instead of CSI.
				if i+2 >= len(b) {
					if !final {
						return events, i
					}
					i += 2 // Alt+O
					continue
				}
				switch f := b[i+2]; {
				case f >= 'A' && f <= 'D':
					press(arrow(f))
					i += 3
				case f >= 0x40 && f <= 0x7e:
					i += 3 // F1 to F4, Home, End, keypad: not game keys
				default:
					i += 2 // Alt+O; f is read on its own
				}

			case next == 0x1b:
				press(KeyEscape) // the first of two; the second is read next
				i++

			default: // Alt chord
				size := 1
				if next >= utf8.RuneSelf {
					if !final && !utf8.FullRune(b[i+1:]) {
						return events, i
					}
					_, size = utf8.DecodeRune(b[i+1:])
				}
				i += 1 + size
			}
		case c == 0x03:
			press(KeyCtrlC)
			i++
		case c == 0x1a:
			press(keyCtrlZ)
			i++
		case c == '\r':
			press(KeyEnter)
			i++
		case c == 0x7f || c == 0x08:
			press(KeyBackspace)
			i++
		case c == '\t':
			press(KeyTab)
			i++
		case c < 0x20:
			i++
		default:
			if !final && !utf8.FullRune(b[i:]) {
				return events, i
			}
			r, size := utf8.DecodeRune(b[i:])
			if !(r == utf8.RuneError && size == 1) {
				press(Key(r))
			}
			i += size
		}
	}
	return events, i
}

// keyParser decodes input that arrives in pieces. Terminals send a key as a short
// sequence of bytes, and a read can end in the middle of one; the parser keeps the
// unfinished tail and joins it to the next piece.
//
// A tail that is a lone ESC cannot be told from the Escape key until a moment has
// passed without more bytes: the caller checks Pending after each Feed, waits, and
// calls Flush if nothing came.
type keyParser struct {
	pending []byte
}

// Feed decodes the next piece of input after whatever was left over from the last.
func (p *keyParser) Feed(b []byte) []keyEvent {
	data := b
	if len(p.pending) > 0 {
		data = append(p.pending, b...)
	}

	events, n := parse(data, false)
	rest := data[n:]
	if len(rest) > maxPendingInput {
		// No key sequence is this long, so this is garbage that would only grow.
		events = append(events, keyEvent{KeyUnknown, kindReset})
		rest = nil
	}
	p.pending = append(p.pending[:0], rest...)
	return events
}

// Pending reports whether an unfinished sequence is being held.
func (p *keyParser) Pending() bool {
	return len(p.pending) > 0
}

// Flush gives up waiting: it decodes the held bytes as the end of the input.
func (p *keyParser) Flush() []keyEvent {
	events, _ := parse(p.pending, true)
	p.pending = p.pending[:0]
	return events
}

// appendCSI decodes a CSI sequence (parameter bytes plus final byte) and appends
// its key event, if it is one we understand. Anything else is dropped.
func appendCSI(events []keyEvent, params []byte, final byte) []keyEvent {
	// '<' to '?' as the first parameter byte marks a private sequence, such as a
	// protocol reply, not a key.
	if len(params) > 0 && params[0] >= '<' && params[0] <= '?' {
		return events
	}

	// Focus reports are a bare CSI I or CSI O.
	if len(params) == 0 {
		switch final {
		case 'I':
			return append(events, keyEvent{KeyUnknown, kindFocusGained})
		case 'O':
			return append(events, keyEvent{KeyUnknown, kindFocusLost})
		}
	}

	fields := csiParams(params)

	switch {
	case final >= 'A' && final <= 'D':
		// Arrows ignore modifiers so ctrl+right still turns right.
		return append(events, keyEvent{arrow(final), eventKind(fields)})
	case final == 'u':
		return appendKitty(events, fields)
	}
	return events
}

// appendKitty decodes a kitty keyboard protocol key event: CSI code ; mods:event u.
func appendKitty(events []keyEvent, fields [][]int) []keyEvent {
	code := param(fields, 0, 0)
	// The protocol sends modifiers as 1 plus their bits; absent means none.
	mods := max(param(fields, 1, 0), 1) - 1

	var key Key
	switch {
	case mods&shortcutMods == modCtrl && code == 'c':
		key = KeyCtrlC
	case mods&shortcutMods == modCtrl && code == 'z':
		key = keyCtrlZ
	case mods&shortcutMods != 0:
		return events // a shortcut like ctrl+w is not a game key
	case code == 27:
		key = KeyEscape
	case code == 13:
		key = KeyEnter
	case code == 9:
		key = KeyTab
	case code == 127:
		key = KeyBackspace
	case code >= 32 && code < privateUseStart && utf8.ValidRune(rune(code)):
		key = Key(code)
	default:
		return events
	}
	return append(events, keyEvent{key, eventKind(fields)})
}

// eventKind reads the event type from the second sub-value of the modifiers field.
func eventKind(fields [][]int) keyKind {
	switch param(fields, 1, 1) {
	case 2:
		return kindRepeat
	case 3:
		return kindRelease
	default:
		return kindPress
	}
}

// csiParams splits CSI parameter bytes such as "97;5:3" into fields of sub-values.
// Empty or unreadable values come back as 0, which every caller treats as "not given".
func csiParams(p []byte) [][]int {
	var fields [][]int
	for _, f := range strings.Split(string(p), ";") {
		var subs []int
		for _, s := range strings.Split(f, ":") {
			n, err := strconv.Atoi(s)
			if err != nil || n < 0 {
				n = 0
			}
			subs = append(subs, n)
		}
		fields = append(fields, subs)
	}
	return fields
}

// param returns sub-value s of field f, or 0 if it is absent.
func param(fields [][]int, f, s int) int {
	if f >= len(fields) || s >= len(fields[f]) {
		return 0
	}
	return fields[f][s]
}
