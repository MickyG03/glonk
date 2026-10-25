package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Spin this off into a background thread
	go func() {
		for range 50 {
			time.Sleep(100 * time.Millisecond) // Move every 100ms
			ball.X += 1
		}
	}()

	scene.Animate(ctx)
}
