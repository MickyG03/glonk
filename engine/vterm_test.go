package engine

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// vcell is one cell of a virtual terminal: what is shown and the style it was
// written in.
type vcell struct {
	r  rune
	st style
}

// vterm is a small terminal for tests. It plays the bytes glonk writes onto a grid
// the way a real terminal would, so a test can ask what a person would see on screen
// instead of what bytes were sent. It knows only what glonk sends and fails the test
// on anything else, including a control character that was meant to be text.
//
// Autowrap is off, as glonk leaves it: a character written in the last column stays
// there and the cursor does not move on.
type vterm struct {
	t       testing.TB
	w, h    int
	grid    []vcell
	x, y    int
	pen     style
	pending []byte // the start of a sequence or character a Write left unfinished
	inSync  bool   // between ESC[?2026h and ESC[?2026l
	syncs   int    // how many synchronized updates have been completed
}

func newVterm(t testing.TB, w, h int) *vterm {
	v := &vterm{t: t, w: w, h: h, grid: make([]vcell, w*h)}
	v.eraseAll()
	return v
}

func (v *vterm) eraseAll() {
	for i := range v.grid {
		v.grid[i] = vcell{r: ' '}
	}
}

// at returns the cell at (x, y).
func (v *vterm) at(x, y int) vcell {
	return v.grid[y*v.w+x]
}

// Write plays p on the terminal. A sequence or character cut off at the end of p is
// kept and finished by the next Write.
func (v *vterm) Write(p []byte) (int, error) {
	v.t.Helper()
	data := append(v.pending, p...)
	v.pending = nil
	for len(data) > 0 {
		if data[0] == 0x1b {
			n, done := v.escape(data)
			if !done {
				break
			}
			data = data[n:]
			continue
		}
		if !utf8.FullRune(data) {
			break
		}
		r, size := utf8.DecodeRune(data)
		// U+FFFD itself is printable (and what a rune with no encoding is sent as); a
		// byte that is not UTF-8 at all decodes to it with a size of one.
		if r < 0x20 || r == 0x7f || (r == utf8.RuneError && size == 1) {
			v.t.Fatalf("wanted only printable text, got %q (%q)", r, data[:size])
		}
		v.put(r)
		data = data[size:]
	}
	v.pending = append([]byte(nil), data...)
	return len(p), nil
}

// put writes r at the cursor in the current style.
func (v *vterm) put(r rune) {
	v.grid[v.y*v.w+v.x] = vcell{r: r, st: v.pen}
	if v.x < v.w-1 {
		v.x++
	}
}

// escape plays the control sequence at the start of data. It reports how many bytes
// it used, or done=false when the sequence is not finished yet.
func (v *vterm) escape(data []byte) (n int, done bool) {
	v.t.Helper()
	if len(data) < 2 {
		return 0, false
	}
	if data[1] != '[' {
		v.t.Fatalf("wanted a CSI sequence, got %q", data)
	}
	i := 2
	for i < len(data) && data[i] >= 0x30 && data[i] <= 0x3f {
		i++
	}
	if i >= len(data) {
		return 0, false
	}
	params, final := string(data[2:i]), data[i]
	if final < 0x40 || final > 0x7e {
		v.t.Fatalf("wanted a final byte, got %q in %q", final, data[:i+1])
	}
	v.csi(params, final)
	return i + 1, true
}

// number reads the n-th (from 0) ';'-separated parameter, or def when it is empty.
func (v *vterm) number(params string, n, def int) int {
	v.t.Helper()
	parts := strings.Split(params, ";")
	if n >= len(parts) || parts[n] == "" {
		return def
	}
	val, err := strconv.Atoi(parts[n])
	if err != nil {
		v.t.Fatalf("wanted a number, got %q in %q", parts[n], params)
	}
	return val
}

func (v *vterm) csi(params string, final byte) {
	v.t.Helper()
	seq := fmt.Sprintf("ESC[%s%c", params, final)
	switch {
	case strings.HasPrefix(params, "?") && (final == 'h' || final == 'l'):
		mode, on := strings.TrimPrefix(params, "?"), final == 'h'
		switch mode {
		case "2026":
			if on == v.inSync {
				v.t.Fatalf("wanted synchronized updates not to nest or end unstarted, got %s", seq)
			}
			v.inSync = on
			if !on {
				v.syncs++
			}
		case "25", "7", "1049", "1004":
			// Cursor, autowrap, screen and focus modes: nothing a frame depends on.
		default:
			v.t.Fatalf("wanted a known mode, got %s", seq)
		}
	case params == "" || params[0] != '?':
		switch final {
		case 'H':
			row, col := v.number(params, 0, 1), v.number(params, 1, 1)
			v.y = min(max(row-1, 0), v.h-1)
			v.x = min(max(col-1, 0), v.w-1)
		case 'C':
			v.x = min(v.x+max(v.number(params, 0, 1), 1), v.w-1)
		case 'J':
			if v.number(params, 0, 0) != 2 {
				v.t.Fatalf("wanted only a whole-display erase, got %s", seq)
			}
			v.eraseAll()
		case 'm':
			v.pen = applySGR(v.t, v.pen, "\x1b["+params+"m")
		default:
			v.t.Fatalf("wanted a known sequence, got %s", seq)
		}
	default:
		v.t.Fatalf("wanted a known sequence, got %s", seq)
	}
}

// shownRune is the character a terminal ends up showing for a cell holding r: the safe
// character, and U+FFFD for a value that has no UTF-8 encoding at all.
func shownRune(r rune) rune {
	r = safeRune(r)
	if !utf8.ValidRune(r) {
		return utf8.RuneError
	}
	return r
}

// screenDiff compares what the terminal shows with what the frame should look like
// when its top left maxW x maxH cells are presented: each cell as the terminal would
// be sent it (a safe character, an RGB color as the nearest palette color), and
// everything outside that region blank. It returns a description of the first
// difference, or "" when they are the same.
func (v *vterm) screenDiff(fb *FrameBuffer, maxW, maxH int) string {
	regionW := min(fb.width, max(maxW, 0))
	regionH := min(fb.height, max(maxH, 0))
	for y := range v.h {
		for x := range v.w {
			want := vcell{r: ' '}
			if x < regionW && y < regionH {
				c := fb.frame[y*fb.width+x]
				want = vcell{r: shownRune(c.Character), st: asWritten(c.style())}
			}
			if got := v.at(x, y); got != want {
				return fmt.Sprintf("at (%d, %d): wanted %q in %+v, got %q in %+v", x, y, want.r, want.st, got.r, got.st)
			}
		}
	}
	return ""
}

func TestVtermPlaysCursorMovesAndText(t *testing.T) {
	v := newVterm(t, 5, 3)

	v.Write([]byte("\x1b[2;3Hab\x1b[1;1Hz\x1b[3;5Hq"))

	if got := []rune{v.at(2, 1).r, v.at(3, 1).r, v.at(0, 0).r, v.at(4, 2).r}; string(got) != "abzq" {
		t.Errorf("wanted abzq at the cursor positions, got %q", string(got))
	}
	if v.at(1, 1).r != ' ' {
		t.Errorf("wanted the cells nothing was written to to stay blank")
	}
}

func TestVtermForwardMoveStopsAtTheEdge(t *testing.T) {
	v := newVterm(t, 5, 1)

	v.Write([]byte("\x1b[1;1H\x1b[2Ca\x1b[9Cb\x1b[Cc"))

	if got := v.at(2, 0).r; got != 'a' {
		t.Errorf("wanted a two cells in, got %q", got)
	}
	// b went to the last column, where the cursor stays, so c overwrote it.
	if got := v.at(4, 0).r; got != 'c' {
		t.Errorf("wanted the last column to hold the last character written, got %q", got)
	}
}

func TestVtermLastColumnKeepsTheCursorThere(t *testing.T) {
	v := newVterm(t, 3, 2)

	v.Write([]byte("\x1b[1;1Habcdef"))

	if got := string([]rune{v.at(0, 0).r, v.at(1, 0).r, v.at(2, 0).r}); got != "abf" {
		t.Errorf("wanted autowrap off to overwrite the last column, got %q", got)
	}
	if v.at(0, 1).r != ' ' {
		t.Errorf("wanted nothing to wrap onto the next row")
	}
}

func TestVtermStyleSticksUntilChanged(t *testing.T) {
	v := newVterm(t, 6, 2)

	v.Write([]byte("\x1b[1;1H\x1b[1;38;5;208ma\x1b[2;1Hb\x1b[0mc"))

	want := style{fg: Palette(208), attrs: Bold}
	if v.at(0, 0).st != want || v.at(0, 1).st != want {
		t.Errorf("wanted the style to survive a cursor jump, got %+v and %+v", v.at(0, 0).st, v.at(0, 1).st)
	}
	if v.at(1, 1).st != (style{}) {
		t.Errorf("wanted a reset to end it, got %+v", v.at(1, 1).st)
	}
}

func TestVtermEraseKeepsTheCursorAndClearsTheCells(t *testing.T) {
	v := newVterm(t, 4, 2)
	v.Write([]byte("\x1b[1;1H\x1b[38;5;1;1mabcd"))
	v.pen = style{}

	v.Write([]byte("\x1b[2J"))

	if got := v.at(0, 0); got != (vcell{r: ' '}) {
		t.Errorf("wanted a blank cell after the erase, got %+v", got)
	}
	if v.x != 3 {
		t.Errorf("wanted the cursor left where it was, got column %d", v.x)
	}
}

func TestVtermFinishesASequenceCutByAWrite(t *testing.T) {
	v := newVterm(t, 8, 1)

	v.Write([]byte("\x1b[1;"))
	v.Write([]byte("3H\xe2\x94"))
	v.Write([]byte("\x80"))

	if got := v.at(2, 0).r; got != '─' {
		t.Errorf("wanted a box character at column 3, got %q", got)
	}
}

func TestVtermTracksSynchronizedUpdates(t *testing.T) {
	v := newVterm(t, 2, 1)

	v.Write([]byte("\x1b[?2026h"))
	if !v.inSync {
		t.Fatalf("wanted an update to be open")
	}
	v.Write([]byte("a\x1b[?2026l"))

	if v.inSync || v.syncs != 1 {
		t.Errorf("wanted one completed update and none open, got inSync=%v syncs=%d", v.inSync, v.syncs)
	}
}

func TestVtermScreenDiffSeesEveryKindOfDifference(t *testing.T) {
	fb, err := NewFrameBuffer(3, 2)
	if err != nil {
		t.Fatal(err)
	}
	fb.Set(1, 0, Cell{Character: 'x', Foreground: Palette(1), Attrs: Bold})
	fb.Set(2, 1, Cell{Character: '\x1b', Background: RGB(255, 0, 0)}) // shown as ?, red

	same := newVterm(t, 3, 2)
	fb.present(same)
	if d := same.screenDiff(fb, 3, 2); d != "" {
		t.Fatalf("wanted the presented frame to match itself, got %s", d)
	}

	for name, edit := range map[string]func(v *vterm){
		"character": func(v *vterm) { v.grid[1].r = 'y' },
		"colour":    func(v *vterm) { v.grid[1].st.fg = Palette(2) },
		"attribute": func(v *vterm) { v.grid[1].st.attrs = 0 },
		"stray":     func(v *vterm) { v.grid[0].r = 'z' },
	} {
		v := newVterm(t, 3, 2)
		fb.present(v)
		edit(v)
		if v.screenDiff(fb, 3, 2) == "" {
			t.Errorf("%s: wanted a difference to be found", name)
		}
	}
}
