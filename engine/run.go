package engine

import (
	"context"
	"errors"
	"os"
	"time"
)

// Config sets up Run. A Width or Height of 0 fills the terminal on that axis
// and follows resizes; a nonzero value fixes that axis. FPS must be positive.
type Config struct {
	Width, Height, FPS int
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

	err = tty.Setup()
	if err != nil {
		return err
	}
	defer tty.Teardown()
	input := readInput(os.Stdin)

	ticker := time.NewTicker(time.Second / time.Duration(cfg.FPS))
	defer ticker.Stop()

	last := time.Now()

	in := newInput()

	dynamic := cfg.Width == 0 || cfg.Height == 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			dt := now.Sub(last).Seconds()
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
			for _, k := range parseKeys(b) {
				if k == KeyCtrlC {
					return nil
				}
				in.press(k, time.Now())
			}
		}
	}
}
