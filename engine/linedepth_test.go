package engine

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

// depthRow returns the depth stored along row y of a buffer, one number for each column.
func depthRow(fb *FrameBuffer, y int) []float64 {
	return slices.Clone(fb.depth[y*fb.width : (y+1)*fb.width])
}

// The line is made of the cells DrawLine makes, whatever the ends: reversed, steep,
// diagonal, a point, partly or wholly off the buffer, and far beyond it.
func TestDrawLineDepthDrawsTheCellsDrawLineDoes(t *testing.T) {
	const w, h = 12, 7
	rng := rand.New(rand.NewPCG(5, 6))
	mark := Cell{Character: '#'}
	coord := func() int {
		switch rng.IntN(8) {
		case 0:
			return rng.IntN(2_000_000_000) - 1_000_000_000
		case 1:
			return math.MaxInt - rng.IntN(5)
		case 2:
			return math.MinInt + rng.IntN(5)
		}
		return rng.IntN(60) - 20
	}

	for range 3000 {
		x0, y0, x1, y1 := coord(), coord(), coord(), coord()
		z0, z1 := 0.1+rng.Float64()*50, 0.1+rng.Float64()*50
		plain, withDepth := newFB(t, w, h), newFB(t, w, h)

		plain.DrawLine(x0, y0, x1, y1, mark)
		withDepth.DrawLineDepth(x0, y0, z0, x1, y1, z1, mark)

		if !slices.Equal(plain.frame, withDepth.frame) {
			t.Fatalf("line (%d,%d)-(%d,%d): wanted the cells DrawLine draws\n%s\nwanted\n%s", x0, y0, x1, y1, withDepth.String(), plain.String())
		}
		for i, c := range withDepth.frame {
			if (c == mark) != (withDepth.depth[i] < math.Inf(1)) {
				t.Fatalf("line (%d,%d)-(%d,%d): wanted depth stored in exactly the cells drawn, cell %d differs", x0, y0, x1, y1, i)
			}
		}
	}
}

// With the same distance at both ends the depth is that distance all along; with
// different distances it is perspective-correct, which these values show: halfway along
// a line from 1 to 3 away, the point is 1.5 away, not 2, because it is 1/z that runs
// evenly. A depth that ran evenly itself would give 1.25, 1.5 and 1.75 cells in.
func TestDrawLineDepthIsPerspectiveCorrect(t *testing.T) {
	tests := []struct {
		name   string
		z0, z1 float64
		want   [5]float64 // the five cells of a line 4 steps long
	}{
		{"equal ends", 2, 2, [5]float64{2, 2, 2, 2, 2}},
		{"1 to 3", 1, 3, [5]float64{1, 1.2, 1.5, 2, 3}},
		{"3 to 1", 3, 1, [5]float64{3, 2, 1.5, 1.2, 1}},
		{"1 to 5", 1, 5, [5]float64{1, 1 / (0.75 + 0.25/5), 1 / (0.5 + 0.5/5), 1 / (0.25 + 0.75/5), 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := newFB(t, 5, 1)

			fb.DrawLineDepth(0, 0, tt.z0, 4, 0, tt.z1, Cell{Character: '#'})

			got := depthRow(fb, 0)
			for i, want := range tt.want {
				if math.Abs(got[i]-want) > 1e-12 {
					t.Errorf("cell %d: wanted depth %v, got %v (all: %v)", i, want, got[i], got)
				}
			}
		})
	}
}

// The end cells have exactly the depths given, not a rounding of them.
func TestDrawLineDepthEndsAreExact(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))
	for range 500 {
		x0, y0, x1, y1 := rng.IntN(12), rng.IntN(7), rng.IntN(12), rng.IntN(7)
		if x0 == x1 && y0 == y1 {
			continue // a point has the nearer end, tested on its own
		}
		z0, z1 := 0.01+rng.Float64()*100, 0.01+rng.Float64()*100
		fb := newFB(t, 12, 7)

		fb.DrawLineDepth(x0, y0, z0, x1, y1, z1, Cell{Character: '#'})

		if got := fb.depth[y0*12+x0]; got != z0 {
			t.Fatalf("line (%d,%d)-(%d,%d): wanted the first cell at exactly %v, got %v", x0, y0, x1, y1, z0, got)
		}
		if got := fb.depth[y1*12+x1]; got != z1 {
			t.Fatalf("line (%d,%d)-(%d,%d): wanted the last cell at exactly %v, got %v", x0, y0, x1, y1, z1, got)
		}
	}
}

// A line drawn from either end has the same depth in every cell. This is checked where
// the cells are the same both ways round: straight and exactly diagonal lines.
func TestDrawLineDepthDoesNotDependOnWhichEndItStartsAt(t *testing.T) {
	rng := rand.New(rand.NewPCG(9, 10))
	for range 300 {
		x0, y0 := rng.IntN(12), rng.IntN(7)
		dx, dy := rng.IntN(3)-1, rng.IntN(3)-1 // along a row, a column or a diagonal
		steps := rng.IntN(7)
		x1, y1 := x0+dx*steps, y0+dy*steps
		if x1 < 0 || x1 >= 12 || y1 < 0 || y1 >= 7 {
			continue
		}
		z0, z1 := 0.01+rng.Float64()*100, 0.01+rng.Float64()*100
		forward, backward := newFB(t, 12, 7), newFB(t, 12, 7)

		forward.DrawLineDepth(x0, y0, z0, x1, y1, z1, Cell{Character: '#'})
		backward.DrawLineDepth(x1, y1, z1, x0, y0, z0, Cell{Character: '#'})

		for i := range forward.depth {
			f, b := forward.depth[i], backward.depth[i]
			if math.IsInf(f, 1) != math.IsInf(b, 1) || (!math.IsInf(f, 1) && math.Abs(f-b) > 1e-12*max(f, b)) {
				t.Fatalf("line (%d,%d)-(%d,%d): cell %d is %v one way and %v the other", x0, y0, x1, y1, i, f, b)
			}
		}
	}
}

// The depth test is SetDepth's: strictly nearer wins, a tie loses, and a cell that
// loses is left as it was, depth and all.
func TestDrawLineDepthIsHiddenByNearerAndHidesFarther(t *testing.T) {
	fb := newFB(t, 5, 1)
	// The line's depths along its five cells are 1, 1.2, 1.5, 2, 3.
	for x, z := range []float64{1, 1.0999, 2, 2, 5} {
		fb.SetDepth(x, 0, z, Cell{Character: 'o'})
	}

	fb.DrawLineDepth(0, 0, 1, 4, 0, 3, Cell{Character: '#'})

	got := []rune{}
	for x := range 5 {
		got = append(got, fb.At(x, 0).Character)
	}
	// Cell 0 ties at 1 and keeps its own; cell 1 has something at 1.0999, nearer than 1.2;
	// cell 2 is nearer than 2; cell 3 ties at 2; cell 4 is nearer than 5.
	if want := []rune("oo#o#"); !slices.Equal(got, want) {
		t.Errorf("wanted %q, got %q", string(want), string(got))
	}
	wantDepth := []float64{1, 1.0999, 1.5, 2, 3}
	if d := depthRow(fb, 0); !slices.Equal(d, wantDepth) {
		t.Errorf("wanted depths %v, got %v", wantDepth, d)
	}
}

// Where a line passes through a triangle it goes from in front of it to behind it:
// seen from the camera it is cut off exactly there, and the cell where they are equally
// far away is the triangle's (a tie).
func TestDrawLineDepthGoesBehindASurfaceItPasses(t *testing.T) {
	fb := newFB(t, 5, 1)
	big := [6]float64{-100, -100, 300, -100, -100, 300}
	fb.FillTriangle(big[0], big[1], 2, big[2], big[3], 2, big[4], big[5], 2, Cell{Character: 'f'})

	// 1/z runs evenly from 1 to 1/3; it is 1/2 three quarters of the way along.
	fb.DrawLineDepth(0, 0, 1, 4, 0, 3, Cell{Character: '#'})

	got := []rune{}
	for x := range 5 {
		got = append(got, fb.At(x, 0).Character)
	}
	if want := []rune("###ff"); !slices.Equal(got, want) {
		t.Errorf("wanted %q (the line in front for 3 cells, then the surface), got %q", string(want), string(got))
	}
}

// The line is stored at the depth it has everywhere along its whole length, even when
// most of it is outside the buffer.
func TestDrawLineDepthKeepsTheDepthOfTheWholeLine(t *testing.T) {
	const w = 20
	fb := newFB(t, w, 5)

	// Cells x = 0 to 19 are steps 1000 to 1019 of a line 1020 steps long.
	fb.DrawLineDepth(-1000, 2, 1, 20, 2, 5, Cell{Character: '#'})

	for x := range w {
		tt := float64(1000+x) / 1020
		want := 1 / ((1-tt)/1 + tt/5)
		if got := fb.depth[2*w+x]; math.Abs(got-want) > 1e-9*want {
			t.Fatalf("cell %d: wanted depth %v, got %v", x, want, got)
		}
	}
	// And the same from the other end.
	back := newFB(t, w, 5)
	back.DrawLineDepth(20, 2, 5, -1000, 2, 1, Cell{Character: '#'})
	for x := range w {
		if math.Abs(back.depth[2*w+x]-fb.depth[2*w+x]) > 1e-9*fb.depth[2*w+x] {
			t.Fatalf("cell %d: wanted the same depth both ways round, got %v and %v", x, fb.depth[2*w+x], back.depth[2*w+x])
		}
	}
}

func TestDrawLineDepthOfAPointIsTheNearerEnd(t *testing.T) {
	for _, tt := range []struct{ z0, z1, want float64 }{{1, 5, 1}, {5, 1, 1}, {2, 2, 2}} {
		fb := newFB(t, 3, 3)

		fb.DrawLineDepth(1, 1, tt.z0, 1, 1, tt.z1, Cell{Character: '#'})

		if got := fb.depth[1*3+1]; got != tt.want || fb.At(1, 1).Character != '#' {
			t.Errorf("ends %v and %v: wanted a cell at depth %v, got %v", tt.z0, tt.z1, tt.want, got)
		}
	}
	off := newFB(t, 3, 3)
	off.DrawLineDepth(7, 7, 1, 7, 7, 2, Cell{Character: '#'})
	if !slices.Equal(off.frame, newFB(t, 3, 3).frame) {
		t.Errorf("wanted a point outside the buffer to draw nothing")
	}
}

func TestDrawLineDepthWithABadDepthDrawsNothing(t *testing.T) {
	nan, inf := math.NaN(), math.Inf(1)
	for _, z := range [][2]float64{{0, 1}, {1, 0}, {-1, 1}, {1, -1}, {nan, 1}, {1, nan}, {0, 0}, {-inf, 1}, {inf, inf}} {
		fb := newFB(t, 5, 1)

		fb.DrawLineDepth(0, 0, z[0], 4, 0, z[1], Cell{Character: '#'})

		if !slices.Equal(fb.frame, newFB(t, 5, 1).frame) {
			t.Errorf("depths %v: wanted nothing drawn, got %q", z, fb.String())
		}
		for _, d := range fb.depth {
			if !math.IsInf(d, 1) {
				t.Errorf("depths %v: wanted the depth buffer untouched, got %v", z, fb.depth)
				break
			}
		}
	}
}

// An end that is infinitely far is allowed: the cells before it have finite depth, and
// the cell at the end is infinitely far, which is not nearer than anything.
func TestDrawLineDepthToInfinityStopsShort(t *testing.T) {
	fb := newFB(t, 5, 1)

	fb.DrawLineDepth(0, 0, 1, 4, 0, math.Inf(1), Cell{Character: '#'})

	got := []rune{}
	for x := range 5 {
		got = append(got, fb.At(x, 0).Character)
	}
	if want := []rune("####" + string(BlankCell.Character)); !slices.Equal(got, want) {
		t.Errorf("wanted %q, got %q", string(want), string(got))
	}
	// 1/z runs from 1 to 0, so z is 1, 4/3, 2, 4: the distance doubles each step toward the end.
	for x, want := range []float64{1, 4.0 / 3, 2, 4} {
		if got := fb.depth[x]; math.Abs(got-want) > 1e-12 {
			t.Errorf("cell %d: wanted depth %v, got %v", x, want, got)
		}
	}
}

// Depths at the limits of a float cannot be added without overflow, but the answer is
// always between the ends, and every cell of the line is still drawn.
func TestDrawLineDepthAtTheLimitsOfAFloat(t *testing.T) {
	tiny, huge := 5e-324, math.MaxFloat64
	for _, z := range [][2]float64{{tiny, huge}, {huge, tiny}, {tiny, tiny}, {huge, huge}, {tiny, 1}, {1, huge}, {1e-300, 1e300}} {
		fb := newFB(t, 9, 1)

		fb.DrawLineDepth(0, 0, z[0], 8, 0, z[1], Cell{Character: '#'})

		lo, hi := min(z[0], z[1]), max(z[0], z[1])
		for x := range 9 {
			if fb.At(x, 0).Character != '#' {
				t.Fatalf("depths %v: wanted every cell drawn, cell %d is not", z, x)
			}
			if d := fb.depth[x]; !(d >= lo && d <= hi) {
				t.Fatalf("depths %v: wanted cell %d between the ends, got %v", z, x, d)
			}
		}
	}
}

// DrawLine itself is untouched by having a depth-aware sibling.
func TestDrawLineStillIgnoresDepth(t *testing.T) {
	fb := newFB(t, 5, 1)
	for x := range 5 {
		fb.SetDepth(x, 0, 1, Cell{Character: 'o'})
	}

	fb.DrawLine(0, 0, 4, 0, Cell{Character: '#'})

	if got := fb.String(); got != "#####" {
		t.Errorf("wanted DrawLine to draw over nearer cells, got %q", got)
	}
	for _, d := range depthRow(fb, 0) {
		if d != 1 {
			t.Errorf("wanted DrawLine to leave depth alone, got %v", depthRow(fb, 0))
			break
		}
	}
}

func BenchmarkDrawLine(b *testing.B) {
	fb, err := NewFrameBuffer(200, 60)
	if err != nil {
		b.Fatal(err)
	}
	cell := Cell{Character: '#'}

	b.Run("DrawLine", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			fb.Clear()
			fb.DrawLine(0, 0, 199, 59, cell)
			fb.DrawLine(10, 50, 190, 3, cell)
			fb.DrawLine(-300, 30, 500, 31, cell)
		}
	})
	b.Run("DrawLineDepth", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			fb.Clear()
			fb.DrawLineDepth(0, 0, 1, 199, 59, 9, cell)
			fb.DrawLineDepth(10, 50, 2, 190, 3, 6, cell)
			fb.DrawLineDepth(-300, 30, 1, 500, 31, 20, cell)
		}
	})
}
