package engine

import (
	"strings"
	"time"
)

const DrawOffset int = 2

type Scene struct {
	objects     []*SceneObject
	frameBuffer *FrameBuffer
}

func NewScene(w, h int) (*Scene, error) {
	fb, err := NewFrameBuffer(w, h)
	if err != nil {
		return nil, err
	}

	return &Scene{frameBuffer: fb}, nil
}

func (_ *Scene) HideCursor() {
	sb := NewSequenceBuilder()
	sb.Print(HIDE_CURSOR)
	sb.Execute()
}

func (s *Scene) Animate(fps int) {
	for {
		s.frameBuffer.Clear()
		s.DrawObjects()
		s.Render()
		time.Sleep(time.Second / time.Duration(fps))
	}
}

func (s *Scene) DrawBoundary() {
	width := s.frameBuffer.width
	height := s.frameBuffer.height

	sb := NewSequenceBuilder()
	sb.MoveTo(1, 1)
	sb.Print("┌")
	sb.Print(strings.Repeat("─", width))
	sb.Print("┐")

	for i := range height {
		sb.MoveTo(i+DrawOffset, 1)
		sb.Print("│")
		sb.MoveTo(i+DrawOffset, width+DrawOffset)
		sb.Print("│")
	}
	sb.MoveTo(height+DrawOffset, 1)

	sb.Print("└")
	sb.Print(strings.Repeat("─", width))
	sb.Print("┘")

	sb.Execute()
}

func (s *Scene) DrawObjects() {
	for _, obj := range s.objects {
		s.frameBuffer.Set(obj.X, obj.Y, Cell{Character: obj.Value})
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
		sb.MoveTo(y+DrawOffset, DrawOffset)
		for x := range width {
			sb.Print(string(s.frameBuffer.frame[y*width+x].Character))
		}
	}

	sb.Execute()
}
