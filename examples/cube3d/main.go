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
	if in.Pressed(engine.KeyEscape) {
		return engine.ErrQuit
	}

	const moveSpeed = 3.0
	const turnSpeed = 1.5

	move := geometry.Vector3{}
	if in.IsDown('w') || in.IsDown(engine.KeyUp) {
		move = move.Add(g.cam.Forward())
	}
	if in.IsDown('s') || in.IsDown(engine.KeyDown) {
		move = move.Sub(g.cam.Forward())
	}
	if in.IsDown('d') {
		move = move.Add(g.cam.Right())
	}
	if in.IsDown('a') {
		move = move.Sub(g.cam.Right())
	}
	// Normalized so moving diagonally is not ~1.41x faster.
	g.cam.Position = g.cam.Position.Add(move.Normalize().Scale(moveSpeed * dt))

	if in.IsDown('e') || in.IsDown(engine.KeyRight) {
		g.cam.Yaw += turnSpeed * dt
	}
	if in.IsDown('q') || in.IsDown(engine.KeyLeft) {
		g.cam.Yaw -= turnSpeed * dt
	}

	g.cube.Yaw += 1.0 * dt
	g.cube.Pitch += 0.6 * dt
	g.fps = 0.9*g.fps + 0.1*(1/dt)
	return nil
}

func (g *cubeGame) Draw(fb *engine.FrameBuffer) {
	g.cube.Draw(fb, g.cam, engine.Cell{Character: '#', Color: 208})

	// HUD after the 3D pass so the scene never draws over it.
	fb.DrawBox(0, 0, fb.Width(), fb.Height(), 0)
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
		FPS: 30,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
