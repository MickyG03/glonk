package engine

import (
	"math"
	"testing"
)

func newTestCanvas(t *testing.T, cols, rows int) (*FrameBuffer, *Canvas) {
	t.Helper()
	// The default for Seamless depends on the terminal running the tests.
	t.Setenv("TERM_PROGRAM", "")
	fb, err := NewFrameBuffer(cols, rows)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	return fb, fb.Canvas()
}

func TestCanvasSize(t *testing.T) {
	_, c := newTestCanvas(t, 5, 3)

	if c.Width() != 5 || c.Height() != 6 {
		t.Errorf("wanted 5x6 pixels for 5x3 cells, got %dx%d", c.Width(), c.Height())
	}
}

func TestCanvasIsSharedAndStable(t *testing.T) {
	fb, c := newTestCanvas(t, 2, 2)

	if fb.Canvas() != c {
		t.Errorf("wanted the same canvas on every call")
	}

	other, err := NewFrameBuffer(2, 2)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	if other.Canvas() == c {
		t.Errorf("wanted each frame buffer to have its own canvas")
	}
}

func TestCanvasSetAndAt(t *testing.T) {
	_, c := newTestCanvas(t, 3, 2)

	c.Set(2, 3, 42)

	if got := c.At(2, 3); got != 42 {
		t.Errorf("wanted 42, got %d", got)
	}
	if got := c.At(1, 3); got != 0 {
		t.Errorf("wanted an untouched pixel empty, got %d", got)
	}
}

func TestCanvasBounds(t *testing.T) {
	_, c := newTestCanvas(t, 2, 2)

	for _, p := range [][2]int{{-1, 0}, {0, -1}, {2, 0}, {0, 4}, {2, 4}} {
		c.Set(p[0], p[1], 9)
		if c.SetDepth(p[0], p[1], 1, 9) {
			t.Errorf("wanted SetDepth at %v to be ignored, got landed", p)
		}
		if got := c.At(p[0], p[1]); got != 0 {
			t.Errorf("wanted At%v to be empty, got %d", p, got)
		}
	}
	for y := range c.Height() {
		for x := range c.Width() {
			if c.At(x, y) != 0 {
				t.Errorf("wanted every pixel still empty, got %d at (%d,%d)", c.At(x, y), x, y)
			}
		}
	}
}

func TestCanvasSetDepth(t *testing.T) {
	tests := []struct {
		name      string
		first     float64
		second    float64
		wantWrote bool
		want      uint8
	}{
		{"nearer second wins", 5, 2, true, 2},
		{"farther second loses", 2, 5, false, 1},
		{"tie keeps the first", 3, 3, false, 1},
		{"nearer than a huge depth", 1e300, 1, true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, c := newTestCanvas(t, 2, 2)

			if !c.SetDepth(1, 1, tt.first, 1) {
				t.Fatalf("wanted the first write to an empty pixel to land")
			}
			if got := c.SetDepth(1, 1, tt.second, 2); got != tt.wantWrote {
				t.Errorf("wanted second write landed=%v, got %v", tt.wantWrote, got)
			}
			if got := c.At(1, 1); got != tt.want {
				t.Errorf("wanted %d, got %d", tt.want, got)
			}
		})
	}
}

func TestCanvasSetDepthEdgeCases(t *testing.T) {
	_, c := newTestCanvas(t, 1, 1)

	if c.SetDepth(0, 0, math.NaN(), 7) {
		t.Errorf("wanted a NaN depth to be rejected, got landed")
	}
	if got := c.At(0, 0); got != 0 {
		t.Errorf("wanted the pixel to stay empty, got %d", got)
	}
	if !c.SetDepth(0, 0, math.MaxFloat64, 7) {
		t.Errorf("wanted any finite depth to beat an empty pixel, got rejected")
	}
}

func TestCanvasSetIgnoresDepth(t *testing.T) {
	_, c := newTestCanvas(t, 1, 1)

	c.SetDepth(0, 0, 1, 1)
	c.Set(0, 0, 2)
	if got := c.At(0, 0); got != 2 {
		t.Errorf("wanted Set to overwrite regardless of depth, got %d", got)
	}

	if c.SetDepth(0, 0, 2, 3) {
		t.Errorf("wanted Set to leave depth at 1, but a farther write landed")
	}
	if !c.SetDepth(0, 0, 0.5, 4) {
		t.Errorf("wanted Set to leave depth at 1, but a nearer write was rejected")
	}
}

func TestCanvasClear(t *testing.T) {
	_, c := newTestCanvas(t, 2, 2)

	c.SetDepth(1, 2, 1, 5)
	c.Clear()

	if got := c.At(1, 2); got != 0 {
		t.Errorf("wanted an empty pixel after Clear, got %d", got)
	}
	if !c.SetDepth(1, 2, 100, 6) {
		t.Errorf("wanted a farther write to land after Clear, got rejected")
	}
}

func TestFrameBufferClearClearsCanvas(t *testing.T) {
	fb, c := newTestCanvas(t, 2, 2)

	c.SetDepth(0, 0, 1, 5)
	fb.Clear()

	if got := c.At(0, 0); got != 0 {
		t.Errorf("wanted FrameBuffer.Clear to empty the canvas, got %d", got)
	}
	if !c.SetDepth(0, 0, 100, 6) {
		t.Errorf("wanted FrameBuffer.Clear to reset canvas depth, got rejected")
	}
}

func TestCanvasCompose(t *testing.T) {
	tests := []struct {
		name        string
		top, bottom uint8
		want        Cell
	}{
		{"both empty leaves the cell alone", 0, 0, Cell{Character: 'x', Color: 9, Background: 8}},
		{"equal pixels are a solid background", 5, 5, Cell{Character: ' ', Background: 5}},
		{"two colors", 5, 6, Cell{Character: '▀', Color: 5, Background: 6}},
		{"two colors, other way round", 6, 5, Cell{Character: '▀', Color: 6, Background: 5}},
		{"top only", 5, 0, Cell{Character: '▀', Color: 5}},
		{"bottom only", 0, 6, Cell{Character: '▄', Color: 6}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, c := newTestCanvas(t, 3, 2)
			sentinel := Cell{Character: 'x', Color: 9, Background: 8}
			for y := range 2 {
				for x := range 3 {
					fb.Set(x, y, sentinel)
				}
			}

			// Cell (1,1) covers pixel rows 2 and 3.
			c.Set(1, 2, tt.top)
			c.Set(1, 3, tt.bottom)
			c.Compose()

			if got := fb.At(1, 1); got != tt.want {
				t.Errorf("wanted %+v, got %+v", tt.want, got)
			}
			for _, p := range [][2]int{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {2, 1}} {
				if got := fb.At(p[0], p[1]); got != sentinel {
					t.Errorf("wanted cell %v untouched, got %+v", p, got)
				}
			}
		})
	}
}

func TestCanvasComposeMapsPixelRowsToCells(t *testing.T) {
	fb, c := newTestCanvas(t, 2, 3)

	// One pixel per cell row, always on the top line of the cell.
	c.Set(0, 0, 1)
	c.Set(0, 2, 2)
	c.Set(0, 4, 3)
	// And one on the bottom line only.
	c.Set(1, 5, 4)
	c.Compose()

	want := []Cell{
		{Character: '▀', Color: 1},
		{Character: '▀', Color: 2},
		{Character: '▀', Color: 3},
	}
	for row, w := range want {
		if got := fb.At(0, row); got != w {
			t.Errorf("wanted %+v at (0,%d), got %+v", w, row, got)
		}
	}
	if got, w := fb.At(1, 2), (Cell{Character: '▄', Color: 4}); got != w {
		t.Errorf("wanted %+v at (1,2), got %+v", w, got)
	}
	if got := fb.At(1, 0); got != BlankCell {
		t.Errorf("wanted an untouched cell blank, got %+v", got)
	}
}

// Anything drawn after Compose lands on top; anything drawn before is covered.
func TestCanvasComposeLayering(t *testing.T) {
	fb, c := newTestCanvas(t, 2, 1)

	fb.Set(0, 0, Cell{Character: 'a'})
	fb.Set(1, 0, Cell{Character: 'b'})
	c.Set(0, 0, 5)
	c.Compose()
	fb.Set(1, 0, Cell{Character: 'h'})

	if got := fb.At(0, 0); got.Character != '▀' {
		t.Errorf("wanted the pixels to cover what was drawn before, got %+v", got)
	}
	if got := fb.At(1, 0); got.Character != 'h' {
		t.Errorf("wanted what is drawn after to stay on top, got %+v", got)
	}
}

func TestCanvasFillTriangleUsesPixelRows(t *testing.T) {
	fb, c := newTestCanvas(t, 4, 2)

	// A pixel triangle covering the left half of every pixel row.
	c.FillTriangle(0, 0, 1, 2, 0, 1, 0, 4, 1, 7)

	// Row 0 of pixels is 2 wide, row 1 is 1 wide, then 1 wide, then none.
	want := [][]uint8{
		{7, 7, 0, 0},
		{7, 0, 0, 0},
		{7, 0, 0, 0},
		{0, 0, 0, 0},
	}
	for y, row := range want {
		for x, w := range row {
			if got := c.At(x, y); got != w {
				t.Errorf("wanted pixel (%d,%d)=%d, got %d", x, y, w, got)
			}
		}
	}

	c.Compose()
	if got := fb.At(0, 0); got != (Cell{Character: ' ', Background: 7}) {
		t.Errorf("wanted cell (0,0) solid after compose, got %+v", got)
	}
}

func TestCanvasFillTriangleDepth(t *testing.T) {
	_, c := newTestCanvas(t, 2, 2)
	big := [6]float64{-100, -100, 300, -100, -100, 300}

	c.FillTriangle(big[0], big[1], 5, big[2], big[3], 5, big[4], big[5], 5, 1)
	c.FillTriangle(big[0], big[1], 2, big[2], big[3], 2, big[4], big[5], 2, 2)
	c.FillTriangle(big[0], big[1], 9, big[2], big[3], 9, big[4], big[5], 9, 3)

	for y := range c.Height() {
		for x := range c.Width() {
			if got := c.At(x, y); got != 2 {
				t.Errorf("wanted the nearest triangle at (%d,%d), got %d", x, y, got)
			}
		}
	}
}

func TestCanvasSeamlessDefault(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"apple terminal", "Apple_Terminal", true},
		{"unset", "", false},
		{"ghostty", "ghostty", false},
		{"iterm", "iTerm.app", false},
		{"tmux", "tmux", false},
		{"different case is not apple terminal", "apple_terminal", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM_PROGRAM", tt.env)

			fb, err := NewFrameBuffer(2, 2)
			if err != nil {
				t.Fatalf("unable to initialize frame buffer: %v", err)
			}

			if got := fb.Canvas().Seamless; got != tt.want {
				t.Errorf("wanted Seamless=%v for TERM_PROGRAM=%q, got %v", tt.want, tt.env, got)
			}
		})
	}
}

func TestCanvasSeamlessOverridesDetection(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "Apple_Terminal")
	fb, err := NewFrameBuffer(1, 1)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	c := fb.Canvas()
	c.Seamless = false
	c.Set(0, 0, 5)
	c.Compose()

	if got := fb.At(0, 0); got != (Cell{Character: '▀', Color: 5}) {
		t.Errorf("wanted exact mode after turning Seamless off, got %+v", got)
	}
}

func TestCanvasComposeSeamless(t *testing.T) {
	type px struct {
		color uint8
		depth float64 // 0 means draw with Set, which leaves the depth at infinity
	}
	set := func(c *Canvas, x, y int, p px) {
		switch {
		case p.color == 0:
		case p.depth == 0:
			c.Set(x, y, p.color)
		default:
			c.SetDepth(x, y, p.depth, p.color)
		}
	}

	tests := []struct {
		name        string
		top, bottom px
		want        Cell
	}{
		{"both empty leaves the cell alone", px{}, px{}, Cell{Character: 'x', Color: 9, Background: 8}},
		{"top only", px{5, 3}, px{}, Cell{Character: ' ', Background: 5}},
		{"bottom only", px{}, px{6, 3}, Cell{Character: ' ', Background: 6}},
		{"equal colors", px{5, 3}, px{5, 4}, Cell{Character: ' ', Background: 5}},
		{"top is nearer", px{5, 2}, px{6, 3}, Cell{Character: ' ', Background: 5}},
		{"bottom is nearer", px{5, 3}, px{6, 2}, Cell{Character: ' ', Background: 6}},
		{"equal depth goes to the top", px{5, 3}, px{6, 3}, Cell{Character: ' ', Background: 5}},
		{"depthless pixels go to the top", px{5, 0}, px{6, 0}, Cell{Character: ' ', Background: 5}},
		{"a depth beats a depthless pixel below", px{5, 0}, px{6, 9}, Cell{Character: ' ', Background: 6}},
		{"a depth beats a depthless pixel above", px{5, 9}, px{6, 0}, Cell{Character: ' ', Background: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, c := newTestCanvas(t, 3, 2)
			c.Seamless = true
			sentinel := Cell{Character: 'x', Color: 9, Background: 8}
			for y := range 2 {
				for x := range 3 {
					fb.Set(x, y, sentinel)
				}
			}

			set(c, 1, 2, tt.top)
			set(c, 1, 3, tt.bottom)
			c.Compose()

			if got := fb.At(1, 1); got != tt.want {
				t.Errorf("wanted %+v, got %+v", tt.want, got)
			}
			for _, p := range [][2]int{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {2, 1}} {
				if got := fb.At(p[0], p[1]); got != sentinel {
					t.Errorf("wanted cell %v untouched, got %+v", p, got)
				}
			}
		})
	}
}

// Seamless output has no glyphs at all, whatever the pixels are.
func TestCanvasComposeSeamlessNeverUsesBlockCharacters(t *testing.T) {
	fb, c := newTestCanvas(t, 6, 4)
	c.Seamless = true

	for y := range c.Height() {
		for x := range c.Width() {
			if (x+y)%3 != 0 {
				c.Set(x, y, uint8(1+(x*7+y*3)%200))
			}
		}
	}
	c.Compose()

	for y := range fb.Height() {
		for x := range fb.Width() {
			if got := fb.At(x, y); got.Character == '▀' || got.Character == '▄' || got.Color != 0 {
				t.Fatalf("wanted only solid cells, got %+v at (%d,%d)", got, x, y)
			}
		}
	}
}

// Erasing a pixel with Set(0) leaves its old depth behind; an erased pixel must not win.
func TestCanvasComposeSeamlessIgnoresErasedPixelDepth(t *testing.T) {
	fb, c := newTestCanvas(t, 1, 1)
	c.Seamless = true

	c.SetDepth(0, 0, 5, 3)
	c.SetDepth(0, 1, 1, 6)
	c.Set(0, 1, 0)
	c.Compose()

	if got, want := fb.At(0, 0), (Cell{Character: ' ', Background: 3}); got != want {
		t.Errorf("wanted %+v, got %+v", want, got)
	}
}
