// Package exitcode is the exit-status contract both tools follow: 0 the command
// did what it was asked, 1 it was asked for something sensible and could not do
// it, 2 the request itself was wrong. A script that retries on a transient
// cluster failure must not retry a typo, and one code for both makes that
// undecidable.
//
// The line between 1 and 2 is "would a different command line have helped?".
// The caller decides which errors are usage errors by marking them with Usagef or
// AsUsage; an unmarked error is a failure. The package never prints and never
// exits: a tool's main calls os.Exit(exitcode.Code(err)) itself, and a tool whose
// handlers return an int uses the constants directly.
package exitcode

import (
	"errors"
	"fmt"
	"os/exec"
)

// The three exit statuses. They are public contract: scripts branch on them.
const (
	OK      = 0
	Failure = 1
	Usage   = 2
)

// usageError marks an error as being about the command line rather than about the
// world. It is a wrapper rather than a sentinel because every one of these errors
// says something specific -- which flag, which word, which platform -- and the
// classification must not cost that message.
type usageError struct{ err error }

func (e usageError) Error() string { return e.err.Error() }
func (e usageError) Unwrap() error { return e.err }

// Usagef builds a usage error. A %w verb in format keeps the cause in the chain.
func Usagef(format string, args ...any) error {
	return usageError{fmt.Errorf(format, args...)}
}

// AsUsage marks an error produced elsewhere -- a flag parser, an argument
// validator, a value parser -- as a usage error, passing nil through and never
// double-wrapping.
func AsUsage(err error) error {
	if err == nil || IsUsage(err) {
		return err
	}
	return usageError{err}
}

// IsUsage reports whether err carries the usage marker anywhere in its chain.
func IsUsage(err error) bool {
	var u usageError
	return errors.As(err, &u)
}

// childExitError marks an error whose OWN exit status is the answer, so the tool
// reports the child's code rather than its own 1. It is for commands that hand the
// terminal over to an interactive session, where the operator's last command inside
// the session is what they are asking about, and collapsing every non-zero session
// to 1 throws that away.
//
// It is deliberately not for anything else. Every other command runs kubectl or a
// container engine as an implementation detail, and leaking those codes would make
// a tool's contract "whatever the engine felt like", which is unusable in a script
// and impossible to document.
//
// A passed-through status can of course BE 2, and then it is indistinguishable from
// the tool's own usage code. That is inherent to a passthrough rather than a defect:
// the operator asked for the session's answer, which is also why the passthrough is
// confined to the commands that hand the terminal over.
type childExitError struct{ err error }

func (e childExitError) Error() string { return e.err.Error() }
func (e childExitError) Unwrap() error { return e.err }

// ChildExit marks an interactive session's failure for status passthrough. nil
// stays nil; the message and the cause are kept.
func ChildExit(err error) error {
	if err == nil {
		return nil
	}
	return childExitError{err}
}

// Code classifies an error into the exit status the process should end with: nil
// is OK, a usage error is Usage (checked first, so it beats a child status), a
// ChildExit-marked session with a positive status passes that status through, and
// anything else is Failure. The result is never negative and never 0 for a non-nil
// error.
func Code(err error) int {
	switch {
	case err == nil:
		return OK
	case IsUsage(err):
		return Usage
	default:
		if code, ok := ChildStatus(err); ok {
			return code
		}
		return Failure
	}
}

// ChildStatus reads the exit status out of a ChildExit-marked error that wraps a
// real *exec.ExitError. ok is false for an unmarked error, for a marked error with
// no exit status behind it, and for a child killed by a signal; code is meaningful
// only when ok is true.
//
// A child killed by a signal reports -1, which os.Exit cannot take, so the status is
// returned with `code > 0` as the ok answer rather than through a branch: the
// signalled case is unreachable on Windows (TerminateProcess sets a real status), so
// a branch for it would be code no local coverage run could ever enter. One
// expression says the same thing and is exercised by every caller that gets here.
func ChildStatus(err error) (code int, ok bool) {
	var marked childExitError
	if !errors.As(err, &marked) {
		return 0, false
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		return 0, false
	}
	code = ee.ExitCode()
	return code, code > 0
}
