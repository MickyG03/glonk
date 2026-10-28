package engine

import (
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
