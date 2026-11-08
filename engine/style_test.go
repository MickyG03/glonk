package engine

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
)

func TestAttrBitsAreDistinctAndLeaveRoom(t *testing.T) {
	all := []Attr{Bold, Dim, Underline, Reverse}
	var seen Attr
	for _, a := range all {
		if a == 0 || a&(a-1) != 0 {
			t.Errorf("wanted %d to be a single bit", a)
		}
		if seen&a != 0 {
			t.Errorf("wanted %d not to share a bit with another attribute", a)
		}
		seen |= a
	}
	if seen != 0b1111 {
		t.Errorf("wanted the four attributes to take bits 0 to 3, got %b", seen)
	}
}

func TestAttrHas(t *testing.T) {
	both := Bold | Underline
	tests := []struct {
		set, want Attr
		has       bool
	}{
		{both, Bold, true},
		{both, Underline, true},
		{both, Bold | Underline, true},
		{both, Dim, false},
		{both, Bold | Dim, false},
		{0, Bold, false},
		{0, 0, true},
		{both, 0, true},
	}
	for _, tt := range tests {
		if got := tt.set.Has(tt.want); got != tt.has {
			t.Errorf("wanted %b.Has(%b) = %v, got %v", tt.set, tt.want, tt.has, got)
		}
	}
}

func styleBytes(from, to style) string {
	sb := newSequenceBuilder()
	sb.SetStyle(from, to)
	return sb.String()
}

func TestSetStyleBytes(t *testing.T) {
	red, blue := Palette(1), Palette(4)
	tests := []struct {
		name     string
		from, to style
		want     string
	}{
		// What the renderer wrote before styles existed.
		{"nothing changes", style{fg: red}, style{fg: red}, ""},
		{"default to default", style{}, style{}, ""},
		{"foreground from default", style{}, style{fg: Palette(208)}, "\x1b[38;5;208m"},
		{"background from default", style{}, style{bg: blue}, "\x1b[48;5;4m"},
		{"both from default", style{}, style{fg: red, bg: blue}, "\x1b[38;5;1;48;5;4m"},
		{"foreground changes", style{fg: red}, style{fg: blue}, "\x1b[38;5;4m"},
		{"foreground back to default", style{fg: red, bg: blue}, style{bg: blue}, "\x1b[39m"},
		{"background back to default", style{fg: red, bg: blue}, style{fg: red}, "\x1b[49m"},
		{"both back to default is one reset", style{fg: red, bg: blue}, style{}, "\x1b[0m"},
		{"one color back to default is one reset", style{fg: red}, style{}, "\x1b[0m"},
		{"black is a real color, not the default", style{}, style{fg: Palette(0)}, "\x1b[38;5;0m"},

		// Attributes.
		{"bold on", style{}, style{attrs: Bold}, "\x1b[1m"},
		{"dim on", style{}, style{attrs: Dim}, "\x1b[2m"},
		{"underline on", style{}, style{attrs: Underline}, "\x1b[4m"},
		{"reverse on", style{}, style{attrs: Reverse}, "\x1b[7m"},
		{"two on, in a fixed order", style{}, style{attrs: Reverse | Bold}, "\x1b[1;7m"},
		{"one more on", style{attrs: Bold}, style{attrs: Bold | Underline}, "\x1b[4m"},
		{"attribute and color together", style{}, style{fg: red, attrs: Bold}, "\x1b[1;38;5;1m"},
		{"color changes, attribute stays", style{fg: red, attrs: Bold}, style{fg: blue, attrs: Bold}, "\x1b[38;5;4m"},
		{"all attributes off is one reset", style{attrs: Bold | Underline}, style{}, "\x1b[0m"},
		{"one attribute off starts over", style{attrs: Bold | Underline}, style{attrs: Bold}, "\x1b[0;1m"},
		{"attribute off keeps the colors", style{fg: red, attrs: Bold}, style{fg: red}, "\x1b[0;38;5;1m"},
		{"attribute off with other colors", style{fg: red, bg: blue, attrs: Dim}, style{fg: red, bg: blue}, "\x1b[0;38;5;1;48;5;4m"},
		{"attribute swapped", style{attrs: Bold}, style{attrs: Underline}, "\x1b[0;4m"},

		// RGB is written as the nearest palette color until true color is supported.
		{"rgb foreground", style{}, style{fg: RGB(255, 135, 0)}, "\x1b[38;5;208m"},
		{"rgb background", style{}, style{bg: RGB(0, 0, 0)}, "\x1b[48;5;16m"},
		{"rgb to the same palette color", style{fg: Palette(208)}, style{fg: RGB(255, 135, 0)}, "\x1b[38;5;208m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := styleBytes(tt.from, tt.to); got != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, got)
			}
		})
	}
}

// applySGR plays a string of SGR sequences on a style, the way a terminal would, and
// fails on any code it does not know. It understands the codes this package could
// plausibly write, not just the ones it does.
func applySGR(t testing.TB, s style, seq string) style {
	t.Helper()
	for seq != "" {
		if !strings.HasPrefix(seq, "\x1b[") {
			t.Fatalf("wanted an escape sequence, got %q", seq)
		}
		end := strings.IndexByte(seq, 'm')
		if end < 0 {
			t.Fatalf("wanted the sequence to end in m, got %q", seq)
		}
		params := strings.Split(seq[2:end], ";")
		seq = seq[end+1:]

		for i := 0; i < len(params); i++ {
			code, err := strconv.Atoi(params[i])
			if err != nil {
				t.Fatalf("wanted a number, got %q", params[i])
			}
			switch code {
			case 0:
				s = style{}
			case 1:
				s.attrs |= Bold
			case 2:
				s.attrs |= Dim
			case 4:
				s.attrs |= Underline
			case 7:
				s.attrs |= Reverse
			case 22:
				s.attrs &^= Bold | Dim
			case 24:
				s.attrs &^= Underline
			case 27:
				s.attrs &^= Reverse
			case 39:
				s.fg = Color{}
			case 49:
				s.bg = Color{}
			case 38, 48:
				if i+2 >= len(params) || params[i+1] != "5" {
					t.Fatalf("wanted 38;5;n or 48;5;n, got %v", params)
				}
				n, err := strconv.Atoi(params[i+2])
				if err != nil || n < 0 || n > 255 {
					t.Fatalf("wanted a palette index, got %q", params[i+2])
				}
				if code == 38 {
					s.fg = Palette(uint8(n))
				} else {
					s.bg = Palette(uint8(n))
				}
				i += 2
			default:
				t.Fatalf("wanted a known SGR code, got %d in %q", code, params)
			}
		}
	}
	return s
}

// asWritten is s as the terminal ends up holding it: an RGB color is written as the
// nearest palette color.
func asWritten(s style) style {
	for _, c := range []*Color{&s.fg, &s.bg} {
		if n, ok := c.paletteIndex(); ok {
			*c = Palette(n)
		}
	}
	return s
}

func randomColor(rng *rand.Rand) Color {
	switch rng.IntN(4) {
	case 0:
		return Color{}
	case 1:
		return Palette(uint8(rng.IntN(256)))
	case 2:
		return Palette(uint8(rng.IntN(3))) // a few repeats, so equal colors turn up often
	}
	return RGB(uint8(rng.IntN(256)), uint8(rng.IntN(256)), uint8(rng.IntN(256)))
}

func randomStyle(rng *rand.Rand) style {
	return style{fg: randomColor(rng), bg: randomColor(rng), attrs: Attr(rng.IntN(16))}
}

// Whatever the two styles, playing what SetStyle wrote on the first leaves the second.
func TestSetStyleLeavesTheTarget(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 5))
	for range 50000 {
		from, to := randomStyle(rng), randomStyle(rng)
		from, want := asWritten(from), asWritten(to)

		if got := applySGR(t, from, styleBytes(from, to)); got != want {
			t.Fatalf("wanted %+v to become %+v, got %+v from %q", from, want, got, styleBytes(from, to))
		}
	}
}

func TestSetStyleWritesNothingWhenNothingChanges(t *testing.T) {
	rng := rand.New(rand.NewPCG(9, 9))
	for range 2000 {
		s := randomStyle(rng)
		if got := styleBytes(s, s); got != "" {
			t.Fatalf("wanted nothing for the same style %+v, got %q", s, got)
		}
	}
}

// The reset is the shortest way to the plain style, and nothing else may use it
// when only colors or added attributes change.
func TestSetStyleOnlyResetsWhenItMust(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for range 20000 {
		from, to := randomStyle(rng), randomStyle(rng)
		got := styleBytes(from, to)
		mustReset := to == (style{}) || from.attrs&^to.attrs != 0
		// A reset is the code 0 first in the sequence, alone or with others.
		resets := got == "\x1b[0m" || strings.HasPrefix(got, "\x1b[0;")
		if mustReset != resets && from != to {
			t.Fatalf("from %+v to %+v: wanted reset=%v, got %q", from, to, mustReset, got)
		}
	}
}

func FuzzSetStyle(f *testing.F) {
	f.Add(uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0))
	f.Add(uint8(1), uint8(5), uint8(0), uint8(0), uint8(0xf), uint8(2), uint8(10), uint8(20), uint8(30), uint8(0), uint8(0), uint8(1))
	f.Add(uint8(2), uint8(255), uint8(128), uint8(0), uint8(1), uint8(1), uint8(4), uint8(0), uint8(0), uint8(1), uint8(0), uint8(3))

	color := func(kind, a, b, c uint8) Color {
		switch kind % 3 {
		case 1:
			return Palette(a)
		case 2:
			return RGB(a, b, c)
		}
		return Color{}
	}

	f.Fuzz(func(t *testing.T, k1, a1, b1, c1, at1, k2, a2, b2, c2, k3, at2, k4 uint8) {
		from := style{fg: color(k1, a1, b1, c1), bg: color(k2, a2, b2, c2), attrs: Attr(at1 & 15)}
		to := style{fg: color(k3, a1, b2, c1), bg: color(k4, a2, b1, c2), attrs: Attr(at2 & 15)}
		from, want := asWritten(from), asWritten(to)

		if got := applySGR(t, from, styleBytes(from, to)); got != want {
			t.Fatalf("wanted %+v to become %+v, got %+v", from, want, got)
		}
	})
}
