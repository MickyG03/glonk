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

// rowText returns the characters of row y.
func rowText(fb *engine.FrameBuffer, y int) string {
	var sb strings.Builder
	for x := range fb.Width() {
		sb.WriteRune(fb.At(x, y).Character)
	}
	return sb.String()
}

func TestBallMovesByVelocityTimesTime(t *testing.T) {
	g := newBounceGame()
	if err := g.Update(0.1, &engine.Input{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.ball.X != 13 || g.ball.Y != 11.5 {
		t.Errorf("wanted the ball at (13, 11.5), got (%v, %v)", g.ball.X, g.ball.Y)
	}
}

func TestBallBouncesOffEachWall(t *testing.T) {
	// The default size is 64x24, so the ball lives in columns 1 to 62 and rows 1 to 22.
	tests := []struct {
		name           string
		x, y           float64
		vx, vy         float64
		wantX, wantY   float64
		wantVxPositive bool
		wantVyPositive bool
	}{
		{"right wall", 62.5, 10, 30, 15, 62, 10, false, true},
		{"left wall", 0.5, 10, -30, 15, 1, 10, true, true},
		{"bottom wall", 10, 22.5, 30, 15, 10, 22, true, false},
		{"top wall", 10, 0.5, 30, -15, 10, 1, true, true},
		{"inside moves freely", 10, 10, 30, 15, 10, 10, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newBounceGame()
			g.ball.X, g.ball.Y, g.vx, g.vy = tt.x, tt.y, tt.vx, tt.vy

			if err := g.Update(0, &engine.Input{}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if g.ball.X != tt.wantX || g.ball.Y != tt.wantY {
				t.Errorf("wanted the ball at (%v, %v), got (%v, %v)", tt.wantX, tt.wantY, g.ball.X, g.ball.Y)
			}
			if (g.vx > 0) != tt.wantVxPositive || (g.vy > 0) != tt.wantVyPositive {
				t.Errorf("wanted velocity signs (%v, %v), got (%v, %v)", tt.wantVxPositive, tt.wantVyPositive, g.vx, g.vy)
			}
		})
	}
}

// The box is the size of the window: the ball bounces off the walls of the frame
// that was last drawn, not off a fixed size.
func TestBallBouncesInsideTheFrameLastDrawn(t *testing.T) {
	g := newBounceGame()

	g.Draw(newFrame(t, 100, 30))
	g.ball.X, g.ball.Y, g.vx = 90, 10, 30
	if err := g.Update(0.01, &engine.Input{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.vx <= 0 {
		t.Errorf("wanted the ball to fly on past the old 64 columns in a 100 column frame, vx=%v", g.vx)
	}

	g.Draw(newFrame(t, 40, 12))
	if err := g.Update(0, &engine.Input{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.ball.X > 38 || g.ball.Y > 10 || g.ball.X < 1 || g.ball.Y < 1 {
		t.Errorf("wanted the ball back inside a 40x12 frame after it shrank, got (%v, %v)", g.ball.X, g.ball.Y)
	}
}

func TestDrawShowsTheBoxTheBallAndTheKey(t *testing.T) {
	g := newBounceGame()
	fb := newFrame(t, 40, 12)
	g.Draw(fb)

	if fb.At(0, 0).Character != '┌' || fb.At(39, 0).Character != '┐' || fb.At(0, 11).Character != '└' || fb.At(39, 11).Character != '┘' {
		t.Errorf("wanted a box around the whole frame, corners were %q %q %q %q",
			fb.At(0, 0).Character, fb.At(39, 0).Character, fb.At(0, 11).Character, fb.At(39, 11).Character)
	}
	if got := rowText(fb, 11); !strings.Contains(got, " Esc quits ") {
		t.Errorf("wanted the key hint on the bottom border, got %q", got)
	}
	if got := fb.At(int(g.ball.X), int(g.ball.Y)).Character; got != 'o' {
		t.Errorf("wanted the ball drawn, got %q", got)
	}
}

func TestDrawInTinyFramesKeepsTheCornersAndDoesNotPanic(t *testing.T) {
	for _, size := range [][2]int{{1, 1}, {2, 2}, {3, 3}, {5, 5}, {10, 4}, {14, 4}, {15, 4}} {
		fb := newFrame(t, size[0], size[1])
		g := newBounceGame()
		g.Draw(fb)
		if err := g.Update(0.1, &engine.Input{}); err != nil {
			t.Fatalf("%v: unexpected error: %v", size, err)
		}

		if size[0] >= 2 && size[1] >= 2 && fb.At(size[0]-1, size[1]-1).Character != '┘' {
			t.Errorf("%v: wanted the bottom right corner kept, got %q", size, fb.At(size[0]-1, size[1]-1).Character)
		}
	}
}
