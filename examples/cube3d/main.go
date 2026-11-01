package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
	"github.com/kyrokohan/glonk/mesh"
)

type cubeGame struct {
	cam  mesh.Camera
	cube mesh.Mesh
	fps  float64
}

func (g *cubeGame) Update(dt float64, in *engine.Input) error {
	if in.Pressed('q') {
		return engine.ErrQuit
	}
	g.cube.Yaw += 1.0 * dt
	g.cube.Pitch += 0.6 * dt
	g.fps = 0.9*g.fps + 0.1*(1/dt)
	return nil
}

func (g *cubeGame) Draw(fb *engine.FrameBuffer) {
	w := fb.Width() - 1
	h := fb.Height() - 1
	fb.DrawBox(0, 0, w, h, 0)

	g.cube.Draw(fb, g.cam, engine.Cell{Character: '#', Color: 208})
	fb.DrawText(1, 1, fmt.Sprintf("%.0f fps", g.fps), 0)
}

func main() {
	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer ctxCancel()

	game := &cubeGame{
		cam:  mesh.Camera{Position: geometry.Vector3{Z: -4}, FOV: math.Pi / 2, Near: 0.1},
		cube: mesh.Cube(2),
		fps:  30,
	}

	err := engine.Run(ctx, game, engine.Config{
		Width:  80,
		Height: 24,
		FPS:    30,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
