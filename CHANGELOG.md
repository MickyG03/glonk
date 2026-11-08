# Changelog

All notable changes to glonk are written here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project will follow [Semantic Versioning](https://semver.org/) from v0.1.0 on.

## Unreleased

There has been no release yet. The API is pre-1.0 and **will** change before v0.1.0: `Cell` and its colors, `Game` and `Run`, the 3D API (transforms, lights, materials), test helpers and the 2D API are all planned to be reshaped. The README lists the known limits of what exists today.

### Added

**engine**
- `Run` drives a `Game` at a fixed frame rate (1 to 1000) and owns the terminal: alternate screen, raw mode, restored on return, on errors and on panics in game code. The frame fills the terminal and follows resizes, or has a fixed size and shows the part that fits.
- `FrameBuffer` of `Cell`s with a `Foreground` and `Background` `Color` and text `Attrs` (`Bold`, `Dim`, `Underline`, `Reverse`); `DrawLine`, `DrawBox`, `DrawText`, and `FillTriangle` with a depth buffer.
- `Color` is a type: the terminal default (the zero value), `Palette(n)` for the 256-color palette, or `RGB(r, g, b)`. RGB is shown as the nearest palette color until true color output is added.
- `Canvas`: a pixel layer with two pixels per cell, composed into half-block characters, with a `Seamless` mode for terminals that draw seams. `Shade` maps a color to a brighter or darker one.
- Keyboard `Input` with held keys and one-shot presses. Terminals with the kitty keyboard protocol report real key releases, so several keys can be held at once; everywhere else held keys fall back to timing. Escape sequences split across reads are reassembled, a lone Escape counts after 50 ms, Alt chords are ignored, and keys typed while glonk starts up are kept.
- Terminal safety: `SIGINT`, `SIGTERM` and `SIGHUP` end `Run` cleanly (a second signal kills a stuck game); Ctrl-Z suspends and `fg` resumes; `Run` leaves stdin alone once it returns; line wrapping is off while a program runs.
- Output safety: every cell is checked before it is sent, so control characters and escape sequences become `?`, wide characters become `?`, combining marks are skipped by `DrawText`, and NUL is a space.

**mesh**
- `Mesh` with edges and triangular faces; `Cube`, `Grid`, `LoadOBJ` and `LoadOBJLimit` (which bounds untrusted input); `Fit`, `Bounds`, `Validate` and `Clone`.
- Perspective `Camera` with yaw, pitch, field of view and a near plane (fields left at zero use defaults); wireframe `Draw`; `DrawShaded` and `DrawShadedPixels` with a directional `Light`, depth testing, back-face culling and near-plane clipping; `DoubleSided` for models wound the other way.
- Drawing does not panic on bad meshes or cameras: edges and faces that name a missing vertex are skipped, and a camera with nonsense values sees nothing.

**sprite**, **geometry**
- Single-cell `Sprite`s in drawing `Layer`s; `Vector2` and `Vector3`.

**Everything else**
- Three example programs (`bounce2d`, `cube3d` and a model `viewer`), runnable examples in the package documentation, fuzz tests for the parsers, rasterizers and clipping, and a test that keeps the documentation complete.
- Continuous integration (format, vet and race tests on Linux and macOS for Go 1.26 and 1.27, cross-compilation, 32-bit tests), Dependabot, a security policy and contributing notes.
