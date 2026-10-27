package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/sprite"
)

type bounceGame struct {
	ball  *sprite.Sprite
	layer *sprite.Layer
}

func (bg *bounceGame) Update(dt float64) {
	bg.ball.X += 10.0 * dt
}

func (bg *bounceGame) Draw(fb *engine.FrameBuffer) {
	bg.layer.Draw(fb)
}

func main() {
	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer ctxCancel()

	game := &bounceGame{
		ball:  sprite.New(10, 10, 'O'),
		layer: sprite.NewLayer(),
	}

	game.layer.Add(game.ball)

	err := engine.Run(ctx, game, engine.Config{
		Width:  64,
		Height: 24,
		FPS:    30,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
