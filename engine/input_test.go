package engine

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func TestReadBlocking(t *testing.T) {
	r := iotest.DataErrReader(iotest.OneByteReader(strings.NewReader("ab\x03")))
	var chunks [][]byte
	input, _ := readBlocking(r)
	for chunk := range input {
		chunks = append(chunks, chunk)
	}
	got := bytes.Join(chunks, nil)

	if !bytes.Equal([]byte("ab\x03"), got) {
		t.Errorf("wanted %q, got %q", "ab\x03", got)
	}
}

func TestInputIsDown(t *testing.T) {
	t0 := time.Unix(0, 0)
	ms := time.Millisecond

	tests := []struct {
		name    string
		presses []time.Duration
		frameAt time.Duration
		want    bool
	}{
		{"never pressed", nil, 0, false},
		{"just pressed", []time.Duration{0}, 0, true},
		{"inside window", []time.Duration{0}, 149 * ms, true},
		{"at timeout", []time.Duration{0}, holdTimeout, false},
		{"repeat refreshes", []time.Duration{0, 100 * ms}, 200 * ms, true},
		{"repeat expires", []time.Duration{0, 100 * ms}, 250 * ms, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := newInput(false)
			for _, at := range tt.presses {
				in.event(keyEvent{'w', kindPress}, t0.Add(at))
			}
			in.beginFrame(t0.Add(tt.frameAt))

			if got := in.IsDown('w'); got != tt.want {
				t.Errorf("wanted IsDown('w')=%v, got %v", tt.want, got)
			}
		})
	}
}

func TestInputIsDownPerKey(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(false)
	in.event(keyEvent{'a', kindPress}, t0)
	in.beginFrame(t0)

	if !in.IsDown('a') {
		t.Errorf("wanted IsDown('a')=true, got false")
	}
	if in.IsDown('b') {
		t.Errorf("wanted IsDown('b')=false, got true")
	}
}

func TestInputIsDownBeforeFirstFrame(t *testing.T) {
	// now and an unseen key are both the zero time here, so without the
	// lookup check every key would read as held.
	in := newInput(false)

	if in.IsDown('w') {
		t.Errorf("wanted IsDown('w')=false before any frame, got true")
	}
}

func TestInputPressed(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(false)
	in.event(keyEvent{'q', kindPress}, t0)
	in.beginFrame(t0)

	if !in.Pressed('q') {
		t.Fatalf("wanted Pressed('q')=true before endFrame, got false")
	}
	if in.Pressed('x') {
		t.Errorf("wanted Pressed('x')=false, got true")
	}

	in.endFrame()

	if in.Pressed('q') {
		t.Errorf("wanted Pressed('q')=false after endFrame, got true")
	}
	// Pressed is per frame, IsDown is time based; clearing one must not touch the other.
	if !in.IsDown('q') {
		t.Errorf("wanted IsDown('q')=true after endFrame, got false")
	}
}

func TestInputExactHoldsTwoKeys(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(true)

	// W goes down, then E: with exact state W stays held although only E repeats.
	in.event(keyEvent{'w', kindPress}, t0)
	in.event(keyEvent{'e', kindPress}, t0.Add(100*time.Millisecond))
	in.beginFrame(t0.Add(120 * time.Millisecond))
	in.endFrame()

	// Ten seconds later, with no further events, both are still down.
	in.beginFrame(t0.Add(10 * time.Second))
	if !in.IsDown('w') || !in.IsDown('e') {
		t.Errorf("wanted w and e both down, got w=%v e=%v", in.IsDown('w'), in.IsDown('e'))
	}
	if in.IsDown('x') {
		t.Errorf("wanted IsDown('x')=false, got true")
	}
}

func TestInputExactRelease(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(true)

	in.event(keyEvent{'w', kindPress}, t0)
	in.event(keyEvent{'e', kindPress}, t0)
	in.beginFrame(t0)
	in.endFrame()

	in.event(keyEvent{'w', kindRelease}, t0.Add(time.Second))
	in.beginFrame(t0.Add(time.Second))

	if in.IsDown('w') {
		t.Errorf("wanted IsDown('w')=false after release, got true")
	}
	if !in.IsDown('e') {
		t.Errorf("wanted IsDown('e')=true while still held, got false")
	}
}

func TestInputExactRepeatKeepsHeld(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(true)

	in.event(keyEvent{'w', kindPress}, t0)
	in.beginFrame(t0)
	in.endFrame()

	in.event(keyEvent{'w', kindRepeat}, t0.Add(500*time.Millisecond))
	in.beginFrame(t0.Add(500 * time.Millisecond))

	if !in.IsDown('w') {
		t.Errorf("wanted IsDown('w')=true on repeat, got false")
	}
	if !in.Pressed('w') {
		t.Errorf("wanted Pressed('w')=true on repeat, got false")
	}
}

// A tap can start and end inside one frame; Update must still see the key down once.
func TestInputExactSameFrameTap(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(true)

	in.event(keyEvent{'w', kindPress}, t0)
	in.event(keyEvent{'w', kindRelease}, t0.Add(10*time.Millisecond))
	in.beginFrame(t0.Add(16 * time.Millisecond))

	if !in.IsDown('w') {
		t.Errorf("wanted IsDown('w')=true for a tap inside the frame, got false")
	}

	in.endFrame()
	in.beginFrame(t0.Add(33 * time.Millisecond))

	if in.IsDown('w') {
		t.Errorf("wanted IsDown('w')=false the frame after the tap, got true")
	}
}

func TestInputReleaseAll(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(true)

	in.event(keyEvent{'w', kindPress}, t0)
	in.event(keyEvent{'e', kindPress}, t0)
	in.beginFrame(t0)
	in.endFrame()

	in.releaseAll()
	in.beginFrame(t0.Add(time.Second))

	if in.IsDown('w') || in.IsDown('e') {
		t.Errorf("wanted nothing down after releaseAll, got w=%v e=%v", in.IsDown('w'), in.IsDown('e'))
	}
}

// Without release events the old timing rule must stay in force.
func TestInputLegacyIgnoresHeld(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(false)

	in.event(keyEvent{'w', kindPress}, t0)
	in.beginFrame(t0.Add(10 * time.Second))

	if in.IsDown('w') {
		t.Errorf("wanted IsDown('w')=false long after a legacy press, got true")
	}
}

func TestInputFocusLostReleasesHeldKeys(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(true)

	in.event(keyEvent{'w', kindPress}, t0)
	in.event(keyEvent{'e', kindPress}, t0)
	in.beginFrame(t0)
	in.endFrame()

	in.event(keyEvent{KeyUnknown, kindFocusLost}, t0.Add(time.Second))
	in.beginFrame(t0.Add(time.Second))

	if in.IsDown('w') || in.IsDown('e') {
		t.Errorf("wanted nothing down after focus loss, got w=%v e=%v", in.IsDown('w'), in.IsDown('e'))
	}
}

func TestInputFocusGainedKeepsState(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput(true)

	in.event(keyEvent{'w', kindPress}, t0)
	in.beginFrame(t0)
	in.endFrame()

	in.event(keyEvent{KeyUnknown, kindFocusGained}, t0.Add(time.Second))
	in.beginFrame(t0.Add(time.Second))

	if !in.IsDown('w') {
		t.Errorf("wanted w still down after focus gain, got false")
	}
	if in.Pressed(KeyUnknown) || in.IsDown(KeyUnknown) {
		t.Errorf("wanted a focus event to leave no key state, got pressed=%v down=%v", in.Pressed(KeyUnknown), in.IsDown(KeyUnknown))
	}
}

// ---- the input feed ----

func newTestFeed(exact bool) (*inputFeed, *Input) {
	in := newInput(exact)
	f := newInputFeed(in)
	f.wait = 5 * time.Millisecond // short, so the tests can wait for it
	return f, in
}

// The other tests shorten the wait, so this is the only one that sees the real default.
func TestInputFeedWaitsLongEnoughForARealSequence(t *testing.T) {
	f := newInputFeed(newInput(false))
	if f.wait != escapeTimeout {
		t.Errorf("wanted new feed to wait escapeTimeout (%v), got %v", escapeTimeout, f.wait)
	}
	if escapeTimeout < 20*time.Millisecond {
		t.Errorf("wanted a wait that outlasts a slow link's gap between reads, got %v", escapeTimeout)
	}
}

func fired(f *inputFeed, within time.Duration) bool {
	select {
	case <-f.expired():
		return true
	case <-time.After(within):
		return false
	}
}

func TestInputFeedLoneEscapeBecomesEscapeAfterTheWait(t *testing.T) {
	f, in := newTestFeed(false)
	t0 := time.Unix(0, 0)

	if f.feed([]byte("\x1b"), t0) {
		t.Fatalf("wanted a lone ESC not to quit")
	}
	if in.Pressed(KeyEscape) {
		t.Fatalf("wanted Escape not pressed before the wait is over")
	}
	if !fired(f, time.Second) {
		t.Fatalf("wanted the wait to end")
	}

	if f.expire(t0.Add(50 * time.Millisecond)) {
		t.Fatalf("wanted Escape not to quit")
	}
	if !in.Pressed(KeyEscape) {
		t.Errorf("wanted Escape pressed once the wait is over")
	}
}

func TestInputFeedEscapeSequenceInTwoReadsIsNotEscape(t *testing.T) {
	f, in := newTestFeed(false)
	t0 := time.Unix(0, 0)

	f.feed([]byte("\x1b"), t0)
	f.feed([]byte("[A"), t0.Add(2*time.Millisecond))

	if !in.Pressed(KeyUp) {
		t.Errorf("wanted Up pressed")
	}
	if in.Pressed(KeyEscape) {
		t.Errorf("wanted Escape not pressed when the sequence completed")
	}
	if fired(f, 40*time.Millisecond) {
		t.Errorf("wanted the wait cancelled once the sequence completed")
	}
}

func TestInputFeedTimerOnlyRunsWhileSomethingIsHeld(t *testing.T) {
	f, _ := newTestFeed(false)
	t0 := time.Unix(0, 0)

	f.feed([]byte("wasd"), t0)
	if fired(f, 40*time.Millisecond) {
		t.Errorf("wanted no wait for complete input")
	}

	f.feed([]byte("\x1b[1"), t0)
	if !fired(f, time.Second) {
		t.Errorf("wanted a wait for an unfinished sequence")
	}
	f.expire(t0) // the wait ended, so the held bytes are given up on

	// Each read that leaves something held restarts the wait.
	f.feed([]byte("\x1b"), t0)
	f.feed([]byte("["), t0)
	if !fired(f, time.Second) {
		t.Errorf("wanted the wait to end after the last read")
	}
}

// A tick left over from an earlier wait must not end a later one early. Timers built
// for Go 1.23 or later cannot leave one behind; stop drains the channel for programs
// built with an older go.mod, which this toolchain can no longer emulate, so this
// guards the behavior rather than the drain.
func TestInputFeedStaleTickDoesNotEndALaterWait(t *testing.T) {
	f, _ := newTestFeed(false)
	f.wait = time.Millisecond
	t0 := time.Unix(0, 0)

	f.feed([]byte("\x1b"), t0)
	time.Sleep(20 * time.Millisecond) // the first wait ends and its tick sits in the channel

	f.wait = time.Hour
	f.feed([]byte("\x1b"), t0) // a new wait starts
	if fired(f, 30*time.Millisecond) {
		t.Errorf("wanted the old tick gone, but it ended the new wait")
	}
}

func TestInputFeedSplitKitlyReleaseIsNotStuck(t *testing.T) {
	f, in := newTestFeed(true)
	t0 := time.Unix(0, 0)

	f.feed([]byte("\x1b[119u"), t0)
	in.beginFrame(t0)
	if !in.IsDown('w') {
		t.Fatalf("wanted w held after the press")
	}
	in.endFrame()

	f.feed([]byte("\x1b[119;1"), t0.Add(time.Second))
	f.feed([]byte(":3u"), t0.Add(time.Second))
	in.beginFrame(t0.Add(time.Second))
	if in.IsDown('w') {
		t.Errorf("wanted w released by the release that arrived in two reads")
	}
}

func TestInputFeedDamagedStreamForgetsHeldKeys(t *testing.T) {
	f, in := newTestFeed(true)
	t0 := time.Unix(0, 0)

	f.feed([]byte("\x1b[119u\x1b[101u"), t0)
	in.beginFrame(t0)
	in.endFrame()

	// A release is cut off and its tail never comes.
	f.feed([]byte("\x1b[119;1"), t0)
	f.expire(t0.Add(50 * time.Millisecond))

	in.beginFrame(t0.Add(50 * time.Millisecond))
	if in.IsDown('w') || in.IsDown('e') {
		t.Errorf("wanted every key forgotten after a damaged stream, got w=%v e=%v", in.IsDown('w'), in.IsDown('e'))
	}
}

func TestInputFeedCtrlC(t *testing.T) {
	tests := []struct {
		name  string
		chunk string
		quit  bool
	}{
		{"legacy ctrl-c", "\x03", true},
		{"legacy ctrl-c among keys", "ab\x03cd", true},
		{"kitty ctrl-c press", "\x1b[99;5u", true},
		{"kitty ctrl-c repeat", "\x1b[99;5:2u", true},
		{"kitty ctrl-c release", "\x1b[99;5:3u", false},
		{"plain c", "c", false},
		{"ctrl-w", "\x1b[119;5u", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := newTestFeed(true)
			if got := f.feed([]byte(tt.chunk), time.Unix(0, 0)); got != tt.quit {
				t.Errorf("wanted quit=%v, got %v", tt.quit, got)
			}
		})
	}

	// Ctrl-C split across reads still quits.
	f, _ := newTestFeed(true)
	if f.feed([]byte("\x1b[99;"), time.Unix(0, 0)) {
		t.Fatalf("wanted no quit on half a sequence")
	}
	if !f.feed([]byte("5u"), time.Unix(0, 0)) {
		t.Errorf("wanted quit once the sequence completed")
	}
}

func TestInputFeedEventsKeepTheirReadTime(t *testing.T) {
	f, in := newTestFeed(false)
	t0 := time.Unix(100, 0)

	f.feed([]byte("w"), t0)
	in.beginFrame(t0.Add(100 * time.Millisecond))
	if !in.IsDown('w') {
		t.Errorf("wanted w down 100 ms after it arrived")
	}
	in.beginFrame(t0.Add(holdTimeout))
	if in.IsDown('w') {
		t.Errorf("wanted w no longer down once the hold timeout passed")
	}
}

func TestReadBlockingReadsBigChunksAtOnce(t *testing.T) {
	data := strings.Repeat("w", 1000)

	var chunks []int
	input, _ := readBlocking(strings.NewReader(data))
	for chunk := range input {
		chunks = append(chunks, len(chunk))
	}
	if len(chunks) != 1 || chunks[0] != 1000 {
		t.Errorf("wanted one chunk of 1000 bytes, got %v", chunks)
	}

	big := strings.Repeat("w", 3*readBuffer+17)
	total := 0
	input, _ = readBlocking(strings.NewReader(big))
	for chunk := range input {
		if len(chunk) > readBuffer {
			t.Errorf("wanted chunks of at most %d bytes, got %d", readBuffer, len(chunk))
		}
		total += len(chunk)
	}
	if total != len(big) {
		t.Errorf("wanted %d bytes in all, got %d", len(big), total)
	}
}

func TestReadBlockingStopFreesAChunkNobodyReceives(t *testing.T) {
	r, w := io.Pipe()
	input, stop := readBlocking(r)

	go w.Write([]byte("x")) // read, then stuck handing it over
	time.Sleep(50 * time.Millisecond)
	stop()
	stop() // safe to repeat

	select {
	case _, ok := <-input:
		if ok {
			t.Errorf("wanted the dropped chunk not to be delivered after stop")
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("wanted the channel closed once stop released the pending chunk")
	}
	w.Close()
}
