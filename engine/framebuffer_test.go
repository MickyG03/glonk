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
			[][]Cell{{{Character: 'a', Color: 1}, {Character: 'b', Color: 1}, {Character: 'c', Color: 0}}},
			"\x1b[1;1H\x1b[38;5;1mab\x1b[0mc",
		},
		{
			"color changes",
			[][]Cell{{{Character: 'a', Color: 1}, {Character: 'b', Color: 2}}},
			"\x1b[1;1H\x1b[38;5;1ma\x1b[38;5;2mb\x1b[0m",
		},
		{
			"row ends colored",
			[][]Cell{{{Character: 'a', Color: 2}}, {{Character: 'b', Color: 2}}},
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
		{"four bytes", '😀', "😀"},
		{"surrogate becomes replacement", 0xD800, "�"},
		{"negative becomes replacement", -1, "�"},
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
						c.Color = uint8(1 + x%200)
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
