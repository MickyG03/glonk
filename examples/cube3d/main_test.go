package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
	"github.com/kyrokohan/glonk/mesh"
)

func newFrame(t *testing.T, w, h int) *engine.FrameBuffer {
	t.Helper()
	fb, err := engine.NewFrameBuffer(w, h)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	return fb
}

// row returns the text of row y.
func row(fb *engine.FrameBuffer, y int) string {
	return strings.Split(fb.String(), "\n")[y]
}

// B flips the pixel edges once per press, and the setting then stays across frames.
func TestEdgeKeyFlipsTheEdgeStyleOncePerPress(t *testing.T) {
	t.Setenv("TERM_PROGRAM", "") // the default style depends on the terminal
	g := newCubeGame()
	g.halfBlocks = true
	kb := engine.NewKeyboard()

	fb := newFrame(t, 80, 24)
	frame := func(pressB bool) string {
		if pressB {
			kb.Press('b')
		}
		if err := g.Update(1.0/30, kb.Input()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		kb.EndFrame()
		kb.Release('b')
		fb.Clear()
		g.Draw(fb)
		return row(fb, 1)
	}

	if got := frame(false); !strings.Contains(got, "exact edges") {
		t.Fatalf("wanted exact edges to start with, got %q", got)
	}
	if got := frame(true); !strings.Contains(got, "solid edges") {
		t.Errorf("wanted solid edges after a press, got %q", got)
	}
	if g.flipEdges {
		t.Errorf("wanted the press used up by the frame that applied it")
	}
	if got := frame(false); !strings.Contains(got, "solid edges") {
		t.Errorf("wanted solid edges kept on the next frame, got %q", got)
	}
	if got := frame(true); !strings.Contains(got, "exact edges") {
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
			kb := engine.NewKeyboard()
			kb.Press(tt.down...)
			if got := keyDown(kb.Input().IsDown, tt.letter); got != tt.want {
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

	if err := g.Update(0.1, engine.NewKeyboard().Input()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.cam != before {
		t.Errorf("wanted the camera unmoved, got %+v from %+v", g.cam, before)
	}
}

// frame holds keys for one Update of dt seconds and ends the frame, as Run does.
func frame(t *testing.T, g *cubeGame, kb *engine.Keyboard, dt float64) error {
	t.Helper()
	err := g.Update(dt, kb.Input())
	kb.EndFrame()
	return err
}

func TestWalkingMovesAlongWhereTheCameraFaces(t *testing.T) {
	const dt = 0.1
	const distance = 3.0 * dt // moveSpeed * dt

	for _, yaw := range []float64{0, 0.7, -2} {
		for _, tt := range []struct {
			name string
			keys []engine.Key
			dir  func(mesh.Camera) geometry.Vector3
		}{
			{"W forward", []engine.Key{'w'}, func(c mesh.Camera) geometry.Vector3 { return c.Forward() }},
			{"S back", []engine.Key{'s'}, func(c mesh.Camera) geometry.Vector3 { return c.Forward().Neg() }},
			{"D right", []engine.Key{'d'}, func(c mesh.Camera) geometry.Vector3 { return c.Right() }},
			{"A left", []engine.Key{'a'}, func(c mesh.Camera) geometry.Vector3 { return c.Right().Neg() }},
			{"capital W", []engine.Key{'W'}, func(c mesh.Camera) geometry.Vector3 { return c.Forward() }},
			{"W and D together go diagonally, no faster", []engine.Key{'w', 'd'}, func(c mesh.Camera) geometry.Vector3 { return c.Forward().Add(c.Right()).Normalize() }},
			{"W and S cancel", []engine.Key{'w', 's'}, func(mesh.Camera) geometry.Vector3 { return geometry.Vector3{} }},
		} {
			t.Run(fmt.Sprintf("%s at yaw %v", tt.name, yaw), func(t *testing.T) {
				g := newCubeGame()
				g.cam.Yaw = yaw
				start := g.cam.Position
				kb := engine.NewKeyboard()
				kb.Press(tt.keys...)

				if err := frame(t, g, kb, dt); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				want := start.Add(tt.dir(g.cam).Scale(distance))
				if got := g.cam.Position; got.Sub(want).Length() > 1e-9 {
					t.Errorf("wanted the camera at %v, got %v", want, got)
				}
			})
		}
	}
}

func TestWalkingStopsWhenTheKeyIsReleased(t *testing.T) {
	g := newCubeGame()
	kb := engine.NewKeyboard()
	kb.Press('w')
	frame(t, g, kb, 0.1)
	moved := g.cam.Position

	frame(t, g, kb, 0.1) // still held: keeps going
	if g.cam.Position == moved {
		t.Fatalf("wanted the camera to keep moving while W is held")
	}
	kb.Release('w')
	stopped := g.cam.Position
	frame(t, g, kb, 0.1)

	if g.cam.Position != stopped {
		t.Errorf("wanted the camera to stay put after W is released, got %v from %v", g.cam.Position, stopped)
	}
}

func TestWalkingStopsAtTheInvisibleWalls(t *testing.T) {
	for _, tt := range []struct {
		name  string
		key   engine.Key
		yaw   float64
		check func(geometry.Vector3) bool
	}{
		{"north", 'w', 0, func(p geometry.Vector3) bool { return p.Z == walkLimit }},
		{"south", 's', 0, func(p geometry.Vector3) bool { return p.Z == -walkLimit }},
		{"east", 'd', 0, func(p geometry.Vector3) bool { return p.X == walkLimit }},
		{"west", 'a', 0, func(p geometry.Vector3) bool { return p.X == -walkLimit }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := newCubeGame()
			kb := engine.NewKeyboard()
			kb.Press(tt.key)
			for range 200 {
				frame(t, g, kb, 0.25)
			}
			if !tt.check(g.cam.Position) {
				t.Errorf("wanted the camera stopped at the wall, got %v", g.cam.Position)
			}
		})
	}
}

func TestPitchDoesNotLiftTheWalker(t *testing.T) {
	g := newCubeGame()
	g.cam.Pitch = 1
	y := g.cam.Position.Y
	kb := engine.NewKeyboard()
	kb.Press('w')

	frame(t, g, kb, 0.1)

	if g.cam.Position.Y != y {
		t.Errorf("wanted walking to stay level however the camera looks, got y %v from %v", g.cam.Position.Y, y)
	}
}

func TestTurningAndLooking(t *testing.T) {
	const dt = 0.1
	tests := []struct {
		name       string
		keys       []engine.Key
		dYaw, dPit float64
	}{
		{"E turns right", []engine.Key{'e'}, 1.5 * dt, 0},
		{"capital E turns right", []engine.Key{'E'}, 1.5 * dt, 0},
		{"right arrow turns right", []engine.Key{engine.KeyRight}, 1.5 * dt, 0},
		{"Q turns left", []engine.Key{'q'}, -1.5 * dt, 0},
		{"left arrow turns left", []engine.Key{engine.KeyLeft}, -1.5 * dt, 0},
		{"Q and E cancel", []engine.Key{'q', 'e'}, 0, 0},
		{"up arrow looks up", []engine.Key{engine.KeyUp}, 0, 1.0 * dt},
		{"down arrow looks down", []engine.Key{engine.KeyDown}, 0, -1.0 * dt},
		{"up and down cancel", []engine.Key{engine.KeyUp, engine.KeyDown}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newCubeGame()
			kb := engine.NewKeyboard()
			kb.Press(tt.keys...)

			frame(t, g, kb, dt)

			if math.Abs(g.cam.Yaw-tt.dYaw) > 1e-9 || math.Abs(g.cam.Pitch-tt.dPit) > 1e-9 {
				t.Errorf("wanted yaw %v and pitch %v, got %v and %v", tt.dYaw, tt.dPit, g.cam.Yaw, g.cam.Pitch)
			}
		})
	}
}

func TestLookingStopsJustShortOfStraightUpOrDown(t *testing.T) {
	for _, tt := range []struct {
		key  engine.Key
		want float64
	}{{engine.KeyUp, 1.5}, {engine.KeyDown, -1.5}} {
		g := newCubeGame()
		kb := engine.NewKeyboard()
		kb.Press(tt.key)
		for range 100 {
			frame(t, g, kb, 0.1)
		}
		if g.cam.Pitch != tt.want {
			t.Errorf("key %v: wanted the pitch held at %v, got %v", tt.key, tt.want, g.cam.Pitch)
		}
	}
}

func TestEscapeQuits(t *testing.T) {
	g := newCubeGame()
	kb := engine.NewKeyboard()
	start := g.cam

	if err := frame(t, g, kb, 0.1); err != nil {
		t.Fatalf("wanted no error without Escape, got %v", err)
	}
	kb.Press(engine.KeyEscape)
	if err := frame(t, g, kb, 0.1); !errors.Is(err, engine.ErrQuit) {
		t.Errorf("wanted ErrQuit, got %v", err)
	}
	if g.cam != start {
		t.Errorf("wanted nothing to move on the quitting frame, got %+v", g.cam)
	}
}

// M is a press, not a hold: holding it down must not flicker the mode.
func TestModeKeySwitchesOncePerPress(t *testing.T) {
	for _, key := range []engine.Key{'m', 'M'} {
		g := newCubeGame()
		kb := engine.NewKeyboard()

		kb.Press(key)
		frame(t, g, kb, 0.1)
		if !g.halfBlocks {
			t.Fatalf("key %q: wanted half blocks after the first press", key)
		}
		for range 5 { // held down, no new press
			frame(t, g, kb, 0.1)
		}
		if !g.halfBlocks {
			t.Errorf("key %q: wanted holding the key to change nothing", key)
		}
		kb.Release(key)
		kb.Press(key)
		frame(t, g, kb, 0.1)
		if g.halfBlocks {
			t.Errorf("key %q: wanted a second press to switch back to glyphs", key)
		}
	}
}

func TestEdgeKeyDoesNothingInGlyphMode(t *testing.T) {
	g := newCubeGame()
	kb := engine.NewKeyboard()
	kb.Press('b')

	frame(t, g, kb, 0.1)

	if g.flipEdges {
		t.Errorf("wanted B ignored while drawing glyphs")
	}
}
