package sprite_test

import (
	"fmt"
	"math"
	"strings"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/sprite"
)

func ExampleLayer() {
	fb, err := engine.NewFrameBuffer(12, 5)
	if err != nil {
		panic(err)
	}
	fb.DrawBox(0, 0, fb.Width(), fb.Height(), 0)

	layer := sprite.NewLayer()
	layer.Add(sprite.New(1.9, 1, 'a')) // a fraction is floored: column 1
	layer.Add(sprite.New(7, 3, 'b'))
	layer.Add(sprite.New(-3, 2, 'c'))         // off screen: not drawn
	layer.Add(sprite.New(math.NaN(), 2, 'd')) // no position: not drawn
	layer.Draw(fb)

	for y := range fb.Height() {
		row := make([]rune, fb.Width())
		for x := range row {
			row[x] = fb.At(x, y).Character
		}
		fmt.Println(strings.TrimRight(string(row), " "))
	}
	// Output:
	// ┌──────────┐
	// │a         │
	// │          │
	// │      b   │
	// └──────────┘
}
