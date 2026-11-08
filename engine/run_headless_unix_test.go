//go:build unix

package engine

import (
	"context"
	"io"
	"os"
	"os/signal"
	"runtime"
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

// leavesNoGoroutines runs f and fails if goroutines started during it are still going
// shortly after.
func leavesNoGoroutines(t *testing.T, f func()) {
	t.Helper()
	before := runtime.NumGoroutine()
	f()
	waitFor(t, "Run's goroutines to end", func() bool { return runtime.NumGoroutine() <= before })
}

// panicGame panics in the method named by where.
type panicGame struct{ where string }

func (g panicGame) Resize(int, int) {
	if g.where == "Resize" {
		panic("boom")
	}
}

func (g panicGame) Update(float64, *Input) error {
	if g.where == "Update" {
		panic("boom")
	}
	return nil
}

func (g panicGame) Draw(*FrameBuffer) {
	if g.where == "Draw" {
		panic("boom")
	}
}

func TestRunLeavesNoGoroutinesBehind(t *testing.T) {
	ends := map[string]func(ctx context.Context, cancel func()) Game{
		"the game quits": func(context.Context, func()) Game {
			return scriptedGame{update: func(float64, *Input) error { return ErrQuit }}
		},
		"the context is cancelled": func(_ context.Context, cancel func()) Game {
			return scriptedGame{update: func(float64, *Input) error { cancel(); return nil }}
		},
		"the game fails": func(context.Context, func()) Game {
			return scriptedGame{update: func(float64, *Input) error { return io.ErrUnexpectedEOF }}
		},
	}
	for name, makeGame := range ends {
		t.Run(name, func(t *testing.T) {
			r, w, err := os.Pipe() // a file, so the reader can be stopped
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			defer w.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			leavesNoGoroutines(t, func() {
				Run(ctx, makeGame(ctx, cancel), Config{FPS: maxFPS, Width: 3, Height: 2, Input: r, Output: io.Discard})
			})
		})
	}
}

func TestRunPanicsOnWithItsGoroutinesStopped(t *testing.T) {
	for _, where := range []string{"Resize", "Update", "Draw"} {
		t.Run(where, func(t *testing.T) {
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			defer w.Close()

			var got any
			leavesNoGoroutines(t, func() {
				defer func() { got = recover() }()
				Run(context.Background(), panicGame{where}, Config{FPS: maxFPS, Width: 3, Height: 2, Input: r, Output: io.Discard})
			})

			if got != "boom" {
				t.Errorf("wanted the game's panic to reach the caller, got %v", got)
			}
		})
	}
}
