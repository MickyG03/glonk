# glonk

A terminal rendering engine for Go. Draw 2D, 3D, or both in the same frame, using plain terminal cells.

> Early and pre-1.0: the API may change.

## What it does

- **2D:** cells with 256-color, lines, boxes, text, and sprites.
- **3D:** wireframe or solid meshes through a perspective camera, with near-plane clipping. Solid meshes are flat-shaded by a directional light into an ASCII brightness ramp, hide each other by depth, and skip faces turned away.
- **Half blocks:** a pixel `Canvas` with two square-ish pixels per cell, drawn with `▀` and a background color, for twice the vertical detail and shading by color instead of by character.
- **Mixed:** draw a 3D scene first, then a 2D HUD on top, into the same frame.
- **Input:** held keys and one-shot presses, including arrows. Terminals with the [kitty keyboard protocol](https://sw.kovidgoyal.net/kitty/keyboard-protocol/) (Ghostty, kitty, WezTerm, foot, iTerm2, Alacritty) report real key releases, so you can hold several keys at once.
- **Resizing:** the frame follows the terminal window.

## Install

```
go get github.com/kyrokohan/glonk
```

Requires Go 1.27.2 or newer (see `go.mod`).

## Run the examples

```
make run                              # bounce2d: a ball in a box (Esc quits)
make run EXAMPLE=./examples/cube3d    # cube3d: walk around a floor with six shaded cubes
```

cube3d controls:

| Keys | Action |
|---|---|
| W / S, Up / Down | move forward / back |
| A / D | strafe left / right |
| Q / E, Left / Right | turn left / right |
| M | switch between character shading and half-block pixels |
| B | in half-block mode, switch between exact and solid edges |
| Esc | quit |

Ctrl-C quits any program built on glonk.

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyrokohan/glonk/engine"
)

type game struct{ x float64 }

func (g *game) Update(dt float64, in *engine.Input) error {
	if in.Pressed(engine.KeyEscape) {
		return engine.ErrQuit
	}
	g.x += 20 * dt // cells per second
	return nil
}

func (g *game) Draw(fb *engine.FrameBuffer) {
	fb.DrawBox(0, 0, fb.Width(), fb.Height(), 0)
	fb.Set(int(g.x)%fb.Width(), fb.Height()/2, engine.Cell{Character: '@', Color: 208})
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Width and Height of 0 fill the terminal and follow resizes.
	if err := engine.Run(ctx, &game{}, engine.Config{FPS: 30}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

`Run` owns the terminal. Each frame it calls `Update` (move things, read input) and then `Draw` (write cells into the `FrameBuffer`), then prints the frame. The terminal is restored when `Run` returns. Run `go doc github.com/kyrokohan/glonk/engine` for the full API.

## Packages

| Package | Purpose |
|---|---|
| `engine` | The loop, `FrameBuffer`, pixel `Canvas`, drawing helpers, keyboard `Input` |
| `geometry` | 2D and 3D vector math |
| `mesh` | 3D meshes (wireframe and shaded), lights, and the perspective `Camera` |
| `sprite` | Single-cell sprites drawn in layers |

## Compatibility and limits

Built to work in most terminals: it uses only the alternate screen, cursor movement, and 256-color escape codes. It is developed on macOS and tested mainly in Ghostty.

- **Without key releases, held keys are a timing guess.** Most terminals report presses and repeats but never releases, so a key counts as held until about 150 ms pass without a repeat, with a short pause when you first hold one. glonk detects the kitty keyboard protocol at startup and uses it where it exists; everywhere else this is what you get.
- **Without key releases, only one key stays held at a time.** The OS repeats just the most recently pressed key, so holding W and then pressing E stops W and you cannot walk and turn together (Terminal.app, for one). Set `Config.LegacyKeys` to force this behavior if a terminal misbehaves with the protocol.
- **Color is 256-color only.** `Cell.Color` 0 is the terminal default and 1 to 255 are palette indices.
- **Box drawing and half blocks need a Unicode font.** `DrawBox` uses `┌─┐│└┘` and the pixel canvas uses `▀▄`; nearly every monospace font has them. Character shading (`DrawShaded`) needs only ASCII and stays the default in the examples.
- **Apple's Terminal draws a faint line along half-block edges.** Its block characters fall a little short of the cell. glonk detects it (`TERM_PROGRAM=Apple_Terminal`) and `Canvas.Compose` then draws solid cells instead of half blocks, so there is no line but silhouettes are a pixel blockier. Set `Canvas.Seamless` yourself to override, for instance inside tmux, which hides that variable.
- **A fixed-size frame larger than the terminal garbles.** Leave `Width` and `Height` at 0 to fill the terminal.
- **Resizing can flicker for one frame.**
- **Windows is untested.**

## Planned

Ideas, not promises: loading OBJ models.

## License

MIT. See [LICENSE](LICENSE).
