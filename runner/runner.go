// Package runner is the one seam every external command crosses: kubectl, oc, helm,
// docker, podman, systemctl and the rest. Builders return a Cmd; only code that holds a
// Runner executes one. Exec is the real Runner, Echo prints what would run, and
// runnertest.Fake records calls for tests.
//
// A Runner has two methods. Run executes a command and hands its two output streams to
// the writers the caller passes. Attach hands a terminal to an interactive session.
// Every capture mode is a helper over Run (Combined, Split, Output), so a fake
// implements two methods and supports all of them.
//
// How each tool's old calls map onto this one:
//
//	solmq-conn-util  Run(cmd)             -> Combined(ctx, r, cmd)
//	                 RunSplit(cmd)        -> Split(ctx, r, cmd)
//	                 Stream(ctx, cmd, ..) -> r.Run(ctx, cmd, stdout, stderr)
//	                 Attach(cmd, files)   -> r.Attach(cmd, files)
//	solace-util      Run, RunInput, RunEnv -> r.Run(ctx, cmd, os.Stdout, os.Stderr),
//	                                         with Cmd.Stdin or Cmd.Env set
//	                 Output, OutputInput   -> Output(ctx, r, cmd, os.Stderr)
//	                 RunInteractive        -> r.Attach(cmd, os.Stdin, os.Stdout, os.Stderr)
//
// Cancellation. Exec interrupts a running command when its context is done, and kills
// it if it is still running when the grace period ends. A cancelled run is reported as
// an error wrapping ctx.Err(), never as success: whether an operator's Ctrl-C was a
// clean end (a log follow) or a failure (a deploy) is the caller's call. Attach takes no
// context: the operator's terminal owns an interactive session, and the runner never
// interrupts or kills one.
//
// Nothing here prints an argument, the stdin bytes or an environment value in an error,
// and Echo masks every environment value. Output policy -- a final ok/error line, a
// verbose trail of commands -- is the tool's, built on Announce and the return values.
package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hafio/go-deploy-utility/exitcode"
)

// Cmd is one external command: what a builder returns and a Runner executes.
type Cmd struct {
	// Argv is the program and its arguments. Argv[0] is resolved through Resolve
	// immediately before it runs. There is no shell: each element is one argument.
	Argv []string

	// Stdin, when non-nil, is fed to the child's standard input; even an empty,
	// non-nil slice is fed (the child reads EOF at once). nil means no input. It is
	// how a secret reaches a child without passing through an argv, so it is never
	// printed: Echo shows only its size.
	Stdin []byte

	// Env holds extra KEY=VALUE entries appended to this process's environment for
	// the child only; a later entry wins over an inherited one of the same name. The
	// values are secrets as often as not, so they are never printed: Echo shows the
	// keys with every value masked.
	Env []string
}

// Check reports a Cmd no runner may execute: an empty Argv, an empty program name, or an
// Env entry that is not KEY=VALUE. Every Runner calls it first, so a builder bug fails
// the same way under a test fake as in production.
func (c Cmd) Check() error {
	if len(c.Argv) == 0 {
		return errors.New("runner: Cmd.Argv is empty; a builder must name the program in Argv[0]")
	}
	if c.Argv[0] == "" {
		return errors.New("runner: Cmd.Argv[0] is empty; a builder must name the program in Argv[0]")
	}
	for i, e := range c.Env {
		if strings.IndexByte(e, '=') <= 0 {
			return fmt.Errorf("runner: %s: Cmd.Env[%d] is not KEY=VALUE (the entry is not shown, in case it is a "+
				"secret value); add environment entries as KEY=VALUE", c.Argv[0], i)
		}
	}
	return nil
}

// CheckRun is Check plus the writer rule Run relies on: a typed-nil *os.File would reach
// the child as a broken descriptor rather than the null device a nil io.Writer gives.
func CheckRun(c Cmd, stdout, stderr io.Writer) error {
	if err := c.Check(); err != nil {
		return err
	}
	for _, w := range []struct {
		name string
		w    io.Writer
	}{{"stdout", stdout}, {"stderr", stderr}} {
		if f, ok := w.w.(*os.File); ok && f == nil {
			return fmt.Errorf("runner: %s: the %s writer is a nil *os.File; pass a nil io.Writer to discard the "+
				"stream, or an open file", c.Argv[0], w.name)
		}
	}
	return nil
}

// CheckAttach is Check plus what an attached session needs: no Stdin bytes, because the
// session reads the terminal, and all three files.
func CheckAttach(c Cmd, stdin, stdout, stderr *os.File) error {
	if err := c.Check(); err != nil {
		return err
	}
	if c.Stdin != nil {
		return fmt.Errorf("runner: %s: an attached session takes the terminal's own stdin, so Cmd.Stdin must be nil",
			c.Argv[0])
	}
	if stdin == nil || stdout == nil || stderr == nil {
		return fmt.Errorf("runner: %s: an attached session needs all three of stdin, stdout and stderr", c.Argv[0])
	}
	return nil
}

// Runner executes Cmds. Exec, Echo and runnertest.Fake implement it.
type Runner interface {
	// Run executes c until it exits or ctx is done, writing the child's output to
	// stdout and stderr as it arrives. A nil writer discards its stream. Pass the same
	// writer value for both to merge them in arrival order; two different writers
	// must each be safe on their own. A non-zero exit is an error that wraps
	// *exec.ExitError. ctx must not be nil.
	Run(ctx context.Context, c Cmd, stdout, stderr io.Writer) error

	// Attach runs c with the three files handed to the child as they are, not through
	// pipes, so a terminal stays a terminal. It returns the child's exit code with a
	// nil error once the child has run: whether a non-zero session is a failure is
	// the caller's call. An error means the session never ran.
	Attach(c Cmd, stdin, stdout, stderr *os.File) (int, error)
}

// DefaultGrace is how long Exec waits after interrupting a cancelled command before it
// kills it, and how long it waits for the output streams to drain after the command
// exits.
const DefaultGrace = 2 * time.Second

// Exec is the real Runner, backed by os/exec. The zero value is ready to use.
type Exec struct {
	// Announce, when set, is called with the resolved path and the arguments right
	// before each command starts. nil is silent. A tool's verbose mode installs one.
	Announce func(path string, args []string)

	// Grace overrides DefaultGrace when positive.
	Grace time.Duration
}

func (e Exec) grace() time.Duration {
	if e.Grace > 0 {
		return e.Grace
	}
	return DefaultGrace
}

func (e Exec) announce(path string, args []string) {
	if e.Announce != nil {
		e.Announce(path, args)
	}
}

// Resolve turns a program name into the absolute path that will run. Exec resolves
// through it immediately before running, and a tool may resolve through it up front to
// report which binaries it will use, so the path reported and the path run come from the
// same rule.
//
// exec.ErrDot is an error here, not a fallback: Go reports it when a bare name resolved
// relative to the current directory, which is the "binary shipped beside the env file"
// case. Any other relative result -- an explicit ./name, or ErrDot switched off with
// GODEBUG -- is refused for the same reason.
func Resolve(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		// LookPath returns the path it found alongside ErrDot, so the message can name
		// the file that would have run.
		if errors.Is(err, exec.ErrDot) {
			return "", fmt.Errorf("refusing to run %q from the current directory: it resolved to %q, which is "+
				"not on your PATH -- a binary shipped beside an env file must never run implicitly; "+
				"install it, or add its directory to PATH", name, path)
		}
		return "", fmt.Errorf("%s: not found on PATH: %w", name, err)
	}
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("refusing to run %q: it resolved to the relative path %q, which depends on the "+
			"current directory -- name the program bare so PATH finds it, or give its absolute path", name, path)
	}
	return path, nil
}

// applyEnv adds c.Env to the inherited environment. An empty Env leaves cmd.Env nil, which
// inherits the environment unchanged.
func applyEnv(cmd *exec.Cmd, env []string) {
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
}

// Run implements Runner. A context that is already done starts nothing and announces
// nothing.
func (e Exec) Run(ctx context.Context, c Cmd, stdout, stderr io.Writer) error {
	if err := CheckRun(c, stdout, stderr); err != nil {
		return err
	}
	name := c.Argv[0]
	path, err := Resolve(name)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%s: not started: %w", name, err)
	}
	e.announce(path, c.Argv[1:])
	cmd := exec.CommandContext(ctx, path, c.Argv[1:]...)
	// Interrupt first, so the child can stop cleanly; WaitDelay kills it if it is
	// still running when the grace period ends. Where an interrupt cannot be sent
	// (Windows), Signal fails and the kill comes at the end of the grace period.
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	cmd.WaitDelay = e.grace()
	if c.Stdin != nil {
		cmd.Stdin = bytes.NewReader(c.Stdin)
	}
	applyEnv(cmd, c.Env)
	cmd.Stdout, cmd.Stderr = stdout, stderr
	return classify(ctx, name, cmd.Run())
}

// classify words the result of a run.
//
// os/exec returns ErrWaitDelay only after the command exited successfully while
// something it started still held its output open past the grace period. The command
// itself succeeded; what that leftover process writes later is lost, and today's
// alternative is waiting for it forever.
func classify(ctx context.Context, name string, err error) error {
	if err == nil || errors.Is(err, exec.ErrWaitDelay) {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		if errors.Is(err, ctxErr) {
			return fmt.Errorf("%s: stopped: %w", name, ctxErr)
		}
		return fmt.Errorf("%s: stopped: %w: %w", name, ctxErr, err)
	}
	return fmt.Errorf("%s: %w", name, err)
}

// Attach implements Runner. There is no Cancel and no WaitDelay: a session the operator
// is still typing into is never interrupted or killed by the runner.
func (e Exec) Attach(c Cmd, stdin, stdout, stderr *os.File) (int, error) {
	if err := CheckAttach(c, stdin, stdout, stderr); err != nil {
		return 0, err
	}
	name := c.Argv[0]
	path, err := Resolve(name)
	if err != nil {
		return 0, err
	}
	e.announce(path, c.Argv[1:])
	cmd := exec.Command(path, c.Argv[1:]...)
	applyEnv(cmd, c.Env)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	err = cmd.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		// A child killed by a signal reports -1, which no caller can exit with.
		if code := ee.ExitCode(); code > 0 {
			return code, nil
		}
		return exitcode.Failure, nil
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return 0, nil
}

// Combined runs c with stdout and stderr merged into one string, in arrival order. The
// output is returned even when the run fails: a CLI usually says why in it.
func Combined(ctx context.Context, r Runner, c Cmd) (string, error) {
	var buf bytes.Buffer
	err := r.Run(ctx, c, &buf, &buf)
	return buf.String(), err
}

// Split runs c with each stream captured on its own, for output that is parsed: a
// warning on stderr must not corrupt JSON on stdout. Both are returned even when the
// run fails.
func Split(ctx context.Context, r Runner, c Cmd) (stdout, stderr string, err error) {
	var out, errOut bytes.Buffer
	err = r.Run(ctx, c, &out, &errOut)
	return out.String(), errOut.String(), err
}

// Output runs c capturing stdout while stderr goes to the given writer as it arrives,
// so an engine's diagnostics reach the operator whether or not its answer is being
// read. The captured bytes are returned even when the run fails; they are nil when
// nothing was written.
func Output(ctx context.Context, r Runner, c Cmd, stderr io.Writer) ([]byte, error) {
	var out bytes.Buffer
	err := r.Run(ctx, c, &out, stderr)
	return out.Bytes(), err
}
