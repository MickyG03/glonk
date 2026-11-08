package engine

import (
	"math"
	"testing"
	"time"
)

func TestResolveSize(t *testing.T) {
	tests := []struct {
		name                   string
		cfgW, cfgH, trmW, trmH int
		wantW, wantH           int
	}{
		{"both fixed", 64, 24, 80, 40, 64, 24},
		{"both fill", 0, 0, 80, 40, 80, 40},
		{"width fixed only", 64, 0, 80, 40, 64, 40},
		{"height fixed only", 0, 24, 80, 40, 80, 24},
		{"terminal reports zero", 0, 0, 0, 0, 0, 0},
		// NewFrameBuffer rejects negatives; they must reach it, not be replaced by the terminal's size.
		{"negative passes through", -5, 0, 80, 40, -5, 40},
		{"negative height passes through", 0, -5, 80, 40, 80, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := resolveSize(tt.cfgW, tt.cfgH, tt.trmW, tt.trmH)

			if w != tt.wantW || h != tt.wantH {
				t.Errorf("wanted (%d,%d), got (%d,%d)", tt.wantW, tt.wantH, w, h)
			}
		})
	}
}

func TestFrameDT(t *testing.T) {
	t0 := time.Unix(0, 0)

	tests := []struct {
		name string
		gap  time.Duration
		want float64
	}{
		{"normal frame", 33 * time.Millisecond, 0.033},
		{"no time passed", 0, 0},
		{"exactly at the cap", 250 * time.Millisecond, maxDT},
		{"just over the cap", 251 * time.Millisecond, maxDT},
		{"suspended for ten minutes", 10 * time.Minute, maxDT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frameDT(t0.Add(tt.gap), t0)

			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCheckConfigFPS(t *testing.T) {
	tests := []struct {
		fps     int
		wantErr bool
	}{
		{math.MinInt, true},
		{-1, true},
		{0, true},
		{1, false},
		{30, false},
		{60, false},
		{maxFPS, false},
		{maxFPS + 1, true},
		{2_000_000_000, true},
		{math.MaxInt, true},
	}

	for _, tt := range tests {
		err := checkConfig(Config{FPS: tt.fps})
		if (err != nil) != tt.wantErr {
			t.Errorf("fps %d: wanted error=%v, got %v", tt.fps, tt.wantErr, err)
		}
	}
}

func TestCheckConfigNamesTheRange(t *testing.T) {
	err := checkConfig(Config{FPS: 5000})
	if err == nil {
		t.Fatalf("wanted an error for 5000 fps")
	}
	if want := "fps must be between 1 and 1000, got 5000"; err.Error() != want {
		t.Errorf("wanted %q, got %q", want, err)
	}
}

// Every accepted rate must give the frame timer a positive interval; a zero one
// makes time.NewTicker panic.
func TestEveryAcceptedFPSHasAPositiveInterval(t *testing.T) {
	for fps := 1; fps <= maxFPS; fps++ {
		if time.Second/time.Duration(fps) <= 0 {
			t.Fatalf("wanted a positive interval for %d fps", fps)
		}
	}
}
