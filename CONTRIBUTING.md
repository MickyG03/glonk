# Contributing

Issues and pull requests are welcome. glonk is a one-person project, so replies can take a while.

## Before you start

- For a bug, a typo or a small improvement, go ahead and open a pull request.
- For anything larger, open an issue first. The API is pre-1.0 and may still change, so a big change is best agreed before you write it.
- To report a security problem, do not use an issue. See [SECURITY.md](SECURITY.md).

## Setting up

You need Go 1.26 or newer. The race detector needs cgo, so a C compiler must be installed.

```
make check    # formatting, go vet, and all tests with the race detector
make fuzz     # fuzz every fuzz test for FUZZTIME each (default 10s)
make run      # try an example; EXAMPLE=./examples/cube3d picks another
```

`make check` is what CI runs, on Linux and macOS with Go 1.26 and 1.27. CI also vets for Windows, FreeBSD and 32-bit ARM, builds for Windows, and runs the tests as a 32-bit program.

## Making a change

- Keep a change to one thing. Commit messages are `type: subject` with the type one of `feat`, `fix`, `refactor`, `docs`, `test` or `chore`, and a lowercase subject in the imperative, under about 70 characters.
- Behavior changes come with tests. For a bug, write the test first and see it fail.
- Exported names need a doc comment, and a doc comment must not name an unexported identifier, because readers of the documentation cannot look it up. Examples in `example_test.go` files need an `// Output:` comment so that `go test` runs them. `engine/docs_test.go` checks all of this.
- The first sentence of a package comment is shown as written in pkg.go.dev lists, so it must not contain `[Link]` brackets.
- What `Run` writes to the terminal is tested by replaying the bytes onto a small virtual terminal (`engine/vterm_test.go`) and requiring that it shows exactly the frame; do the same for a change to how frames are drawn.
- Text and cell contents reach the terminal only through the frame buffer, which checks every cell. Do not send them by another route. Control sequences are built in one place, `engine/escape.go`.
- Comments say why, not what.

## License

By contributing you agree that your work is released under the [MIT License](LICENSE) like the rest of the project.
