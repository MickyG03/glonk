//go:build !unix

package engine

import "os"

// readInput streams what is typed on f in chunks until f ends, then closes the
// channel. Without select there is no way to stop a read that is waiting, so
// stop cannot keep the goroutine from taking the next input; see readBlocking.
func readInput(f *os.File) (<-chan []byte, func()) {
	return readBlocking(f)
}
