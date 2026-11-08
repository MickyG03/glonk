// Command bounce2d is a ball bouncing around inside a box that fills the window,
// drawn as a sprite.
//
//	go run ./examples/bounce2d
//
// Esc quits.
package main

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/sprite"
)

// hint is the text on the bottom border.
const hint = " Esc quits "

type bounceGame struct {
	// width and height are the size of the frame last drawn into. Update is not told
	// the frame size, so it bounces inside the one Draw saw; after a resize the ball
	// can be one cell off for a frame.
	width, height int
	vx, vy        float64
	ball          *sprite.Sprite
	layer         *sprite.Layer
}

// newBounceGame returns the game main runs. Its size is only a guess for the first
// frame, which Draw replaces with the real one.
func newBounceGame() *bounceGame {
	game := &bounceGame{
		vx:     30,
		vy:     15,
		width:  64,
		height: 24,
		ball:   sprite.New(10, 10, 'o'),
		layer:  sprite.NewLayer(),
	}
	game.layer.Add(game.ball)
	return game
}

func (bg *bounceGame) Update(dt float64, in *engine.Input) error {
	if in.Pressed(engine.KeyEscape) {
		return engine.ErrQuit
	}
	bg.ball.X += bg.vx * dt
	bg.ball.Y += bg.vy * dt

	if bg.ball.X >= float64(bg.width-2) {
		bg.vx = -math.Abs(bg.vx)
		bg.ball.X = float64(bg.width - 2)
	}
	if bg.ball.X < 1 {
		bg.vx = math.Abs(bg.vx)
		bg.ball.X = 1
	}

	if bg.ball.Y >= float64(bg.height-2) {
		bg.vy = -math.Abs(bg.vy)
		bg.ball.Y = float64(bg.height - 2)
	}
	if bg.ball.Y < 1 {
		bg.vy = math.Abs(bg.vy)
		bg.ball.Y = 1
	}

	return nil
}

func (bg *bounceGame) Draw(fb *engine.FrameBuffer) {
	bg.width, bg.height = fb.Width(), fb.Height()

	fb.DrawBox(0, 0, fb.Width(), fb.Height(), engine.Color{})
	bg.layer.Draw(fb)

	// On the bottom border, where the ball never goes. Skipped when it would run into
	// the corners.
	if fb.Width() >= len(hint)+4 {
		fb.DrawText(2, fb.Height()-1, hint, engine.Color{})
	}
}

func main() {
	// Width and Height of 0 fill the terminal and follow resizes.
	err := engine.Run(context.Background(), newBounceGame(), engine.Config{FPS: 30})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
