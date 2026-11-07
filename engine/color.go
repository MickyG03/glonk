package engine

import "math"

// paletteRGB is the red, green and blue of each of the 256 palette colors, as
// xterm defines them: 16 system colors, a 6x6x6 color cube, and 24 grays.
var paletteRGB = buildPalette()

func buildPalette() [256][3]int {
	var p [256][3]int

	// Colors 0 to 15 are whatever the user's theme says; these are xterm's defaults.
	system := [16][3]int{
		{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0},
		{0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229},
		{127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
		{92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
	}
	copy(p[:16], system[:])

	levels := [6]int{0, 95, 135, 175, 215, 255}
	for i := range 216 {
		p[16+i] = [3]int{levels[i/36], levels[(i/6)%6], levels[i%6]}
	}

	for i := range 24 {
		g := 8 + 10*i
		p[232+i] = [3]int{g, g, g}
	}
	return p
}

// Shade returns the palette color closest to color scaled to the given brightness:
// 1 is color as it is, 0 is black. Intensity outside 0 to 1 is clamped and NaN is
// treated as 0. Colors 0 to 15 vary with the user's theme, so they are read as
// xterm's defaults, and color 0 is read as black. The result is always in 16 to
// 255, never 0, so a shaded pixel stays visible; black comes out as 16.
func Shade(color uint8, intensity float64) uint8 {
	if math.IsNaN(intensity) {
		intensity = 0
	}
	intensity = min(max(intensity, 0), 1)

	base := paletteRGB[color]
	target := [3]float64{
		float64(base[0]) * intensity,
		float64(base[1]) * intensity,
		float64(base[2]) * intensity,
	}

	best, bestDist := 16, math.Inf(1)
	for i := 16; i < 256; i++ {
		var dist float64
		for c := range 3 {
			d := target[c] - float64(paletteRGB[i][c])
			dist += d * d
		}
		if dist < bestDist {
			best, bestDist = i, dist
		}
	}
	return uint8(best)
}
