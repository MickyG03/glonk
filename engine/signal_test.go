//go:build unix

package engine

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

const helperEnv = "GLONK_SIGNAL_HELPER"

// TestSignalHelper is not a test: the tests below run it in a child process,
// because the thing under test is how a whole process reacts to signals.
func TestSignalHelper(t *testing.T) {
	mode := os.Getenv(helperEnv)
	if mode == "" {
		t.Skip("only runs as a child of the signal tests")
	}

	ctx, cancel := context.WithCancel(context.Background())
	stop := notifyQuit(cancel)
	if mode == "stopped" {
		stop()
	}
	fmt.Println("ready")

	select {
	case <-ctx.Done():
		fmt.Println("cancelled")
	case <-time.After(time.Minute):
	}
	// A program stuck in shutdown: only a second signal can end it now.
	time.Sleep(time.Minute)
}

// child is a running helper process and the lines it printed.
type child struct {
	cmd   *exec.Cmd
	lines chan string
	done  chan error
}

// startChild runs the helper in the given mode. ignore lists signals to start it
// with ignored, the way nohup does.
func startChild(t *testing.T, mode string, ignore ...string) *child {
	t.Helper()

	args := []string{os.Args[0], "-test.run=^TestSignalHelper$"}
	cmd := exec.Command(args[0], args[1:]...)
	if len(ignore) > 0 {
		trap := "trap '' "
		for _, s := range ignore {
			trap += s + " "
		}
		// exec keeps ignored signals ignored.
		cmd = exec.Command("sh", append([]string{"-c", trap + `; exec "$0" "$@"`}, args...)...)
	}
	cmd.Env = append(os.Environ(), helperEnv+"="+mode)

	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	c := &child{cmd: cmd, lines: make(chan string, 16), done: make(chan error, 1)}
	go func() {
		sc := bufio.NewScanner(out)
		for sc.Scan() {
			c.lines <- sc.Text()
		}
		c.done <- cmd.Wait()
	}()
	t.Cleanup(func() { cmd.Process.Kill() })

	c.expect(t, "ready")
	return c
}

func (c *child) expect(t *testing.T, want string) {
	t.Helper()
	select {
	case got := <-c.lines:
		if got != want {
			t.Fatalf("wanted the child to print %q, got %q", want, got)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("wanted the child to print %q, got nothing in 5 s", want)
	}
}

// expectSilence fails if the child prints anything or exits within d.
func (c *child) expectSilence(t *testing.T, d time.Duration) {
	t.Helper()
	select {
	case got := <-c.lines:
		t.Fatalf("wanted the child to stay quiet, it printed %q", got)
	case err := <-c.done:
		t.Fatalf("wanted the child to keep running, it ended: %v", err)
	case <-time.After(d):
	}
}

// diedOf waits for the child to end and returns the signal that killed it.
func (c *child) diedOf(t *testing.T) syscall.Signal {
	t.Helper()
	select {
	case <-c.done:
	case <-time.After(5 * time.Second):
		t.Fatalf("wanted the child to end, still running after 5 s")
	}
	status, ok := c.cmd.ProcessState.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() {
		t.Fatalf("wanted the child killed by a signal, got %v", c.cmd.ProcessState)
	}
	return status.Signal()
}

func (c *child) signal(t *testing.T, s syscall.Signal) {
	t.Helper()
	if err := c.cmd.Process.Signal(s); err != nil {
		t.Fatal(err)
	}
}

func TestQuitSignalCancelsOnceThenTheNextOneKills(t *testing.T) {
	for _, s := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP} {
		t.Run(s.String(), func(t *testing.T) {
			c := startChild(t, "quit")

			c.signal(t, s)
			c.expect(t, "cancelled")
			c.expectSilence(t, 200*time.Millisecond) // a first signal does not kill

			c.signal(t, s)
			if got := c.diedOf(t); got != s {
				t.Errorf("wanted the second %v to kill the stuck program, it died of %v", s, got)
			}
		})
	}
}

func TestQuitSignalsLeaveAnIgnoredSignalIgnored(t *testing.T) {
	c := startChild(t, "quit", "HUP")

	c.signal(t, syscall.SIGHUP)
	c.expectSilence(t, 300*time.Millisecond) // nohup'd: HUP means nothing

	c.signal(t, syscall.SIGTERM) // the others still work
	c.expect(t, "cancelled")
}

func TestQuitSignalsLeaveSigquitAlone(t *testing.T) {
	// SIGQUIT is how a developer asks a stuck Go program for its stack, so it keeps
	// Go's default (a dump and exit status 2), as the Run docs promise.
	c := startChild(t, "quit")

	c.signal(t, syscall.SIGQUIT)
	select {
	case <-c.done:
	case <-time.After(5 * time.Second):
		t.Fatalf("wanted SIGQUIT to end the child, still running after 5 s")
	}
	if got := c.cmd.ProcessState.ExitCode(); got != 2 {
		t.Errorf("wanted SIGQUIT to end the child with Go's stack dump (exit 2), got %v", c.cmd.ProcessState)
	}
}

func TestQuitSignalsStopWhenToldTo(t *testing.T) {
	c := startChild(t, "stopped")

	c.signal(t, syscall.SIGTERM)
	if got := c.diedOf(t); got != syscall.SIGTERM {
		t.Errorf("wanted SIGTERM to kill once stop was called, it died of %v", got)
	}
}
