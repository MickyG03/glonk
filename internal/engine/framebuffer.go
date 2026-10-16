package engine

import "errors"

type FrameBuffer struct {
	width, height int
	frame         []Cell
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
	if x >= fb.width || x < 0 || y >= fb.height || y < 0 {
		return
	}

	fb.frame[y*fb.width+x] = cell
}
