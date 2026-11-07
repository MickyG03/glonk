package engine

import (
	"math"
	"math/rand/v2"
	"testing"
)

type point struct {
	X, Y int
}

func TestDrawLine(t *testing.T) {
	width := 10
	height := 10

	var tests = []struct {
		name           string
		x0, y0, x1, y1 int
		want           []point
	}{
		{
			"point",
			2, 2, 2, 2,
			[]point{{2, 2}},
		},
		{
			"horizontal",
			0, 0, 3, 0,
			[]point{{0, 0}, {1, 0}, {2, 0}, {3, 0}},
		},
		{
			"vertical",
			0, 0, 0, 3,
			[]point{{0, 0}, {0, 1}, {0, 2}, {0, 3}},
		},
		{
			"diagonal",
			0, 0, 3, 3,
			[]point{{0, 0}, {1, 1}, {2, 2}, {3, 3}},
		},
		{
			"steep",
			0, 0, 1, 3,
			[]point{{0, 0}, {0, 1}, {1, 2}, {1, 3}},
		},
		{
			"reversed",
			3, 0, 0, 0,
			[]point{{0, 0}, {1, 0}, {2, 0}, {3, 0}},
		},
		{
			"partly-off-screen",
			-2, 0, 2, 0,
			[]point{{0, 0}, {1, 0}, {2, 0}},
		},
		{
			"tie",
			0, 0, 1, 2,
			[]point{{0, 0}, {1, 1}, {1, 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nfb, err := NewFrameBuffer(width, height)
			if err != nil {
				t.Fatalf("failed to initialize framebuffer: %v", err)
			}

			nfb.DrawLine(tt.x0, tt.y0, tt.x1, tt.y1, Cell{Character: '#'})

			wantCells := map[int]Cell{}
			for _, p := range tt.want {
				wantCells[getIndex(p.X, p.Y, width)] = Cell{Character: '#'}
			}

			assertFrameEquals(t, wantCells, nfb)
		})
	}
}

type pointCell struct {
	X, Y  int
	Value rune
}

func TestDrawText(t *testing.T) {
	width := 10
	height := 10
	var color uint8 = 7

	var tests = []struct {
		name string
		x, y int
		text string
		want []pointCell
	}{
		{
			"basic", 1, 1, "hi",
			[]pointCell{{1, 1, 'h'}, {2, 1, 'i'}},
		},
		{
			"multibyte", 0, 0, "é→",
			[]pointCell{{0, 0, 'é'}, {1, 0, '→'}},
		},
		{
			"clips-right", 8, 0, "abcd",
			[]pointCell{{8, 0, 'a'}, {9, 0, 'b'}},
		},
		{
			"clips-left", -1, 0, "ab",
			[]pointCell{{0, 0, 'b'}},
		},
		{
			"empty", 0, 0, "",
			[]pointCell{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nfb, err := NewFrameBuffer(width, height)
			if err != nil {
				t.Fatalf("failed to initialize framebuffer: %v", err)
			}

			nfb.DrawText(tt.x, tt.y, tt.text, color)

			wantCells := map[int]Cell{}
			for _, p := range tt.want {
				wantCells[getIndex(p.X, p.Y, width)] = Cell{Character: p.Value, Color: color}
			}

			assertFrameEquals(t, wantCells, nfb)
		})
	}
}

func assertRows(t *testing.T, fb *FrameBuffer, rows []string, color uint8) {
	t.Helper()

	if len(rows) != fb.Height() {
		t.Fatalf("wanted len(rows)=%v, got %v", fb.Height(), len(rows))
	}

	for y := range len(rows) {
		runes := []rune(rows[y])

		if len(runes) != fb.Width() {
			t.Fatalf("wanted len(runes)=%v, got %v", fb.Width(), len(runes))
		}

		for x := range runes {
			r := runes[x]
			got := fb.At(x, y)
			if r == ' ' && got != BlankCell {
				t.Errorf("wanted %v at (%d,%d), got %v", BlankCell, x, y, got)
			}
			if r != ' ' {
				if r != got.Character {
					t.Errorf("wanted %v at (%d,%d), got %v", r, x, y, got.Character)
				}

				if color != got.Color {
					t.Errorf("wanted %v at (%d,%d), got %v", color, x, y, got.Color)
				}
			}
		}
	}
}

func TestDrawBox(t *testing.T) {
	var tests = []struct {
		name       string
		color      uint8
		bw, bh     int
		x, y, w, h int
		want       []string
	}{
		{
			"fills buffer", 0, 4, 3, 0, 0, 4, 3,
			[]string{"┌──┐", "│  │", "└──┘"},
		},
		{
			"offset, not square", 5, 6, 5, 1, 1, 4, 3,
			[]string{"      ", " ┌──┐ ", " │  │ ", " └──┘ ", "      "},
		},
		{
			"too small", 0, 3, 3, 0, 0, 1, 3,
			[]string{"   ", "   ", "   "},
		},
		{
			"clipped", 0, 3, 3, -1, -1, 3, 3,
			[]string{" │ ", "─┘ ", "   "},
		},
		{
			"too short", 0, 3, 3, 0, 0, 3, 1,
			[]string{"   ", "   ", "   "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nfb, err := NewFrameBuffer(tt.bw, tt.bh)
			if err != nil {
				t.Fatalf("failed to initialize framebuffer: %v", err)
			}

			nfb.DrawBox(tt.x, tt.y, tt.w, tt.h, tt.color)

			assertRows(t, nfb, tt.want, tt.color)
		})
	}
}

func TestFillTriangle(t *testing.T) {
	mark := Cell{Character: '#'}

	tests := []struct {
		name string
		w, h int
		// corners as x0,y0, x1,y1, x2,y2; every corner gets depth 1
		c    [6]float64
		want []string
	}{
		{
			"right triangle", 4, 4, [6]float64{0, 0, 4, 0, 0, 4},
			[]string{"####", "### ", "##  ", "#   "},
		},
		{
			"same triangle, other winding", 4, 4, [6]float64{0, 0, 0, 4, 4, 0},
			[]string{"####", "### ", "##  ", "#   "},
		},
		{
			"mirrored", 4, 4, [6]float64{4, 0, 0, 0, 4, 4},
			[]string{"####", " ###", "  ##", "   #"},
		},
		{
			"corners far off screen", 4, 4, [6]float64{-2, -2, 6, -2, -2, 6},
			[]string{"####", "### ", "##  ", "#   "},
		},
		{
			"last corner is the rightmost", 4, 4, [6]float64{0, 0, 0, 4, 4, 2},
			[]string{"#   ", "### ", "### ", "#   "},
		},
		{
			"last corner is the leftmost", 4, 4, [6]float64{4, 0, 4, 4, 0, 2},
			[]string{"   #", " ###", " ###", "   #"},
		},
		{
			"last corner is the topmost", 4, 4, [6]float64{4, 4, 0, 4, 0, 0},
			[]string{"#   ", "##  ", "### ", "####"},
		},
		{
			"last corner is the bottommost", 4, 4, [6]float64{0, 0, 4, 0, 0, 4},
			[]string{"####", "### ", "##  ", "#   "},
		},
		{
			"first corner is the bottommost", 4, 4, [6]float64{0, 4, 4, 0, 0, 0},
			[]string{"####", "### ", "##  ", "#   "},
		},
		{
			"covers everything", 3, 2, [6]float64{-100, -100, 300, -100, -100, 300},
			[]string{"###", "###"},
		},
		{
			"astronomical corners", 3, 2, [6]float64{-1e12, -1e12, 3e12, -1e12, -1e12, 3e12},
			[]string{"###", "###"},
		},
		{
			"entirely off screen", 4, 4, [6]float64{10, 10, 20, 10, 10, 20},
			[]string{"    ", "    ", "    ", "    "},
		},
		{
			"left of the screen", 4, 4, [6]float64{-10, 0, -5, 0, -10, 4},
			[]string{"    ", "    ", "    ", "    "},
		},
		{
			"above the screen", 4, 4, [6]float64{0, -10, 4, -10, 0, -5},
			[]string{"    ", "    ", "    ", "    "},
		},
		{
			"thin sliver between cell centres", 4, 4, [6]float64{0, 1.6, 4, 1.6, 0, 1.9},
			[]string{"    ", "    ", "    ", "    "},
		},
		{
			"one cell", 3, 3, [6]float64{1, 1, 2, 1, 1, 2},
			[]string{"   ", " # ", "   "},
		},
		{
			"collinear points", 4, 4, [6]float64{0, 0, 2, 2, 4, 4},
			[]string{"    ", "    ", "    ", "    "},
		},
		{
			"all corners equal", 4, 4, [6]float64{1, 1, 1, 1, 1, 1},
			[]string{"    ", "    ", "    ", "    "},
		},
		{
			"NaN corner", 4, 4, [6]float64{0, 0, math.NaN(), 0, 0, 4},
			[]string{"    ", "    ", "    ", "    "},
		},
		{
			"infinite corner", 4, 4, [6]float64{0, 0, math.Inf(1), 0, 0, 4},
			[]string{"    ", "    ", "    ", "    "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := NewFrameBuffer(tt.w, tt.h)
			if err != nil {
				t.Fatalf("unable to initialize frame buffer: %v", err)
			}

			c := tt.c
			fb.FillTriangle(c[0], c[1], 1, c[2], c[3], 1, c[4], c[5], 1, mark)

			assertRows(t, fb, tt.want, 0)
		})
	}
}

// Two triangles sharing an edge must tile a rectangle with no seam between them.
func TestFillTriangleSharedEdgeLeavesNoGap(t *testing.T) {
	mark := Cell{Character: '#'}

	for _, size := range []int{1, 2, 3, 4, 7, 16, 33} {
		fb, err := NewFrameBuffer(size, size)
		if err != nil {
			t.Fatalf("unable to initialize frame buffer: %v", err)
		}
		s := float64(size)

		fb.FillTriangle(0, 0, 1, s, 0, 1, s, s, 1, mark)
		fb.FillTriangle(0, 0, 1, s, s, 1, 0, s, 1, mark)

		for y := range size {
			for x := range size {
				if fb.At(x, y) != mark {
					t.Fatalf("size %d: wanted every cell filled, gap at (%d,%d)", size, x, y)
				}
			}
		}
	}
}

func TestFillTriangleDepth(t *testing.T) {
	near := Cell{Character: 'n'}
	far := Cell{Character: 'f'}
	// Each covers the whole 3x2 buffer.
	big := [6]float64{-100, -100, 300, -100, -100, 300}

	tests := []struct {
		name  string
		first Cell
		fz    float64
		next  Cell
		nz    float64
		want  rune
	}{
		{"near drawn second wins", far, 5, near, 2, 'n'},
		{"near drawn first stays", near, 2, far, 5, 'n'},
		{"tie keeps the first", near, 3, far, 3, 'n'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb, err := NewFrameBuffer(3, 2)
			if err != nil {
				t.Fatalf("unable to initialize frame buffer: %v", err)
			}

			fb.FillTriangle(big[0], big[1], tt.fz, big[2], big[3], tt.fz, big[4], big[5], tt.fz, tt.first)
			fb.FillTriangle(big[0], big[1], tt.nz, big[2], big[3], tt.nz, big[4], big[5], tt.nz, tt.next)

			for y := range 2 {
				for x := range 3 {
					if got := fb.At(x, y).Character; got != tt.want {
						t.Errorf("wanted %q at (%d,%d), got %q", tt.want, x, y, got)
					}
				}
			}
		})
	}
}

func TestFillTriangleOnlyHidesWhereCovered(t *testing.T) {
	fb, err := NewFrameBuffer(4, 4)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	// A far square behind a near triangle that covers its upper left half.
	fb.FillTriangle(0, 0, 9, 4, 0, 9, 4, 4, 9, Cell{Character: 'f'})
	fb.FillTriangle(0, 0, 9, 4, 4, 9, 0, 4, 9, Cell{Character: 'f'})
	fb.FillTriangle(0, 0, 1, 4, 0, 1, 0, 4, 1, Cell{Character: 'n'})

	assertRows(t, fb, []string{"nnnn", "nnnf", "nnff", "nfff"}, 0)
}

// 1/z, not z, varies linearly on screen; interpolating z would give 1.375 here.
func TestFillTrianglePerspectiveDepth(t *testing.T) {
	fb, err := NewFrameBuffer(8, 8)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}

	fb.FillTriangle(0, 0, 1, 8, 0, 3, 0, 8, 1, Cell{Character: '#'})

	// Cell (1,1) has centre (1.5,1.5): weights 0.625, 0.1875, 0.1875.
	want := 1 / (0.625/1 + 0.1875/3 + 0.1875/1)
	if got := fb.depth[1*8+1]; math.Abs(got-want) > 1e-9 {
		t.Errorf("wanted depth %v, got %v", want, got)
	}
}

func TestFillTriangleRejectsBadDepth(t *testing.T) {
	for _, z := range []float64{0, -1, math.NaN(), math.Inf(-1)} {
		fb, err := NewFrameBuffer(4, 4)
		if err != nil {
			t.Fatalf("unable to initialize frame buffer: %v", err)
		}

		fb.FillTriangle(-100, -100, z, 300, -100, 1, -100, 300, 1, Cell{Character: '#'})

		assertRows(t, fb, []string{"    ", "    ", "    ", "    "}, 0)
	}
}

// A cell centre lying exactly on the edge two triangles share must be drawn by at
// least one of them. Rounding makes the two edge tests disagree by about 1e-16,
// so without a tolerance some such cells fall through the seam.
func TestFillTriangleSeamThroughCellCentre(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	mark := Cell{Character: '#'}

	for trial := range 2000 {
		fb, err := NewFrameBuffer(8, 8)
		if err != nil {
			t.Fatalf("unable to initialize frame buffer: %v", err)
		}

		// The shared edge runs from c-d to c+d, through the centre of cell (3,4).
		cx, cy := 3.5, 4.5
		dx, dy := rng.Float64()*2-1, rng.Float64()*2-1
		if dx == 0 && dy == 0 {
			continue
		}
		// One far corner on each side of the edge, at random places along it.
		ax, ay := cx-dy*(0.2+rng.Float64())+dx*(rng.Float64()-0.5), cy+dx*(0.2+rng.Float64())+dy*(rng.Float64()-0.5)
		bx, by := cx+dy*(0.2+rng.Float64())+dx*(rng.Float64()-0.5), cy-dx*(0.2+rng.Float64())+dy*(rng.Float64()-0.5)

		fb.FillTriangle(cx-dx, cy-dy, 1, cx+dx, cy+dy, 1, ax, ay, 1, mark)
		fb.FillTriangle(cx+dx, cy+dy, 1, cx-dx, cy-dy, 1, bx, by, 1, mark)

		if fb.At(3, 4) != mark {
			t.Fatalf("trial %d: wanted the seam cell filled, got a gap (edge %v,%v to %v,%v)", trial, cx-dx, cy-dy, cx+dx, cy+dy)
		}
	}
}

// The tolerance is for rounding, not for growing the triangle: a cell whose centre is a
// full unit outside a long edge must stay empty.
func TestFillTriangleEdgeIsTight(t *testing.T) {
	fb, err := NewFrameBuffer(60, 60)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	mark := Cell{Character: '#'}

	// The long edge is x+y=100; cell centres (49.5,50.5) and (50.5,49.5) sit exactly on it.
	fb.FillTriangle(0, 0, 1, 100, 0, 1, 0, 100, 1, mark)

	for _, tc := range []struct {
		x, y int
		want bool
	}{
		{49, 49, true},  // x+y of the centre is 99, inside
		{49, 50, true},  // on the edge
		{50, 49, true},  // on the edge
		{50, 50, false}, // 101, outside
		{51, 49, false}, // 101, outside
	} {
		if got := fb.At(tc.x, tc.y) == mark; got != tc.want {
			t.Errorf("wanted cell (%d,%d) filled=%v, got %v", tc.x, tc.y, tc.want, got)
		}
	}
}
