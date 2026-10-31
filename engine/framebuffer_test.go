package engine

import (
	"fmt"
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
