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
