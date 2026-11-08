//go:build unix

package engine

import (
	"bytes"
	"errors"
	"os"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// pollInterval is how often the reader looks up from waiting to see whether it was
// told to stop. It is also the most stop can take.
const pollInterval = 20 * time.Millisecond

// selectLimit is the first descriptor number select cannot watch (FD_SETSIZE).
const selectLimit = 1024

// readInput streams what is typed on f in chunks until f ends or fails, then
// closes the channel. stop ends the stream and returns only once the reader has
// stopped, so nothing reads f after it returns: the next reader of f gets
// everything typed from then on. A chunk read but not yet received when stop is
// called is dropped.
//
// The reader asks select whether input is waiting and reads only if it is, so
// no read is ever stuck holding input that belongs to someone else.
func readInput(f *os.File) (<-chan []byte, func()) {
	// shortcut: select watches one descriptor, which must be below FD_SETSIZE; fall
	// back to the blocking reader above that, and revisit if A2 adds Config.Input.
	fd := int(f.Fd())
	if fd >= selectLimit {
		return readBlocking(f)
	}

	ch := make(chan []byte)
	done := make(chan struct{})
	exited := make(chan struct{})

	go func() {
		defer close(exited)
		defer close(ch)
		buf := make([]byte, readBuffer)
		for {
			select {
			case <-done:
				return
			default:
			}

			ready, err := waitReadable(fd)
			if err != nil {
				return
			}
			if !ready {
				continue
			}

			n, err := unix.Read(fd, buf)
			if retryable(err) {
				continue
			}
			if err != nil || n <= 0 { // n == 0 is the end of input
				return
			}
			select {
			case ch <- bytes.Clone(buf[:n]):
			case <-done:
				return
			}
		}
	}()

	var once sync.Once
	return ch, func() {
		once.Do(func() { close(done) })
		<-exited
	}
}

// waitReadable waits up to pollInterval for fd to have input or to reach its end.
func waitReadable(fd int) (bool, error) {
	var set unix.FdSet
	set.Set(fd)
	timeout := unix.NsecToTimeval(int64(pollInterval))

	n, err := unix.Select(fd+1, &set, nil, nil, &timeout)
	return selectReady(n, err)
}

// selectReady reads the result of select. A signal that interrupts the wait is
// not a failure, only a wait in which nothing arrived.
func selectReady(n int, err error) (bool, error) {
	if errors.Is(err, unix.EINTR) {
		return false, nil
	}
	return n > 0 && err == nil, err
}

// retryable reports whether a failed read just needs repeating: a signal arrived
// first, or another reader of the same descriptor got there before us.
func retryable(err error) bool {
	return errors.Is(err, unix.EINTR) || errors.Is(err, unix.EAGAIN)
}
