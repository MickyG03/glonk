package sprite

import (
	"math"
	"testing"

	"github.com/kyrokohan/glonk/engine"
	"github.com/kyrokohan/glonk/geometry"
)

const (
	frameW = 10
	frameH = 5
)

func newFrame(t *testing.T) *engine.FrameBuffer {
	t.Helper()
	fb, err := engine.NewFrameBuffer(frameW, frameH)
	if err != nil {
		t.Fatalf("unable to initialize frame buffer: %v", err)
	}
	return fb
}

// drawn lists the cells that are not blank, as "x,y".
func drawn(fb *engine.FrameBuffer) map[[2]int]rune {
	got := map[[2]int]rune{}
	for y := range fb.Height() {
		for x := range fb.Width() {
			if c := fb.At(x, y); c != engine.BlankCell {
				got[[2]int{x, y}] = c.Character
			}
		}
	}
	return got
}

func TestLayerDrawFloorsThePosition(t *testing.T) {
	tests := []struct {
		name string
		x, y float64
		want map[[2]int]rune // nil: nothing drawn
	}{
		{"origin", 0, 0, map[[2]int]rune{{0, 0}: 'X'}},
		{"fraction floors down", 2.9, 1.1, map[[2]int]rune{{2, 1}: 'X'}},
		{"last cell", frameW - 0.1, frameH - 0.1, map[[2]int]rune{{frameW - 1, frameH - 1}: 'X'}},
		{"just left of the frame", -0.1, 0, nil},
		{"just above the frame", 0, -0.1, nil},
		{"one past the right edge", frameW, 0, nil},
		{"one past the bottom edge", 0, frameH, nil},
		{"far away", 1000, -1000, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLayer()
			l.Add(New(geometry.Vector2{X: tt.x, Y: tt.y}, 'X'))
			fb := newFrame(t)
			l.Draw(fb)

			got := drawn(fb)
			if len(got) != len(tt.want) {
				t.Fatalf("wanted %v drawn, got %v", tt.want, got)
			}
			for pos, r := range tt.want {
				if got[pos] != r {
					t.Errorf("wanted %q at %v, got %v", r, pos, got)
				}
			}
		})
	}
}

// What int(NaN) and int(1e300) give differs between CPUs: on some the result is
// cell 0, so a sprite with no real position must be stopped before converting.
func TestLayerDrawSkipsSpritesWithoutARealPosition(t *testing.T) {
	nan, inf := math.NaN(), math.Inf(1)

	tests := []struct {
		name string
		x, y float64
	}{
		{"NaN x", nan, 1},
		{"NaN y", 1, nan},
		{"NaN both", nan, nan},
		{"+Inf x", inf, 1},
		{"-Inf x", -inf, 1},
		{"+Inf y", 1, inf},
		{"-Inf y", 1, -inf},
		{"1e300", 1e300, 1e300},
		{"-1e300", -1e300, 1},
		{"largest float", math.MaxFloat64, 0},
		{"just past int64", 1 << 63, 0},
		{"smallest int64", -(1 << 63), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLayer()
			l.Add(New(geometry.Vector2{X: tt.x, Y: tt.y}, 'X'))
			fb := newFrame(t)
			l.Draw(fb)

			if got := drawn(fb); len(got) != 0 {
				t.Errorf("wanted nothing drawn, got %v", got)
			}
		})
	}
}

func TestLayerIgnoresNilSprites(t *testing.T) {
	l := NewLayer()
	l.Add(New(geometry.Vector2{X: 1, Y: 1}, 'A'))
	l.Add(nil)
	l.Add(New(geometry.Vector2{X: 2, Y: 1}, 'B'))

	fb := newFrame(t)
	l.Draw(fb)

	want := map[[2]int]rune{{1, 1}: 'A', {2, 1}: 'B'}
	got := drawn(fb)
	if len(got) != len(want) || got[[2]int{1, 1}] != 'A' || got[[2]int{2, 1}] != 'B' {
		t.Errorf("wanted %v, got %v", want, got)
	}
}

func TestLayerDrawsAfterABadSprite(t *testing.T) {
	l := NewLayer()
	l.Add(New(geometry.Vector2{X: math.NaN(), Y: 0}, 'N'))
	l.Add(New(geometry.Vector2{X: 3, Y: 2}, 'G'))

	fb := newFrame(t)
	l.Draw(fb)

	got := drawn(fb)
	if len(got) != 1 || got[[2]int{3, 2}] != 'G' {
		t.Errorf("wanted only G at (3, 2), got %v", got)
	}
}

func TestLayerLaterSpritesDrawOverEarlierOnes(t *testing.T) {
	l := NewLayer()
	l.Add(New(geometry.Vector2{X: 4, Y: 4}, 'a'))
	l.Add(New(geometry.Vector2{X: 4.5, Y: 4.5}, 'b'))

	fb := newFrame(t)
	l.Draw(fb)

	if got := fb.At(4, 4).Character; got != 'b' {
		t.Errorf("wanted b on top, got %q", got)
	}
}
