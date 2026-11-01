package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/sprite"
)

type bounceGame struct {
	width, height int
	vx, vy        float64
	ball          *sprite.Sprite
	layer         *sprite.Layer
}

func (bg *bounceGame) Update(dt float64) {
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
}

func (bg *bounceGame) Draw(fb *engine.FrameBuffer) {
	w := fb.Width() - 1
	h := fb.Height() - 1
	fb.DrawBox(0, 0, w, h, 0)

	bg.layer.Draw(fb)
}

func main() {
	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer ctxCancel()

	game := &bounceGame{
		vx:     30,
		vy:     15,
		width:  64,
		height: 24,
		ball:   sprite.New(10, 10, 'o'),
		layer:  sprite.NewLayer(),
	}

	game.layer.Add(game.ball)

	err := engine.Run(ctx, game, engine.Config{
		Width:  game.width,
		Height: game.height,
		FPS:    30,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
