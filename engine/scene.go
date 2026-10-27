package engine

import (
	"context"
	"errors"
	"math"
	"time"
)

type Scene struct {
	objects     []*SceneObject
	frameBuffer *FrameBuffer
	fps         uint
}

func NewScene(w, h int, fps uint) (*Scene, error) {
	fb, err := NewFrameBuffer(w, h)
	if err != nil {
		return nil, err
	}

	if fps == 0 {
		return nil, errors.New("fps cannot be 0")
	}

	return &Scene{frameBuffer: fb, fps: fps}, nil
}

func (s *Scene) Animate(ctx context.Context, update func(dt float64)) {
	ticker := time.NewTicker(time.Second / time.Duration(s.fps))
	defer ticker.Stop()

	last := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			dt := now.Sub(last).Seconds()
			last = now
			update(dt)
			s.frameBuffer.Clear()
			s.DrawObjects()
			s.Render()
		}
	}
}

func (s *Scene) DrawObjects() {
	for _, obj := range s.objects {
		s.frameBuffer.Set(int(math.Floor(obj.X)), int(math.Floor(obj.Y)), Cell{Character: obj.Value})
	}
}

func (s *Scene) AddObject(obj *SceneObject) {
	s.objects = append(s.objects, obj)
}

func (s *Scene) Render() {
	width := s.frameBuffer.width
	height := s.frameBuffer.height

	sb := NewSequenceBuilder()

	for y := range height {
		sb.MoveTo(y+1, 1)
		for x := range width {
			sb.Print(string(s.frameBuffer.frame[y*width+x].Character))
		}
	}

	sb.Execute()
}
