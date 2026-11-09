package engine

import "testing"

func env(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func TestDetectTrueColor(t *testing.T) {
	tests := []struct {
		name string
		vars map[string]string
		want bool
	}{
		{"COLORTERM truecolor", map[string]string{"COLORTERM": "truecolor"}, true},
		{"COLORTERM 24bit", map[string]string{"COLORTERM": "24bit"}, true},
		{"COLORTERM in capitals", map[string]string{"COLORTERM": "TrueColor"}, true},
		{"COLORTERM with a 256-color TERM", map[string]string{"COLORTERM": "truecolor", "TERM": "xterm-256color"}, true},
		{"TERM direct", map[string]string{"TERM": "xterm-direct"}, true},
		{"TERM truecolor", map[string]string{"TERM": "screen-truecolor"}, true},
		{"TERM 24bit", map[string]string{"TERM": "xterm-24bit"}, true},
		{"TERM kitty alone does not say", map[string]string{"TERM": "xterm-kitty"}, false},
		{"nothing set", map[string]string{}, false},
		{"256 colors", map[string]string{"TERM": "xterm-256color"}, false},
		{"COLORTERM yes means a few colors, not 24-bit", map[string]string{"COLORTERM": "yes"}, false},
		{"COLORTERM 256color", map[string]string{"COLORTERM": "256color"}, false},
		{"COLORTERM empty", map[string]string{"COLORTERM": ""}, false},
		{"dumb", map[string]string{"TERM": "dumb"}, false},
		{"a word containing direct but not ending in it", map[string]string{"TERM": "directx-256color"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectTrueColor(env(tt.vars)); got != tt.want {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

func TestColorModeDecidesWhetherToWriteTrueColor(t *testing.T) {
	truecolor := env(map[string]string{"COLORTERM": "truecolor"})
	plain := env(map[string]string{"TERM": "xterm-256color"})
	tests := []struct {
		name     string
		mode     ColorMode
		headless bool
		vars     func(string) string
		want     bool
	}{
		{"auto on a true color terminal", ColorsAuto, false, truecolor, true},
		{"auto on a 256-color terminal", ColorsAuto, false, plain, false},
		{"auto without a terminal ignores the environment", ColorsAuto, true, truecolor, false},
		{"256 colors forced on a true color terminal", Colors256, false, truecolor, false},
		{"true color forced on a plain terminal", ColorsTrue, false, plain, true},
		{"true color forced without a terminal", ColorsTrue, true, plain, true},
		{"256 colors forced without a terminal", Colors256, true, truecolor, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.useTrueColor(tt.headless, tt.vars); got != tt.want {
				t.Errorf("wanted %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCheckConfigRejectsAColorModeThatDoesNotExist(t *testing.T) {
	for _, mode := range []ColorMode{ColorsAuto, Colors256, ColorsTrue} {
		if err := checkConfig(Config{FPS: 30, Colors: mode}); err != nil {
			t.Errorf("mode %d: wanted it accepted, got %v", mode, err)
		}
	}
	if err := checkConfig(Config{FPS: 30, Colors: ColorsTrue + 1}); err == nil {
		t.Errorf("wanted an unknown mode refused")
	}
}
