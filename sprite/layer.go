package sprite

import (
	"math"

	"github.com/kyrokohan/glonk/engine"
)

// Layer is an ordered set of sprites drawn together; later sprites draw over earlier ones.
type Layer struct {
	sprites []*Sprite
}

// NewLayer returns an empty Layer.
func NewLayer() *Layer {
	return &Layer{}
}

// Add appends obj to the layer. A nil sprite is ignored.
func (l *Layer) Add(obj *Sprite) {
	if obj == nil {
		return
	}
	l.sprites = append(l.sprites, obj)
}

// Draw sets each sprite's cell in fb, clipping anything off screen. A sprite whose
// Position has a NaN or infinite X or Y is not drawn, like one that is off screen.
func (l *Layer) Draw(fb *engine.FrameBuffer) {
	w, h := float64(fb.Width()), float64(fb.Height())

	for _, sp := range l.sprites {
		x, y := math.Floor(sp.Position.X), math.Floor(sp.Position.Y)

		// Checked before converting to int, and written so that NaN fails: what
		// int(NaN) or int(1e300) gives differs from one CPU to the next.
		if !(x >= 0 && x < w && y >= 0 && y < h) {
			continue
		}
		fb.Set(int(x), int(y), engine.Cell{Character: sp.Value})
	}
}
