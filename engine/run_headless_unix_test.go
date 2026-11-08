//go:build unix

package engine

import (
	"context"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// A headless Run does not own the process, so a quit signal is none of its business.
// The test catches the signal itself, which keeps the process alive if Run did not.
func TestRunHeadlessLeavesQuitSignalsAlone(t *testing.T) {
	mine := make(chan os.Signal, 1)
	signal.Notify(mine, syscall.SIGHUP)
	defer signal.Stop(mine)

	var updates atomic.Int64
	game := scriptedGame{update: func(float64, *Input) error { updates.Add(1); return nil }}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := startHeadless(ctx, game, Config{Input: strings.NewReader(""), Output: io.Discard})
	waitFor(t, "the first frame", func() bool { return updates.Load() > 0 })

	if err := syscall.Kill(os.Getpid(), syscall.SIGHUP); err != nil {
		t.Fatal(err)
	}
	<-mine

	select {
	case err := <-done:
		t.Fatalf("wanted Run to carry on after SIGHUP, but it returned %v", err)
	case <-time.After(200 * time.Millisecond):
	}
	cancel()
	finished(t, done)
}
