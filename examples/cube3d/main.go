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

const (
	gridSize      = 40
	gridDivisions = 20 // 2x2 floor cells, so a size-2 cube fills one
	floorY        = -1.5
	walkLimit     = gridSize/2 - 0.5 // invisible walls, half a unit inside the floor edge
)

// sun lights the cubes from above, to one side and slightly ahead of the start.
var sun = mesh.Light{Direction: geometry.Vector3{X: 0.4, Y: -1, Z: 0.6}, Ambient: 0.2}

type object struct {
	mesh  mesh.Mesh
	color uint8
}

type cubeGame struct {
	cam   mesh.Camera
	floor mesh.Mesh
	cubes []object
	fps   float64

	// halfBlocks draws the cubes as pixels, two per cell, shaded by color. Off, they
	// are drawn one per cell, shaded by character. M toggles it.
	halfBlocks bool

	// B flips the edge style of the pixels. The library picks a default for the
	// terminal; once B is pressed we keep our own choice, and reapply it when a
	// resize gives us a fresh canvas.
	edgesPending bool // B was pressed and the next half-block frame should flip
	edgesChosen  bool
	edgesSolid   bool
}

// newCube returns a cube of the given size resting on the floor at (x, z).
func newCube(size, x, z float64, color uint8) object {
	m := mesh.Cube(size)
	m.Position = geometry.Vector3{X: x, Y: floorY + size/2, Z: z}
	return object{mesh: m, color: color}
}

func (g *cubeGame) Update(dt float64, in *engine.Input) error {
	if in.Pressed(engine.KeyEscape) {
		return engine.ErrQuit
	}
	// Without release events shift+m arrives as 'M'.
	if in.Pressed('m') || in.Pressed('M') {
		g.halfBlocks = !g.halfBlocks
	}
	if g.halfBlocks && (in.Pressed('b') || in.Pressed('B')) {
		g.edgesPending = true
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
	g.cam.Position.X = max(-walkLimit, min(walkLimit, g.cam.Position.X))
	g.cam.Position.Z = max(-walkLimit, min(walkLimit, g.cam.Position.Z))

	if in.IsDown('e') || in.IsDown(engine.KeyRight) {
		g.cam.Yaw += turnSpeed * dt
	}
	if in.IsDown('q') || in.IsDown(engine.KeyLeft) {
		g.cam.Yaw -= turnSpeed * dt
	}

	g.fps = 0.9*g.fps + 0.1*(1/dt)
	return nil
}

func (g *cubeGame) Draw(fb *engine.FrameBuffer) {
	// The floor is wireframe and ignores depth, so draw it first; the cubes then
	// cover it where they stand, and hide each other by depth.
	g.floor.Draw(fb, g.cam, engine.Cell{Character: '.', Color: 240})

	mode := "glyphs"
	if g.halfBlocks {
		canvas := fb.Canvas()
		switch {
		case g.edgesPending:
			canvas.Seamless = !canvas.Seamless
			g.edgesSolid, g.edgesChosen, g.edgesPending = canvas.Seamless, true, false
		case g.edgesChosen:
			canvas.Seamless = g.edgesSolid
		}
		mode = "half-blocks, exact edges"
		if canvas.Seamless {
			mode = "half-blocks, solid edges"
		}
		for _, o := range g.cubes {
			o.mesh.DrawShadedPixels(canvas, g.cam, sun, o.color)
		}
		// Compose before the HUD so the HUD lands on top of the pixels.
		canvas.Compose()
	} else {
		for _, o := range g.cubes {
			o.mesh.DrawShaded(fb, g.cam, sun, o.color)
		}
	}

	fb.DrawBox(0, 0, fb.Width(), fb.Height(), 0)
	fb.DrawText(1, 1, fmt.Sprintf("%.0f fps  %s", g.fps, mode), 0)
	fb.DrawText(1, fb.Height()-2, "WASD move  QE turn  M mode  B edges  Esc quit", 0)
}

func main() {
	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer ctxCancel()

	floor := mesh.Grid(gridSize, gridDivisions)
	floor.Position.Y = floorY

	game := &cubeGame{
		cam:   mesh.Camera{Position: geometry.Vector3{Z: -4}, FOV: math.Pi / 2, Near: 0.1},
		floor: floor,
		cubes: []object{
			newCube(2, 0, 6, 208),
			newCube(1, -4, 4, 39),
			newCube(3, 6, 10, 118),
			newCube(1.5, -8, 12, 201),
			newCube(2, 10, -4, 39),
			newCube(1, -6, -8, 118),
		},
		fps: 30,
	}

	err := engine.Run(ctx, game, engine.Config{
		FPS: 30,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
