package engine

import "math"

// Block characters for pixel pairs. The upper half block paints the top half of a
// cell in the character color and the bottom half in the background color.
const (
	upperHalfBlock = '▀'
	lowerHalfBlock = '▄'
)

// Canvas is a grid of pixels twice as tall as the FrameBuffer it belongs to: each
// terminal cell holds two stacked pixels, which Compose turns into half-block
// characters. Cells are about twice as tall as wide, so the pixels come out roughly
// square. A pixel is a palette color, with 0 meaning empty, so black must be 16.
//
// Get one with FrameBuffer.Canvas. It keeps its own depth per pixel, separate from
// the FrameBuffer's, so solids drawn on a Canvas hide each other but not the cells.
type Canvas struct {
	fb            *FrameBuffer
	width, height int
	pixels        []uint8
	depth         []float64
}

func newCanvas(fb *FrameBuffer) *Canvas {
	c := &Canvas{
		fb:     fb,
		width:  fb.width,
		height: fb.height * 2,
		pixels: make([]uint8, fb.width*fb.height*2),
		depth:  make([]float64, fb.width*fb.height*2),
	}
	c.Clear()
	return c
}

// Canvas returns the pixel layer of this FrameBuffer, creating it on first use.
// It is cleared whenever the FrameBuffer is.
func (fb *FrameBuffer) Canvas() *Canvas {
	if fb.canvas == nil {
		fb.canvas = newCanvas(fb)
	}
	return fb.canvas
}

// Width returns the number of pixel columns, the same as the FrameBuffer's.
func (c *Canvas) Width() int {
	return c.width
}

// Height returns the number of pixel rows, twice the FrameBuffer's.
func (c *Canvas) Height() int {
	return c.height
}

// Clear empties every pixel and resets every depth to infinitely far.
func (c *Canvas) Clear() {
	clear(c.pixels)
	for i := range c.depth {
		c.depth[i] = math.Inf(1)
	}
}

func (c *Canvas) inBounds(x, y int) bool {
	return x >= 0 && x < c.width && y >= 0 && y < c.height
}

// Set stores color at pixel (x, y). Writes outside the canvas are ignored.
func (c *Canvas) Set(x, y int, color uint8) {
	if !c.inBounds(x, y) {
		return
	}
	c.pixels[y*c.width+x] = color
}

// At returns the color at pixel (x, y), or 0 (empty) outside the canvas.
func (c *Canvas) At(x, y int) uint8 {
	if !c.inBounds(x, y) {
		return 0
	}
	return c.pixels[y*c.width+x]
}

// SetDepth stores color at pixel (x, y) only if z is nearer than what is already
// there, and reports whether it did. Smaller z is nearer; a tie keeps the existing
// pixel. Writes outside the canvas, and a NaN z, are ignored. Set never reads or
// writes depth.
func (c *Canvas) SetDepth(x, y int, z float64, color uint8) bool {
	if !c.inBounds(x, y) {
		return false
	}

	i := y*c.width + x
	// Written as !(z < ...) so a NaN fails the test instead of slipping through.
	if !(z < c.depth[i]) {
		return false
	}

	c.depth[i] = z
	c.pixels[i] = color
	return true
}

// FillTriangle fills the triangle with corners (x0, y0), (x1, y1), (x2, y2) in
// pixel coordinates, like FrameBuffer.FillTriangle but into pixels. The z values
// are the corners' distances from the camera and must be positive.
func (c *Canvas) FillTriangle(x0, y0, z0, x1, y1, z1, x2, y2, z2 float64, color uint8) {
	scanTriangle(c.width, c.height, x0, y0, z0, x1, y1, z1, x2, y2, z2, func(x, y int, z float64) {
		c.SetDepth(x, y, z, color)
	})
}

// Compose writes the pixels into the FrameBuffer's cells as half-block characters.
// Cells whose two pixels are both empty are left alone; every other cell is
// replaced. Call it at the point in drawing where the pixels should sit: anything
// drawn into the FrameBuffer afterwards lands on top of them.
func (c *Canvas) Compose() {
	fb := c.fb
	for row := range fb.height {
		for x := range fb.width {
			top := c.pixels[(2*row)*c.width+x]
			bottom := c.pixels[(2*row+1)*c.width+x]

			var cell Cell
			switch {
			case top == 0 && bottom == 0:
				continue
			case top == bottom:
				// A space on a background is solid however the font draws blocks.
				cell = Cell{Character: ' ', Background: top}
			case top != 0 && bottom != 0:
				cell = Cell{Character: upperHalfBlock, Color: top, Background: bottom}
			case top != 0:
				cell = Cell{Character: upperHalfBlock, Color: top}
			default:
				cell = Cell{Character: lowerHalfBlock, Color: bottom}
			}
			fb.frame[row*fb.width+x] = cell
		}
	}
}
