package mesh

import "github.com/kyrokohan/glonk/geometry"

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
