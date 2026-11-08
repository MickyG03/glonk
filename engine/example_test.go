package engine_test

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kyrokohan/glonk/engine"
)

// printFrame prints the characters of fb, one line per row.
func printFrame(fb *engine.FrameBuffer) {
	for y := range fb.Height() {
		row := make([]rune, fb.Width())
		for x := range row {
			row[x] = fb.At(x, y).Character
		}
		fmt.Println(strings.TrimRight(string(row), " "))
	}
}

// walker is a game: a dot that crosses the screen until Escape is pressed.
type walker struct{ x float64 }

func (w *walker) Update(dt float64, in *engine.Input) error {
	if in.Pressed(engine.KeyEscape) {
		return engine.ErrQuit
	}
	w.x += 20 * dt // cells per second
	return nil
}

func (w *walker) Draw(fb *engine.FrameBuffer) {
	fb.DrawBox(0, 0, fb.Width(), fb.Height(), 0)
	fb.Set(int(w.x)%fb.Width(), fb.Height()/2, engine.Cell{Character: '@', Color: 208})
}

// Run takes over the terminal until the game quits, the user presses Ctrl-C, or the
// program is told to stop; the terminal is restored when it returns.
func ExampleRun() {
	// Width and Height of 0 fill the terminal and follow resizes.
	err := engine.Run(context.Background(), &walker{}, engine.Config{FPS: 30})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func ExampleFrameBuffer_DrawText() {
	fb, err := engine.NewFrameBuffer(16, 3)
	if err != nil {
		panic(err)
	}

	fb.DrawBox(0, 0, fb.Width(), fb.Height(), 0)
	fb.DrawText(2, 1, "hello, glonk", 0)

	printFrame(fb)
	// Output:
	// ┌──────────────┐
	// │ hello, glonk │
	// └──────────────┘
}

// A Canvas holds two pixels per cell, one above the other, and Compose turns them
// into half-block characters.
func ExampleCanvas() {
	fb, err := engine.NewFrameBuffer(8, 2)
	if err != nil {
		panic(err)
	}

	canvas := fb.Canvas()
	for x := range canvas.Width() {
		canvas.Set(x, x%canvas.Height(), 196) // a zigzag in palette color 196
	}
	canvas.Compose()

	printFrame(fb)
	// Output:
	// ▀▄  ▀▄
	//   ▀▄  ▀▄
}

func ExampleShade() {
	const red = 196
	fmt.Println(engine.Shade(red, 1), engine.Shade(red, 0.5), engine.Shade(red, 0))
	// Output:
	// 196 88 16
}
