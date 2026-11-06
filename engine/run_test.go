package engine

import "testing"

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
