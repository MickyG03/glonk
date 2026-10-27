package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyrokohan/glonk/internal/engine"
)

func main() {
	scene, err := engine.NewScene(64, 24, 30)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer ctxCancel()

	terminal := engine.NewTerminal()
	terminal.Setup()
	defer terminal.Teardown()

	scene.DrawBoundary()

	ball := engine.NewSceneObject(10, 10, 'O')
	scene.AddObject(ball)

	speed := 10.0
	scene.Animate(ctx, func(dt float64) {
		ball.X += speed * dt
	})
}
