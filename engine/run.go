package engine

import (
	"context"
	"errors"
	"os"
	"time"
)

type Config struct {
	Width, Height, FPS int
}

func Run(ctx context.Context, game Game, cfg Config) error {
	if cfg.FPS <= 0 {
		return errors.New("invalid fps value given")
	}

	nfb, err := NewFrameBuffer(cfg.Width, cfg.Height)
	if err != nil {
		return err
	}

	tty := newTerminal(os.Stdin, os.Stdout)
	err = tty.Setup()
	if err != nil {
		return err
	}
	defer tty.Teardown()
	input := readInput(os.Stdin)

	ticker := time.NewTicker(time.Second / time.Duration(cfg.FPS))
	defer ticker.Stop()

	last := time.Now()

	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			dt := now.Sub(last).Seconds()
			last = now
			game.Update(dt)
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
			}
		}
	}
}
