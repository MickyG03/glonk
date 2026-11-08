package mesh

import (
	"encoding/binary"
	"math"
	"slices"
	"strings"
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
		cube.Transform.Position = geometry.Vector3{X: mx, Y: my, Z: mz}
		cube.Transform.Yaw, cube.Transform.Pitch = myaw, mpitch
		light := sunAndFill(geometry.Vector3{X: 0.4, Y: -1, Z: 0.6}, 0.2)

		fb, err := engine.NewFrameBuffer(20, 8)
		if err != nil {
			t.Fatalf("unable to initialize frame buffer: %v", err)
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			cube.Draw(fb, cam, engine.Cell{Character: '#'})
			cube.withColor(engine.Palette(1)).DrawShaded(fb, cam, light)
			cube.DoubleSided = true
			cube.withColor(engine.Palette(1)).DrawShaded(fb, cam, light)
			cube.withColor(engine.Palette(1)).DrawShadedPixels(fb.Canvas(), cam, light)
			fb.Canvas().Compose()
		}()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatalf("wanted drawing to finish, but it is still running for camera %+v and cube at %v", cam, cube.Transform.Position)
		}
	})
}

// Drawing a mesh must finish promptly and not panic whatever its Transform holds, and a
// Transform that cannot be applied must leave the frame blank.
func FuzzDrawTransform(f *testing.F) {
	f.Add(0.0, 0.0, 5.0, 0.3, 0.2, 0.1, 1.0)
	f.Add(0.0, 0.0, 5.0, 0.0, 0.0, 0.0, 0.0)  // Scale unset
	f.Add(0.0, 0.0, 5.0, 0.0, 0.0, 0.0, -1.0) // negative scale
	f.Add(0.0, 0.0, 5.0, 0.0, 0.0, 0.0, 1e300)
	f.Add(0.0, 0.0, 5.0, 0.0, 0.0, 0.0, 1e-300)
	f.Add(0.0, 0.0, 5.0, 1e18, -1e18, 1e18, 1.0)
	f.Add(math.NaN(), 0.0, 5.0, 0.0, 0.0, 0.0, 1.0)
	f.Add(0.0, 0.0, 5.0, math.Inf(1), 0.0, 0.0, 1.0)
	f.Add(0.0, 0.0, 5.0, 0.0, 0.0, math.NaN(), 1.0)
	f.Add(0.0, 0.0, 5.0, 0.0, 0.0, 0.0, math.Inf(1))

	f.Fuzz(func(t *testing.T, px, py, pz, yaw, pitch, roll, scale float64) {
		cube := Cube(2)
		cube.Transform = Transform{Position: geometry.Vector3{X: px, Y: py, Z: pz}, Yaw: yaw, Pitch: pitch, Roll: roll, Scale: scale}
		cam := Camera{Position: geometry.Vector3{Z: -2}}
		light := sunAndFill(geometry.Vector3{X: 0.4, Y: -1, Z: 0.6}, 0.2)
		t.Setenv("TERM_PROGRAM", "") // Canvas.Seamless depends on the terminal running the tests

		fb, err := engine.NewFrameBuffer(20, 8)
		if err != nil {
			t.Fatalf("unable to initialize frame buffer: %v", err)
		}

		done := make(chan struct{})
		go func() {
			defer close(done)
			cube.Draw(fb, cam, engine.Cell{Character: '#'})
			cube.withColor(engine.Palette(1)).DrawShaded(fb, cam, light)
			cube.withColor(engine.Palette(1)).DrawShadedPixels(fb.Canvas(), cam, light)
			fb.Canvas().Compose()
		}()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatalf("wanted drawing to finish, but it is still running for %+v", cube.Transform)
		}

		if !cube.Transform.valid() && len(drawn(fb)) != 0 {
			t.Fatalf("wanted nothing drawn for the invalid %+v", cube.Transform)
		}
		if (cube.Validate() == nil) != cube.Transform.valid() {
			t.Fatalf("wanted Validate to agree with valid() for %+v", cube.Transform)
		}
	})
}

// Shading with any ramp must not panic, and every character it draws is one of the
// ramp's.
func FuzzMaterialRamp(f *testing.F) {
	f.Add("")
	f.Add("#")
	f.Add(".:-=+*#%@")
	f.Add(" ░▒▓█")
	f.Add("a\xffb")
	f.Add("\x1b[2J\x00\t")
	f.Add("中日")
	f.Add(strings.Repeat("ab", 5000))

	f.Fuzz(func(t *testing.T, ramp string) {
		cube := Cube(2)
		cube.Transform = Transform{Position: geometry.Vector3{Z: 5}, Yaw: 0.6, Pitch: -0.4}
		cube.Material = Material{Color: engine.Palette(1), Ramp: ramp}
		fb, err := engine.NewFrameBuffer(20, 8)
		if err != nil {
			t.Fatalf("unable to initialize frame buffer: %v", err)
		}

		cube.DrawShaded(fb, Camera{}, sunAndFill(geometry.Vector3{X: 0.4, Y: -1, Z: 0.6}, 0.2))

		allowed := cube.Material.ramp()
		for _, c := range drawn(fb) {
			if got := fb.At(c[0], c[1]).Character; !slices.Contains(allowed, got) {
				t.Fatalf("wanted a character of the ramp %q, got %q", ramp, got)
			}
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

func FuzzLoadOBJ(f *testing.F) {
	for _, seed := range []string{
		objCube,
		"v 0 0 0\nv 1 0 0\nv 0 1 0\nf 1 2 3\n",
		"v 0 0 0\nv 1 0 0\nv 1 1 0\nv 0 1 0\nf 1/1/1 2//2 3/3 4\nl 1 2 3\n",
		"v 0 0 0\nv 1 0 0\nv 0 1 0\nf -3 -2 -1\nv 1 1 1\nf -1 -2 -3\n",
		"v 0 0 0 1.0\nv 1 0 0 0.5 0.5 0.5\nv 0 1 0 # c\nvt 0 0\nvn 0 0 1\ng a\nusemtl m\nf 1 2 3 # c\n",
		"v 0 0 0\r\nv 1 0 0\r\nv 0 1 0\r\nf 1 2 \\\r\n 3\r\n",
		"v 0 0 0\nv 1 0 0\nl 1 2",
		"v 1e999 0 0\nf 1 1 1\n",
		"v nan 0 0\nv 0 inf 0\nv 0 0 -inf\nf 1 2 3\n",
		// Quadratic before the continuation fix: about 5 s then, milliseconds now.
		strings.Repeat("\\\n", 300000),
		"f 1 2 3\n", "f 0 0 0\n", "v 0 0 0\nf 99999999999999999999 1 1\n", "f \\\n\\\n", "\x00\xff\x1b[",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, text string) {
		var (
			m    Mesh
			err  error
			done = make(chan struct{})
		)
		go func() {
			defer close(done)
			m, err = LoadOBJ(strings.NewReader(text))
		}()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatalf("wanted loading to finish, but it is still running for a %d byte file", len(text))
		}

		if err != nil {
			if len(m.Vertices) != 0 || len(m.Faces) != 0 || len(m.Edges) != 0 {
				t.Fatalf("wanted an empty mesh alongside an error, got %d vertices, %d faces, %d edges", len(m.Vertices), len(m.Faces), len(m.Edges))
			}
			if len(text) > 0 {
				if _, limErr := LoadOBJLimit(strings.NewReader(text), int64(len(text))); limErr == nil || limErr.Error() != err.Error() {
					t.Fatalf("wanted LoadOBJLimit at the exact size to fail the same way, got %v vs %v", limErr, err)
				}
			}
			return
		}

		if len(m.Faces) == 0 && len(m.Edges) == 0 {
			t.Fatalf("wanted a successful load to have faces or lines")
		}
		for _, v := range m.Vertices {
			if !finite(v.X, v.Y, v.Z) {
				t.Fatalf("wanted finite vertices, got %v", v)
			}
		}
		for _, face := range m.Faces {
			for _, i := range face {
				if i < 0 || i >= len(m.Vertices) {
					t.Fatalf("wanted face indices in range of %d vertices, got %v", len(m.Vertices), face)
				}
			}
		}
		seen := map[[2]int]bool{}
		for _, e := range m.Edges {
			if e[0] < 0 || e[1] >= len(m.Vertices) || e[0] >= e[1] {
				t.Fatalf("wanted edges as sorted distinct in-range pairs of %d vertices, got %v", len(m.Vertices), e)
			}
			if seen[e] {
				t.Fatalf("wanted no duplicate edges, got %v twice", e)
			}
			seen[e] = true
		}

		// Loading is a pure function of the text.
		again, err := LoadOBJ(strings.NewReader(text))
		if err != nil || !slices.Equal(again.Vertices, m.Vertices) || !slices.Equal(again.Faces, m.Faces) || !slices.Equal(again.Edges, m.Edges) {
			t.Fatalf("wanted loading twice to give the same mesh")
		}

		// The limit variant agrees with the unlimited one at the exact size, and one
		// byte under it never succeeds.
		if exact, exactErr := LoadOBJLimit(strings.NewReader(text), int64(len(text))); exactErr != nil || !slices.Equal(exact.Vertices, m.Vertices) || !slices.Equal(exact.Faces, m.Faces) || !slices.Equal(exact.Edges, m.Edges) {
			t.Fatalf("wanted LoadOBJLimit at the exact size to match LoadOBJ, got err %v", exactErr)
		}
		if len(text) >= 2 {
			if _, underErr := LoadOBJLimit(strings.NewReader(text), int64(len(text)-1)); underErr == nil {
				t.Fatalf("wanted a limit one byte under the size to fail")
			}
		}

		// A loaded model must be safe to fit and draw.
		if len(m.Faces) > 2000 {
			return
		}
		fitted := m.Fit(2)
		fitted.DoubleSided = true
		fb, ferr := engine.NewFrameBuffer(20, 8)
		if ferr != nil {
			t.Fatalf("unable to initialize frame buffer: %v", ferr)
		}
		cam := Camera{Position: geometry.Vector3{Z: -5}, FOV: math.Pi / 2, Near: 0.1}
		fitted.Draw(fb, cam, engine.Cell{Character: '#'})
		fitted.withColor(engine.Palette(1)).DrawShaded(fb, cam, sunAndFill(geometry.Vector3{Y: -1}, 0.2))
		fitted.withColor(engine.Palette(1)).DrawShadedPixels(fb.Canvas(), cam, sunAndFill(geometry.Vector3{Y: -1}, 0.2))
	})
}

// A mesh whose edges and faces name any number at all, vertices or not, must draw
// without panicking.
func FuzzDrawBadMesh(f *testing.F) {
	f.Add(uint8(3), []byte{0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0})
	f.Add(uint8(0), []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	f.Add(uint8(4), []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0x80})
	f.Add(uint8(2), []byte{})

	f.Fuzz(func(t *testing.T, vertices uint8, data []byte) {
		// 15 integers: three edges (two each) and three faces (three each).
		var n [15]int
		for i := range n {
			if 8*i+8 <= len(data) {
				n[i] = int(int64(binary.LittleEndian.Uint64(data[8*i:])))
			}
		}
		m := Mesh{
			Edges: [][2]int{{n[0], n[1]}, {n[2], n[3]}, {n[4], n[5]}},
			Faces: [][3]int{{n[6], n[7], n[8]}, {n[9], n[10], n[11]}, {n[12], n[13], n[14]}},
		}
		for i := range int(vertices % 5) {
			m.Vertices = append(m.Vertices, geometry.Vector3{X: float64(i) - 1, Y: float64(i%2) - 0.5, Z: 4 + float64(i)})
		}

		fb, err := engine.NewFrameBuffer(20, 8)
		if err != nil {
			t.Fatalf("unable to initialize frame buffer: %v", err)
		}
		cam := Camera{FOV: math.Pi / 2, Near: 0.1}
		light := sunAndFill(geometry.Vector3{Y: -1}, 0.2)

		done := make(chan struct{})
		go func() {
			defer close(done)
			m.Draw(fb, cam, engine.Cell{Character: '#'})
			m.withColor(engine.Palette(1)).DrawShaded(fb, cam, light)
			m.DoubleSided = true
			m.withColor(engine.Palette(1)).DrawShaded(fb, cam, light)
			m.withColor(engine.Palette(1)).DrawShadedPixels(fb.Canvas(), cam, light)
			fb.Canvas().Compose()
		}()

		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatalf("wanted drawing to finish for %d vertices, edges %v and faces %v", len(m.Vertices), m.Edges, m.Faces)
		}
	})
}
