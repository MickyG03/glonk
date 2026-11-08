package engine

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"testing"
)

func getIndex(x, y, w int) int {
	return y*w + x
}

func assertFrameEquals(t *testing.T, assertion map[int]Cell, fb *FrameBuffer) {
	t.Helper()
	for i, cell := range fb.frame {
		if val, ok := assertion[i]; ok {
			if cell != val {
				t.Fatalf("wanted cell[%d]=%v, got %v", i, val, cell)
			}
		} else if cell != BlankCell {
			t.Fatalf("wanted nfb.frame[%d]=BlankCell, got %v", i, cell)
		}
	}
}

func TestNewFrameBufferRejectsBadSizes(t *testing.T) {
	var tests = []struct {
		w, h  int
		valid bool
	}{
		{0, 5, false},
		{5, 0, false},
		{-1, 5, false},
		{10, 100, true},
		{999, 1, true},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d,%d", tt.w, tt.h)
		t.Run(testname, func(t *testing.T) {
			nfb, err := NewFrameBuffer(tt.w, tt.h)

			if err != nil {
				if tt.valid {
					t.Errorf("wanted err=nil, got %v", err)
				}
				if nfb != nil {
					t.Errorf("wanted nfb=nil, got %v", nfb)
				}
			} else {
				if !tt.valid {
					t.Errorf("wanted err!=nil, got %v", err)
				}
			}
		})
	}
}

func TestNewFrameBufferOutput(t *testing.T) {
	var tests = []struct {
		w, h int
		want int
	}{
		{10, 5, 50},
		{5, 10, 50},
		{100, 100, 10000},
		{99, 10, 990},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d,%d", tt.w, tt.h)
		t.Run(testname, func(t *testing.T) {
			nfb, err := NewFrameBuffer(tt.w, tt.h)
			if err != nil {
				t.Fatalf("wanted err=nil, got %v", err)
			}
			if nfb.width != tt.w {
				t.Fatalf("wanted nfb.width=%v, got %v", tt.w, nfb.width)
			}
			if nfb.height != tt.h {
				t.Fatalf("wanted nfb.height=%v, got %v", tt.h, nfb.height)
			}
			if nfbLength := len(nfb.frame); nfbLength != tt.want {
				t.Errorf("wanted nfb.frame=%v, got %v", tt.want, nfbLength)
			}
			for _, cell := range nfb.frame {
				if cell != BlankCell {
					t.Fatalf("wanted nfb.frame[*]=BlankCell, got %v", cell)
				}
			}
		})
	}
}

func TestSetBounds(t *testing.T) {
	fixedWidth := 40
	fixedHeight := 20

	var tests = []struct {
		x, y     int
		inBounds bool
	}{
		{0, 0, true},
		{fixedWidth - 1, 0, true},
		{0, fixedHeight - 1, true},
		{fixedWidth - 1, fixedHeight - 1, true},

		{fixedWidth, 0, false},
		{0, fixedHeight, false},
		{fixedWidth, fixedHeight, false},
		{-1, 0, false},
		{0, -1, false},
	}

	tempCell := Cell{
		Character: 'X',
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d,%d", tt.x, tt.y)
		t.Run(testname, func(t *testing.T) {
			nfb, err := NewFrameBuffer(fixedWidth, fixedHeight)
			if err != nil {
				t.Fatalf("wanted err=nil, got %v", err)
			}

			nfb.Set(tt.x, tt.y, tempCell)

			idx := getIndex(tt.x, tt.y, nfb.width)
			idxCellMap := map[int]Cell{}

			if tt.inBounds {
				idxCellMap[idx] = tempCell
			}

			assertFrameEquals(t, idxCellMap, nfb)

		})
	}
}

func TestClear(t *testing.T) {
	fixedWidth := 40
	fixedHeight := 20
	nfb, err := NewFrameBuffer(fixedWidth, fixedHeight)
	if err != nil {
		t.Fatalf("wanted err=nil, got %v", err)
	}

	tempCell := Cell{
		Character: 'X',
	}

	tempCellsMap := map[int]Cell{}

	nfb.Set(10, 10, tempCell)
	tempCellsMap[getIndex(10, 10, fixedWidth)] = tempCell

	nfb.Set(0, 0, tempCell)
	tempCellsMap[getIndex(0, 0, fixedWidth)] = tempCell

	nfb.Set(fixedWidth-1, fixedHeight-1, tempCell)
	tempCellsMap[getIndex(fixedWidth-1, fixedHeight-1, fixedWidth)] = tempCell

	assertFrameEquals(t, tempCellsMap, nfb)

	nfb.Clear()

	assertFrameEquals(t, map[int]Cell{}, nfb)
}

func TestAt(t *testing.T) {
	fixedWidth := 40
	fixedHeight := 20
	nfb, err := NewFrameBuffer(fixedWidth, fixedHeight)
	if err != nil {
		t.Fatalf("wanted err=nil, got %v", err)
	}

	tempCell := Cell{
		Character: 'X',
	}

	nfb.Set(10, 10, tempCell)
	nfb.Set(0, 1, tempCell)

	cell := nfb.At(10, 10)
	if cell != tempCell {
		t.Errorf("wanted cell=%v, got %v", tempCell, cell)
	}

	cell = nfb.At(fixedWidth, 0)
	if cell != BlankCell {
		t.Errorf("wanted cell=%v, got %v", BlankCell, cell)
	}

	cell = nfb.At(-1, 0)
	if cell != BlankCell {
		t.Errorf("wanted cell=%v, got %v", BlankCell, cell)
	}
}

func TestPresent(t *testing.T) {
	fb, err := NewFrameBuffer(3, 2)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	fb.Set(0, 0, Cell{Character: 'a'})
	fb.Set(2, 1, Cell{Character: 'b'})

	buf := bytes.Buffer{}
	err = fb.present(&buf)
	if err != nil {
		t.Fatalf("could not present: %v", err)
	}

	want := "\x1b[1;1Ha  \x1b[2;1H  b"

	if buf.String() != want {
		t.Errorf("wanted %q, got %q", want, buf.String())
	}
}

func TestPresentColor(t *testing.T) {
	tests := []struct {
		name string
		rows [][]Cell
		want string
	}{
		{
			"same color across",
			[][]Cell{{{Character: 'a', Foreground: pix(1)}, {Character: 'b', Foreground: pix(1)}, {Character: 'c', Foreground: pix(0)}}},
			"\x1b[1;1H\x1b[38;5;1mab\x1b[0mc",
		},
		{
			"color changes",
			[][]Cell{{{Character: 'a', Foreground: pix(1)}, {Character: 'b', Foreground: pix(2)}}},
			"\x1b[1;1H\x1b[38;5;1ma\x1b[38;5;2mb\x1b[0m",
		},
		{
			"row ends colored",
			[][]Cell{{{Character: 'a', Foreground: pix(2)}}, {{Character: 'b', Foreground: pix(2)}}},
			"\x1b[1;1H\x1b[38;5;2ma\x1b[0m\x1b[2;1H\x1b[38;5;2mb\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := NewFrameBuffer(len(tt.rows[0]), len(tt.rows))
			if err != nil {
				t.Fatalf("unable to initialize frame buffer: %v", err)
			}

			for y := range len(tt.rows) {
				for x := range len(tt.rows[0]) {
					fb.Set(x, y, tt.rows[y][x])
				}
			}

			buf := bytes.Buffer{}
			err = fb.present(&buf)
			if err != nil {
				t.Fatalf("could not present: %v", err)
			}

			if buf.String() != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, buf.String())
			}
		})
	}
}

func TestSetDepth(t *testing.T) {
	near := Cell{Character: 'n'}
	far := Cell{Character: 'f'}

	tests := []struct {
		name      string
		first     float64
		second    float64
		wantWrote bool // whether the second write lands
		want      Cell
	}{
		{"nearer second wins", 5, 2, true, far},
		{"farther second loses", 2, 5, false, near},
		{"tie keeps the first", 3, 3, false, near},
		{"negative depths order too", -1, -2, true, far},
		{"nearer than infinity", 1e300, 1, true, far},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := NewFrameBuffer(2, 2)
			if err != nil {
				t.Fatalf("unable to initialize frame buffer: %v", err)
			}

			if !fb.SetDepth(1, 1, tt.first, near) {
				t.Fatalf("wanted the first write to an empty cell to land")
			}
			if got := fb.SetDepth(1, 1, tt.second, far); got != tt.wantWrote {
				t.Errorf("wanted second write landed=%v, got %v", tt.wantWrote, got)
			}
			if got := fb.At(1, 1); got != tt.want {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

func TestSetDepthFirstWriteBeatsEmptyCell(t *testing.T) {
	fb, err := NewFrameBuffer(1, 1)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	// A fresh cell is infinitely far, so even an enormous depth must land.
	if !fb.SetDepth(0, 0, math.MaxFloat64, Cell{Character: 'x'}) {
		t.Errorf("wanted a write into an empty cell to land, got false")
	}
}

func TestSetDepthBounds(t *testing.T) {
	fb, err := NewFrameBuffer(2, 2)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	for _, p := range [][2]int{{-1, 0}, {0, -1}, {2, 0}, {0, 2}, {2, 2}} {
		if fb.SetDepth(p[0], p[1], 1, Cell{Character: 'x'}) {
			t.Errorf("wanted a write at %v to be ignored, got landed", p)
		}
	}
	assertFrameEquals(t, map[int]Cell{}, fb)
}

func TestSetDepthRejectsNaN(t *testing.T) {
	fb, err := NewFrameBuffer(1, 1)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	if fb.SetDepth(0, 0, math.NaN(), Cell{Character: 'x'}) {
		t.Errorf("wanted a NaN depth to be rejected, got landed")
	}
	if got := fb.At(0, 0); got != BlankCell {
		t.Errorf("wanted BlankCell, got %v", got)
	}
}

func TestClearResetsDepth(t *testing.T) {
	fb, err := NewFrameBuffer(1, 1)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	fb.SetDepth(0, 0, 1, Cell{Character: 'a'})
	fb.Clear()

	if !fb.SetDepth(0, 0, 100, Cell{Character: 'b'}) {
		t.Errorf("wanted a farther write to land after Clear, got rejected")
	}
	if got := fb.At(0, 0); got.Character != 'b' {
		t.Errorf("wanted 'b', got %v", got)
	}
}

// 2D drawing and HUDs use Set, which must neither read nor write depth.
func TestSetIgnoresDepth(t *testing.T) {
	fb, err := NewFrameBuffer(1, 1)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	fb.SetDepth(0, 0, 1, Cell{Character: 'a'})
	fb.Set(0, 0, Cell{Character: 'h'})
	if got := fb.At(0, 0); got.Character != 'h' {
		t.Errorf("wanted Set to overwrite regardless of depth, got %v", got)
	}

	// Set left the depth at 1: a farther surface is still rejected, a nearer one lands.
	if fb.SetDepth(0, 0, 2, Cell{Character: 'b'}) {
		t.Errorf("wanted Set to leave depth at 1, but a farther write landed")
	}
	if !fb.SetDepth(0, 0, 0.5, Cell{Character: 'c'}) {
		t.Errorf("wanted Set to leave depth at 1, but a nearer write was rejected")
	}
}

func TestPresentBackground(t *testing.T) {
	c := func(r rune, fg, bg uint8) Cell { return Cell{Character: r, Foreground: pix(fg), Background: pix(bg)} }

	tests := []struct {
		name string
		rows [][]Cell
		want string
	}{
		{
			"background only",
			[][]Cell{{c('a', 0, 4)}},
			"\x1b[1;1H\x1b[48;5;4ma\x1b[0m",
		},
		{
			"foreground then background",
			[][]Cell{{c('a', 1, 4)}},
			"\x1b[1;1H\x1b[38;5;1m\x1b[48;5;4ma\x1b[0m",
		},
		{
			"same background run is set once",
			[][]Cell{{c('a', 0, 4), c('b', 0, 4)}},
			"\x1b[1;1H\x1b[48;5;4mab\x1b[0m",
		},
		{
			"background changes, foreground kept",
			[][]Cell{{c('a', 1, 4), c('b', 1, 5)}},
			"\x1b[1;1H\x1b[38;5;1m\x1b[48;5;4ma\x1b[48;5;5mb\x1b[0m",
		},
		{
			"foreground changes, background kept",
			[][]Cell{{c('a', 1, 4), c('b', 2, 4)}},
			"\x1b[1;1H\x1b[38;5;1m\x1b[48;5;4ma\x1b[38;5;2mb\x1b[0m",
		},
		{
			"foreground back to default, background kept",
			[][]Cell{{c('a', 1, 4), c('b', 0, 4)}},
			"\x1b[1;1H\x1b[38;5;1m\x1b[48;5;4ma\x1b[39mb\x1b[0m",
		},
		{
			"background back to default, foreground kept",
			[][]Cell{{c('a', 1, 4), c('b', 1, 0)}},
			"\x1b[1;1H\x1b[38;5;1m\x1b[48;5;4ma\x1b[49mb\x1b[0m",
		},
		{
			"both back to default",
			[][]Cell{{c('a', 1, 4), c('b', 0, 0)}},
			"\x1b[1;1H\x1b[38;5;1m\x1b[48;5;4ma\x1b[0mb",
		},
		{
			"colors swap sides",
			[][]Cell{{c('a', 1, 0), c('b', 0, 2)}},
			"\x1b[1;1H\x1b[38;5;1ma\x1b[39m\x1b[48;5;2mb\x1b[0m",
		},
		{
			"style does not leak into the next row",
			[][]Cell{{c('a', 0, 4)}, {c('b', 0, 4)}},
			"\x1b[1;1H\x1b[48;5;4ma\x1b[0m\x1b[2;1H\x1b[48;5;4mb\x1b[0m",
		},
		{
			"all default emits no style",
			[][]Cell{{c('a', 0, 0), c('b', 0, 0)}},
			"\x1b[1;1Hab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := NewFrameBuffer(len(tt.rows[0]), len(tt.rows))
			if err != nil {
				t.Fatalf("unable to initialize frame buffer: %v", err)
			}
			for y, row := range tt.rows {
				for x, cell := range row {
					fb.Set(x, y, cell)
				}
			}

			buf := bytes.Buffer{}
			if err := fb.present(&buf); err != nil {
				t.Fatalf("could not present: %v", err)
			}

			if buf.String() != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, buf.String())
			}
		})
	}
}

// Wide runes and invalid ones must come out as the bytes string(rune) would produce.
func TestPresentRunes(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want string
	}{
		{"ascii", 'a', "a"},
		{"two bytes", 'é', "é"},
		{"three bytes", '─', "─"},
		{"half block", '▀', "▀"},
		{"braille", '⣿', "⣿"},
		{"private use", '\ue0b0', "\ue0b0"},
		{"nul is a space", 0, " "},
		{"escape", 0x1b, "?"},
		{"newline", '\n', "?"},
		{"tab", '\t', "?"},
		{"bell", 0x07, "?"},
		{"delete", 0x7f, "?"},
		{"c1 csi", 0x9b, "?"},
		{"bidi override", 0x202e, "?"},
		{"zero width joiner", 0x200d, "?"},
		{"combining accent", 0x0301, "?"},
		{"wide cjk", '中', "?"},
		{"wide fullwidth", 'Ａ', "?"},
		{"wide emoji", '😀', "?"},
		{"unassigned", 0x0378, "?"},
		{"surrogate", 0xD800, "?"},
		{"negative", -1, "?"},
		{"past unicode", 0x110000, "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := NewFrameBuffer(1, 1)
			if err != nil {
				t.Fatalf("unable to initialize frame buffer: %v", err)
			}
			fb.Set(0, 0, Cell{Character: tt.r})

			buf := bytes.Buffer{}
			if err := fb.present(&buf); err != nil {
				t.Fatalf("could not present: %v", err)
			}

			want := "\x1b[1;1H" + tt.want
			if buf.String() != want {
				t.Errorf("wanted %q, got %q", want, buf.String())
			}
		})
	}
}

// A cell with a colour but no character is a coloured blank, not a dropped column.
func TestPresentColouredCellWithoutCharacterKeepsItsColumn(t *testing.T) {
	fb, err := NewFrameBuffer(3, 1)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	fb.Set(0, 0, Cell{Character: 'a'})
	fb.Set(1, 0, Cell{Background: pix(4)})
	fb.Set(2, 0, Cell{Character: 'b'})

	buf := bytes.Buffer{}
	if err := fb.present(&buf); err != nil {
		t.Fatalf("could not present: %v", err)
	}

	want := "\x1b[1;1Ha\x1b[48;5;4m \x1b[0mb"
	if buf.String() != want {
		t.Errorf("wanted %q, got %q", want, buf.String())
	}
}

func BenchmarkPresent(b *testing.B) {
	for _, size := range []struct{ w, h int }{{80, 24}, {200, 60}} {
		b.Run(fmt.Sprintf("%dx%d", size.w, size.h), func(b *testing.B) {
			fb, err := NewFrameBuffer(size.w, size.h)
			if err != nil {
				b.Fatalf("unable to initialize frame buffer: %v", err)
			}
			for y := range size.h {
				for x := range size.w {
					c := Cell{Character: rune('a' + (x+y)%26)}
					if x%7 == 0 {
						c.Foreground = pix(uint8(1 + x%200))
					}
					fb.Set(x, y, c)
				}
			}

			b.ReportAllocs()
			for b.Loop() {
				if err := fb.present(io.Discard); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkPresentBlocks is a frame of half blocks, as the pixel canvas makes: every
// cell is outside printable ASCII.
func BenchmarkPresentBlocks(b *testing.B) {
	fb, err := NewFrameBuffer(200, 60)
	if err != nil {
		b.Fatalf("unable to initialize frame buffer: %v", err)
	}
	for y := range 60 {
		for x := range 200 {
			c := Cell{Character: '▀', Foreground: pix(uint8(1 + (x+y)%200)), Background: pix(uint8(1 + (x*y)%200))}
			if (x+y)%3 == 0 {
				c.Character = '▄'
			}
			fb.Set(x, y, c)
		}
	}

	b.ReportAllocs()
	for b.Loop() {
		if err := fb.present(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}

// lettered is a 4x3 frame holding a..l, row by row.
func lettered(t *testing.T) *FrameBuffer {
	t.Helper()
	fb, err := NewFrameBuffer(4, 3)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	for i := range 12 {
		fb.Set(i%4, i/4, Cell{Character: rune('a' + i)})
	}
	return fb
}

func TestPresentClipped(t *testing.T) {
	tests := []struct {
		name       string
		maxW, maxH int
		want       string
	}{
		{"exactly the frame", 4, 3, "\x1b[1;1Habcd\x1b[2;1Hefgh\x1b[3;1Hijkl"},
		{"a terminal larger than the frame shows all of it", 100, 50, "\x1b[1;1Habcd\x1b[2;1Hefgh\x1b[3;1Hijkl"},
		{"narrower", 2, 3, "\x1b[1;1Hab\x1b[2;1Hef\x1b[3;1Hij"},
		{"shorter", 4, 2, "\x1b[1;1Habcd\x1b[2;1Hefgh"},
		{"narrower and shorter", 3, 1, "\x1b[1;1Habc"},
		{"one cell", 1, 1, "\x1b[1;1Ha"},
		{"no width", 0, 3, "\x1b[1;1H\x1b[2;1H\x1b[3;1H"},
		{"no height", 4, 0, ""},
		{"negative size", -3, -1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := lettered(t).presentClipped(&buf, tt.maxW, tt.maxH); err != nil {
				t.Fatalf("could not present: %v", err)
			}
			if buf.String() != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, buf.String())
			}
		})
	}
}

func TestPresentIsPresentClippedToTheWholeFrame(t *testing.T) {
	fb := lettered(t)
	fb.Set(1, 1, Cell{Character: 'x', Foreground: pix(3), Background: pix(4)})

	var a, b bytes.Buffer
	if err := fb.present(&a); err != nil {
		t.Fatal(err)
	}
	if err := fb.presentClipped(&b, fb.Width(), fb.Height()); err != nil {
		t.Fatal(err)
	}
	if a.String() != b.String() {
		t.Errorf("wanted the same bytes, got %q and %q", a.String(), b.String())
	}
}

// A colour set before the cut must be reset at the end of the row, as for a full row.
func TestPresentClippedResetsColoursAtTheCut(t *testing.T) {
	fb := lettered(t)
	fb.Set(0, 0, Cell{Character: 'a', Foreground: pix(5)})
	fb.Set(1, 0, Cell{Character: 'b', Foreground: pix(5)})

	var buf bytes.Buffer
	if err := fb.presentClipped(&buf, 2, 1); err != nil {
		t.Fatal(err)
	}
	if want := "\x1b[1;1H\x1b[38;5;5mab\x1b[0m"; buf.String() != want {
		t.Errorf("wanted %q, got %q", want, buf.String())
	}
}

// Cells past the cut must not shift the ones before the next row.
func TestPresentClippedReadsRowsAtTheFullWidth(t *testing.T) {
	var buf bytes.Buffer
	if err := lettered(t).presentClipped(&buf, 1, 3); err != nil {
		t.Fatal(err)
	}
	if want := "\x1b[1;1Ha\x1b[2;1He\x1b[3;1Hi"; buf.String() != want {
		t.Errorf("wanted %q, got %q", want, buf.String())
	}
}

func presented(t *testing.T, w, h int, cells map[[2]int]Cell) string {
	t.Helper()
	fb, err := NewFrameBuffer(w, h)
	if err != nil {
		t.Fatal(err)
	}
	for p, c := range cells {
		fb.Set(p[0], p[1], c)
	}
	var buf bytes.Buffer
	if err := fb.present(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestPresentWritesAttributes(t *testing.T) {
	tests := []struct {
		name  string
		cells map[[2]int]Cell
		want  string
	}{
		{"bold run then plain", map[[2]int]Cell{{0, 0}: {Character: 'a', Attrs: Bold}, {1, 0}: {Character: 'b', Attrs: Bold}}, "\x1b[1;1H\x1b[1mab\x1b[0m "},
		{"bold in the middle", map[[2]int]Cell{{1, 0}: {Character: 'b', Attrs: Bold}}, "\x1b[1;1H \x1b[1mb\x1b[0m "},
		{"underline", map[[2]int]Cell{{0, 0}: {Character: 'a', Attrs: Underline}}, "\x1b[1;1H\x1b[4ma\x1b[0m  "},
		{"dim", map[[2]int]Cell{{0, 0}: {Character: 'a', Attrs: Dim}}, "\x1b[1;1H\x1b[2ma\x1b[0m  "},
		{"reverse", map[[2]int]Cell{{0, 0}: {Character: 'a', Attrs: Reverse}}, "\x1b[1;1H\x1b[7ma\x1b[0m  "},
		{"bold and underline", map[[2]int]Cell{{0, 0}: {Character: 'a', Attrs: Bold | Underline}}, "\x1b[1;1H\x1b[1m\x1b[4ma\x1b[0m  "},
		{"attribute with color", map[[2]int]Cell{{0, 0}: {Character: 'a', Foreground: Palette(1), Attrs: Bold}}, "\x1b[1;1H\x1b[1m\x1b[38;5;1ma\x1b[0m  "},
		{"one attribute off starts over", map[[2]int]Cell{{0, 0}: {Character: 'a', Attrs: Bold | Underline}, {1, 0}: {Character: 'b', Attrs: Bold}}, "\x1b[1;1H\x1b[1m\x1b[4ma\x1b[0m\x1b[1mb\x1b[0m "},
		{"attributes reset at the end of the row", map[[2]int]Cell{{2, 0}: {Character: 'c', Attrs: Bold}, {0, 1}: {Character: 'x'}}, "\x1b[1;1H  \x1b[1mc\x1b[0m\x1b[2;1Hx  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := 1
			if len(tt.cells) > 0 {
				for p := range tt.cells {
					rows = max(rows, p[1]+1)
				}
			}
			// Blank cells in the table are spaces; fill the rest of the row to 3 wide.
			if got := presented(t, 3, rows, tt.cells); got != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, got)
			}
		})
	}
}

func TestPresentWritesRGBAsTheNearestPaletteColor(t *testing.T) {
	got := presented(t, 2, 1, map[[2]int]Cell{{0, 0}: {Character: 'a', Foreground: RGB(255, 135, 0), Background: RGB(0, 0, 0)}})
	if want := "\x1b[1;1H\x1b[38;5;208m\x1b[48;5;16ma\x1b[0m "; got != want {
		t.Errorf("wanted %q, got %q", want, got)
	}
}

func TestPresentBlackIsAColorAndNotTheDefault(t *testing.T) {
	got := presented(t, 2, 1, map[[2]int]Cell{{0, 0}: {Character: 'a', Foreground: Palette(0), Background: Palette(0)}})
	if want := "\x1b[1;1H\x1b[38;5;0m\x1b[48;5;0ma\x1b[0m "; got != want {
		t.Errorf("wanted %q, got %q", want, got)
	}
}

func TestCellKeepsItsColorsAndAttributes(t *testing.T) {
	fb, err := NewFrameBuffer(2, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := Cell{Character: 'q', Foreground: RGB(1, 2, 3), Background: Palette(9), Attrs: Bold | Reverse}
	fb.Set(1, 0, want)

	if got := fb.At(1, 0); got != want {
		t.Errorf("wanted %+v, got %+v", want, got)
	}
	if !fb.At(1, 0).Attrs.Has(Reverse) || fb.At(1, 0).Attrs.Has(Underline) {
		t.Errorf("wanted exactly Bold and Reverse kept")
	}
}

func TestDrawTextAndBoxUseTheColorGiven(t *testing.T) {
	fb, err := NewFrameBuffer(6, 3)
	if err != nil {
		t.Fatal(err)
	}
	fb.DrawBox(0, 0, 6, 3, RGB(9, 8, 7))
	fb.DrawText(1, 1, "hi", Palette(5))

	if got := fb.At(0, 0).Foreground; got != RGB(9, 8, 7) {
		t.Errorf("wanted the box corner in the RGB color, got %v", got)
	}
	if got := fb.At(1, 1); got.Foreground != Palette(5) || got.Background != (Color{}) || got.Attrs != 0 {
		t.Errorf("wanted text in palette 5 on the default background, plain, got %+v", got)
	}

	fb.DrawText(1, 1, "hi", Color{})
	if got := fb.At(1, 1).Foreground; !got.IsDefault() {
		t.Errorf("wanted the zero Color to give the default color, got %v", got)
	}
}

func TestFrameBufferString(t *testing.T) {
	ch := func(r rune) Cell { return Cell{Character: r} }
	frame := func(w, h int, cells map[[2]int]Cell) *FrameBuffer {
		fb, err := NewFrameBuffer(w, h)
		if err != nil {
			t.Fatal(err)
		}
		for at, c := range cells {
			fb.Set(at[0], at[1], c)
		}
		return fb
	}

	tests := []struct {
		name string
		fb   *FrameBuffer
		want string
	}{
		{"blank is empty lines", frame(4, 3, nil), "\n\n"},
		{"one row", frame(4, 1, map[[2]int]Cell{{0, 0}: ch('a'), {1, 0}: ch('b')}), "ab"},
		{"cells land where they were set", frame(3, 3, map[[2]int]Cell{{0, 0}: ch('a'), {2, 1}: ch('b'), {1, 2}: ch('c')}), "a\n  b\n c"},
		{"inner spaces stay, trailing ones go", frame(6, 1, map[[2]int]Cell{{0, 0}: ch('a'), {3, 0}: ch('b')}), "a  b"},
		{"leading spaces stay", frame(4, 1, map[[2]int]Cell{{2, 0}: ch('x')}), "  x"},
		{"a drawn space is trimmed like a blank one", frame(3, 1, map[[2]int]Cell{{0, 0}: ch('a'), {1, 0}: ch(' '), {2, 0}: ch(' ')}), "a"},
		{"NUL is a space", frame(3, 1, map[[2]int]Cell{{0, 0}: ch('a'), {1, 0}: ch(0), {2, 0}: ch('b')}), "a b"},
		{"a trailing NUL is trimmed", frame(3, 1, map[[2]int]Cell{{0, 0}: ch('a'), {1, 0}: ch(0)}), "a"},
		{"control characters are question marks", frame(4, 1, map[[2]int]Cell{{0, 0}: ch('\x1b'), {1, 0}: ch('\n'), {2, 0}: ch(0x7f)}), "???"},
		{"wide characters are question marks", frame(2, 1, map[[2]int]Cell{{0, 0}: ch('中'), {1, 0}: ch('😀')}), "??"},
		{"box drawing is kept", frame(3, 1, map[[2]int]Cell{{0, 0}: ch('┌'), {1, 0}: ch('─'), {2, 0}: ch('┐')}), "┌─┐"},
		{"colors and attributes are not shown", frame(2, 1, map[[2]int]Cell{{0, 0}: {Character: 'a', Foreground: Palette(1), Background: Palette(2), Attrs: Bold | Reverse}, {1, 0}: {Background: Palette(4)}}), "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fb.String(); got != tt.want {
				t.Errorf("wanted %q, got %q", tt.want, got)
			}
		})
	}
}

func TestFrameBufferStringShowsCanvasPixelsOnlyOnceComposed(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "") // Canvas.Seamless depends on the terminal running the tests
	fb, err := NewFrameBuffer(3, 2)
	if err != nil {
		t.Fatal(err)
	}
	fb.Canvas().Set(1, 0, Palette(1)) // the upper half of cell (1, 0)

	if got := fb.String(); got != "\n" {
		t.Errorf("wanted the pixel not shown before Compose, got %q", got)
	}
	fb.Canvas().Compose()
	if got := fb.String(); got != " ▀\n" {
		t.Errorf("wanted the composed half block, got %q", got)
	}
}

func TestFrameBufferStringPrintsWithFmt(t *testing.T) {
	fb, err := NewFrameBuffer(3, 2)
	if err != nil {
		t.Fatal(err)
	}
	fb.DrawText(0, 0, "hi", Color{})

	if got := fmt.Sprint(fb); got != "hi\n" {
		t.Errorf("wanted fmt to print the frame, got %q", got)
	}
	if got := fmt.Sprintf("%v|%s", fb, fb); got != "hi\n|hi\n" {
		t.Errorf("wanted %%v and %%s to print the frame, got %q", got)
	}
}

func TestFrameBufferStringOfAnEmptyFrame(t *testing.T) {
	if got := (&FrameBuffer{}).String(); got != "" {
		t.Errorf("wanted nothing for a frame with no rows, got %q", got)
	}
}
