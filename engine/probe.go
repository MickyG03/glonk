package engine

import (
	"io"
	"time"
)

// parseProbeReply scans b for the answers to the two probe queries. kitty is
// true if it holds a protocol reply (ESC [ ? flags u). done is true if it holds
// the device-attributes reply (a CSI sequence ending in c), which every terminal
// sends after the protocol reply, if any, and so marks the end of the wait.
func parseProbeReply(b []byte) (kitty, done bool) {
	for i := 0; i+1 < len(b); i++ {
		if b[i] != 0x1b || b[i+1] != '[' {
			continue
		}

		j := i + 2
		private := j < len(b) && b[j] == '?'
		for j < len(b) && b[j] >= 0x20 && b[j] <= 0x3f {
			j++
		}
		if j >= len(b) {
			break // cut off; the rest may arrive in the next read
		}

		switch {
		case private && b[j] == 'u':
			kitty = true
		case b[j] == 'c':
			done = true
		}
		i = j
	}
	return kitty, done
}

// probeKitty asks the terminal whether it speaks the kitty keyboard protocol.
// It writes the protocol query followed by a device-attributes query to w and
// reads the answers from in. The wait ends when the attributes reply arrives, so
// a terminal without the protocol answers almost at once; one that answers
// neither costs the full timeout. Keystrokes typed during the probe are lost.
func probeKitty(in <-chan []byte, w io.Writer, timeout time.Duration) (bool, error) {
	sb := newSequenceBuilder()
	sb.QueryKeyboardProtocol()
	sb.QueryDeviceAttributes()
	if err := sb.Flush(w); err != nil {
		return false, err
	}

	// Replies can be split across reads, so keep everything and rescan it.
	var got []byte
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case b, ok := <-in:
			if !ok {
				kitty, _ := parseProbeReply(got)
				return kitty, nil
			}
			got = append(got, b...)
			if kitty, done := parseProbeReply(got); done {
				return kitty, nil
			}
		case <-timer.C:
			kitty, _ := parseProbeReply(got)
			return kitty, nil
		}
	}
}
