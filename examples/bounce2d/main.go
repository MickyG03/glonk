// Command bounce2d is a ball bouncing around inside a box, drawn as a sprite.
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

type bounceGame struct {
	width, height int
	vx, vy        float64
	ball          *sprite.Sprite
	layer         *sprite.Layer
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
	fb.DrawBox(0, 0, fb.Width(), fb.Height(), 0)

	bg.layer.Draw(fb)
}

func main() {
	game := &bounceGame{
		vx:     30,
		vy:     15,
		width:  64,
		height: 24,
		ball:   sprite.New(10, 10, 'o'),
		layer:  sprite.NewLayer(),
	}

	game.layer.Add(game.ball)

	err := engine.Run(context.Background(), game, engine.Config{
		Width:  game.width,
		Height: game.height,
		FPS:    30,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
