package engine

import (
	"bytes"
	"encoding/binary"
	"math"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// seedInputs are real terminal inputs, legacy and kitty, to start the fuzzer from.
var seedInputs = []string{
	"", "w", "wa", "\x1b", "\x1b[", "\x1b[A", "\x1b[D", "\x1bOB", "\x1bO", "\x1b[1;5C",
	"\x1b[15~x", "\x1bx", "\x03\r\x7f\t", "\x08", "é", "\xff", "\x01",
	"\x1b[119u", "\x1b[119;1:2u", "\x1b[119;1:3u", "\x1b[119;2u", "\x1b[27u", "\x1b[13;1:3u",
	"\x1b[99;5u", "\x1b[99;69u", "\x1a", "\x1b[122;5u", "\x1b[122;5:3u", "\x1b[119;5u", "\x1b[1;1:3A", "\x1b[1;2:2D", "\x1b[57441;1u",
	"\x1b[?0u", "\x1b[?11u\x1b[?62;4c", "\x1b[I", "\x1b[O", "\x1b[119u\x1b[O\x1b[119;1:3u",
	"\x1b[;5u", "\x1b[1;u", "\x1b[99999999999999999999u", "\x1b[-5u", "\x1b[119;:3u",
}

func FuzzParseEvents(f *testing.F) {
	for _, s := range seedInputs {
		f.Add([]byte(s))
	}

	special := map[Key]bool{
		KeyUnknown: true, KeyUp: true, KeyDown: true, KeyLeft: true, KeyRight: true,
		KeyEscape: true, KeyEnter: true, KeyBackspace: true, KeyTab: true, KeyCtrlC: true, keyCtrlZ: true,
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		events := parseEvents(b)

		if len(events) > len(b) {
			t.Fatalf("wanted at most one event per input byte, got %d events for %d bytes", len(events), len(b))
		}

		for _, e := range events {
			switch e.kind {
			case kindPress, kindRepeat, kindRelease:
				if special[e.key] {
					if e.key == KeyUnknown {
						// Only an arrow-shaped sequence can yield an unknown key, and arrow() never does.
						t.Fatalf("wanted no KeyUnknown on a key event, got %+v from %q", e, b)
					}
				} else if e.key < 32 && e.key != 0 || !utf8.ValidRune(rune(e.key)) {
					t.Fatalf("wanted a printable rune or a special key, got %+v from %q", e, b)
				}
			case kindFocusGained, kindFocusLost, kindReset:
				if e.key != KeyUnknown {
					t.Fatalf("wanted focus and reset events to carry no key, got %+v from %q", e, b)
				}
			default:
				t.Fatalf("wanted a known event kind, got %+v from %q", e, b)
			}

			if e.kind == kindRepeat || e.kind == kindRelease {
				// Plain bytes only ever press; repeats and releases need a CSI sequence.
				if !strings.Contains(string(b), "\x1b[") {
					t.Fatalf("wanted repeat/release only from a CSI sequence, got %+v from %q", e, b)
				}
			}
		}

		// Parsing must not depend on anything but the bytes.
		again := parseEvents(b)
		if len(again) != len(events) {
			t.Fatalf("wanted parsing to be deterministic, got %d then %d events", len(events), len(again))
		}
		for i := range events {
			if events[i] != again[i] {
				t.Fatalf("wanted parsing to be deterministic, got %+v then %+v", events[i], again[i])
			}
		}
	})
}

func FuzzParseProbeReply(f *testing.F) {
	for _, s := range seedInputs {
		f.Add([]byte(s))
	}
	f.Add([]byte("\x1b[?0u\x1b[?62;4c"))
	f.Add([]byte("\x1b[?62;4c"))
	f.Add([]byte("\x1b[?"))

	f.Fuzz(func(t *testing.T, b []byte) {
		kitty, done := parseProbeReply(b)

		if kitty && !strings.Contains(string(b), "\x1b[?") {
			t.Fatalf("wanted a protocol reply only where ESC [ ? appears, got kitty for %q", b)
		}
		if kitty && !strings.Contains(string(b), "u") {
			t.Fatalf("wanted a protocol reply to end in u, got kitty for %q", b)
		}
		if done && !strings.Contains(string(b), "c") {
			t.Fatalf("wanted an attributes reply to end in c, got done for %q", b)
		}

		// Whatever came before, a complete pair of replies at the end must be seen.
		k2, d2 := parseProbeReply(append(append([]byte{}, b...), "\x1b[?0u\x1b[?62;4c"...))
		if !d2 && !k2 && !done && !kitty {
			// An unterminated sequence in b can swallow the appended replies; that is allowed.
			// What is not allowed is losing a reply that was already complete.
			return
		}
		if kitty && !k2 {
			t.Fatalf("wanted appending data to keep a reply already found, lost kitty for %q", b)
		}
		if done && !d2 {
			t.Fatalf("wanted appending data to keep a reply already found, lost done for %q", b)
		}
	})
}

func FuzzCSIParams(f *testing.F) {
	for _, s := range []string{"", "1", "1;5", "119;1:3", ";", "::", ";;;", "99999999999999999999", "-5", "1;a;3", "1:2:3:4"} {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, p []byte) {
		fields := csiParams(p)

		if want := strings.Count(string(p), ";") + 1; len(fields) != want {
			t.Fatalf("wanted %d fields for %q, got %d", want, p, len(fields))
		}
		for i, field := range fields {
			if len(field) == 0 {
				t.Fatalf("wanted every field to have a value, field %d of %q is empty", i, p)
			}
			for _, v := range field {
				if v < 0 {
					t.Fatalf("wanted no negative values, got %d in %q", v, p)
				}
			}
		}

		// param must be safe for any index, in or out of range.
		for f := -1; f < len(fields)+2; f++ {
			for s := -1; s < 4; s++ {
				if f >= 0 && s >= 0 {
					_ = param(fields, f, s)
				}
			}
		}
	})
}

func finite(vs ...float64) bool {
	for _, v := range vs {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}

func FuzzFillTriangle(f *testing.F) {
	nan, inf := math.NaN(), math.Inf(1)
	f.Add(0.0, 0.0, 1.0, 4.0, 0.0, 1.0, 0.0, 4.0, 1.0)
	f.Add(-100.0, -100.0, 1.0, 300.0, -100.0, 2.0, -100.0, 300.0, 3.0)
	f.Add(-1e12, -1e12, 1.0, 3e12, -1e12, 1.0, -1e12, 3e12, 1.0)
	f.Add(0.0, 0.0, 1.0, nan, 0.0, 1.0, 0.0, 4.0, 1.0)
	f.Add(0.0, 0.0, 1.0, inf, 0.0, 1.0, 0.0, 4.0, 1.0)
	f.Add(0.0, 0.0, 0.0, 4.0, 0.0, 1.0, 0.0, 4.0, 1.0)
	f.Add(0.0, 0.0, 1.0, 2.0, 2.0, 1.0, 4.0, 4.0, 1.0)
	f.Add(1e300, 1e300, 1e-300, -1e300, 1e300, 1e300, 0.0, -1e300, 1.0)

	f.Fuzz(func(t *testing.T, x0, y0, z0, x1, y1, z1, x2, y2, z2 float64) {
		const w, h = 7, 5
		mark := Cell{Character: '#'}

		fb, err := NewFrameBuffer(w, h)
		if err != nil {
			t.Fatalf("unable to initialize frame buffer: %v", err)
		}
		fb.FillTriangle(x0, y0, z0, x1, y1, z1, x2, y2, z2, mark)

		_, c := fb.Canvas().Width(), fb.Canvas()
		c.FillTriangle(x0, y0, z0, x1, y1, z1, x2, y2, z2, 7)

		// Bad depths must draw nothing at all.
		if !(z0 > 0 && z1 > 0 && z2 > 0) {
			for y := range h {
				for x := range w {
					if fb.At(x, y) != BlankCell {
						t.Fatalf("wanted a bad depth to draw nothing, got a cell at (%d,%d)", x, y)
					}
				}
			}
			for y := range c.Height() {
				for x := range w {
					if c.At(x, y) != 0 {
						t.Fatalf("wanted a bad depth to draw nothing on the canvas, got a pixel at (%d,%d)", x, y)
					}
				}
			}
		}

		// With sane inputs, only cells inside the triangle's bounding box may be drawn.
		if finite(x0, y0, x1, y1, x2, y2) && z0 > 0 && z1 > 0 && z2 > 0 {
			minX, maxX := math.Min(x0, math.Min(x1, x2)), math.Max(x0, math.Max(x1, x2))
			minY, maxY := math.Min(y0, math.Min(y1, y2)), math.Max(y0, math.Max(y1, y2))
			for y := range h {
				for x := range w {
					if fb.At(x, y) == mark && (float64(x) > maxX || float64(x+1) < minX || float64(y) > maxY || float64(y+1) < minY) {
						t.Fatalf("wanted drawn cells inside the bounding box, got (%d,%d) outside x[%v,%v] y[%v,%v]", x, y, minX, maxX, minY, maxY)
					}
				}
			}
		}
	})
}

func FuzzShade(f *testing.F) {
	f.Add(uint8(208), 0.5)
	f.Add(uint8(0), 0.0)
	f.Add(uint8(255), 1.0)
	f.Add(uint8(16), math.NaN())
	f.Add(uint8(231), math.Inf(1))
	f.Add(uint8(1), math.Inf(-1))
	f.Add(uint8(9), -1e300)

	f.Fuzz(func(t *testing.T, color uint8, intensity float64) {
		got := Shade(color, intensity)

		if got < 16 {
			t.Fatalf("wanted a color in 16..255, got %d for color %d at %v", got, color, intensity)
		}
		if again := Shade(color, intensity); again != got {
			t.Fatalf("wanted Shade to be deterministic, got %d then %d", got, again)
		}
	})
}

func FuzzKeyParser(f *testing.F) {
	for _, s := range seedInputs {
		f.Add([]byte(s), uint8(1), uint8(4))
		f.Add([]byte(s), uint8(0), uint8(255))
	}
	f.Add([]byte("\x1b[119;1:3u\x1b[101u\xe2\x82\xac\x1bOA\x1bx"), uint8(7), uint8(9))
	f.Add(append([]byte("\x1b["), bytes.Repeat([]byte("1"), 300)...), uint8(100), uint8(200))

	f.Fuzz(func(t *testing.T, b []byte, cut1, cut2 uint8) {
		n := len(b)
		a, c := 0, 0
		if n > 0 {
			a, c = int(cut1)%(n+1), int(cut2)%(n+1)
		}
		if a > c {
			a, c = c, a
		}

		feed := func(pieces ...[]byte) []keyEvent {
			var p keyParser
			var events []keyEvent
			for _, piece := range pieces {
				events = append(events, p.Feed(piece)...)
				if len(p.pending) > maxPendingInput {
					t.Fatalf("wanted at most %d bytes held, got %d", maxPendingInput, len(p.pending))
				}
			}
			events = append(events, p.Flush()...)
			if p.Pending() {
				t.Fatalf("wanted nothing held after a flush")
			}
			return events
		}

		split := feed(b[:a], b[a:c], b[c:])
		bytewise := make([][]byte, n)
		for i := range b {
			bytewise[i] = b[i : i+1]
		}
		single := feed(bytewise...)

		// A sequence longer than the hold limit is discarded by design, so the split
		// readings only have to match the whole when nothing could exceed it.
		if n <= maxPendingInput {
			whole := parseEvents(b)
			if !slices.Equal(split, whole) {
				t.Fatalf("wanted %q cut at %d and %d to give %v, got %v", b, a, c, whole, split)
			}
			if !slices.Equal(single, whole) {
				t.Fatalf("wanted %q one byte at a time to give %v, got %v", b, whole, single)
			}
		}
	})
}

// ownSequence matches the escape sequences present writes itself: cursor moves and
// colours.
var ownSequence = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func FuzzPresent(f *testing.F) {
	seed := func(runes ...rune) {
		var b []byte
		for _, r := range runes {
			b = binary.LittleEndian.AppendUint32(b, uint32(r))
		}
		f.Add(b)
	}
	seed('a', 'b', 0, 0x1b, '\n', 0x9b, 0x202e, '中', '😀', 0x301, 0xd800, -1, 0x110000, '─', '▀', '⣿')
	seed(0x1b, '[', '2', 'J', 0x1b, '[', 'H', 0x07, 0x7f, '\t', '\r', 0x08, 0x0e, 0x0f, 0x90, 0x9d)
	seed(0x200d, 0x200b, 0x2060, 0xfeff, 0xfe0f, 0x20dd, 0xe0b0, 0xf8ff, 0xfffe, 0xffff, 0xfdd0, 0x3000, 0xff21, 0x1100, 0x1f600, 0x2fffd)

	f.Fuzz(func(t *testing.T, data []byte) {
		const w, h = 8, 2
		fb, err := NewFrameBuffer(w, h)
		if err != nil {
			t.Fatal(err)
		}
		for i := range w * h {
			r := 'x'
			if 4*i+4 <= len(data) {
				r = rune(int32(binary.LittleEndian.Uint32(data[4*i:])))
			}
			fb.Set(i%w, i/w, Cell{Character: r, Color: uint8(i * 17), Background: uint8(i * 29)})
		}

		var buf bytes.Buffer
		if err := fb.present(&buf); err != nil {
			t.Fatal(err)
		}

		// What is left once glonk's own sequences are removed is the drawn text. It
		// must be one visible, narrow character per cell, with nothing a terminal
		// could act on, however hostile the cells were.
		out := ownSequence.ReplaceAll(buf.Bytes(), nil)
		if !utf8.Valid(out) {
			t.Fatalf("wanted valid UTF-8, got %q", out)
		}
		runes := []rune(string(out))
		if len(runes) != w*h {
			t.Fatalf("wanted %d characters, one per cell, got %d in %q", w*h, len(runes), out)
		}
		for _, r := range runes {
			if unicode.IsControl(r) || unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf) {
				t.Fatalf("wanted no control or zero-width character, got %U in %q", r, out)
			}
			if !unicode.IsGraphic(r) && !unicode.Is(unicode.Co, r) {
				t.Fatalf("wanted a printable character, got %U in %q", r, out)
			}
		}
	})
}

func FuzzDrawText(f *testing.F) {
	for _, s := range []string{
		"", "hello", "\x1b[2J", "a\tb\nc", "e\u0301", "中文", "😀👍🏽", "\u202eevil", "a\x00b", "\xff\xfe", "\u009b31m",
		"👨\u200d👩\u200d👧", "\ufe0f", "x\u0301\u0301\u0301y",
	} {
		f.Add(s, 0)
		f.Add(s, -2)
		f.Add(s, 5)
	}

	f.Fuzz(func(t *testing.T, text string, x int) {
		fb, err := NewFrameBuffer(8, 1)
		if err != nil {
			t.Fatal(err)
		}
		fb.DrawText(x, 0, text, 7)

		// Everything stored is already safe to send: no control, no zero-width, and
		// a character is either what was asked for or the stand-in.
		for col := range 8 {
			c := fb.At(col, 0).Character
			if c == 0 || unicode.IsControl(c) || unicode.In(c, unicode.Mn, unicode.Me, unicode.Cf) {
				t.Fatalf("wanted a printable character at column %d, got %U from %q", col, c, text)
			}
			if c != ' ' && c != '?' && !strings.ContainsRune(text, c) {
				t.Fatalf("wanted column %d to hold a character of %q or '?', got %U", col, text, c)
			}
		}
	})
}
