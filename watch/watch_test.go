package watch

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// clock starts at a fixed time and moves on by the interval each call.
func clock() func() time.Time {
	t := time.Date(2026, 11, 4, 9, 30, 0, 0, time.UTC)
	return func() time.Time {
		now := t
		t = t.Add(time.Second)
		return now
	}
}

// stopAfter is a Sleep seam that records each pause and cancels after n of them.
func stopAfter(n int, cancel context.CancelFunc, slept *[]time.Duration) func(context.Context, time.Duration) error {
	return func(ctx context.Context, d time.Duration) error {
		*slept = append(*slept, d)
		if len(*slept) >= n {
			cancel()
			return ctx.Err()
		}
		return nil
	}
}

// run drives Loop for n refreshes and returns what it wrote, what it warned and the
// pauses.
func run(t *testing.T, redraw bool, n int, refresh func(time.Time) error) (string, []string, []time.Duration) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out bytes.Buffer
	var warned []string
	var slept []time.Duration
	err := Loop(ctx, Options{
		Every:  time.Second,
		Title:  "solace-util broker status --all",
		Redraw: redraw,
		Out:    &out,
		Warn:   func(msg string) { warned = append(warned, msg) },
		Now:    clock(),
		Sleep:  stopAfter(n, cancel, &slept),
	}, refresh)
	if err != nil {
		t.Fatalf("Loop = %v, want nil when the context ends", err)
	}
	return out.String(), warned, slept
}

// TestLoopRedraws carries solace-util's TestWatchRedrawsInPlaceOnATerminal: each
// refresh clears the screen and writes the header, with the refresh's own time.
func TestLoopRedraws(t *testing.T) {
	var seen []time.Time
	out, warned, slept := run(t, true, 2, func(now time.Time) error {
		seen = append(seen, now)
		return nil
	})
	want := ClearScreen + "Every 1s: solace-util broker status --all   2026-11-04 09:30:00   Ctrl-C to stop\n\n" +
		ClearScreen + "Every 1s: solace-util broker status --all   2026-11-04 09:30:01   Ctrl-C to stop\n\n"
	if out != want {
		t.Errorf("out =\n%q\nwant\n%q", out, want)
	}
	if len(seen) != 2 || !seen[1].Equal(seen[0].Add(time.Second)) {
		t.Errorf("refresh times = %v, want each pass's own time", seen)
	}
	if len(warned) != 0 || len(slept) != 2 || slept[0] != time.Second {
		t.Errorf("warned %v, slept %v; want no warnings and two 1s pauses", warned, slept)
	}
}

// TestLoopAppendsSections carries solace-util's TestWatchAppendsWhenNotATerminal and
// TestWatchNeverRedrawsIntoAPipe: no escape bytes, a 60-column section line per
// refresh.
func TestLoopAppendsSections(t *testing.T) {
	out, _, _ := run(t, false, 2, func(time.Time) error { return nil })
	if strings.Contains(out, "\x1b") {
		t.Errorf("an escape byte reached a pipe: %q", out)
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 2 || lines[0] != "=== 2026-11-04 09:30:00 "+strings.Repeat("=", 36) || len(lines[1]) != 60 {
		t.Errorf("sections = %q, want one 60-column line per refresh", lines)
	}
	if got := section(strings.Repeat("x", 70)); got != "=== "+strings.Repeat("x", 70)+" " {
		t.Errorf("a long title is not padded: %q", got)
	}
}

// TestLoopRetriesAfterAFailedRefresh carries solace-util's
// TestWatchKeepsGoingAfterAFailedRefresh: the failure is warned with the retry
// interval and the next refresh still runs.
func TestLoopRetriesAfterAFailedRefresh(t *testing.T) {
	calls := 0
	_, warned, _ := run(t, false, 2, func(time.Time) error {
		calls++
		if calls == 1 {
			return errors.New("engine unreachable")
		}
		return nil
	})
	if calls != 2 || len(warned) != 1 || warned[0] != "engine unreachable -- trying again in 1s" {
		t.Errorf("calls %d, warned %q; want a warning and a second refresh", calls, warned)
	}
}

// TestLoopEndsQuietlyOnCancel carries solace-util's
// TestWatchEndsQuietlyWhenInterrupted: a refresh cut short by Ctrl-C returns nil
// with no retry message, and nothing is printed on the way out.
func TestLoopEndsQuietlyOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var out bytes.Buffer
	var warned []string
	err := Loop(ctx, Options{Every: time.Second, Out: &out, Warn: func(m string) { warned = append(warned, m) }},
		func(time.Time) error {
			cancel()
			return context.Canceled
		})
	if err != nil || len(warned) != 0 {
		t.Errorf("Loop = %v, warned %q; want nil and no warning", err, warned)
	}
	if !strings.HasPrefix(out.String(), "=== ") || strings.Count(out.String(), "\n") != 1 {
		t.Errorf("out = %q, want exactly one section line and nothing after", out.String())
	}
}

// TestDefaultSleepHonoursCancel carries solace-util's TestWatchSleepHonoursCtrlC:
// the default pause ends at once when the context does, and runs out otherwise.
func TestDefaultSleepHonoursCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if err := Sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) || time.Since(start) > time.Second {
		t.Errorf("Sleep on a done context = %v after %s, want Canceled at once", err, time.Since(start))
	}
	if err := Sleep(context.Background(), time.Millisecond); err != nil {
		t.Errorf("Sleep that runs out = %v, want nil", err)
	}
	// Loop falls back to it when no Sleep is given.
	ctx, cancel = context.WithCancel(context.Background())
	var out bytes.Buffer
	err := Loop(ctx, Options{Every: time.Second, Out: &out, Warn: func(string) {}}, func(time.Time) error {
		cancel()
		return nil
	})
	if err != nil {
		t.Errorf("Loop = %v, want nil once the default sleep sees the cancel", err)
	}
}

func TestLoopRefusesBadOptions(t *testing.T) {
	ok := func(time.Time) error {
		t.Error("refresh must not run with options that cannot work")
		return nil
	}
	var out bytes.Buffer
	warn := func(string) {}
	for name, c := range map[string]struct {
		o       Options
		refresh func(time.Time) error
	}{
		"interval below 1s": {Options{Every: 500 * time.Millisecond, Out: &out, Warn: warn}, ok},
		"no Out":            {Options{Every: time.Second, Warn: warn}, ok},
		"no Warn":           {Options{Every: time.Second, Out: &out}, ok},
		"no refresh":        {Options{Every: time.Second, Out: &out, Warn: warn}, nil},
	} {
		if err := Loop(context.Background(), c.o, c.refresh); err == nil {
			t.Errorf("%s: Loop must refuse", name)
		}
	}
	if out.Len() != 0 {
		t.Errorf("nothing may be written before the options are checked: %q", out.String())
	}
}

func TestCheckInterval(t *testing.T) {
	for d, ok := range map[time.Duration]bool{
		time.Second: true, DefaultInterval: true, time.Hour: true,
		999 * time.Millisecond: false, 0: false, -time.Second: false, time.Hour + time.Second: false,
	} {
		if err := CheckInterval(d); (err == nil) != ok {
			t.Errorf("CheckInterval(%s) = %v, want ok %v", d, err, ok)
		}
	}
	if err := CheckInterval(0); !strings.Contains(err.Error(), "1s to 1h0m0s") {
		t.Errorf("err = %v, want it to name the range", err)
	}
}

// TestCanRedraw: a regular file is never redrawn; a character device (the null
// device on Unix) is, once its console takes escape sequences. On Windows the null
// device is a character device but no console, so enableVT refuses it and the loop
// appends.
func TestCanRedraw(t *testing.T) {
	f, err := os.Create(filepath.Join(t.TempDir(), "out"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if CanRedraw(f) {
		t.Error("a regular file must get appended sections, not redraws")
	}
	if CanRedraw(nil) {
		t.Error("no file, no redraw")
	}
	closed, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()
	if CanRedraw(closed) {
		t.Error("a file that cannot be asked must not be redrawn")
	}
	null, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer null.Close()
	if fi, _ := null.Stat(); fi.Mode()&os.ModeCharDevice != 0 {
		if want := runtime.GOOS != "windows"; CanRedraw(null) != want {
			t.Errorf("CanRedraw(%s) = %v on %s, want %v", os.DevNull, !want, runtime.GOOS, want)
		}
	}
}
