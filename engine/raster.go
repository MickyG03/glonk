package engine

import "math"

func intAbs(a int) int {
	if a >= 0 {
		return a
	}
	return -a
}

// DrawLine draws a line from (x0, y0) to (x1, y1)
// using Bresenham's algorithm
func (fb *FrameBuffer) DrawLine(x0, y0, x1, y1 int, cell Cell) {
	dx := intAbs(x1 - x0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}

	dy := -intAbs(y1 - y0)
	sy := -1
	if y0 < y1 {
		sy = 1
	}

	e := dx + dy

	for {
		fb.Set(x0, y0, cell)

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * e
		if e2 >= dy {
			e += dy
			x0 += sx
		}
		if e2 <= dx {
			e += dx
			y0 += sy
		}
	}
}

// DrawText draws a text on the screen. Off-screen text is clipped.
func (fb *FrameBuffer) DrawText(x, y int, text string, color uint8) {
	col := x

	for _, r := range text {
		if col >= fb.width {
			break
		}
		fb.Set(col, y, Cell{Character: r, Color: color})
		col++
	}
}

// DrawBox draws a w x h border at (x, y);
// boxes smaller than 2x2 draw nothing
func (fb *FrameBuffer) DrawBox(x, y, w, h int, color uint8) {
	if w < 2 || h < 2 {
		return
	}
	x1 := x + w - 1
	y1 := y + h - 1

	fb.DrawLine(x, y, x1, y, Cell{Character: '─', Color: color})
	fb.DrawLine(x, y1, x1, y1, Cell{Character: '─', Color: color})
	fb.DrawLine(x, y, x, y1, Cell{Character: '│', Color: color})
	fb.DrawLine(x1, y, x1, y1, Cell{Character: '│', Color: color})
	fb.Set(x, y, Cell{Character: '┌', Color: color})
	fb.Set(x1, y, Cell{Character: '┐', Color: color})
	fb.Set(x, y1, Cell{Character: '└', Color: color})
	fb.Set(x1, y1, Cell{Character: '┘', Color: color})
}

// triangleEps lets a cell centre that sits exactly on an edge count as inside
// despite rounding, so triangles that share an edge leave no gaps between them.
const triangleEps = 1e-9

// FillTriangle fills the triangle with corners (x0, y0), (x1, y1), (x2, y2) in
// cell coordinates, drawing cell into every cell whose centre lies inside it.
// z0, z1 and z2 are the corners' distances from the camera and must be positive;
// each cell is drawn with SetDepth, so nearer surfaces win. Either winding works.
// A triangle with no area, or with a non-positive or NaN corner depth, draws nothing.
func (fb *FrameBuffer) FillTriangle(x0, y0, z0, x1, y1, z1, x2, y2, z2 float64, cell Cell) {
	scanTriangle(fb.width, fb.height, x0, y0, z0, x1, y1, z1, x2, y2, z2, func(x, y int, z float64) {
		fb.SetDepth(x, y, z, cell)
	})
}

// scanTriangle walks the w x h grid and calls plot for every cell whose centre
// lies inside the triangle, with the interpolated depth there. It does the
// geometry for every kind of target; the caller decides what a cell holds.
func scanTriangle(w, h int, x0, y0, z0, x1, y1, z1, x2, y2, z2 float64, plot func(x, y int, z float64)) {
	if !(z0 > 0 && z1 > 0 && z2 > 0) {
		return
	}

	// Twice the signed area. The sign depends on winding; dividing by it below
	// makes the weights positive inside the triangle either way.
	area := (x1-x0)*(y2-y0) - (x2-x0)*(y1-y0)
	if !(math.Abs(area) > 0) {
		return
	}

	// Clamp as floats before converting: corners can be astronomically far off
	// screen, and a float that big has no valid int.
	fw, fh := float64(w), float64(h)
	left := int(max(0, min(math.Floor(min(x0, x1, x2)), fw)))
	right := int(max(-1, min(math.Ceil(max(x0, x1, x2)), fw-1)))
	top := int(max(0, min(math.Floor(min(y0, y1, y2)), fh)))
	bottom := int(max(-1, min(math.Ceil(max(y0, y1, y2)), fh-1)))

	for y := top; y <= bottom; y++ {
		for x := left; x <= right; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5

			w0 := ((x1-px)*(y2-py) - (x2-px)*(y1-py)) / area
			w1 := ((x2-px)*(y0-py) - (x0-px)*(y2-py)) / area
			w2 := ((x0-px)*(y1-py) - (x1-px)*(y0-py)) / area
			if w0 < -triangleEps || w1 < -triangleEps || w2 < -triangleEps {
				continue
			}

			// 1/z is the quantity that varies linearly across the screen, so
			// interpolate it and invert; interpolating z itself would skew depth.
			invZ := w0/z0 + w1/z1 + w2/z2
			plot(x, y, 1/invZ)
		}
	}
}
