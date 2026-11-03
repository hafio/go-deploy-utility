// Package runnertest is the recording fake both tools test against instead of starting
// processes. A Fake implements runner.Runner: it records every call, answers each from
// its Script, and writes the scripted output to whatever writers the caller passed, so
// runner.Combined, runner.Split and runner.Output all work against it unchanged.
//
// Scripts match by position (Seq, for tests that know the call order) or by argv content
// (FailOn, or any function of the Call, for tests that know only which command matters).
//
// A Fake applies the same checks as runner.Exec first. A refused call is neither recorded
// nor passed to the Script, so a builder bug fails a tool test the way it would fail in
// production.
package runnertest

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/hafio/go-deploy-utility/runner"
)

// Call is one recorded call. Argv, Stdin and Env are copies, so a caller that reuses its
// slices cannot rewrite the record.
type Call struct {
	Argv  []string
	Stdin []byte // nil when the Cmd had none
	Env   []string

	// Stdout and Stderr are the writers passed to Run, so a test can tell a streamed
	// call (os.Stdout) from a captured one. Both are nil for Attach.
	Stdout, Stderr io.Writer

	// Attach is true for a call made through Attach.
	Attach bool
}

// Has reports whether any argv token contains sub.
func (c Call) Has(sub string) bool {
	for _, a := range c.Argv {
		if strings.Contains(a, sub) {
			return true
		}
	}
	return false
}

// String renders the argv for a failure message, quoted as Echo prints it.
func (c Call) String() string {
	if len(c.Argv) == 0 {
		return ""
	}
	return runner.Quote(c.Argv[0], c.Argv[1:]...)
}

func (c Call) clone() Call {
	c.Argv = cloneStrings(c.Argv)
	c.Env = cloneStrings(c.Env)
	if c.Stdin != nil {
		c.Stdin = append([]byte{}, c.Stdin...)
	}
	return c
}

func cloneStrings(s []string) []string {
	if s == nil {
		return nil
	}
	return append([]string{}, s...)
}

// Reply is a Script's answer to one call.
type Reply struct {
	// Stdout and Stderr are written to the caller's writers (Run only).
	Stdout, Stderr string

	// Err is returned as the call's error.
	Err error

	// Code is the exit code Attach returns.
	Code int
}

// Fake is a recording runner.Runner. The zero value answers every call with the zero
// Reply. A Fake is safe for concurrent use; use it by pointer.
type Fake struct {
	// Script answers call n (zero-based, in arrival order). nil answers every call
	// with the zero Reply. It runs outside the Fake's lock, so it may call Calls.
	Script func(n int, c Call) Reply

	mu    sync.Mutex
	calls []Call
}

// record stores c and returns its index.
func (f *Fake) record(c Call) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, c.clone())
	return len(f.calls) - 1
}

func (f *Fake) reply(n int, c Call) Reply {
	if f.Script == nil {
		return Reply{}
	}
	return f.Script(n, c.clone())
}

// Run implements runner.Runner. A done ctx is refused as runner.Exec refuses it: the call
// is not recorded.
func (f *Fake) Run(ctx context.Context, c runner.Cmd, stdout, stderr io.Writer) error {
	if err := runner.CheckRun(c, stdout, stderr); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%s: not started: %w", c.Argv[0], err)
	}
	call := Call{Argv: c.Argv, Stdin: c.Stdin, Env: c.Env, Stdout: stdout, Stderr: stderr}
	r := f.reply(f.record(call), call)
	if r.Stdout != "" && stdout != nil {
		_, _ = io.WriteString(stdout, r.Stdout)
	}
	if r.Stderr != "" && stderr != nil {
		_, _ = io.WriteString(stderr, r.Stderr)
	}
	return r.Err
}

// Attach implements runner.Runner. The files are not touched.
func (f *Fake) Attach(c runner.Cmd, stdin, stdout, stderr *os.File) (int, error) {
	if err := runner.CheckAttach(c, stdin, stdout, stderr); err != nil {
		return 0, err
	}
	call := Call{Argv: c.Argv, Env: c.Env, Attach: true}
	r := f.reply(f.record(call), call)
	return r.Code, r.Err
}

// Calls returns a copy of every recorded call, in arrival order.
func (f *Fake) Calls() []Call {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]Call, len(f.calls))
	for i, c := range f.calls {
		out[i] = c.clone()
	}
	return out
}

// Seq answers call n with replies[n], and every call past the end with the zero Reply.
func Seq(replies ...Reply) func(int, Call) Reply {
	replies = append([]Reply{}, replies...)
	return func(n int, _ Call) Reply {
		if n < len(replies) {
			return replies[n]
		}
		return Reply{}
	}
}

// FailOn answers a call whose argv has a token containing sub with err, and every other
// call with the zero Reply.
func FailOn(sub string, err error) func(int, Call) Reply {
	return func(_ int, c Call) Reply {
		if c.Has(sub) {
			return Reply{Err: err}
		}
		return Reply{}
	}
}
