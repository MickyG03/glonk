package engine

import (
	"math"
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
			if got := Shade(tt.color, tt.intensity); got != tt.want {
				t.Errorf("wanted %d, got %d", tt.want, got)
			}
		})
	}
}

func TestShadeEveryColor(t *testing.T) {
	brightest := func(c uint8) int {
		p := paletteRGB[c]
		return max(p[0], p[1], p[2])
	}

	for c := range 256 {
		color := uint8(c)

		if got := Shade(color, 0); got != 16 {
			t.Errorf("color %d: wanted black (16) at intensity 0, got %d", c, got)
		}

		// Cube and gray colors are exact; the theme-dependent ones snap to the nearest.
		if c >= 16 {
			if got := Shade(color, 1); got != color {
				t.Errorf("color %d: wanted itself at intensity 1, got %d", c, got)
			}
		}

		// Nearest-color matching can wobble between neighboring steps, but a quarter
		// of the range apart a brighter request must not give a dimmer color.
		for _, lo := range []float64{0, 0.25, 0.5, 0.75} {
			dim, bright := Shade(color, lo), Shade(color, lo+0.25)
			if brightest(dim) > brightest(bright) {
				t.Errorf("color %d: wanted %v to be no brighter than %v, got %d vs %d", c, lo, lo+0.25, dim, bright)
			}
		}

		for _, i := range []float64{0, 0.001, 0.3, 0.7, 1} {
			if Shade(color, i) < 16 {
				t.Errorf("color %d at %v: wanted a color in 16..255, got %d", c, i, Shade(color, i))
			}
		}
	}
}
