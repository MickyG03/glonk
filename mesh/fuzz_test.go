package mesh

import (
	"math"
	"testing"
	"time"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

func FuzzClipTriangleNear(f *testing.F) {
	f.Add(0.0, 0.0, -1.0, 2.0, 0.0, 3.0, 0.0, 2.0, 3.0, 0.1)
	f.Add(0.0, 0.0, 3.0, 2.0, 0.0, -1.0, 0.0, 2.0, -1.0, 0.1)
	f.Add(0.0, 0.0, 0.1, 2.0, 0.0, 0.1, 0.0, 2.0, -1.0, 0.1)
	f.Add(0.0, 0.0, 0.1, 2.0, 0.0, 0.1, 0.0, 2.0, 0.1, 0.1)
	f.Add(1e50, -1e50, 1e50, -1e50, 1e50, -1e50, 5.0, 5.0, 5.0, 1e-50)
	f.Add(math.NaN(), 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.1)

	f.Fuzz(func(t *testing.T, ax, ay, az, bx, by, bz, cx, cy, cz, near float64) {
		a, b, c := geometry.Vector3{X: ax, Y: ay, Z: az}, geometry.Vector3{X: bx, Y: by, Z: bz}, geometry.Vector3{X: cx, Y: cy, Z: cz}

		out := clipTriangleNear(a, b, c, near) // must never panic, whatever the inputs

		// The geometric promises only make sense for sane numbers.
		const huge = 1e100
		sane := finite(ax, ay, az, bx, by, bz, cx, cy, cz, near)
		for _, v := range []float64{ax, ay, az, bx, by, bz, cx, cy, cz, near} {
			sane = sane && math.Abs(v) < huge
		}
		if !sane {
			return
		}

		if len(out) > 2 {
			t.Fatalf("wanted at most 2 triangles, got %d", len(out))
		}

		origNormal := b.Sub(a).Cross(c.Sub(a))
		origArea := origNormal.Length() / 2

		var total float64
		for _, tri := range out {
			for _, p := range tri {
				if !finite(p.X, p.Y, p.Z) {
					t.Fatalf("wanted finite vertices, got %v", p)
				}
				if p.Z < near {
					t.Fatalf("wanted every vertex at Z>=%v, got %v", near, p)
				}
				// The cut piece lies inside the original triangle's bounding box.
				slack := 1e-9 * (1 + math.Abs(near) + math.Abs(az) + math.Abs(bz) + math.Abs(cz) + math.Abs(ax) + math.Abs(bx) + math.Abs(cx) + math.Abs(ay) + math.Abs(by) + math.Abs(cy))
				if p.X < min(ax, bx, cx)-slack || p.X > max(ax, bx, cx)+slack ||
					p.Y < min(ay, by, cy)-slack || p.Y > max(ay, by, cy)+slack ||
					p.Z < min(az, bz, cz)-slack || p.Z > max(az, bz, cz)+slack {
					t.Fatalf("wanted the piece inside the original's bounding box, got %v", p)
				}
			}
			total += tri[1].Sub(tri[0]).Cross(tri[2].Sub(tri[0])).Length() / 2
		}

		// Cross products of long, nearly parallel edges cancel almost everything, so
		// the error in an area is relative to the edge lengths squared, not to the area.
		edge := math.Max(b.Sub(a).Length(), math.Max(c.Sub(b).Length(), a.Sub(c).Length()))
		if total > origArea*(1+1e-6)+1e-9*(1+edge*edge) {
			t.Fatalf("wanted the clipped area %v to be no more than the original %v", total, origArea)
		}

		// A triangle wholly in front is returned as it was.
		if az >= near && bz >= near && cz >= near {
			if len(out) != 1 || out[0] != [3]geometry.Vector3{a, b, c} {
				t.Fatalf("wanted a triangle in front returned unchanged, got %v", out)
			}
		}
		// A triangle wholly behind is gone.
		if az < near && bz < near && cz < near && len(out) != 0 {
			t.Fatalf("wanted a triangle behind to vanish, got %v", out)
		}
	})
}

// Drawing must finish promptly and not panic whatever the camera, light and mesh
// numbers are, including nonsense ones such as a zero or negative Near.
func FuzzDraw(f *testing.F) {
	f.Add(0.0, 0.0, -4.0, 0.0, 0.0, 1.57, 0.1, 0.0, 0.0, 0.0, 2.0, 0.3, 0.2)
	f.Add(0.0, 0.0, 0.0, 0.0, 0.0, 1.57, 0.0, 0.0, 0.0, 0.0, 2.0, 0.0, 0.0)  // Near unset
	f.Add(0.0, 0.0, 0.0, 0.0, 0.0, 1.57, -1.0, 0.0, 0.0, 0.0, 2.0, 0.0, 0.0) // negative Near
	f.Add(0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.1, 0.0, 0.0, 5.0, 2.0, 0.0, 0.0)   // zero FOV
	f.Add(0.0, 0.0, 0.0, 0.0, 0.0, math.Pi, 0.1, 0.0, 0.0, 5.0, 2.0, 0.0, 0.0)
	f.Add(0.0, 0.0, 0.0, 0.0, 1.0, 1.57, 1e-12, 0.5, 0.0, 1e-9, 2.0, 0.0, 0.0) // grazing the camera
	f.Add(math.NaN(), 0.0, 0.0, 0.0, 0.0, 1.57, 0.1, 0.0, 0.0, 5.0, 2.0, 0.0, 0.0)
	f.Add(0.0, 0.0, 0.0, 0.0, 0.0, 1.57, 0.1, 0.0, 0.0, 5.0, 1e300, 0.0, 0.0)

	f.Fuzz(func(t *testing.T, cx, cy, cz, yaw, pitch, fov, near, mx, my, mz, size, myaw, mpitch float64) {
		cam := Camera{Position: geometry.Vector3{X: cx, Y: cy, Z: cz}, Yaw: yaw, Pitch: pitch, FOV: fov, Near: near}
		cube := Cube(size)
		cube.Position = geometry.Vector3{X: mx, Y: my, Z: mz}
		cube.Yaw, cube.Pitch = myaw, mpitch
		light := Light{Direction: geometry.Vector3{X: 0.4, Y: -1, Z: 0.6}, Ambient: 0.2}

		fb, err := engine.NewFrameBuffer(20, 8)
		if err != nil {
			t.Fatalf("unable to initialize frame buffer: %v", err)
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			cube.Draw(fb, cam, engine.Cell{Character: '#'})
			cube.DrawShaded(fb, cam, light, 1)
			cube.DrawShadedPixels(fb.Canvas(), cam, light, 1)
			fb.Canvas().Compose()
		}()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatalf("wanted drawing to finish, but it is still running for camera %+v and cube at %v", cam, cube.Position)
		}
	})
}

func FuzzClipSegment(f *testing.F) {
	f.Add(2.0, 3.0, 8.0, 7.0, 0.0, 0.0, 10.0, 10.0)
	f.Add(-10.0, 5.0, 20.0, 5.0, 0.0, 0.0, 10.0, 10.0)
	f.Add(-5.0, 5.0, 5.0, -5.0, 0.0, 0.0, 10.0, 10.0)
	f.Add(-1e100, 5.0, 1e100, 5.0, 0.0, 0.0, 10.0, 10.0)
	f.Add(math.NaN(), 0.0, 1.0, 1.0, 0.0, 0.0, 10.0, 10.0)
	f.Add(0.0, 0.0, 5.0, 5.0, 10.0, 10.0, 0.0, 0.0) // inverted rectangle

	f.Fuzz(func(t *testing.T, x0, y0, x1, y1, minX, minY, maxX, maxY float64) {
		cx0, cy0, cx1, cy1, ok := clipSegment(x0, y0, x1, y1, minX, minY, maxX, maxY) // never panics
		if !ok {
			return
		}

		sane := finite(x0, y0, x1, y1, minX, minY, maxX, maxY) && minX <= maxX && minY <= maxY
		for _, v := range []float64{minX, minY, maxX, maxY} {
			sane = sane && math.Abs(v) < 1e100
		}
		if !sane {
			return
		}

		for _, v := range []float64{cx0, cy0, cx1, cy1} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("wanted finite ends, got (%v,%v)-(%v,%v)", cx0, cy0, cx1, cy1)
			}
		}

		// Float error scales with the numbers involved.
		scale := math.Max(math.Max(math.Abs(x0), math.Abs(y0)), math.Max(math.Abs(x1), math.Abs(y1)))
		scale = math.Max(scale, math.Max(math.Max(math.Abs(minX), math.Abs(minY)), math.Max(math.Abs(maxX), math.Abs(maxY))))
		tol := 1e-9 + 1e-12*scale

		if cx0 < minX-tol || cx0 > maxX+tol || cx1 < minX-tol || cx1 > maxX+tol ||
			cy0 < minY-tol || cy0 > maxY+tol || cy1 < minY-tol || cy1 > maxY+tol {
			t.Fatalf("wanted ends inside [%v,%v]x[%v,%v], got (%v,%v)-(%v,%v)", minX, maxX, minY, maxY, cx0, cy0, cx1, cy1)
		}

		// The piece keeps the segment's direction: it does not flip or leave the line.
		dx, dy := x1-x0, y1-y0
		if (cx1-cx0)*dx+(cy1-cy0)*dy < -tol*(1+math.Abs(dx)+math.Abs(dy)) {
			t.Fatalf("wanted the piece to run the same way as the segment, got (%v,%v)-(%v,%v) for (%v,%v)-(%v,%v)", cx0, cy0, cx1, cy1, x0, y0, x1, y1)
		}
		cross := (cx0-x0)*dy - (cy0-y0)*dx
		if math.Abs(cross) > 1e-6*(1+math.Abs(dx)+math.Abs(dy))*(1+scale) {
			t.Fatalf("wanted the piece on the original line, off by %v", cross)
		}
	})
}
