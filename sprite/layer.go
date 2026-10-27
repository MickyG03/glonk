package sprite

import (
	"math"

	"github.com/kyrokohan/glonk/engine"
)

type Layer struct {
	sprites []*Sprite
}

func NewLayer() *Layer {
	return &Layer{}
}

func (l *Layer) Add(obj *Sprite) {
	l.sprites = append(l.sprites, obj)
}

func (l *Layer) Draw(fb *engine.FrameBuffer) {
	for _, sp := range l.sprites {
		fb.Set(int(math.Floor(sp.X)), int(math.Floor(sp.Y)), engine.Cell{Character: sp.Value})
	}
}
