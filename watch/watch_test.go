package watch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
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

// TestLoopTrimsTheFailureBeforeTheRetry: an error that carries a command's stderr
// ends in a newline, which is trimmed so the retry stays on the same line.
func TestLoopTrimsTheFailureBeforeTheRetry(t *testing.T) {
	_, warned, _ := run(t, false, 1, func(time.Time) error {
		return errors.New("inspecting solmq-connector: engine down\n")
	})
	if len(warned) != 1 || warned[0] != "inspecting solmq-connector: engine down -- trying again in 1s" {
		t.Errorf("warned %q, want the error trimmed and the retry on its line", warned)
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

// writes keeps each Write as its own entry, and adds it to log when log is set,
// so a test can count the writes a frame took and see where Warn fell among them.
type writes struct {
	calls []string
	log   *[]string
}

func (w *writes) Write(p []byte) (int, error) {
	w.calls = append(w.calls, string(p))
	if w.log != nil {
		*w.log = append(*w.log, "out: "+string(p))
	}
	return len(p), nil
}

// String is everything written so far, which is what a terminal would show.
func (w *writes) String() string { return strings.Join(w.calls, "") }

// runFrames drives LoopFrames for n refreshes and returns each write, what it
// warned and the pauses.
func runFrames(t *testing.T, redraw bool, n int, refresh func(time.Time, io.Writer) error) ([]string, []string, []time.Duration) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out writes
	var warned []string
	var slept []time.Duration
	err := LoopFrames(ctx, Options{
		Every:  time.Second,
		Title:  "solace-util broker status --all",
		Redraw: redraw,
		Out:    &out,
		Warn:   func(msg string) { warned = append(warned, msg) },
		Now:    clock(),
		Sleep:  stopAfter(n, cancel, &slept),
	}, refresh)
	if err != nil {
		t.Fatalf("LoopFrames = %v, want nil when the context ends", err)
	}
	return out.calls, warned, slept
}

// TestLoopFramesRedraws mirrors TestLoopRedraws: each frame is the clear, the
// header with the time its pass started and the report, in a single write.
func TestLoopFramesRedraws(t *testing.T) {
	var seen []time.Time
	calls, warned, slept := runFrames(t, true, 2, func(now time.Time, frame io.Writer) error {
		seen = append(seen, now)
		fmt.Fprintf(frame, "report %d\n", len(seen))
		return nil
	})
	want := []string{
		ClearScreen + "Every 1s: solace-util broker status --all   2026-11-04 09:30:00   Ctrl-C to stop\n\nreport 1\n",
		ClearScreen + "Every 1s: solace-util broker status --all   2026-11-04 09:30:01   Ctrl-C to stop\n\nreport 2\n",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("writes =\n%q\nwant one write per frame:\n%q", calls, want)
	}
	if len(seen) != 2 || !seen[1].Equal(seen[0].Add(time.Second)) {
		t.Errorf("refresh times = %v, want each pass's own time", seen)
	}
	if len(warned) != 0 || len(slept) != 2 || slept[0] != time.Second {
		t.Errorf("warned %v, slept %v; want no warnings and two 1s pauses", warned, slept)
	}
}

// TestLoopFramesAppendsSections mirrors TestLoopAppendsSections: off a terminal
// each write is the 60-column section line and the report, with no escape byte.
func TestLoopFramesAppendsSections(t *testing.T) {
	calls, _, _ := runFrames(t, false, 2, func(_ time.Time, frame io.Writer) error {
		io.WriteString(frame, "report\n")
		return nil
	})
	if len(calls) != 2 || calls[0] != "=== 2026-11-04 09:30:00 "+strings.Repeat("=", 36)+"\nreport\n" {
		t.Fatalf("writes = %q, want one section line and report per write", calls)
	}
	head, rest, _ := strings.Cut(calls[1], "\n")
	if strings.Contains(calls[0]+calls[1], "\x1b") || len(head) != 60 || rest != "report\n" {
		t.Errorf("second write = %q, want a 60-column section line, the report and no escape byte", calls[1])
	}
}

// TestLoopFramesKeepsTheLastFrameUntilTheNextIsCollected carries solmq-conn-util's
// TestWatchStatusKeepsTheLastReportUntilTheNextIsCollected: nothing is written
// before the first refresh returns, and while the second runs the screen holds
// exactly the first frame.
func TestLoopFramesKeepsTheLastFrameUntilTheNextIsCollected(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var out writes
	var slept []time.Duration
	var onScreen []string
	var writesSoFar []int
	err := LoopFrames(ctx, Options{
		Every: time.Second, Title: "t", Redraw: true, Out: &out, Warn: func(string) {},
		Now: clock(), Sleep: stopAfter(2, cancel, &slept),
	}, func(_ time.Time, frame io.Writer) error {
		onScreen = append(onScreen, out.String())
		writesSoFar = append(writesSoFar, len(out.calls))
		fmt.Fprintf(frame, "report %d\n", len(onScreen))
		return nil
	})
	if err != nil || len(onScreen) != 2 {
		t.Fatalf("LoopFrames = %v after %d refreshes, want nil after 2", err, len(onScreen))
	}
	if onScreen[0] != "" {
		t.Errorf("nothing may be written before the first refresh returns, got %q", onScreen[0])
	}
	if writesSoFar[1] != 1 || strings.Count(onScreen[1], ClearScreen) != 1 || !strings.HasSuffix(onScreen[1], "report 1\n") {
		t.Errorf("during the second refresh the screen should be exactly the first frame, got %d writes:\n%q",
			writesSoFar[1], onScreen[1])
	}
}

// TestLoopFramesWarnsAfterTheHeader mirrors TestLoopRetriesAfterAFailedRefresh: a
// failed refresh writes its header, with what it wrote before failing, and only
// then warns, so the clear cannot wipe the warning; the next refresh still runs.
func TestLoopFramesWarnsAfterTheHeader(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var events []string
	out := writes{log: &events}
	var slept []time.Duration
	calls := 0
	err := LoopFrames(ctx, Options{
		Every: time.Second, Title: "t", Redraw: true, Out: &out,
		Warn: func(msg string) { events = append(events, "warn: "+msg) },
		Now:  clock(), Sleep: stopAfter(2, cancel, &slept),
	}, func(_ time.Time, frame io.Writer) error {
		calls++
		if calls == 1 {
			io.WriteString(frame, "partial\n")
			return errors.New("engine unreachable")
		}
		io.WriteString(frame, "report\n")
		return nil
	})
	want := []string{
		"out: " + ClearScreen + "Every 1s: t   2026-11-04 09:30:00   Ctrl-C to stop\n\npartial\n",
		"warn: engine unreachable -- trying again in 1s",
		"out: " + ClearScreen + "Every 1s: t   2026-11-04 09:30:01   Ctrl-C to stop\n\nreport\n",
	}
	if err != nil || calls != 2 || !reflect.DeepEqual(events, want) {
		t.Errorf("LoopFrames = %v after %d refreshes, events =\n%q\nwant\n%q", err, calls, events, want)
	}
}

// TestLoopFramesDrawsNothingForARefreshCutShort carries solmq-conn-util's
// TestWatchStatusDrawsNothingForARefreshCutShort and mirrors
// TestLoopEndsQuietlyOnCancel: a refresh that returns once the context is done
// writes nothing, not even the header, warns nothing and ends the loop with nil
// without another pause, whether it returns nil or an error.
func TestLoopFramesDrawsNothingForARefreshCutShort(t *testing.T) {
	for name, ret := range map[string]error{"returns nil": nil, "returns an error": context.Canceled} {
		ctx, cancel := context.WithCancel(context.Background())
		var out writes
		var warned []string
		slept, calls := 0, 0
		err := LoopFrames(ctx, Options{
			Every: time.Second, Redraw: true, Out: &out,
			Warn: func(m string) { warned = append(warned, m) },
			Sleep: func(context.Context, time.Duration) error {
				slept++
				return errors.New("no pause expected")
			},
		}, func(_ time.Time, frame io.Writer) error {
			calls++
			io.WriteString(frame, "half a report")
			cancel()
			return ret
		})
		cancel()
		if err != nil || calls != 1 || slept != 0 || len(warned) != 0 {
			t.Errorf("%s: LoopFrames = %v after %d refreshes and %d pauses, warned %q; want nil, 1, 0 and none",
				name, err, calls, slept, warned)
		}
		if len(out.calls) != 0 {
			t.Errorf("%s: a refresh cut short must write nothing, got %q", name, out.calls)
		}
	}
}

// TestLoopFramesRefusesBadOptions mirrors TestLoopRefusesBadOptions.
func TestLoopFramesRefusesBadOptions(t *testing.T) {
	ok := func(time.Time, io.Writer) error {
		t.Error("refresh must not run with options that cannot work")
		return nil
	}
	var out bytes.Buffer
	warn := func(string) {}
	for name, c := range map[string]struct {
		o       Options
		refresh func(time.Time, io.Writer) error
	}{
		"interval below 1s": {Options{Every: 500 * time.Millisecond, Out: &out, Warn: warn}, ok},
		"no Out":            {Options{Every: time.Second, Warn: warn}, ok},
		"no Warn":           {Options{Every: time.Second, Out: &out}, ok},
		"no refresh":        {Options{Every: time.Second, Out: &out, Warn: warn}, nil},
	} {
		if err := LoopFrames(context.Background(), c.o, c.refresh); err == nil {
			t.Errorf("%s: LoopFrames must refuse", name)
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
