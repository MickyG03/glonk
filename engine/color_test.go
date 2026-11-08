package engine

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestPaletteRGB(t *testing.T) {
	tests := []struct {
		index uint8
		want  [3]int
	}{
		{0, [3]int{0, 0, 0}},
		{1, [3]int{205, 0, 0}},
		{9, [3]int{255, 0, 0}},
		{15, [3]int{255, 255, 255}},
		{16, [3]int{0, 0, 0}},
		{17, [3]int{0, 0, 95}},
		{21, [3]int{0, 0, 255}},
		{22, [3]int{0, 95, 0}},
		{46, [3]int{0, 255, 0}},
		{52, [3]int{95, 0, 0}},
		{196, [3]int{255, 0, 0}},
		{208, [3]int{255, 135, 0}},
		{226, [3]int{255, 255, 0}},
		{231, [3]int{255, 255, 255}},
		{232, [3]int{8, 8, 8}},
		{244, [3]int{128, 128, 128}},
		{255, [3]int{238, 238, 238}},
	}

	for _, tt := range tests {
		if got := paletteRGB[tt.index]; got != tt.want {
			t.Errorf("wanted color %d to be %v, got %v", tt.index, tt.want, got)
		}
	}
}

func TestShade(t *testing.T) {
	tests := []struct {
		name      string
		color     uint8
		intensity float64
		want      uint8
	}{
		{"full intensity keeps a cube color", 196, 1, 196},
		{"full intensity keeps orange", 208, 1, 208},
		{"full intensity keeps a gray", 240, 1, 240},
		{"half red lands on dark red", 196, 0.5, 88},
		{"half green lands on dark green", 46, 0.5, 28},
		{"half white lands on mid gray", 231, 0.5, 244},
		{"zero is black", 208, 0, 16},
		{"below zero clamps to black", 208, -4, 16},
		{"above one clamps to the color", 208, 4, 208},
		{"NaN is black", 208, math.NaN(), 16},
		{"negative infinity is black", 208, math.Inf(-1), 16},
		{"positive infinity is the color", 208, math.Inf(1), 208},
		{"color 0 is read as black", 0, 1, 16},
		{"a system color maps to the nearest cube color", 1, 1, 160},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Shade(Palette(tt.color), tt.intensity); got != Palette(tt.want) {
				t.Errorf("wanted %v, got %v", Palette(tt.want), got)
			}
		})
	}
}

func TestShadeOfTheDefaultColorIsReadAsBlack(t *testing.T) {
	for _, intensity := range []float64{0, 0.5, 1} {
		if got := Shade(Color{}, intensity); got != Palette(16) {
			t.Errorf("wanted black (palette 16) at intensity %v, got %v", intensity, got)
		}
	}
}

func TestShadeOfAnRGBColorScalesItExactly(t *testing.T) {
	tests := []struct {
		name      string
		color     Color
		intensity float64
		want      Color
	}{
		{"full intensity keeps it", RGB(200, 100, 50), 1, RGB(200, 100, 50)},
		{"half", RGB(200, 100, 50), 0.5, RGB(100, 50, 25)},
		{"zero is black", RGB(200, 100, 50), 0, RGB(0, 0, 0)},
		{"rounds to the nearest", RGB(101, 3, 255), 0.5, RGB(51, 2, 128)},
		{"clamps above one", RGB(200, 100, 50), 9, RGB(200, 100, 50)},
		{"clamps below zero", RGB(200, 100, 50), -9, RGB(0, 0, 0)},
		{"NaN is black", RGB(200, 100, 50), math.NaN(), RGB(0, 0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Shade(tt.color, tt.intensity); got != tt.want {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

func TestShadeEveryColor(t *testing.T) {
	// shaded is Shade of palette color c, as a palette index.
	shaded := func(c uint8, intensity float64) int {
		got := Shade(Palette(c), intensity)
		index, ok := got.paletteIndex()
		if !ok || got.kind() != colorPalette {
			t.Fatalf("color %d at %v: wanted a palette color, got %v", c, intensity, got)
		}
		return int(index)
	}
	brightest := func(index int) int {
		p := paletteRGB[index]
		return max(p[0], p[1], p[2])
	}

	for c := range 256 {
		color := uint8(c)

		if got := shaded(color, 0); got != 16 {
			t.Errorf("color %d: wanted black (16) at intensity 0, got %d", c, got)
		}

		// Cube and gray colors are exact; the theme-dependent ones snap to the nearest.
		if c >= 16 {
			if got := shaded(color, 1); got != c {
				t.Errorf("color %d: wanted itself at intensity 1, got %d", c, got)
			}
		}

		// Nearest-color matching can wobble between neighboring steps, but a quarter
		// of the range apart a brighter request must not give a dimmer color.
		for _, lo := range []float64{0, 0.25, 0.5, 0.75} {
			dim, bright := shaded(color, lo), shaded(color, lo+0.25)
			if brightest(dim) > brightest(bright) {
				t.Errorf("color %d: wanted %v to be no brighter than %v, got %d vs %d", c, lo, lo+0.25, dim, bright)
			}
		}

		for _, i := range []float64{0, 0.001, 0.3, 0.7, 1} {
			if got := shaded(color, i); got < 16 {
				t.Errorf("color %d at %v: wanted a color in 16..255, got %d", c, i, got)
			}
		}
	}
}

func TestColorZeroValueIsTheDefault(t *testing.T) {
	var zero Color
	if !zero.IsDefault() {
		t.Errorf("wanted the zero Color to be the default")
	}
	if zero != (Color{}) {
		t.Errorf("wanted Color{} and the zero value to be equal")
	}
	if Palette(0).IsDefault() {
		t.Errorf("wanted Palette(0), black, not to be the default")
	}
	if RGB(0, 0, 0).IsDefault() {
		t.Errorf("wanted RGB(0, 0, 0) not to be the default")
	}
}

func TestColorsCompareByWhatTheyAre(t *testing.T) {
	orange, sameOrange := Palette(208), Palette(208)
	if orange != sameOrange {
		t.Errorf("wanted equal palette colors to be equal")
	}
	if Palette(208) == Palette(209) {
		t.Errorf("wanted different palette colors to differ")
	}
	rgb, sameRGB := RGB(1, 2, 3), RGB(1, 2, 3)
	if rgb != sameRGB {
		t.Errorf("wanted equal RGB colors to be equal")
	}
	if RGB(1, 2, 3) == RGB(3, 2, 1) {
		t.Errorf("wanted RGB colors with swapped channels to differ")
	}
	// A palette color and the RGB color that looks like it are different colors: one
	// follows the user's theme (0 to 15), the other does not.
	if Palette(196) == RGB(255, 0, 0) {
		t.Errorf("wanted a palette color and an RGB color to differ")
	}
	if Palette(0) == (Color{}) {
		t.Errorf("wanted black to differ from the default")
	}
}

func TestColorString(t *testing.T) {
	tests := []struct {
		color Color
		want  string
	}{
		{Color{}, "default"},
		{Palette(0), "palette(0)"},
		{Palette(208), "palette(208)"},
		{Palette(255), "palette(255)"},
		{RGB(255, 135, 0), "#ff8700"},
		{RGB(0, 0, 0), "#000000"},
		{RGB(1, 2, 3), "#010203"},
	}
	for _, tt := range tests {
		if got := tt.color.String(); got != tt.want {
			t.Errorf("wanted %q, got %q", tt.want, got)
		}
	}
}

func TestColorRGB(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		want  [3]uint8
	}{
		{"default is read as black", Color{}, [3]uint8{0, 0, 0}},
		{"palette black", Palette(0), [3]uint8{0, 0, 0}},
		{"palette red", Palette(196), [3]uint8{255, 0, 0}},
		{"palette orange", Palette(208), [3]uint8{255, 135, 0}},
		{"palette gray", Palette(244), [3]uint8{128, 128, 128}},
		{"rgb as it is", RGB(12, 34, 56), [3]uint8{12, 34, 56}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b := tt.color.rgb()
			if got := [3]uint8{r, g, b}; got != tt.want {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

// nearestPaletteScan is what Shade used to do: try every palette color from 16 to 255
// and keep the closest, the lowest index on a tie. nearestPalette must agree with it.
func nearestPaletteScan(r, g, b float64) uint8 {
	best, bestDist := 16, math.Inf(1)
	for i := 16; i < 256; i++ {
		dr, dg, db := r-float64(paletteRGB[i][0]), g-float64(paletteRGB[i][1]), b-float64(paletteRGB[i][2])
		if dist := dr*dr + dg*dg + db*db; dist < bestDist {
			best, bestDist = i, dist
		}
	}
	return uint8(best)
}

func TestNearestPaletteFindsEveryPaletteColor(t *testing.T) {
	for i := 16; i < 256; i++ {
		p := paletteRGB[i]
		if got := nearestPalette(float64(p[0]), float64(p[1]), float64(p[2])); int(got) != i && paletteRGB[got] != p {
			t.Errorf("wanted palette color %d (%v) to find itself, got %d (%v)", i, p, got, paletteRGB[got])
		}
	}
}

// distance is the squared distance from (r, g, b) to palette color i.
func distance(r, g, b float64, i uint8) float64 {
	dr, dg, db := r-float64(paletteRGB[i][0]), g-float64(paletteRGB[i][1]), b-float64(paletteRGB[i][2])
	return dr*dr + dg*dg + db*db
}

// sameNearest reports whether got is as good an answer as want: the same color, or
// one just as near up to rounding error (an input that is a tie in exact arithmetic
// but not in floating point can go either way).
func sameNearest(r, g, b float64, got, want uint8) bool {
	if got == want {
		return true
	}
	dg, dw := distance(r, g, b, got), distance(r, g, b, want)
	return math.Abs(dg-dw) <= 1e-9*(1+dw)
}

// Where the arithmetic is exact (whole numbers and halves) the answer is the scan's
// answer to the letter, ties included: the lowest index.
func TestNearestPaletteMatchesTheScanExactly(t *testing.T) {
	check := func(r, g, b float64) {
		t.Helper()
		if got, want := nearestPalette(r, g, b), nearestPaletteScan(r, g, b); got != want {
			t.Fatalf("wanted palette %d, got %d, for (%v, %v, %v)", want, got, r, g, b)
		}
	}

	// A grid of whole numbers, including the ends.
	for r := 0; r <= 255; r += 5 {
		for g := 0; g <= 255; g += 5 {
			for b := 0; b <= 255; b += 5 {
				check(float64(r), float64(g), float64(b))
			}
		}
	}

	// Values on and next to the boundaries between cube levels, in every combination:
	// the ties and near-ties.
	edges := []float64{0, 47, 47.5, 48, 114.5, 115, 115.5, 154.5, 155, 155.5, 194.5, 195, 195.5, 234.5, 235, 235.5, 255}
	for _, r := range edges {
		for _, g := range edges {
			for _, b := range edges {
				check(r, g, b)
			}
		}
	}

	// Gray ties: whole-number channels whose sum puts the average exactly halfway
	// between two gray levels (sum - 24 is 15 more than a multiple of 30).
	for sum := 39; sum <= 765; sum += 30 {
		for r := 0; r <= 255; r += 17 {
			for g := 0; g <= 255; g += 17 {
				if b := sum - r - g; b >= 0 && b <= 255 {
					check(float64(r), float64(g), float64(b))
				}
			}
		}
	}
}

// Anywhere else the answer is a nearest color, as near as the scan's.
func TestNearestPaletteIsAsNearAsTheScan(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 11))
	for range 300000 {
		r, g, b := rng.Float64()*255, rng.Float64()*255, rng.Float64()*255
		if got, want := nearestPalette(r, g, b), nearestPaletteScan(r, g, b); !sameNearest(r, g, b, got, want) {
			t.Fatalf("wanted palette %d (or as near), got %d, for (%v, %v, %v)", want, got, r, g, b)
		}
	}
	for sum := 0.0; sum <= 765; sum += 0.5 {
		m := sum / 3
		for _, c := range [][3]float64{{m, m, m}, {m + 1, m - 1, m}, {m * 1.5, m * 0.5, m}} {
			r, g, b := min(max(c[0], 0), 255), min(max(c[1], 0), 255), min(max(c[2], 0), 255)
			if got, want := nearestPalette(r, g, b), nearestPaletteScan(r, g, b); !sameNearest(r, g, b, got, want) {
				t.Fatalf("wanted palette %d (or as near), got %d, for (%v, %v, %v)", want, got, r, g, b)
			}
		}
	}
}

func TestShadeMatchesTheScan(t *testing.T) {
	for c := range 256 {
		for _, intensity := range []float64{0, 0.1, 0.25, 0.5, 0.75, 0.9, 1} {
			base := paletteRGB[c]
			want := nearestPaletteScan(float64(base[0])*intensity, float64(base[1])*intensity, float64(base[2])*intensity)
			r, g, b := float64(base[0])*intensity, float64(base[1])*intensity, float64(base[2])*intensity
			got, _ := Shade(Palette(uint8(c)), intensity).paletteIndex()
			if !sameNearest(r, g, b, got, want) {
				t.Fatalf("wanted Shade(%d, %v) = %d (or as near), got %d", c, intensity, want, got)
			}
		}
	}
}

// A color equally near a cube color and a gray goes to the cube color, which has the
// lower index. (0, 0, 12) is 144 away from black (16) and from the first gray (232).
func TestNearestPaletteTieBetweenCubeAndGrayGoesToTheCube(t *testing.T) {
	for _, c := range [][3]float64{{0, 0, 12}, {0, 1, 11}, {0, 2, 10}, {0, 5, 7}, {12, 0, 0}, {6, 6, 0}} {
		got := nearestPalette(c[0], c[1], c[2])
		if got != 16 {
			t.Errorf("wanted the cube color 16 for %v, got %d", c, got)
		}
		if want := nearestPaletteScan(c[0], c[1], c[2]); got != want {
			t.Errorf("wanted what the scan picks (%d) for %v, got %d", want, c, got)
		}
	}
}
