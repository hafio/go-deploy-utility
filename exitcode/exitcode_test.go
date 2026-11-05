package exitcode

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
)

// TestChildExitFallsBackWhenThereIsNoStatus: a marked error with no exit status
// behind it (the engine binary was not found, a signal killed the child) must not
// invent one.
func TestChildExitFallsBackWhenThereIsNoStatus(t *testing.T) {
	if got := Code(ChildExit(nil)); got != OK {
		t.Errorf("ChildExit(nil): got %d, want %d", got, OK)
	}
	if got := Code(ChildExit(errors.New("docker: not found on PATH"))); got != Failure {
		t.Errorf("marked error with no child status: got %d, want %d", got, Failure)
	}
	// A usage error is never overridden by a status: the two markers cannot both
	// apply to one error, and the classification order must be stable if they ever do.
	if got := Code(AsUsage(ChildExit(childStatusError(t, 7)))); got != Usage {
		t.Errorf("usage marker over a child status: got %d, want %d", got, Usage)
	}
}

// TestChildExitKeepsItsMessage is the ChildExit half of TestUsageErrorKeepsItsMessage
// below, and it matters for the same reason: a tool's main PRINTS the error before
// exiting with the code. A marker that swallowed the message would leave a failed
// session reporting a bare status and nothing about what went wrong.
func TestChildExitKeepsItsMessage(t *testing.T) {
	inner := errors.New("kubectl: exit status 7")
	marked := ChildExit(inner)
	if marked.Error() != inner.Error() {
		t.Errorf("ChildExit message: got %q, want %q", marked, inner)
	}
	if !errors.Is(marked, inner) {
		t.Errorf("ChildExit lost the cause: %v", marked)
	}
}

// TestExitCodeIsNeverNegative covers the guard on a signal-killed child, which is the
// one status a process cannot pass through: os.Exit is documented for 0-125 and a
// negative is meaningless to both operating systems the tools run on. Go reports -1
// for a signalled child, so without the guard a killed session would call os.Exit(-1).
//
// The two platforms produce different codes for a Kill -- Unix signals (-1 from
// ExitCode) where Windows terminates with a real status -- so the assertion is the
// portable property rather than a number: whatever comes back is a status a process
// can legally exit with.
func TestExitCodeIsNeverNegative(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run", "^TestHelperExitProcess$")
	cmd.Env = append(os.Environ(), exitCodeVar+"=block")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper: %v", err)
	}
	if err := cmd.Process.Kill(); err != nil {
		// Skipped rather than failed: a Kill this process is not permitted to make
		// says nothing about the guard under test, and a flaky hard failure in a
		// gate is worse than a case that occasionally does not run.
		t.Skipf("could not kill the helper process on this platform: %v", err)
	}
	err := cmd.Wait()
	if err == nil {
		t.Fatal("a killed child reported success")
	}
	got := Code(ChildExit(fmt.Errorf("kubectl: %w", err)))
	if got <= 0 || got > 125 {
		t.Errorf("Code for a killed session = %d, want a status os.Exit can take (1-125)", got)
	}
}

// TestUsageErrorKeepsItsMessage: the classification must cost nothing in what the
// operator reads, since every one of these errors names the flag, word or platform
// that was wrong.
func TestUsageErrorKeepsItsMessage(t *testing.T) {
	const msg = "--pod nope: unknown role"
	if got := Usagef("%s", msg).Error(); got != msg {
		t.Errorf("Usagef message: got %q, want %q", got, msg)
	}
	inner := errors.New(msg)
	marked := AsUsage(inner)
	if marked.Error() != msg || !errors.Is(marked, inner) {
		t.Errorf("AsUsage lost the message or the cause: %v", marked)
	}
	// Marking twice is marking once: AsUsage is applied both at individual sites and
	// over whole call paths, so the two have to compose.
	if twice := AsUsage(marked); twice.Error() != msg || !IsUsage(twice) {
		t.Errorf("double marking changed the error: %v", twice)
	}
	if AsUsage(nil) != nil {
		t.Error("AsUsage(nil) is not nil")
	}
}

// TestHelperExitProcess is not a test: it is the child process childStatusError and
// TestExitCodeIsNeverNegative start. This is the standard os/exec helper-process
// pattern -- the test binary re-executes itself -- because a real *exec.ExitError
// cannot be built by hand (only os/exec fills in its ProcessState) and no one command
// spells "exit 7" on both Windows and Linux.
//
// A numeric value exits with that status. The value "block" instead waits to be
// killed, which is how the signalled-child branch is reached; the sleep is a
// backstop against a Kill that never lands, so a broken test fails on the suite
// timeout rather than hanging forever.
func TestHelperExitProcess(t *testing.T) {
	code := os.Getenv(exitCodeVar)
	switch code {
	case "":
		t.Skip("helper process for childStatusError; nothing to do when run directly")
	case "block":
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	n, err := strconv.Atoi(code)
	if err != nil {
		t.Fatalf("%s=%q is neither a number nor \"block\"", exitCodeVar, code)
	}
	os.Exit(n)
}

const exitCodeVar = "GO_DEPLOY_UTILITY_TEST_CHILD_EXIT_CODE"

// childStatusError returns a real *exec.ExitError carrying the given status.
func childStatusError(t *testing.T, code int) error {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run", "^TestHelperExitProcess$")
	cmd.Env = append(os.Environ(), exitCodeVar+"="+strconv.Itoa(code))
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	err := cmd.Run()
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("helper process did not produce an exit status: %v", err)
	}
	if ee.ExitCode() != code {
		t.Fatalf("helper process exited %d, wanted %d", ee.ExitCode(), code)
	}
	return err
}

// TestCodeClassifiesEveryBranch walks the classifier through each outcome,
// including the passthrough branch the lifted tests never reach on Unix (there the
// killed child reports -1) and the scoping property at classifier level: a real
// status that is NOT marked as a session must not escape.
func TestCodeClassifiesEveryBranch(t *testing.T) {
	seven := childStatusError(t, 7)
	two := childStatusError(t, 2)
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, OK},
		{"plain error", errors.New("cluster unreachable"), Failure},
		{"usage", Usagef("unknown flag %q", "--nope"), Usage},
		{"wrapped usage", fmt.Errorf("deploy: %w", Usagef("bad platform")), Usage},
		{"marked session status passes through", ChildExit(fmt.Errorf("kubectl: %w", seven)), 7},
		{"unmarked real status stays a failure", fmt.Errorf("kubectl: %w", seven), Failure},
		{"a passed-through 2 is indistinguishable from usage", ChildExit(two), 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Code(tc.err); got != tc.want {
				t.Errorf("Code(%v) = %d, want %d", tc.err, got, tc.want)
			}
		})
	}
}

// TestChildStatusReportsOnlyMarkedRealStatus pins each early return separately.
func TestChildStatusReportsOnlyMarkedRealStatus(t *testing.T) {
	seven := childStatusError(t, 7)
	if _, ok := ChildStatus(nil); ok {
		t.Error("ChildStatus(nil) reported a status")
	}
	if _, ok := ChildStatus(ChildExit(errors.New("no process ran"))); ok {
		t.Error("a marked error with no exit status reported one")
	}
	if _, ok := ChildStatus(seven); ok {
		t.Error("an unmarked exit status was reported as a session status")
	}
	code, ok := ChildStatus(ChildExit(seven))
	if !ok || code != 7 {
		t.Errorf("ChildStatus(marked 7) = (%d, %v), want (7, true)", code, ok)
	}
}

// TestExitedCarriesTheStatus: Exited is the status of a runner that hands back the code
// rather than an *exec.ExitError. It reads exactly as os/exec words the same status, is a
// failure until it is marked, and once marked passes the code through like a real one.
func TestExitedCarriesTheStatus(t *testing.T) {
	if err := Exited("docker", 0); err != nil {
		t.Errorf("Exited(0) = %v, want nil", err)
	}
	err := Exited("docker", 3)
	if want := fmt.Errorf("docker: %w", childStatusError(t, 3)).Error(); err.Error() != want {
		t.Errorf("Exited message = %q, want os/exec's %q", err, want)
	}
	if got := Code(err); got != Failure {
		t.Errorf("Code(unmarked Exited) = %d, want %d", got, Failure)
	}
	if got := Code(ChildExit(fmt.Errorf("opening a session: %w", err))); got != 3 {
		t.Errorf("Code(marked Exited) = %d, want 3", got)
	}
	if got := Code(AsUsage(ChildExit(err))); got != Usage {
		t.Errorf("usage marker over Exited: got %d, want %d", got, Usage)
	}
	if code, ok := ChildStatus(ChildExit(Exited("docker", -1))); ok {
		t.Errorf("ChildStatus(marked Exited(-1)) = (%d, true), want ok false", code)
	}
}

// TestUsagefWrapsWithPercentW: a usage error built around a cause keeps that cause
// in the chain and reads exactly like fmt.Errorf.
func TestUsagefWrapsWithPercentW(t *testing.T) {
	cause := errors.New("no such file")
	err := Usagef("load %s: %w", "env.yaml", cause)
	if !errors.Is(err, cause) {
		t.Errorf("Usagef with %%w lost the cause: %v", err)
	}
	if !IsUsage(err) {
		t.Errorf("Usagef result is not a usage error: %v", err)
	}
	if want := fmt.Errorf("load %s: %w", "env.yaml", cause).Error(); err.Error() != want {
		t.Errorf("Usagef message = %q, want %q", err, want)
	}
}

// TestIsUsageAcrossWrapping: the marker is found wherever it sits in the chain.
func TestIsUsageAcrossWrapping(t *testing.T) {
	marked := Usagef("bad role")
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain", errors.New("x"), false},
		{"direct", marked, true},
		{"wrapped with %w", fmt.Errorf("prepare: %w", marked), true},
		{"inside errors.Join", errors.Join(errors.New("other"), marked), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsUsage(tc.err); got != tc.want {
				t.Errorf("IsUsage(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// TestExitValuesAreTheContract: scripts retry on 1 and not on 2, so the numbers
// themselves are public contract, not an implementation detail.
func TestExitValuesAreTheContract(t *testing.T) {
	if OK != 0 || Failure != 1 || Usage != 2 {
		t.Errorf("exit values = %d/%d/%d, want 0/1/2", OK, Failure, Usage)
	}
}

// TestAsUsageNeverDoubleWraps: marking an already-marked error returns the very
// same value, not a wrapper around it.
func TestAsUsageNeverDoubleWraps(t *testing.T) {
	marked := Usagef("bad flag")
	if again := AsUsage(marked); again != marked {
		t.Errorf("AsUsage wrapped an already-marked error: got %#v, want %#v", again, marked)
	}
	wrapped := fmt.Errorf("ctx: %w", marked)
	if again := AsUsage(wrapped); again != wrapped {
		t.Errorf("AsUsage wrapped an error whose chain already carries the marker: %#v", again)
	}
}
