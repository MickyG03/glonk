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

// Add appends obj to the layer.
func (l *Layer) Add(obj *Sprite) {
	l.sprites = append(l.sprites, obj)
}

// Draw sets each sprite's cell in fb, clipping anything off screen.
func (l *Layer) Draw(fb *engine.FrameBuffer) {
	for _, sp := range l.sprites {
		fb.Set(int(math.Floor(sp.X)), int(math.Floor(sp.Y)), engine.Cell{Character: sp.Value})
	}
}
