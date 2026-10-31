package engine

import "errors"

type FrameBuffer struct {
	width, height int
	frame         []Cell
}

func (fb *FrameBuffer) inBounds(x, y int) bool {
	return x >= 0 && x < fb.width && y >= 0 && y < fb.height
}

func NewFrameBuffer(w, h int) (*FrameBuffer, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("width and height must be positive integers")
	}
	allocatedFrame := make([]Cell, w*h)
	fb := &FrameBuffer{
		width:  w,
		height: h,
		frame:  allocatedFrame,
	}
	fb.Clear()

	return fb, nil
}

func (fb *FrameBuffer) Clear() {
	for i := range fb.frame {
		fb.frame[i] = BlankCell
	}
}

func (fb *FrameBuffer) Set(x, y int, cell Cell) {
	if !fb.inBounds(x, y) {
		return
	}

	fb.frame[y*fb.width+x] = cell
}

// At returns the cell at a specific location.
// Returns BlankCell if x and/or y are out of bounds.
func (fb *FrameBuffer) At(x, y int) Cell {
	if !fb.inBounds(x, y) {
		return BlankCell
	}

	return fb.frame[y*fb.width+x]
}

func (fb *FrameBuffer) Width() int {
	return fb.width
}

func (fb *FrameBuffer) Height() int {
	return fb.height
}

func (fb *FrameBuffer) present() {
	width := fb.width
	height := fb.height

	sb := NewSequenceBuilder()

	for y := range height {
		sb.MoveTo(y+1, 1)
		for x := range width {
			sb.Print(string(fb.frame[y*width+x].Character))
		}
	}

	sb.Execute()
}
