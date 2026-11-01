package engine

import (
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
)

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

// parseKeys decodes one chunk of raw terminal input into keys.
// An escape sequence split across two reads is misparsed; that is rare
// locally and fixing it needs a stateful parser with timeouts.
func parseKeys(b []byte) []Key {
	var keys []Key
	i := 0
	for i < len(b) {
		c := b[i]
		switch {
		case c == 0x1b:
			if i+1 >= len(b) {
				keys = append(keys, KeyEscape)
				i++
				continue
			}

			next := b[i+1]

			// Some terminals send arrows as SS3 (ESC O A) instead of CSI.
			if next == 'O' {
				if i+2 < len(b) && b[i+2] >= 'A' && b[i+2] <= 'D' {
					keys = append(keys, arrow(b[i+2]))
					i += 3
					continue
				}
			}

			if next == '[' {
				j := i + 2
				// Skipping parameters lets modified arrows (ESC [1;5C) read as
				// plain arrows and unknown keys (ESC [15~) drop cleanly.
				for j < len(b) && b[j] >= 0x20 && b[j] <= 0x3f {
					j++
				}
				if j >= len(b) {
					i = len(b)
					continue
				}
				if b[j] >= 'A' && b[j] <= 'D' {
					keys = append(keys, arrow(b[j]))
				}
				i = j + 1
				continue
			}

			keys = append(keys, KeyEscape)
			i++
		case c == 0x03:
			keys = append(keys, KeyCtrlC)
			i++
		case c == '\r':
			keys = append(keys, KeyEnter)
			i++
		case c == 0x7f || c == 0x08:
			keys = append(keys, KeyBackspace)
			i++
		case c == '\t':
			keys = append(keys, KeyTab)
			i++
		case c < 0x20:
			i++
		default:
			r, size := utf8.DecodeRune(b[i:])
			if !(r == utf8.RuneError && size == 1) {
				keys = append(keys, Key(r))
			}
			i += size
		}
	}
	return keys
}
