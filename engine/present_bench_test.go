package engine

import (
	"io"
	"math"
	"math/rand/v2"
	"testing"
)

// benchEmitters are the ways to send a frame that the benchmarks compare: every cell
// of every frame, and only the cells that changed.
var benchEmitters = []struct {
	name string
	make func() func(w io.Writer, fb *FrameBuffer) error
}{
	{"full", func() func(io.Writer, *FrameBuffer) error {
		p := newPresenter(false, false)
		return func(w io.Writer, fb *FrameBuffer) error { return p.present(w, fb, fb.width, fb.height) }
	}},
	{"changes", func() func(io.Writer, *FrameBuffer) error {
		p := newPresenter(true, false)
		return func(w io.Writer, fb *FrameBuffer) error { return p.present(w, fb, fb.width, fb.height) }
	}},
	{"full 24-bit", func() func(io.Writer, *FrameBuffer) error {
		p := newPresenter(false, true)
		return func(w io.Writer, fb *FrameBuffer) error { return p.present(w, fb, fb.width, fb.height) }
	}},
	{"changes 24-bit", func() func(io.Writer, *FrameBuffer) error {
		p := newPresenter(true, true)
		return func(w io.Writer, fb *FrameBuffer) error { return p.present(w, fb, fb.width, fb.height) }
	}},
}

// benchScene is a made-up but typical sequence of frames: what moves, and how much.
type benchScene struct {
	name string
	w, h int
	draw func(fb *FrameBuffer, i int)
}

func box(fb *FrameBuffer) {
	fb.DrawBox(0, 0, fb.Width(), fb.Height(), Palette(240))
}

var benchScenes = []benchScene{
	{"static 80x24", 80, 24, func(fb *FrameBuffer, i int) {
		box(fb)
		fb.DrawText(2, 1, "nothing moves here", Palette(208))
	}},
	{"ball in a box 80x24", 80, 24, func(fb *FrameBuffer, i int) {
		box(fb)
		x := 1 + int(math.Abs(math.Mod(float64(i)*1.5, 2*77)-77))
		y := 1 + int(math.Abs(math.Mod(float64(i)*0.7, 2*21)-21))
		fb.Set(x, y, Cell{Character: 'o', Foreground: Palette(208)})
		fb.DrawText(1, 22, "fps "+string(rune('0'+i%10))+string(rune('0'+(i/10)%10)), Color{}) // a counter that changes every frame
	}},
	{"5% of the cells change 80x24", 80, 24, func(fb *FrameBuffer, i int) {
		rng := rand.New(rand.NewPCG(7, 7))
		for y := range fb.Height() {
			for x := range fb.Width() {
				c := Cell{Character: rune('a' + (x+y)%26)}
				if x%7 == 0 {
					c.Foreground = Palette(uint8(1 + x%200))
				}
				fb.Set(x, y, c)
			}
		}
		moving := rand.New(rand.NewPCG(uint64(i), 1))
		for range 96 { // 5% of 1920
			fb.Set(moving.IntN(80), moving.IntN(24), Cell{Character: rune('A' + rng.IntN(26)), Foreground: Palette(uint8(1 + moving.IntN(200)))})
		}
	}},
	{"scrolling log 80x24", 80, 24, func(fb *FrameBuffer, i int) {
		for row := range fb.Height() {
			n := i + row
			fb.DrawText(0, row, "line "+string(rune('0'+(n/100)%10))+string(rune('0'+(n/10)%10))+string(rune('0'+n%10))+": the quick brown fox jumps over the lazy dog", Palette(uint8(1+n%7)))
		}
	}},
	{"half-block triangle 80x24", 80, 24, func(fb *FrameBuffer, i int) {
		c := fb.Canvas()
		a := float64(i) * 0.08
		var px, py [3]float64
		for k := range 3 {
			t := a + float64(k)*2*math.Pi/3
			px[k], py[k] = 40+30*math.Cos(t), 24+20*math.Sin(t)
		}
		c.FillTriangle(px[0], py[0], 1, px[1], py[1], 1, px[2], py[2], 1, Palette(208))
		c.Compose()
	}},
	{"shaded RGB triangle 80x24", 80, 24, func(fb *FrameBuffer, i int) {
		c := fb.Canvas()
		a := float64(i) * 0.08
		var px, py [3]float64
		for k := range 3 {
			t := a + float64(k)*2*math.Pi/3
			px[k], py[k] = 40+30*math.Cos(t), 24+20*math.Sin(t)
		}
		// A different brightness for every pixel row, like a lit surface.
		for y := range c.Height() {
			shade := Shade(RGB(255, 128, 0), float64(y)/float64(c.Height()))
			for x := range c.Width() {
				if pointInTriangle(float64(x)+0.5, float64(y)+0.5, px, py) {
					c.Set(x, y, shade)
				}
			}
		}
		c.Compose()
	}},
	{"everything changes 80x24", 80, 24, func(fb *FrameBuffer, i int) {
		rng := rand.New(rand.NewPCG(uint64(i), 3))
		for y := range fb.Height() {
			for x := range fb.Width() {
				fb.Set(x, y, Cell{Character: rune('a' + rng.IntN(26)), Foreground: Palette(uint8(1 + rng.IntN(200))), Background: Palette(uint8(1 + rng.IntN(200)))})
			}
		}
	}},
	{"half-block noise 200x60", 200, 60, func(fb *FrameBuffer, i int) {
		rng := rand.New(rand.NewPCG(uint64(i), 5))
		for y := range fb.Height() {
			for x := range fb.Width() {
				fb.Set(x, y, Cell{Character: '▀', Foreground: Palette(uint8(1 + rng.IntN(200))), Background: Palette(uint8(1 + rng.IntN(200)))})
			}
		}
	}},
}

// BenchmarkFrames reports the time, the allocations and the bytes written per frame
// for each scene and each way of sending it. The frames are made first and then
// cycled, so the one jump from the last back to the first is part of the average.
func BenchmarkFrames(b *testing.B) {
	const frames = 120
	for _, sc := range benchScenes {
		seq := make([]*FrameBuffer, frames)
		for i := range seq {
			fb, err := NewFrameBuffer(sc.w, sc.h)
			if err != nil {
				b.Fatal(err)
			}
			sc.draw(fb, i)
			seq[i] = fb
		}
		for _, em := range benchEmitters {
			b.Run(sc.name+"/"+em.name, func(b *testing.B) {
				emit := em.make()
				var out countingWriter

				b.ReportAllocs()
				n := 0
				for b.Loop() {
					if err := emit(&out, seq[n%frames]); err != nil {
						b.Fatal(err)
					}
					n++
				}
				b.ReportMetric(float64(out.bytes)/float64(n), "bytes/frame")
			})
		}
	}
}

// pointInTriangle reports whether (x, y) is inside the triangle with corners (px[k], py[k]).
func pointInTriangle(x, y float64, px, py [3]float64) bool {
	sign := func(ax, ay, bx, by float64) float64 { return (x-bx)*(ay-by) - (ax-bx)*(y-by) }
	d1, d2, d3 := sign(px[0], py[0], px[1], py[1]), sign(px[1], py[1], px[2], py[2]), sign(px[2], py[2], px[0], py[0])
	neg := d1 < 0 || d2 < 0 || d3 < 0
	pos := d1 > 0 || d2 > 0 || d3 > 0
	return !(neg && pos)
}
