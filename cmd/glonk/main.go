package main

import (
	"fmt"
	"time"

	"github.com/kyrokohan/glonk/internal/engine"
)

func main() {
	scene, err := engine.NewScene(64, 24)
	if err != nil {
		fmt.Println(fmt.Errorf("%s", err.Error()))
		return
	}

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

	scene.Animate(30)
}
