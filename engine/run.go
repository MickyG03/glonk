package engine

import (
	"context"
	"errors"
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

// Config sets up Run. A Width or Height of 0 fills the terminal on that axis
// and follows resizes; a nonzero value fixes that axis. FPS must be positive.
//
// Where the terminal supports the kitty keyboard protocol, Run uses it so Input
// can tell exactly which keys are down. LegacyKeys turns that off and uses the
// timing-based handling every terminal supports, as an escape hatch for a
// terminal that misbehaves.
type Config struct {
	Width, Height, FPS int
	LegacyKeys         bool
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

// Run takes over the terminal and calls game.Update and game.Draw cfg.FPS times
// a second. It returns nil when ctx is cancelled, Update returns ErrQuit, or the
// user presses Ctrl-C; any other Update error is returned. The terminal is
// restored on return, including when game code panics.
func Run(ctx context.Context, game Game, cfg Config) error {
	if cfg.FPS <= 0 {
		return errors.New("invalid fps value given")
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
	defer tty.Teardown()
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
					nfb, err = NewFrameBuffer(resW, resH)
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
			if keys.feed(b, time.Now()) {
				return nil
			}
		case <-keys.expired():
			if keys.expire(time.Now()) {
				return nil
			}
		}
	}
}
