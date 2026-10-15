package engine

import (
	"errors"
	"strings"
	"time"
)

type Scene struct {
	Width   int
	Height  int
	objects []*SceneObject
}

func NewScene(w, h int) (*Scene, error) {
	if w <= 0 || h <= 0 {
		return nil, errors.New("Width and Height must be positive integers")
	}

	return &Scene{
		Width:  w,
		Height: h,
	}, nil
}

func (_ *Scene) HideCursor() {
	sb := NewSequenceBuilder()
	sb.Print(HIDE_CURSOR)
	sb.Execute()
}

func (s *Scene) Animate(fps int) {
	for {
		s.DrawBoundary()
		s.DrawObjects()
		s.HideCursor()
		time.Sleep(time.Second / time.Duration(fps))
	}
}

func (s *Scene) DrawBoundary() {
	sb := NewSequenceBuilder()

	sb.EraseInDisplay(2)
	sb.EraseInDisplay(3)
	sb.MoveTo(1, 1)
	sb.Print("┌")
	sb.Print(strings.Repeat("─", s.Width))
	sb.Print("┐")

	for i := range s.Height {
		sb.MoveTo(i+2, 1)
		sb.Print("│")
		sb.MoveTo(i+2, s.Width+2)
		sb.Print("│")
	}
	sb.NextLine(1)

	sb.Print("└")
	sb.Print(strings.Repeat("─", s.Width))
	sb.Print("┘")

	sb.Execute()
}

func (s *Scene) DrawObjects() {
	for _, obj := range s.objects {
		if obj.X >= s.Width || obj.X < 0 || obj.Y >= s.Height || obj.Y < 0 {
			continue
		}

		sb := NewSequenceBuilder()

		sb.MoveTo(obj.Y+2, obj.X+2)
		sb.Print(string(obj.Value))

		sb.Execute()
	}
}

func (s *Scene) AddObject(obj *SceneObject) {
	s.objects = append(s.objects, obj)
}
