package engine

import (
	"math"
	"slices"
	"testing"
)

// cellDepth is the depth Compose left in the FrameBuffer for cell (x, y).
func cellDepth(fb *FrameBuffer, x, y int) float64 { return fb.depth[y*fb.width+x] }

// A cell takes the depth of its nearer pixel, whichever of the two it is, and in both
// ways Compose draws a cell.
func TestComposeGivesACellItsNearerPixelsDepth(t *testing.T) {
	tests := []struct {
		name       string
		top, below float64 // depths; 0 means that pixel is not set
		want       float64
	}{
		{"top nearer", 3, 5, 3},
		{"bottom nearer", 5, 3, 3},
		{"equal", 4, 4, 4},
		{"only the top", 5, 0, 5},
		{"only the bottom", 0, 5, 5},
	}
	for _, seamless := range []bool{false, true} {
		for _, tt := range tests {
			name := tt.name
			if seamless {
				name += ", solid cells"
			}
			t.Run(name, func(t *testing.T) {
				fb, c := newTestCanvas(t, 1, 1)
				c.Seamless = seamless
				if tt.top > 0 {
					c.SetDepth(0, 0, tt.top, pix(3))
				}
				if tt.below > 0 {
					c.SetDepth(0, 1, tt.below, pix(6))
				}

				c.Compose()

				if got := cellDepth(fb, 0, 0); got != tt.want {
					t.Errorf("wanted depth %v, got %v", tt.want, got)
				}
				// Depth-tested drawing into the cell: hidden if farther, drawn if nearer.
				if fb.SetDepth(0, 0, tt.want+0.5, Cell{Character: 'x'}) {
					t.Errorf("wanted something farther than the pixels to be hidden")
				}
				if !fb.SetDepth(0, 0, tt.want-0.5, Cell{Character: 'y'}) || fb.At(0, 0).Character != 'y' {
					t.Errorf("wanted something nearer than the pixels to be drawn")
				}
			})
		}
	}
}

// A cell with no pixel is not Compose's: its cell and its depth stay as they were.
func TestComposeLeavesEmptyCellsAndTheirDepthAlone(t *testing.T) {
	fb, c := newTestCanvas(t, 3, 1)
	fb.SetDepth(0, 0, 7, Cell{Character: 'a'})
	fb.SetDepth(2, 0, 9, Cell{Character: 'b'})
	c.SetDepth(1, 0, 2, pix(5))

	c.Compose()

	if got := fb.At(0, 0); got.Character != 'a' || cellDepth(fb, 0, 0) != 7 {
		t.Errorf("wanted cell 0 left as it was at depth 7, got %+v at %v", got, cellDepth(fb, 0, 0))
	}
	if got := fb.At(2, 0); got.Character != 'b' || cellDepth(fb, 2, 0) != 9 {
		t.Errorf("wanted cell 2 left as it was at depth 9, got %+v at %v", got, cellDepth(fb, 2, 0))
	}
	if got := cellDepth(fb, 1, 0); got != 2 {
		t.Errorf("wanted the filled cell at depth 2, got %v", got)
	}
}

// Pixels replace the cells under them, as they always have; the cell's depth goes with
// its content, even if what was there was nearer.
func TestComposeReplacesACellAndItsDepth(t *testing.T) {
	fb, c := newTestCanvas(t, 1, 1)
	fb.SetDepth(0, 0, 1, Cell{Character: 'n'}) // nearer than the pixel
	c.SetDepth(0, 0, 6, pix(4))

	c.Compose()

	if got := fb.At(0, 0); got.Character == 'n' {
		t.Errorf("wanted the pixels to cover the cell, got %+v", got)
	}
	if got := cellDepth(fb, 0, 0); got != 6 {
		t.Errorf("wanted the depth of what is now shown, 6, got %v", got)
	}
}

// An erased pixel leaves its old depth behind in the Canvas, but it is not a pixel any
// more, so it does not count, whichever of the two it is.
func TestComposeIgnoresTheDepthOfAnErasedPixel(t *testing.T) {
	for _, seamless := range []bool{false, true} {
		for _, erasedRow := range []int{0, 1} { // the top pixel, then the bottom one
			fb, c := newTestCanvas(t, 1, 1)
			c.Seamless = seamless
			// The erased pixel was the nearer one; the pixel left is at depth 5.
			c.SetDepth(0, erasedRow, 1, pix(6))
			c.SetDepth(0, 1-erasedRow, 5, pix(3))
			c.Set(0, erasedRow, pix(0))

			c.Compose()

			if got := cellDepth(fb, 0, 0); got != 5 {
				t.Errorf("seamless %v, erased row %d: wanted the depth of the pixel that is there, 5, got %v", seamless, erasedRow, got)
			}
		}
	}
}

// A pixel set without a depth is as far away as can be, so depth-tested drawing lands on it.
func TestComposeOfPixelsWithoutDepthIsInfinitelyFar(t *testing.T) {
	fb, c := newTestCanvas(t, 1, 1)
	c.Set(0, 0, pix(2))

	c.Compose()

	if got := cellDepth(fb, 0, 0); !math.IsInf(got, 1) {
		t.Errorf("wanted an infinite depth, got %v", got)
	}
	if !fb.SetDepth(0, 0, 1e300, Cell{Character: 'x'}) {
		t.Errorf("wanted depth-tested drawing to land on a pixel with no depth")
	}
}

// What the step is for: surfaces and lines drawn into the cells after Compose are hidden
// behind pixel solids that are nearer, and drawn over those that are farther.
func TestDepthTestedDrawingAfterComposeSeesPixelSolids(t *testing.T) {
	big := [6]float64{-100, -100, 300, -100, -100, 300}
	fb, c := newTestCanvas(t, 6, 2) // 6x4 pixels
	// A pixel triangle covering the left three cells, at depth 4.
	c.FillTriangle(3, -100, 4, 3, 300, 4, -300, 100, 4, pix(7))
	c.Compose()
	solid := fb.At(0, 0)
	if solid.Background != pix(7) || fb.At(2, 1).Background != pix(7) || fb.At(3, 0) != BlankCell {
		t.Fatalf("wanted pixels in the left three cells and none beyond, got %+v, %+v and %+v", solid, fb.At(2, 1), fb.At(3, 0))
	}

	// A surface at depth 6 (farther) and then one at depth 2 (nearer), over everything.
	fb.FillTriangle(big[0], big[1], 6, big[2], big[3], 6, big[4], big[5], 6, Cell{Character: 'f'})
	for x := range 6 {
		got := fb.At(x, 0)
		if x < 3 && got != solid {
			t.Errorf("cell %d: wanted the pixels to hide a farther surface, got %+v", x, got)
		}
		if x >= 3 && got.Character != 'f' {
			t.Errorf("cell %d: wanted the surface where there are no pixels, got %+v", x, got)
		}
	}
	fb.FillTriangle(big[0], big[1], 2, big[2], big[3], 2, big[4], big[5], 2, Cell{Character: 'n'})
	for x := range 6 {
		if got := fb.At(x, 0).Character; got != 'n' {
			t.Errorf("cell %d: wanted a nearer surface to cover the pixels, got %q", x, got)
		}
	}
}

func TestLinesAfterComposeGoBehindOrInFrontOfPixelSolids(t *testing.T) {
	fb, c := newTestCanvas(t, 6, 1)
	c.FillTriangle(3, -100, 4, 3, 300, 4, -300, 100, 4, pix(7)) // the left three cells, at depth 4
	c.Compose()
	pixels := fb.At(0, 0)
	if pixels.Background != pix(7) || fb.At(2, 0).Background != pix(7) || fb.At(3, 0) != BlankCell {
		t.Fatalf("wanted pixels in the left three cells and none beyond")
	}

	// A line along the row from depth 9 down to depth 1. 1/z runs evenly, so its depth is
	// 9, 3.46, 2.14, 1.55, 1.22, 1 in the six cells: farther than 4 only in the first.
	fb.DrawLineDepth(0, 0, 9, 5, 0, 1, Cell{Character: '#'})

	if got := fb.At(0, 0); got != pixels {
		t.Errorf("wanted the pixels to hide the line where it is farther, got %+v", got)
	}
	if got := cellDepth(fb, 0, 0); got != 4 {
		t.Errorf("wanted the pixels' depth kept where the line lost, got %v", got)
	}
	for x := 1; x < 6; x++ {
		if got := fb.At(x, 0).Character; got != '#' {
			t.Errorf("cell %d: wanted the line to show, nearer than the pixels or over empty cells, got %q", x, got)
		}
	}
	if got := cellDepth(fb, 1, 0); math.Abs(got-1/(0.8/9+0.2)) > 1e-12 {
		t.Errorf("wanted the line's depth in cell 1, got %v", got)
	}
}

// The depth lasts one frame, like everything else: Clear puts it back.
func TestClearForgetsComposedDepth(t *testing.T) {
	fb, c := newTestCanvas(t, 2, 1)
	c.SetDepth(0, 0, 2, pix(4))
	c.Compose()

	fb.Clear()

	for x := range 2 {
		if got := cellDepth(fb, x, 0); !math.IsInf(got, 1) {
			t.Errorf("cell %d: wanted the depth back to infinite after Clear, got %v", x, got)
		}
	}
	if !slices.Equal(fb.frame, newFB(t, 2, 1).frame) {
		t.Errorf("wanted a blank frame after Clear")
	}
}

// Each cell gets the depth of its own pixels, across rows and columns.
func TestComposeDepthLandsInTheRightCell(t *testing.T) {
	for _, seamless := range []bool{false, true} {
		fb, c := newTestCanvas(t, 3, 2) // 3x4 pixels
		c.Seamless = seamless
		want := map[[2]int]float64{}
		for y := range 2 {
			for x := range 3 {
				if x == 1 && y == 0 {
					continue // left empty
				}
				// A different depth for each pixel, the lower one nearer in odd cells.
				topZ, bottomZ := float64(10+10*y+x), float64(20+10*y+x)
				if (x+y)%2 == 1 {
					topZ, bottomZ = bottomZ, topZ
				}
				c.SetDepth(x, 2*y, topZ, pix(3))
				c.SetDepth(x, 2*y+1, bottomZ, pix(6))
				want[[2]int{x, y}] = min(topZ, bottomZ)
			}
		}

		c.Compose()

		for y := range 2 {
			for x := range 3 {
				w, ok := want[[2]int{x, y}]
				if !ok {
					w = math.Inf(1)
				}
				if got := cellDepth(fb, x, y); got != w {
					t.Errorf("seamless %v, cell (%d,%d): wanted depth %v, got %v", seamless, x, y, w, got)
				}
			}
		}
	}
}

func BenchmarkCompose(b *testing.B) {
	for _, tt := range []struct {
		name     string
		seamless bool
	}{{"half blocks", false}, {"solid cells", true}} {
		fb, err := NewFrameBuffer(200, 60)
		if err != nil {
			b.Fatal(err)
		}
		c := fb.Canvas()
		c.Seamless = tt.seamless
		// Triangles over most of the canvas, so that most cells have pixels.
		c.FillTriangle(0, 0, 2, 200, 0, 5, 0, 120, 3, Palette(33))
		c.FillTriangle(200, 120, 2, 200, 0, 5, 0, 120, 3, Palette(208))

		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				c.Compose()
			}
		})
	}
}
