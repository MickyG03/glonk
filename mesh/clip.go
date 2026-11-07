package mesh

import (
	"math"

	"github.com/kyrokohan/glonk/geometry"
)

// clipTriangleNear cuts the view-space triangle a, b, c against the plane z = near
// and returns the part in front of it as zero, one or two triangles. Winding is
// preserved, and every returned vertex has Z >= near exactly.
func clipTriangleNear(a, b, c geometry.Vector3, near float64) [][3]geometry.Vector3 {
	in := func(v geometry.Vector3) bool { return v.Z >= near }

	// Common cases first: the triangle is wholly in front or wholly behind.
	switch {
	case in(a) && in(b) && in(c):
		return [][3]geometry.Vector3{{a, b, c}}
	case !in(a) && !in(b) && !in(c):
		return nil
	}

	// Sutherland-Hodgman against one plane: walk the edges, keep the vertices in
	// front, and add the crossing point wherever an edge changes sides.
	tri := [3]geometry.Vector3{a, b, c}
	var poly []geometry.Vector3
	for i, cur := range tri {
		next := tri[(i+1)%3]

		if in(cur) {
			poly = append(poly, cur)
		}
		switch {
		case in(cur) && !in(next):
			poly = append(poly, clipNear(next, cur, near))
		case !in(cur) && in(next):
			poly = append(poly, clipNear(cur, next, near))
		}
	}

	// A vertex lying exactly on the plane yields its own crossing point as well;
	// drop the repeats, wrap-around included.
	var clean []geometry.Vector3
	for i, v := range poly {
		if v != poly[(i+len(poly)-1)%len(poly)] {
			clean = append(clean, v)
		}
	}

	// Fan the polygon (three or four corners) into triangles.
	var out [][3]geometry.Vector3
	for i := 1; i+1 < len(clean); i++ {
		out = append(out, [3]geometry.Vector3{clean[0], clean[i], clean[i+1]})
	}
	return out
}

// clipSegment cuts the segment from (x0, y0) to (x1, y1) to the rectangle with
// corners (minX, minY) and (maxX, maxY), keeping its direction, and reports whether
// any of it is inside (Liang-Barsky). The ends it returns lie on the original
// segment. Non-finite input, and coordinates beyond +-1e150 whose differences could
// overflow, are rejected. Past about 1e15 the visible piece is smaller than a float
// can resolve, so it may come back as a point; it is always inside the rectangle.
func clipSegment(x0, y0, x1, y1, minX, minY, maxX, maxY float64) (cx0, cy0, cx1, cy1 float64, ok bool) {
	const limit = 1e150
	for _, v := range [...]float64{x0, y0, x1, y1} {
		if math.IsNaN(v) || math.Abs(v) > limit {
			return 0, 0, 0, 0, false
		}
	}

	dx, dy := x1-x0, y1-y0
	// Each border contributes p (how fast the segment moves toward it) and q (how
	// far the start is from it); the visible part is where all four allow it.
	p := [4]float64{-dx, dx, -dy, dy}
	q := [4]float64{x0 - minX, maxX - x0, y0 - minY, maxY - y0}

	t0, t1 := 0.0, 1.0
	for i := range p {
		if p[i] == 0 {
			if q[i] < 0 {
				return 0, 0, 0, 0, false // parallel to this border and outside it
			}
			continue
		}

		t := q[i] / p[i]
		if p[i] < 0 {
			if t > t1 {
				return 0, 0, 0, 0, false
			}
			t0 = max(t0, t)
		} else {
			if t < t0 {
				return 0, 0, 0, 0, false
			}
			t1 = min(t1, t)
		}
	}

	return x0 + t0*dx, y0 + t0*dy, x0 + t1*dx, y0 + t1*dy, true
}
