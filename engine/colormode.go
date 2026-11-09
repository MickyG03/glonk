package engine

import "strings"

// ColorMode says how many colors Run may use.
type ColorMode uint8

// The color modes. The zero value lets Run decide.
const (
	// ColorsAuto uses true color when the environment says the terminal has it
	// (COLORTERM set to truecolor or 24bit, or a TERM that names true color), and the
	// 256-color palette otherwise. Without a terminal it means the palette, so that
	// what a headless Run writes never depends on the machine it runs on.
	ColorsAuto ColorMode = iota

	// Colors256 sends RGB colors as the nearest of the palette colors 16 to 255.
	Colors256

	// ColorsTrue sends RGB colors exactly, as 24-bit colors. Use it when the
	// terminal can show them but does not say so, as inside tmux or over ssh when
	// COLORTERM does not arrive. A terminal that cannot shows wrong colors.
	ColorsTrue
)

// detectTrueColor reports whether the environment says that the terminal can show
// 24-bit colors: COLORTERM is the standard way to say it, and some TERM names say it too.
func detectTrueColor(getenv func(string) string) bool {
	switch strings.ToLower(getenv("COLORTERM")) {
	case "truecolor", "24bit":
		return true
	}
	term := strings.ToLower(getenv("TERM"))
	return strings.Contains(term, "truecolor") || strings.Contains(term, "24bit") || strings.HasSuffix(term, "-direct")
}

// useTrueColor decides whether a run writes RGB colors exactly.
func (m ColorMode) useTrueColor(headless bool, getenv func(string) string) bool {
	switch m {
	case ColorsTrue:
		return true
	case ColorsAuto:
		return !headless && detectTrueColor(getenv)
	}
	return false
}
