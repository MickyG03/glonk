package engine

import (
	"bytes"
	"io"
)

// readInput streams chunks read from r until it errors
// then closes the channel.
func readInput(r io.Reader) <-chan []byte {
	ch := make(chan []byte)
	// the goroutine stays blocked in Read after Run returns;
	// harmless unless Run is called twice.
	go func() {
		defer close(ch)
		buf := make([]byte, 64)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				ch <- bytes.Clone(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()
	return ch
}
