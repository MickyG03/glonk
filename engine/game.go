package engine

type Game interface {
	Update(dt float64)
	Draw(fb *FrameBuffer)
}
