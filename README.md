# glonk

A terminal rendering engine for Go. Draw 2D, 3D, or both in the same frame, using plain terminal cells.

> Early and pre-1.0: the API may change.

## What it does

- **2D:** cells with foreground and background colors (the 256-color palette or RGB) and text attributes (bold, dim, underline, reverse), lines, boxes, text, and sprites.
- **3D:** wireframe or solid meshes through a perspective camera, with near-plane clipping. Solid meshes are flat-shaded by a directional light into an ASCII brightness ramp, hide each other by depth, and skip faces turned away.
- **Models:** load OBJ files from any `io.Reader` (a file, `//go:embed`, a download), center and scale them, and draw them wireframe or shaded.
- **Half blocks:** a pixel `Canvas` with two square-ish pixels per cell, drawn with `▀` and a background color, for twice the vertical detail and shading by color instead of by character.
- **Mixed:** draw a 3D scene first, then a 2D HUD on top, into the same frame.
- **Input:** held keys and one-shot presses, including arrows. Terminals with the [kitty keyboard protocol](https://sw.kovidgoyal.net/kitty/keyboard-protocol/) (Ghostty, kitty, WezTerm, foot, iTerm2, Alacritty) report real key releases, so you can hold several keys at once.
- **Resizing:** the frame follows the terminal window, and a game with a `Resize(width, height int)` method is told its size.
- **Testing:** `Config.Input` and `Config.Output` take any reader and writer. With ones that are not a terminal, `Run` runs headless at a size you give it, so a game can be driven with key bytes and its frames read back.

## Install

```
go get github.com/kyrokohan/glonk
```

Requires Go 1.26 or newer.

## Run the examples

```
make run                                      # bounce2d: a ball in a box that fills the window (Esc quits)
make run EXAMPLE=./examples/cube3d            # cube3d: walk around a floor with six shaded cubes
make run EXAMPLE=./examples/viewer            # viewer: look at the bundled OBJ models
make run EXAMPLE=./examples/viewer ARGS=my.obj  # viewer: look at your own model
```

cube3d controls:

| Keys | Action |
|---|---|
| W / S | move forward / back |
| A / D | strafe left / right |
| Q / E, Left / Right | turn left / right |
| Up / Down | look up / down |
| M | switch between character shading and half-block pixels |
| B | in half-block mode, switch between exact and solid edges |
| Esc | quit |

viewer controls:

| Keys | Action |
|---|---|
| Arrows | orbit the model |
| + / - or W / S | zoom |
| R | spin on and off |
| N | next model |
| V | wireframe or solid |
| M, B | character shading or half-block pixels; exact or solid pixel edges |
| D | draw both sides of every face (for models that look inside out) |
| C | change color |
| Esc | quit |

Letters work in either case. Ctrl-C quits any program built on glonk.

## Loading models

This sits inside your `Draw` method, with `fb`, `cam` and `light` set up as in the examples; `ExampleLoadOBJ` in `mesh/example_test.go` is a complete, runnable version.

```go
//go:embed ship.obj
var shipOBJ []byte

ship, err := mesh.LoadOBJ(bytes.NewReader(shipOBJ))
if err != nil {
	log.Fatal(err) // names the line that went wrong, e.g. "obj line 12: bad number"
}
ship = ship.Fit(2) // centered on the origin, longest side 2
ship.Transform.Position = geometry.Vector3{Z: 6}

ship.DrawShaded(fb, cam, light, engine.Palette(208)) // or ship.Draw for a wireframe
```

`LoadOBJ` reads any `io.Reader`, so a file, an embedded asset or a download all work. For input you do not control, use `LoadOBJLimit(r, maxBytes)`: it fails with `mesh.ErrOBJTooLarge` instead of reading without end (a plain `io.LimitReader` would quietly cut the model short).

- **Understood:** vertices (`v`), faces (`f`, with or without texture and normal indices, any number of corners, negative indices) and lines (`l`). Faces with more than three corners are split into triangles for shading, and the wireframe shows the original polygons.
- **Skipped:** materials and colors, texture coordinates, normals, groups and curves. Everything in the file becomes one mesh, flat shaded, in the color you pass to `DrawShaded`.
- **Orientation:** OBJ files are right-handed and look down -Z; glonk looks down +Z. `LoadOBJ` flips Z and reverses each face so a model keeps its outside and is not mirrored left to right. Its front faces a camera placed on the -Z side.
- **Inside out or invisible?** Some exporters wind faces the other way. Set `Mesh.DoubleSided = true` to draw both sides.
- **Untrusted files:** malformed input returns an error naming the line, never a panic, and the time taken grows in step with the input size. Memory use is roughly 25 times the input size, so bound the input with `LoadOBJLimit` when you do not trust the source.

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"os"

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
	fb.DrawBox(0, 0, fb.Width(), fb.Height(), engine.Color{}) // the default color
	fb.Set(int(g.x)%fb.Width(), fb.Height()/2, engine.Cell{Character: '@', Foreground: engine.Palette(208)})
}

func main() {
	// Width and Height of 0 fill the terminal and follow resizes.
	if err := engine.Run(context.Background(), &game{}, engine.Config{FPS: 30}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

`Run` owns the terminal. Each frame it calls `Update` (move things, read input) and then `Draw` (write cells into the `FrameBuffer`), then prints the frame. The terminal is restored when `Run` returns, and Ctrl-C, `kill` and a closing window (SIGINT, SIGTERM, SIGHUP) end it cleanly (`Run` returns `nil`), so the game needs no signal handler. When your `context` ends it, `Run` returns the context's error. The full API is in each package's documentation, with runnable examples: `go doc github.com/kyrokohan/glonk/engine`, or the package pages on pkg.go.dev once the module is published.

## Packages

| Package | Purpose |
|---|---|
| `engine` | The loop, `FrameBuffer`, pixel `Canvas`, drawing helpers, keyboard `Input` |
| `geometry` | 2D and 3D vector math |
| `mesh` | 3D meshes (wireframe and shaded), OBJ loading, lights, and the perspective `Camera` |
| `sprite` | Single-cell sprites drawn in layers |

## Compatibility and limits

Built to work in most terminals. The basics need only the alternate screen, cursor movement, and 256-color escape codes; line wrapping is turned off while a program runs, and the kitty keyboard protocol and focus reporting are used where the terminal has them. All of it is undone when `Run` returns. It is developed on macOS and tested mainly in Ghostty.

- **Without key releases, held keys are a timing guess.** Most terminals report presses and repeats but never releases, so a key counts as held until about 150 ms pass without a repeat, with a short pause when you first hold one. glonk detects the kitty keyboard protocol at startup and uses it where it exists; everywhere else this is what you get.
- **Without key releases, only one key stays held at a time.** The OS repeats just the most recently pressed key, so holding W and then pressing E stops W and you cannot walk and turn together (Terminal.app, for one). Set `Config.LegacyKeys` to force this behavior if a terminal misbehaves with the protocol.
- **Escape waits 50 ms, and Alt chords are ignored.** An Escape press and the start of an arrow key are the same byte, so the Escape key counts only after 50 ms with nothing following it. Alt+key arrives as Escape plus the key and is dropped rather than read as Escape.
- **Ctrl-Z suspends, and some endings cannot be cleaned up.** Ctrl-Z restores the terminal, stops the program, and takes the terminal over again on `fg`; Windows has no suspend. `os.Exit`, `log.Fatal`, a panic in a goroutine you started, `kill -9` and `kill -QUIT` skip the cleanup and leave the terminal raw; type `reset` to repair it. After the first quit signal a second one kills a game that is stuck, unless your program catches those signals itself.
- **Bad meshes and cameras draw nothing instead of failing.** Edges and faces that name a missing vertex are skipped (`Mesh.Validate` finds them), a `Camera` field left at 0 uses a default (a 90 degree FOV, a Near of 0.1), a nonsense FOV or Near sees nothing, and sprites at NaN or infinite positions are not drawn.
- **One character is one column.** A cell holds a single character one column wide. Wide characters (CJK, most emoji) are drawn as `?`, combining marks and joiners are skipped by `DrawText`, and control characters, escape sequences in text included, are drawn as `?` too, so nothing a file name or a game contains can send commands to the terminal. A NUL character is shown as a space, so `Cell{Background: engine.Palette(4)}` is a colored blank. Real wide-character support is planned with the redesign of `Cell`.
- **RGB colors are shown as the nearest palette color for now.** A `Color` is the terminal default (the zero value), one of the 256 `Palette` colors, or an `RGB` color, and `Attrs` adds bold, dim, underline and reverse. Output is still 256-color, so an RGB color comes out as the closest of palette colors 16 to 255; true color output is planned. Palette colors 0 to 15 follow the user's terminal theme.
- **Box drawing and half blocks need a Unicode font.** `DrawBox` uses `┌─┐│└┘` and the pixel canvas uses `▀▄`; nearly every monospace font has them. Character shading (`DrawShaded`) needs only ASCII and stays the default in the examples.
- **Apple's Terminal draws a faint line along half-block edges.** Its block characters fall a little short of the cell. glonk detects it (`TERM_PROGRAM=Apple_Terminal`) and `Canvas.Compose` then draws solid cells instead of half blocks, so there is no line but silhouettes are a pixel blockier. Set `Canvas.Seamless` yourself to override, for instance inside tmux, which hides that variable.
- **A fixed-size frame larger than the terminal is cut off.** Only the top-left part that fits is shown, and the screen is cleared when the window changes size. Leave `Width` and `Height` at 0 to fill the terminal.
- **`FPS` runs from 1 to 1000;** `Run` returns an error outside that range. A key pressed during start-up, while glonk asks the terminal about the keyboard protocol, arrives as a tap rather than a hold.
- **Without a terminal there is no terminal handling.** A headless `Run` (see `Config.Output`) treats a key as held from its press to its release, as the kitty keyboard protocol reports them, unless `LegacyKeys` is set; a plain letter with no release then stays held. Ctrl-Z does nothing and quit signals are left to your program. Reading from an input that never ends a read (a pipe nobody closes) leaves one goroutine waiting after `Run` returns.
- **Resizing can flicker for one frame.**
- **Windows is untested.**

## Planned

Ideas, not promises: colors from model material files, smooth shading from model normals, drawing images on the canvas.

## Development

```
make check    # tests (with the race detector) and vet
make fuzz     # fuzz the parsers, rasterizers and clipping; FUZZTIME=1m for longer
```

The race detector needs cgo. Where there is no C compiler, run `make test RACE=`.

## Contributing

Issues and pull requests are welcome; see [CONTRIBUTING.md](CONTRIBUTING.md). To report a security problem privately, see [SECURITY.md](SECURITY.md). Changes are listed in [CHANGELOG.md](CHANGELOG.md).

## License

MIT. See [LICENSE](LICENSE).
