package main

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

func loadSamples(t *testing.T) []model {
	t.Helper()
	models, err := loadModels(nil)
	if err != nil {
		t.Fatalf("unable to load the samples: %v", err)
	}
	return models
}

func TestSamplesLoad(t *testing.T) {
	models := loadSamples(t)

	want := []struct {
		name  string
		verts int
		faces int // triangles, after splitting polygons
	}{
		{"house.obj", 30, 44},
		{"icosphere.obj", 42, 80},
		{"torus.obj", 128, 256},
	}
	if len(models) != len(want) {
		t.Fatalf("wanted %d samples, got %d", len(want), len(models))
	}
	for i, w := range want {
		if models[i].name != w.name || models[i].verts != w.verts || models[i].faces != w.faces {
			t.Errorf("sample %d: wanted %s with %d vertices and %d triangles, got %s with %d and %d",
				i, w.name, w.verts, w.faces, models[i].name, models[i].verts, models[i].faces)
		}
	}
}

// Every sample must be a closed solid wound outward everywhere: positive volume, and
// each directed edge used once with its reverse used once by the neighbouring face.
func TestSamplesAreClosedAndWoundOutward(t *testing.T) {
	for _, m := range loadSamples(t) {
		t.Run(m.name, func(t *testing.T) {
			var volume float64
			directed := map[[2]int]int{}
			for _, f := range m.mesh.Faces {
				a, b, c := m.mesh.Vertices[f[0]], m.mesh.Vertices[f[1]], m.mesh.Vertices[f[2]]
				volume += a.Dot(b.Cross(c)) / 6
				for k := range 3 {
					directed[[2]int{f[k], f[(k+1)%3]}]++
				}
			}

			if volume <= 0 {
				t.Errorf("wanted a positive volume (outward winding), got %v", volume)
			}
			for e, n := range directed {
				if n != 1 || directed[[2]int{e[1], e[0]}] != 1 {
					t.Fatalf("wanted every edge shared by two faces wound opposite ways, edge %v used %d times, reverse %d", e, n, directed[[2]int{e[1], e[0]}])
				}
			}
		})
	}
}

func TestSamplesAreFitted(t *testing.T) {
	for _, m := range loadSamples(t) {
		lo, hi := m.mesh.Bounds()
		size, center := hi.Sub(lo), lo.Add(hi).Scale(0.5)
		longest := math.Max(size.X, math.Max(size.Y, size.Z))

		if math.Abs(longest-modelSize) > 1e-9 {
			t.Errorf("%s: wanted the longest side %v, got %v", m.name, float64(modelSize), longest)
		}
		if center.Length() > 1e-9 {
			t.Errorf("%s: wanted the model centered, got center %v", m.name, center)
		}
	}
}

func writeFile(t *testing.T, name, text string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatalf("unable to write %s: %v", p, err)
	}
	return p
}

func TestLoadModelsFromFiles(t *testing.T) {
	tri := writeFile(t, "tri.obj", "v 0 0 0\nv 4 0 0\nv 0 2 0\nf 1 2 3\n")
	quad := writeFile(t, "quad.obj", "v 0 0 0\nv 1 0 0\nv 1 1 0\nv 0 1 0\nf 1 2 3 4\n")

	models, err := loadModels([]string{tri, quad})
	if err != nil {
		t.Fatalf("unable to load: %v", err)
	}
	if len(models) != 2 || models[0].name != "tri.obj" || models[1].name != "quad.obj" {
		t.Fatalf("wanted tri.obj then quad.obj, got %+v", models)
	}
	if models[0].verts != 3 || models[0].faces != 1 || models[1].faces != 2 {
		t.Errorf("wanted counts 3/1 and 4/2, got %d/%d and %d/%d", models[0].verts, models[0].faces, models[1].verts, models[1].faces)
	}
	if _, size := extents(models[0]); math.Abs(size-modelSize) > 1e-9 {
		t.Errorf("wanted the file fitted to %v, got %v", float64(modelSize), size)
	}
}

func extents(m model) (center geometry.Vector3, longest float64) {
	lo, hi := m.mesh.Bounds()
	size := hi.Sub(lo)
	return lo.Add(hi).Scale(0.5), math.Max(size.X, math.Max(size.Y, size.Z))
}

func TestLoadModelsRelativePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "m.obj"), []byte("v 0 0 0\nv 1 0 0\nv 0 1 0\nf 1 2 3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(old)

	models, err := loadModels([]string{"m.obj"})
	if err != nil || len(models) != 1 {
		t.Fatalf("wanted a model from a relative path, got %v, %v", models, err)
	}
}

func TestLoadModelsErrorsNameTheFile(t *testing.T) {
	good := writeFile(t, "good.obj", "v 0 0 0\nv 1 0 0\nv 0 1 0\nf 1 2 3\n")
	bad := writeFile(t, "bad.obj", "v 0 0 0\nv 1 oops 0\n")
	empty := writeFile(t, "empty.obj", "# nothing\n")
	missing := filepath.Join(t.TempDir(), "missing.obj")

	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"missing file", []string{missing}, []string{"missing.obj"}},
		{"malformed file names the line", []string{bad}, []string{"bad.obj", "obj line 2", "oops"}},
		{"file with no geometry", []string{empty}, []string{"empty.obj", "no faces or lines"}},
		{"one bad file among good ones", []string{good, bad}, []string{"bad.obj", "obj line 2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models, err := loadModels(tt.args)
			if err == nil {
				t.Fatalf("wanted an error, got %d models", len(models))
			}
			for _, frag := range tt.want {
				if !strings.Contains(err.Error(), frag) {
					t.Errorf("wanted the message to contain %q, got %q", frag, err)
				}
			}
		})
	}
}

func drawAll(t *testing.T, v *viewer) *engine.FrameBuffer {
	t.Helper()
	t.Setenv("TERM_PROGRAM", "") // the pixel edge style default depends on the terminal
	fb, err := engine.NewFrameBuffer(80, 24)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	v.Draw(fb)
	return fb
}

// modelCells counts the cells inside the frame border that are not text on the HUD rows.
func modelCells(fb *engine.FrameBuffer) [][2]int {
	var cells [][2]int
	for y := 3; y < fb.Height()-4; y++ {
		for x := 1; x < fb.Width()-1; x++ {
			if fb.At(x, y) != engine.BlankCell {
				cells = append(cells, [2]int{x, y})
			}
		}
	}
	return cells
}

func TestEveryModelDrawsInEveryMode(t *testing.T) {
	modes := []struct {
		name  string
		apply func(v *viewer)
	}{
		{"characters", func(v *viewer) {}},
		{"characters two-sided", func(v *viewer) { v.doubleSided = true }},
		{"wireframe", func(v *viewer) { v.wire = true }},
		{"half-blocks", func(v *viewer) { v.halfBlocks = true }},
		{"half-blocks solid edges", func(v *viewer) { v.halfBlocks, v.edgesChosen, v.edgesSolid = true, true, true }},
		{"half-blocks two-sided", func(v *viewer) { v.halfBlocks, v.doubleSided = true, true }},
	}

	models := loadSamples(t)
	for i := range models {
		for _, mode := range modes {
			t.Run(models[i].name+" "+mode.name, func(t *testing.T) {
				v := newViewer(models)
				v.current = i
				mode.apply(v)

				fb := drawAll(t, v)

				if cells := modelCells(fb); len(cells) < 40 {
					t.Errorf("wanted the model on screen, got %d cells", len(cells))
				}
			})
		}
	}
}

func hudText(fb *engine.FrameBuffer, row int) string {
	var sb strings.Builder
	for x := 1; x < fb.Width()-1; x++ {
		sb.WriteRune(fb.At(x, row).Character)
	}
	return sb.String()
}

func TestHUDNamesTheModelAndMode(t *testing.T) {
	models := loadSamples(t)
	v := newViewer(models)
	v.current = 1

	fb := drawAll(t, v)
	if got := hudText(fb, 1); !strings.Contains(got, "icosphere.obj (2/3)") || !strings.Contains(got, "42 vertices, 80 faces") {
		t.Errorf("wanted the model and its counts on the first line, got %q", got)
	}
	if got := hudText(fb, 2); !strings.Contains(got, "characters") {
		t.Errorf("wanted the mode on the second line, got %q", got)
	}

	v.wire, v.doubleSided = true, true
	fb = drawAll(t, v)
	if got := hudText(fb, 2); !strings.Contains(got, "wireframe, two-sided") {
		t.Errorf("wanted wireframe, two-sided, got %q", got)
	}
}

// The house is lopsided on purpose: the chimney stands on its +X side. Seen from the
// front it must be on the right, and from behind on the left. A mirrored import would
// swap them.
func TestHouseIsNotMirrored(t *testing.T) {
	models := loadSamples(t)
	house := 0
	for i, m := range models {
		if m.name == "house.obj" {
			house = i
		}
	}

	chimneyColumn := func(yaw float64) float64 {
		v := newViewer(models)
		v.current = house
		v.pitch = 0
		v.yaw = yaw
		fb := drawAll(t, v)

		// The chimney is the highest thing, so the topmost row with anything in it.
		cells := modelCells(fb)
		top := cells[0][1]
		for _, c := range cells {
			top = min(top, c[1])
		}
		var sum, n float64
		for _, c := range cells {
			if c[1] == top {
				sum += float64(c[0])
				n++
			}
		}
		return sum / n
	}

	mid := 40.0
	if got := chimneyColumn(0); got <= mid {
		t.Errorf("wanted the chimney right of center from the front, got column %.1f", got)
	}
	if got := chimneyColumn(math.Pi); got >= mid {
		t.Errorf("wanted the chimney left of center from behind, got column %.1f", got)
	}
}

func TestViewerStartsWithAUsefulView(t *testing.T) {
	v := newViewer(loadSamples(t))
	cam := v.camera()

	// The camera is in front of the model (negative Z), looking at the origin.
	if cam.Position.Z >= 0 {
		t.Errorf("wanted the camera in front of the model, got %v", cam.Position)
	}
	toOrigin := cam.Position.Neg().Normalize()
	if got := cam.Look().Dot(toOrigin); math.Abs(got-1) > 1e-9 {
		t.Errorf("wanted the camera to look at the origin, got alignment %v", got)
	}
	if math.Abs(cam.Position.Length()-v.dist) > 1e-9 {
		t.Errorf("wanted the camera %v away, got %v", v.dist, cam.Position.Length())
	}
}

func TestColorAndDrawStyleReachTheScreen(t *testing.T) {
	models := loadSamples(t)

	characters := func(fb *engine.FrameBuffer) map[rune]bool {
		seen := map[rune]bool{}
		for _, c := range modelCells(fb) {
			seen[fb.At(c[0], c[1]).Character] = true
		}
		return seen
	}

	for i, want := range palette {
		v := newViewer(models)
		v.color = i

		fb := drawAll(t, v)
		for _, c := range modelCells(fb) {
			if got := fb.At(c[0], c[1]).Color; got != want {
				t.Fatalf("color %d: wanted palette color %d, got %d at %v", i, want, got, c)
			}
		}
	}

	solid := newViewer(models)
	if got := characters(drawAll(t, solid)); len(got) < 2 {
		t.Errorf("wanted shaded drawing to use several characters, got %v", got)
	}

	wire := newViewer(models)
	wire.wire = true
	if got := characters(drawAll(t, wire)); len(got) != 1 || !got['#'] {
		t.Errorf("wanted a wireframe made only of #, got %v", got)
	}
}

// A triangle wound away from the camera is invisible until two-sided drawing is on.
func TestTwoSidedShowsABackwardsModel(t *testing.T) {
	p := writeFile(t, "backwards.obj", "v 0 0 0\nv 0 1 0\nv 1 0 0\nf 1 2 3\n")
	models, err := loadModels([]string{p})
	if err != nil {
		t.Fatalf("unable to load: %v", err)
	}

	for _, halfBlocks := range []bool{false, true} {
		v := newViewer(models)
		v.pitch, v.yaw, v.halfBlocks = 0, 0, halfBlocks
		if cells := modelCells(drawAll(t, v)); len(cells) != 0 {
			t.Errorf("halfBlocks=%v: wanted the backwards triangle culled, got %d cells", halfBlocks, len(cells))
		}

		v.doubleSided = true
		if cells := modelCells(drawAll(t, v)); len(cells) < 10 {
			t.Errorf("halfBlocks=%v: wanted the backwards triangle shown two-sided, got %d cells", halfBlocks, len(cells))
		}
	}
}

func TestColorReachesHalfBlockPixels(t *testing.T) {
	models := loadSamples(t)

	shades := func(color int) map[uint8]bool {
		v := newViewer(models)
		v.halfBlocks, v.color = true, color
		fb := drawAll(t, v)

		seen := map[uint8]bool{}
		for _, c := range modelCells(fb) {
			cell := fb.At(c[0], c[1])
			seen[cell.Color] = true
			seen[cell.Background] = true
		}
		return seen
	}

	first, second := shades(0), shades(1)
	same := len(first) == len(second)
	for c := range first {
		same = same && second[c]
	}
	if same {
		t.Errorf("wanted different palette colors to give different pixels, both gave %v", first)
	}
}
