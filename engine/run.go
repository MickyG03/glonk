package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

// maxDT caps the time step Update sees, in seconds. After a stall (the machine
// sleeping, or the process stopped from another shell) the real gap can be
// huge, and one giant step would fling everything across the world.
const maxDT = 0.25

// probeTimeout is how long Run waits for a terminal to answer the keyboard
// protocol probe. Terminals answer in milliseconds; this only matters when one never does.
const probeTimeout = 250 * time.Millisecond

// maxFPS is the highest frame rate Run accepts. A terminal cannot show more, and the
// frame timer needs an interval of at least a nanosecond.
const maxFPS = 1000

// Config sets up Run. A Width or Height of 0 fills the terminal on that axis
// and follows resizes; a nonzero value fixes that axis. FPS must be between 1 and
// 1000.
//
// Where the terminal supports the kitty keyboard protocol, Run uses it so Input
// can tell exactly which keys are down. LegacyKeys turns that off and uses the
// timing-based handling every terminal supports, as an escape hatch for a
// terminal that misbehaves.
type Config struct {
	Width, Height, FPS int
	LegacyKeys         bool
}

// checkConfig rejects settings Run cannot honor.
func checkConfig(cfg Config) error {
	if cfg.FPS < 1 || cfg.FPS > maxFPS {
		return fmt.Errorf("fps must be between 1 and %d, got %d", maxFPS, cfg.FPS)
	}
	return nil
}

// frameDT returns the seconds between two frames, capped at maxDT.
func frameDT(now, last time.Time) float64 {
	dt := now.Sub(last).Seconds()
	if dt > maxDT {
		dt = maxDT
	}
	return dt
}

// resolveSize picks the frame size: a nonzero config value wins on its axis,
// otherwise the terminal's. Negatives pass through so NewFrameBuffer rejects them.
func resolveSize(cfgW, cfgH, termW, termH int) (int, int) {
	resW := termW
	if cfgW != 0 {
		resW = cfgW
	}

	resH := termH
	if cfgH != 0 {
		resH = cfgH
	}

	return resW, resH
}

// finish adds a failure to restore the terminal to err, the result of the run.
func finish(err error, tty *terminal) error {
	if restoreErr := tty.Teardown(); restoreErr != nil {
		return errors.Join(err, fmt.Errorf("restoring the terminal: %w", restoreErr))
	}
	return err
}

// handle carries out what the input asked for. done says Run should now return err.
func handle(action inputAction, suspend func() error) (done bool, err error) {
	switch action {
	case actionQuit:
		return true, nil
	case actionSuspend:
		if err := suspend(); err != nil {
			return true, err
		}
	}
	return false, nil
}

// Run takes over the terminal and calls game.Update and game.Draw cfg.FPS times
// a second. It returns nil when ctx is cancelled, Update returns ErrQuit, the
// user presses Ctrl-C, or the process receives SIGINT, SIGTERM or SIGHUP; any
// other Update error is returned. The terminal is restored on return, including
// when game code panics, and nothing reads it afterwards, so the program can read
// its own input once Run is over.
//
// Those signals need no handler of your own. After the first one Run stops
// catching them, so a second takes Go's default action and ends a game stuck in
// Update. Catching them yourself (signal.Notify, signal.NotifyContext) keeps that
// from working. A signal the process was started ignoring (nohup does this to
// SIGHUP) stays ignored.
//
// Ctrl-Z suspends the program: the terminal is restored, the process stops until
// the shell continues it (fg), and Run takes the terminal over again. A stop sent
// from outside (kill -TSTP) is not caught, so it leaves the terminal as it is.
// Windows has no suspend.
//
// If the terminal cannot be restored, Run returns that error too, after the one that
// ended the run, if there was one.
//
// Run cannot restore the terminal if the process ends without returning from it:
// os.Exit or log.Fatal called by the game, a panic in a goroutine of your own,
// SIGKILL, or SIGQUIT (which keeps Go's stack dump). Typing reset repairs the
// terminal.
func Run(ctx context.Context, game Game, cfg Config) (err error) {
	if err := checkConfig(cfg); err != nil {
		return err
	}

	tty := newTerminal(os.Stdin, os.Stdout)
	termW, termH, err := tty.Size()
	if err != nil {
		return err
	}

	resW, resH := resolveSize(cfg.Width, cfg.Height, termW, termH)

	nfb, err := NewFrameBuffer(resW, resH)
	if err != nil {
		return err
	}

	// From here a quit signal must reach the deferred Teardown, not kill the process
	// with the terminal still taken over.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer notifyQuit(cancel)()

	err = tty.Setup()
	if err != nil {
		return err
	}
	defer func() { err = finish(err, tty) }()
	input, stopInput := readInput(os.Stdin)
	defer stopInput() // before Teardown: nothing may read the terminal once Run is over

	exact := false
	if !cfg.LegacyKeys {
		kitty, err := probeKitty(input, tty.out, probeTimeout)
		if err != nil {
			return err
		}
		if kitty {
			if err := tty.EnableKeyboardProtocol(); err != nil {
				return err
			}
			exact = true
		}
	}

	ticker := time.NewTicker(time.Second / time.Duration(cfg.FPS))
	defer ticker.Stop()

	last := time.Now()

	in := newInput(exact)
	keys := newInputFeed(in)
	defer keys.stop()

	// suspend hands the terminal back, stops the process until the shell continues
	// it, then takes the terminal over again.
	suspend := func() error {
		keys.reset()
		if err := tty.Teardown(); err != nil {
			return err
		}
		stopProcess()
		if err := tty.Setup(); err != nil {
			return err
		}
		if exact {
			if err := tty.EnableKeyboardProtocol(); err != nil {
				return err
			}
		}
		last = time.Now() // so the time spent stopped is not a frame
		return nil
	}

	dynamic := cfg.Width == 0 || cfg.Height == 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			dt := frameDT(now, last)
			last = now

			if dynamic {
				termW, termH, err := tty.Size()
				if err != nil {
					return err
				}
				resW, resH := resolveSize(cfg.Width, cfg.Height, termW, termH)
				// Some terminals briefly report 0x0; keep the current frame instead of failing.
				if resW > 0 && resH > 0 && (resW != nfb.Width() || resH != nfb.Height()) {
					nfb, err = nfb.resized(resW, resH)
					if err != nil {
						return err
					}
					err = tty.Clear()
					if err != nil {
						return err
					}
				}
			}

			in.beginFrame(now)
			if err := game.Update(dt, in); err != nil {
				if errors.Is(err, ErrQuit) {
					return nil
				}
				return err
			}
			in.endFrame()
			nfb.Clear()
			game.Draw(nfb)
			if err := nfb.present(tty.out); err != nil {
				return err
			}
		case b, ok := <-input:
			if !ok {
				input = nil
				continue
			}
			if done, err := handle(keys.feed(b, time.Now()), suspend); done {
				return err
			}
		case <-keys.expired():
			if done, err := handle(keys.expire(time.Now()), suspend); done {
				return err
			}
		}
	}
}
