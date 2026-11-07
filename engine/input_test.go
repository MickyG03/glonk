package engine

import (
	"bytes"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func TestReadInput(t *testing.T) {
	r := iotest.DataErrReader(iotest.OneByteReader(strings.NewReader("ab\x03")))
	var chunks [][]byte
	for chunk := range readInput(r) {
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
