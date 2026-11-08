package engine

import (
	"encoding/binary"
	"math/rand/v2"
	"testing"
)

// frameChars is what random test frames are built from: plain text, controls that
// must never reach the terminal, wide and combining characters, box drawing and blocks.
var frameChars = []rune{
	' ', 'a', 'b', 'Z', '~', '#', 0, '\x1b', '\n', '\t', 0x7f, 0x9b, '中', '😀', 0x301, 0x200d,
	'─', '│', '▀', '▄', '⣿', 'é',
}

// randomFrame returns a frame of the given size filled with random cells: any
// character from frameChars, any color (default, palette, RGB) and any attributes.
func randomFrame(t testing.TB, rng *rand.Rand, w, h int) *FrameBuffer {
	t.Helper()
	fb, err := NewFrameBuffer(w, h)
	if err != nil {
		t.Fatal(err)
	}
	for y := range h {
		for x := range w {
			fb.Set(x, y, randomCell(rng))
		}
	}
	return fb
}

func randomCell(rng *rand.Rand) Cell {
	return Cell{
		Character:  frameChars[rng.IntN(len(frameChars))],
		Foreground: randomColor(rng),
		Background: randomColor(rng),
		Attrs:      Attr(rng.IntN(16)),
	}
}

// Whatever a frame holds, playing the bytes present writes onto a terminal must show
// that frame, and only that. This is the check the changes-only presenter is held to
// as well.
func TestPresentReplaysToTheFrame(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	for range 400 {
		w, h := 1+rng.IntN(24), 1+rng.IntN(8)
		fb := randomFrame(t, rng, w, h)
		maxW, maxH := rng.IntN(w+3), rng.IntN(h+3) // the window may be smaller or bigger
		v := newVterm(t, w, h)

		if err := fb.presentClipped(v, maxW, maxH); err != nil {
			t.Fatal(err)
		}

		if d := v.screenDiff(fb, maxW, maxH); d != "" {
			t.Fatalf("%dx%d frame clipped to %dx%d: %s", w, h, maxW, maxH, d)
		}
	}
}

func FuzzPresentReplay(f *testing.F) {
	seed := func(runes ...rune) {
		var b []byte
		for _, r := range runes {
			b = binary.LittleEndian.AppendUint32(b, uint32(r))
		}
		f.Add(b, uint8(3), uint8(2))
	}
	seed('a', 'b', 0, 0x1b, '\n', 0x9b, 0x202e, '中', '😀', 0x301, 0xd800, -1, 0x110000, '─', '▀', '⣿')
	seed('x', 'y', 'z')

	f.Fuzz(func(t *testing.T, data []byte, clipW, clipH uint8) {
		const w, h = 8, 2
		fb, err := NewFrameBuffer(w, h)
		if err != nil {
			t.Fatal(err)
		}
		for i := range w * h {
			r := 'x'
			var fg, bg, attrs byte
			if 4*i+4 <= len(data) {
				r = rune(int32(binary.LittleEndian.Uint32(data[4*i:])))
				fg, bg, attrs = data[4*i], data[(4*i+1)%len(data)], data[(4*i+2)%len(data)]
			}
			fb.Set(i%w, i/w, Cell{Character: r, Foreground: pix(fg), Background: Palette(bg), Attrs: Attr(attrs & 15)})
		}
		maxW, maxH := int(clipW)%(w+3), int(clipH)%(h+3)
		v := newVterm(t, w, h)

		if err := fb.presentClipped(v, maxW, maxH); err != nil {
			t.Fatal(err)
		}

		if d := v.screenDiff(fb, maxW, maxH); d != "" {
			t.Fatalf("clipped to %dx%d: %s", maxW, maxH, d)
		}
	})
}

// countingWriter counts the bytes written and the writes.
type countingWriter struct {
	bytes, writes int
}

func (c *countingWriter) Write(p []byte) (int, error) {
	c.bytes += len(p)
	c.writes++
	return len(p), nil
}
