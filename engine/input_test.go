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
			in := newInput()
			for _, at := range tt.presses {
				in.press('w', t0.Add(at))
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
	in := newInput()
	in.press('a', t0)
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
	in := newInput()

	if in.IsDown('w') {
		t.Errorf("wanted IsDown('w')=false before any frame, got true")
	}
}

func TestInputPressed(t *testing.T) {
	t0 := time.Unix(0, 0)
	in := newInput()
	in.press('q', t0)
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
