package engine

import (
	"errors"
	"io"
)

// FrameBuffer is a width x height grid of Cells that a game draws into each
// frame; Run presents it to the terminal.
type FrameBuffer struct {
	width, height int
	frame         []Cell
}

func (fb *FrameBuffer) inBounds(x, y int) bool {
	return x >= 0 && x < fb.width && y >= 0 && y < fb.height
}

// NewFrameBuffer returns a blank FrameBuffer of w x h cells. Both must be positive.
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

// Clear resets every cell to BlankCell.
func (fb *FrameBuffer) Clear() {
	for i := range fb.frame {
		fb.frame[i] = BlankCell
	}
}

// Set stores cell at (x, y). Writes outside the buffer are ignored.
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

// Width returns the number of columns.
func (fb *FrameBuffer) Width() int {
	return fb.width
}

// Height returns the number of rows.
func (fb *FrameBuffer) Height() int {
	return fb.height
}

func (fb *FrameBuffer) present(w io.Writer) error {
	width := fb.width
	height := fb.height

	sb := newSequenceBuilder()

	for y := range height {
		sb.MoveTo(y+1, 1)
		cur := uint8(0)
		for x := range width {
			cell := fb.frame[y*width+x]

			if cell.Color != cur {
				if cell.Color == 0 {
					sb.ResetStyle()
				} else {
					sb.SetForeground(cell.Color)
				}
				cur = cell.Color
			}

			sb.Print(string(cell.Character))
		}

		if cur != 0 {
			sb.ResetStyle()
		}
	}

	return sb.Flush(w)
}
