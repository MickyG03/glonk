package engine

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
