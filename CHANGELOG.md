# Changelog

All notable changes to glonk are written here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows [Semantic Versioning](https://semver.org/). While the version is below 1.0, a new minor version (0.2.0, say) may change the API.

## [Unreleased]

### Changed

- `Run` sends a terminal only the cells that changed since the last frame, as one synchronized update (terminal mode 2026, which terminals that do not know it ignore), instead of every cell of every frame. A screen that does not change writes nothing, and the example programs write 74% to 100% fewer bytes. The screen is drawn again in full after a resize, after Ctrl-Z, after a failed write and when the terminal reports the window regained focus. A headless `Run` still writes every frame whole. Text that something else prints to the terminal during a run is no longer overwritten by the next frame.
- Escape sequences are built without `fmt` into a buffer that is reused from frame to frame, and a change of colors or attributes is written as one sequence. A frame takes 2 to 7 times less time to produce, allocates nothing once running, and a color-heavy frame is up to 9% smaller.

## [0.1.0] - 2026-11-08

The first release. The API is pre-1.0 and may still change. The README lists the known limits of what exists today.

### Added

**engine**
- `Run` drives a `Game` at a fixed frame rate (1 to 1000) and owns the terminal: alternate screen, raw mode, restored on return, on errors and on panics in game code. The frame fills the terminal and follows resizes, or has a fixed size and shows the part that fits.
- `FrameBuffer` of `Cell`s with a `Foreground` and `Background` `Color` and text `Attrs` (`Bold`, `Dim`, `Underline`, `Reverse`); `DrawLine`, `DrawBox`, `DrawText`, and `FillTriangle` with a depth buffer.
- `Color` is a type: the terminal default (the zero value), `Palette(n)` for the 256-color palette, or `RGB(r, g, b)`. RGB is shown as the nearest palette color until true color output is added.
- `Canvas`: a pixel layer with two pixels per cell, composed into half-block characters, with a `Seamless` mode for terminals that draw seams. `Shade` maps a color to a brighter or darker one.
- Keyboard `Input` with held keys and one-shot presses. Terminals with the kitty keyboard protocol report real key releases, so several keys can be held at once; everywhere else held keys fall back to timing. Escape sequences split across reads are reassembled, a lone Escape counts after 50 ms, Alt chords are ignored, and keys typed while glonk starts up are kept.
- `Config.Input` and `Config.Output` (default `os.Stdin` and `os.Stdout`) for running on another terminal such as `/dev/tty`, or on none: with an input or output that is not a terminal, `Run` runs headless at the given `Width` and `Height`, writes only the frames, and leaves terminal settings and signals alone, so a game can be tested with key bytes in and a buffer out. With neither a terminal nor a size, `Run` returns an error that says so.
- `Resizer`: a game with a `Resize(width, height int)` method is told the frame size before the first `Update` and whenever it changes.
- A written contract for how `Run` ends and what it calls: it returns `nil` when the player quits (Ctrl-C, `ErrQuit`, a quit signal) and the context's error (`context.Canceled`, `context.DeadlineExceeded`) when the context ends it, at once if the context is already done; `Resize`, `Update` and `Draw` are called one at a time from one goroutine, none starts once the context is done, and none is called after `Run` returns.
- Test helpers: `Keyboard` presses and releases keys for a game's `Update` (exact, like a terminal that reports releases, with no clock), and `FrameBuffer` has a `String` method, so a whole screen can be compared as text and `fmt.Println(fb)` prints it.
- Terminal safety: `SIGINT`, `SIGTERM` and `SIGHUP` end `Run` cleanly (a second signal kills a stuck game); Ctrl-Z suspends and `fg` resumes; `Run` leaves stdin alone once it returns; line wrapping is off while a program runs.
- Output safety: every cell is checked before it is sent, so control characters and escape sequences become `?`, wide characters become `?`, combining marks are skipped by `DrawText`, and NUL is a space.

**mesh**
- `Mesh` with edges and triangular faces; `Cube`, `Grid`, `LoadOBJ` and `LoadOBJLimit` (which bounds untrusted input); `Fit`, `Bounds`, `Validate` and `Clone`.
- `Transform` places a mesh: `Position`, `Yaw`, `Pitch`, `Roll` and a uniform `Scale` (0 means 1), in `Mesh.Transform`. Yaw and Pitch turn a model the way they turn a `Camera`: positive yaw swings the front toward +x, positive pitch tips it up. `Transform.Apply` says where a model point ends up. A transform with a number that is not finite, or a negative scale, draws nothing and `Validate` reports it.
- `Material` says how a mesh looks shaded: a `Color` and a `Ramp` of brightness characters (default `.:-=+*#%@`), in `Mesh.Material`; the draw calls no longer take a color.
- `Light` is an interface with `Directional` and `Ambient` lights, and your own; `DrawShaded` and `DrawShadedPixels` take any number. Lights add up without passing full brightness, so one ambient light and one directional light give what the single combined light gave before.
- Perspective `Camera` with yaw, pitch, field of view, a near plane and a `CellAspect` (how much taller than wide a cell is; fields left at zero use defaults), and `Project`, `Forward`, `Right` and `Look` for putting things on screen and moving around; wireframe `Draw`; `DrawShaded` and `DrawShadedPixels` with depth testing, back-face culling and near-plane clipping; `DoubleSided` for models wound the other way.
- Drawing does not panic on bad meshes, transforms or cameras: edges and faces that name a missing vertex are skipped, and a camera with nonsense values sees nothing.

**sprite**, **geometry**
- `Sprite` is an `engine.Cell` (character, colors, attributes) at a `geometry.Vector2` position, drawn in a `Layer` that can `Add`, `Remove`, `Clear` and count (`Len`) its sprites; a sprite replaces the whole cell under it. `Vector2` and `Vector3`.

**Everything else**
- Three example programs (`bounce2d`, `cube3d` and a model `viewer`), runnable examples in the package documentation, fuzz tests for the parsers, rasterizers and clipping, and a test that keeps the documentation complete.
- Continuous integration (format, vet and race tests on Linux and macOS for Go 1.26 and 1.27, cross-compilation, 32-bit tests), Dependabot, a security policy and contributing notes.

[Unreleased]: https://github.com/kyrokohan/glonk/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/kyrokohan/glonk/releases/tag/v0.1.0
