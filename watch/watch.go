// Package watch is the refresh loop behind `status --watch` in solmq-conn-util and
// solace-util. On a terminal each refresh redraws the screen in place; anywhere
// else (a pipe, a file, a console without escape processing) each refresh is
// appended under a timestamped section line, so no escape bytes reach a log. A
// refresh that fails is reported and retried. Ctrl-C ends the loop with success.
//
// Loop writes the header and then lets the refresh print as it goes. LoopFrames
// collects each refresh first and draws it only once it is complete, so the last
// frame stays on screen while the next is collected; use it when a refresh is slow.
package watch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	// ClearScreen moves the cursor home and clears the screen.
	ClearScreen = "\x1b[H\x1b[2J"

	// DefaultInterval, MinInterval and MaxInterval bound the refresh interval.
	DefaultInterval = 5 * time.Second
	MinInterval     = time.Second
	MaxInterval     = time.Hour

	timeLayout   = "2006-01-02 15:04:05"
	sectionWidth = 60
)

// CheckInterval refuses an interval outside MinInterval to MaxInterval.
func CheckInterval(d time.Duration) error {
	if d < MinInterval || d > MaxInterval {
		return fmt.Errorf("watch: the interval %s is outside %s to %s; pick a refresh interval in that range",
			d, MinInterval, MaxInterval)
	}
	return nil
}

// Options configure Loop and LoopFrames.
type Options struct {
	// Every is the pause after each refresh; it must pass CheckInterval.
	Every time.Duration

	// Title follows "Every <interval>: " in the redraw header, such as
	// "solace-util broker status --all".
	Title string

	// Redraw clears the screen before each refresh; otherwise each refresh is
	// appended under a section line. Pass CanRedraw(os.Stdout).
	Redraw bool

	// Out receives the header or section line. Under Loop the refresh writes its
	// own output; under LoopFrames Out receives each whole frame in one Write.
	Out io.Writer

	// Warn receives the message for a failed refresh, on the tool's own warning
	// channel. It must be set.
	Warn func(msg string)

	// Now is the clock, time.Now when nil.
	Now func() time.Time

	// Sleep pauses for d or until ctx is done, returning an error then; Sleep (the
	// function) when nil.
	Sleep func(ctx context.Context, d time.Duration) error
}

// Loop refreshes until ctx is done, which ends it with nil: run it under
// signal.NotifyContext(ctx, os.Interrupt), and Ctrl-C exits 0. Each pass writes the
// header (or the section line), calls refresh with the pass's time, and sleeps
// Every. A refresh error goes to Warn and the loop carries on; a refresh that
// reports its own failure returns nil, so nothing is printed twice. An error
// returned after ctx is done is the interruption itself and ends the loop quietly.
//
// Options that cannot work (an interval out of range, no Out, Warn or refresh) are
// an error before anything runs.
func Loop(ctx context.Context, o Options, refresh func(now time.Time) error) error {
	return loop(ctx, o, refresh != nil, func(t time.Time) error {
		io.WriteString(o.Out, o.head(t))
		return refresh(t)
	})
}

// LoopFrames is Loop for a refresh that writes its frame to the writer it is given
// instead of printing it. Use it instead of Loop when a refresh is slow, such as a
// status report that runs commands inside containers: Loop clears the screen
// before the refresh runs, so the screen sits blank for the whole collection. Loop
// stays the choice for a refresh that is quick or streams its own output.
//
// Each pass collects the refresh into a buffer and then writes the clear-screen,
// the header and the frame to Out in a single Write, or off a terminal the
// section line and the frame; the time shown is when the pass started. Nothing is
// written before the first refresh returns, and the last frame stays on screen
// until the next is complete. A refresh that returns once ctx is done was cut
// short, so it writes nothing at all, not even the header, and the loop ends with
// nil: Ctrl-C leaves the last complete frame in place. A refresh error is written
// first, with the header and whatever the refresh wrote before failing, and then
// goes to Warn, so no clear can wipe the message. Otherwise it is Loop: the same
// Options, checks, pause, retry and quiet end.
func LoopFrames(ctx context.Context, o Options, refresh func(now time.Time, frame io.Writer) error) error {
	var frame, screen bytes.Buffer
	return loop(ctx, o, refresh != nil, func(t time.Time) error {
		frame.Reset()
		err := refresh(t, &frame)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		screen.Reset()
		screen.WriteString(o.head(t))
		screen.Write(frame.Bytes())
		o.Out.Write(screen.Bytes())
		return err
	})
}

// loop checks o and then makes one pass per interval until ctx is done. set says
// whether the caller's refresh function is present. A pass error goes to Warn,
// unless ctx is done, which ends the loop with nil. The error text is trimmed
// first: one that carries a command's stderr often ends in a newline, which would
// push the retry onto a line of its own.
func loop(ctx context.Context, o Options, set bool, pass func(t time.Time) error) error {
	if err := CheckInterval(o.Every); err != nil {
		return err
	}
	if o.Out == nil || o.Warn == nil || !set {
		return errors.New("watch: Options.Out, Options.Warn and the refresh function must all be set")
	}
	now, sleep := o.Now, o.Sleep
	if now == nil {
		now = time.Now
	}
	if sleep == nil {
		sleep = Sleep
	}
	for {
		if err := pass(now()); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			o.Warn(fmt.Sprintf("%s -- trying again in %s", strings.TrimSpace(err.Error()), o.Every))
		}
		if sleep(ctx, o.Every) != nil {
			return nil
		}
	}
}

// head is the clear-screen and header for a pass started at t, or the section
// line when o does not redraw.
func (o Options) head(t time.Time) string {
	if o.Redraw {
		return fmt.Sprintf("%sEvery %s: %s   %s   Ctrl-C to stop\n\n", ClearScreen, o.Every, o.Title,
			t.Format(timeLayout))
	}
	return section(t.Format(timeLayout)) + "\n"
}

// section is "=== <title> " padded with '=' to sectionWidth.
func section(title string) string {
	head := "=== " + title + " "
	if pad := sectionWidth - len(head); pad > 0 {
		head += strings.Repeat("=", pad)
	}
	return head
}

// Sleep waits d, or until ctx is done, when it returns ctx's error.
func Sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// CanRedraw reports whether f is a terminal that can take ClearScreen: a character
// device and, on Windows, a console with escape processing switched on (this turns
// it on when it can). Anything else gets appended sections instead.
func CanRedraw(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	return enableVT(f)
}
