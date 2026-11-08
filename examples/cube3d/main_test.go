package main

import (
	"strings"
	"testing"

	"github.com/kyrokohan/glonk/engine"
)

func newFrame(t *testing.T, w, h int) *engine.FrameBuffer {
	t.Helper()
	fb, err := engine.NewFrameBuffer(w, h)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	return fb
}

// row returns the characters of row y inside the border.
func row(fb *engine.FrameBuffer, y int) string {
	var sb strings.Builder
	for x := 1; x < fb.Width()-1; x++ {
		sb.WriteRune(fb.At(x, y).Character)
	}
	return sb.String()
}

// B flips the pixel edges once per press, and the setting then stays across frames.
func TestEdgeKeyFlipsTheEdgeStyleOncePerPress(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "") // the default style depends on the terminal
	g := newCubeGame()
	g.halfBlocks = true

	fb := newFrame(t, 80, 24)
	draw := func() string {
		fb.Clear()
		g.Draw(fb)
		return row(fb, 1)
	}

	if got := draw(); !strings.Contains(got, "exact edges") {
		t.Fatalf("wanted exact edges to start with, got %q", got)
	}

	g.flipEdges = true
	if got := draw(); !strings.Contains(got, "solid edges") {
		t.Errorf("wanted solid edges after a press, got %q", got)
	}
	if g.flipEdges {
		t.Errorf("wanted the press used up by the frame that applied it")
	}
	if got := draw(); !strings.Contains(got, "solid edges") {
		t.Errorf("wanted solid edges kept on the next frame, got %q", got)
	}

	g.flipEdges = true
	if got := draw(); !strings.Contains(got, "exact edges") {
		t.Errorf("wanted a second press to switch back, got %q", got)
	}
}

func TestSceneDrawsAndStartsWithGlyphs(t *testing.T) {
	g := newCubeGame()
	fb := newFrame(t, 80, 24)
	g.Draw(fb)

	if got := row(fb, 1); !strings.Contains(got, "glyphs") {
		t.Errorf("wanted glyph shading to start with, got %q", got)
	}
	drawn := 0
	for y := 3; y < fb.Height()-4; y++ {
		for x := 1; x < fb.Width()-1; x++ {
			if fb.At(x, y) != engine.BlankCell {
				drawn++
			}
		}
	}
	if drawn == 0 {
		t.Errorf("wanted the floor and cubes on screen, got an empty scene")
	}
}

// held returns a lookup that reports exactly the given keys as down.
func held(keys ...engine.Key) func(engine.Key) bool {
	set := map[engine.Key]bool{}
	for _, k := range keys {
		set[k] = true
	}
	return func(k engine.Key) bool { return set[k] }
}

func TestKeyDownTakesEitherCase(t *testing.T) {
	tests := []struct {
		name   string
		down   []engine.Key
		letter rune
		want   bool
	}{
		{"lower case", []engine.Key{'w'}, 'w', true},
		{"upper case", []engine.Key{'W'}, 'w', true},
		{"both cases", []engine.Key{'w', 'W'}, 'w', true},
		{"nothing down", nil, 'w', false},
		{"another letter", []engine.Key{'s'}, 'w', false},
		{"another letter in upper case", []engine.Key{'S'}, 'w', false},
		{"an arrow is not a letter", []engine.Key{engine.KeyUp}, 'w', false},
		{"q", []engine.Key{'Q'}, 'q', true},
		{"e", []engine.Key{'E'}, 'e', true},
		{"a digit has no other case", []engine.Key{'1'}, '1', true},
		{"a digit that is not down", []engine.Key{'2'}, '1', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyDown(held(tt.down...), tt.letter); got != tt.want {
				t.Errorf("wanted keyDown(%v, %q) = %v, got %v", tt.down, tt.letter, tt.want, got)
			}
		})
	}
}

func TestHelpListsEveryKeyAndFitsANarrowWindow(t *testing.T) {
	for _, width := range []int{80, 40, 30} {
		fb := newFrame(t, width, 24)
		newCubeGame().Draw(fb)

		first, second := row(fb, fb.Height()-3), row(fb, fb.Height()-2)
		if width >= 40 {
			for _, want := range []string{"WASD move", "Q/E turn", "arrows look"} {
				if !strings.Contains(first, want) {
					t.Errorf("width %d: wanted %q in %q", width, want, first)
				}
			}
		}
		for _, want := range []string{"M mode", "B edges", "Esc quit"} {
			if !strings.Contains(second, want) {
				t.Errorf("width %d: wanted %q in %q", width, want, second)
			}
		}
	}
}

// With no keys held nothing moves, however long the frame.
func TestUpdateWithoutKeysStandsStill(t *testing.T) {
	g := newCubeGame()
	before := g.cam

	if err := g.Update(0.1, &engine.Input{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.cam != before {
		t.Errorf("wanted the camera unmoved, got %+v from %+v", g.cam, before)
	}
}
