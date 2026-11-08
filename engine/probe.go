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

// stripProbeReplies returns b without the probe's replies, so what is left is what
// was typed. It follows the same rules as parseProbeReply, and a reply cut off at
// the end of b is kept, since it cannot be told apart from the start of a key.
func stripProbeReplies(b []byte) []byte {
	var typed []byte
	for i := 0; i < len(b); {
		if b[i] == 0x1b && i+1 < len(b) && b[i+1] == '[' {
			j := i + 2
			private := j < len(b) && b[j] == '?'
			for j < len(b) && b[j] >= 0x20 && b[j] <= 0x3f {
				j++
			}
			if j < len(b) && ((private && b[j] == 'u') || b[j] == 'c') {
				i = j + 1
				continue
			}
		}
		typed = append(typed, b[i])
		i++
	}
	return typed
}

// probeKitty asks the terminal whether it speaks the kitty keyboard protocol.
// It writes the protocol query followed by a device-attributes query to w and
// reads the answers from in. The wait ends when the attributes reply arrives, so
// a terminal without the protocol answers almost at once; one that answers
// neither costs the full timeout.
//
// Keys typed during the wait are read along with the replies. They are returned in
// typed, so that they are not lost.
func probeKitty(in <-chan []byte, w io.Writer, timeout time.Duration) (kitty bool, typed []byte, err error) {
	sb := newSequenceBuilder()
	sb.QueryKeyboardProtocol()
	sb.QueryDeviceAttributes()
	if err := sb.Flush(w); err != nil {
		return false, nil, err
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
				return kitty, stripProbeReplies(got), nil
			}
			got = append(got, b...)
			if kitty, done := parseProbeReply(got); done {
				return kitty, stripProbeReplies(got), nil
			}
		case <-timer.C:
			kitty, _ := parseProbeReply(got)
			return kitty, stripProbeReplies(got), nil
		}
	}
}
