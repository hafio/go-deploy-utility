package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// helperEnv switches TestHelperProcess on in the re-executed test binary. Exec leaves
// cmd.Env nil unless Cmd.Env is set, and appends to os.Environ() when it is, so a value
// set with t.Setenv always reaches the child.
const helperEnv = "GO_DEPLOY_UTILITY_TEST_RUNNER_HELPER"

// helperArgv builds a Cmd.Argv that re-executes this test binary as a fake external
// command: TestHelperProcess runs as the child, in the given mode. argv[0] is the
// absolute path of the test binary, which Resolve accepts wherever the test runs from.
func helperArgv(t *testing.T, mode ...string) []string {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	return append([]string{exe, "-test.run=^TestHelperProcess$", "--"}, mode...)
}

// helperOn sets the switch for the duration of the test.
func helperOn(t *testing.T) {
	t.Helper()
	t.Setenv(helperEnv, "1")
}

// TestHelperProcess is not a real test: it is a no-op unless helperEnv is set, which
// only the tests below do for the child they start. Everything after "--" in os.Args is
// the mode and its arguments.
func TestHelperProcess(t *testing.T) {
	if os.Getenv(helperEnv) != "1" {
		return
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) == 0 {
		os.Exit(0)
	}
	switch args[0] {
	case "echo":
		fmt.Fprint(os.Stdout, strings.Join(args[1:], " "))
	case "stdin", "cat":
		_, _ = io.Copy(os.Stdout, os.Stdin)
	case "both":
		fmt.Fprintln(os.Stdout, "stdout-line")
		fmt.Fprintln(os.Stderr, "stderr-line")
	case "fail":
		fmt.Fprintln(os.Stdout, "before-exit")
		os.Exit(3)
	case "drip":
		// Print at once, then outlive any reasonable test: a reader that sees the line
		// saw it while the process was still running.
		fmt.Fprintln(os.Stdout, "drip-line")
		time.Sleep(10 * time.Minute)
	case "env":
		if len(args) > 1 {
			fmt.Fprint(os.Stdout, os.Getenv(args[1]))
			break
		}
		fmt.Fprintf(os.Stdout, "AMBIENT_ONLY=%s OVERRIDE=%s EXTRA_ONLY=%s\n",
			os.Getenv("RUNNER_TEST_AMBIENT_ONLY"),
			os.Getenv("RUNNER_TEST_OVERRIDE"),
			os.Getenv("RUNNER_TEST_EXTRA_ONLY"))
	case "stdiokind":
		// Whether each standard file is a plain regular file. Had os/exec interposed a
		// pipe -- which it does for any io.Writer that is not an *os.File -- these would
		// read as pipes, and a pipe is not a terminal.
		fmt.Fprintf(os.Stdout, "in=%v out=%v err=%v\n",
			regularFile(os.Stdin), regularFile(os.Stdout), regularFile(os.Stderr))
	case "trapint":
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		fmt.Fprintln(os.Stdout, "ready")
		<-c
		fmt.Fprintln(os.Stdout, "interrupted")
	case "ignoreint":
		signal.Ignore(os.Interrupt)
		fmt.Fprintln(os.Stdout, "ready")
		time.Sleep(10 * time.Minute)
	case "orphan":
		// Start a grandchild that inherits stdout and outlives this process, then exit
		// 0 without waiting: the command succeeded, but its output pipe stays open.
		g := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$", "--", "sleep")
		g.Stdout, g.Stderr = os.Stdout, os.Stderr
		if err := g.Start(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "grandchild=%d\nparent-done\n", g.Process.Pid)
	case "sleep":
		time.Sleep(10 * time.Second)
	case "killself":
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Kill()
		time.Sleep(10 * time.Second)
	default:
		fmt.Fprintf(os.Stderr, "TestHelperProcess: unknown mode %q\n", args[0])
		os.Exit(2)
	}
	os.Exit(0)
}

// regularFile reports whether f is a plain file rather than a pipe. Go reports an
// anonymous pipe as ModeNamedPipe on Windows and unix alike.
func regularFile(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode().IsRegular()
}

// attachFiles opens three temp files to stand in for a terminal, since a test has none
// to hand over.
func attachFiles(t *testing.T) (stdin, stdout, stderr *os.File) {
	t.Helper()
	dir := t.TempDir()
	open := func(name string) *os.File {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("creating %s: %v", name, err)
		}
		t.Cleanup(func() { _ = f.Close() })
		return f
	}
	return open("stdin"), open("stdout"), open("stderr")
}

// readFile returns what the child wrote to one of those files.
func readFile(t *testing.T, f *os.File) string {
	t.Helper()
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("reading %s: %v", f.Name(), err)
	}
	return string(b)
}

// captureStdout swaps os.Stdout for a pipe while fn runs and returns whatever fn, or a
// child handed os.Stdout, wrote. Not safe with t.Parallel.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	done := make(chan []byte, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- b
	}()
	fn()
	os.Stdout = old
	_ = w.Close()
	out := <-done
	_ = r.Close()
	return string(out)
}

// writerFunc adapts a function to io.Writer, so a test can observe output at the moment
// it is written rather than after the process has exited.
type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// waitFor returns a writer that closes the returned channel the first time the text
// written so far contains want.
func waitFor(want string) (io.Writer, <-chan struct{}) {
	var once sync.Once
	var mu sync.Mutex
	var got strings.Builder
	seen := make(chan struct{})
	return writerFunc(func(p []byte) (int, error) {
		mu.Lock()
		got.Write(p)
		ok := strings.Contains(got.String(), want)
		mu.Unlock()
		if ok {
			once.Do(func() { close(seen) })
		}
		return len(p), nil
	}), seen
}

// ---- lifted from solmq-conn-util internal/runner/runner_test.go -------------------

// TestOSRunWiresStdinToChild covers text, and binary content of the kind a TLS store
// loaded into podman's secret store is: NUL, CR and bytes that are not UTF-8 must reach
// the child byte for byte.
func TestOSRunWiresStdinToChild(t *testing.T) {
	helperOn(t)
	for _, in := range []string{"hello-stdin\n", "\xfe\xed\xfe\xed\x00\x00\x00\x02\r\n\xff"} {
		out, err := Combined(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "stdin"), Stdin: []byte(in)})
		if err != nil {
			t.Fatalf("Combined returned error: %v (output %q)", err, out)
		}
		if out != in {
			t.Errorf("stdin content not visible to child unchanged: output = %q, want %q", out, in)
		}
	}
}

func TestOSRunCombinesStdoutAndStderr(t *testing.T) {
	helperOn(t)
	out, err := Combined(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "both")})
	if err != nil {
		t.Fatalf("Combined returned error: %v (output %q)", err, out)
	}
	if !strings.Contains(out, "stdout-line") {
		t.Errorf("combined output missing stdout write: %q", out)
	}
	if !strings.Contains(out, "stderr-line") {
		t.Errorf("combined output missing stderr write: %q", out)
	}
}

func TestOSRunNonZeroExitReturnsErrorWithOutput(t *testing.T) {
	helperOn(t)
	out, err := Combined(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "fail")})
	if err == nil {
		t.Fatalf("non-zero child exit must return a non-nil error, output = %q", out)
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 3 {
		t.Errorf("err = %v, want it to wrap the child's *exec.ExitError with status 3", err)
	}
	if !strings.Contains(out, "before-exit") {
		t.Errorf("output written before the failing exit must still be captured: %q", out)
	}
}

// TestOSRunEnvReachesChildAndAmbientInherited pins the Cmd.Env contract: a supplied
// entry reaches the child, an ambient variable is still inherited when Env carries
// unrelated entries, and a supplied entry wins over an ambient one of the same name.
func TestOSRunEnvReachesChildAndAmbientInherited(t *testing.T) {
	helperOn(t)
	t.Setenv("RUNNER_TEST_AMBIENT_ONLY", "ambient-only-value")
	t.Setenv("RUNNER_TEST_OVERRIDE", "ambient-should-lose")
	out, err := Combined(context.Background(), Exec{}, Cmd{
		Argv: helperArgv(t, "env"),
		Env: []string{
			"RUNNER_TEST_OVERRIDE=override-wins",
			"RUNNER_TEST_EXTRA_ONLY=extra-only-value",
		},
	})
	if err != nil {
		t.Fatalf("Combined returned error: %v (output %q)", err, out)
	}
	want := "AMBIENT_ONLY=ambient-only-value OVERRIDE=override-wins EXTRA_ONLY=extra-only-value\n"
	if out != want {
		t.Errorf("output = %q, want %q", out, want)
	}
}

// TestOSRunAcceptsAbsolutePathArgv0 pins that an absolute argv[0] resolves and runs
// exactly like a bare name does.
func TestOSRunAcceptsAbsolutePathArgv0(t *testing.T) {
	argv := helperArgv(t, "stdin")
	if !filepath.IsAbs(argv[0]) {
		t.Fatalf("os.Executable() returned a non-absolute path %q", argv[0])
	}
	helperOn(t)
	out, err := Combined(context.Background(), Exec{}, Cmd{Argv: argv, Stdin: []byte("abs-argv0-ok\n")})
	if err != nil {
		t.Fatalf("Combined with absolute argv[0] returned error: %v (output %q)", err, out)
	}
	if out != "abs-argv0-ok\n" {
		t.Errorf("output = %q", out)
	}
}

// TestOSRunWritesNothingToStderr pins that a run prints nothing of its own: the child's
// streams go only where the caller sent them, and the caller decides what an operator
// sees. os.Stderr is swapped for a pipe to catch anything that leaks.
func TestOSRunWritesNothingToStderr(t *testing.T) {
	helperOn(t)
	argv := helperArgv(t, "stdin")
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w

	out, runErr := Combined(context.Background(), Exec{}, Cmd{Argv: argv, Stdin: []byte("echo-test\n")})

	_ = w.Close()
	os.Stderr = origStderr
	data, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatalf("reading captured stderr: %v", readErr)
	}
	if runErr != nil {
		t.Fatalf("Combined returned error: %v", runErr)
	}
	if got := string(data); got != "" {
		t.Errorf("the run wrote %q to stderr, want nothing", got)
	}
	if !strings.Contains(out, "echo-test") {
		t.Errorf("combined output = %q, want it to carry the child's output", out)
	}
}

// TestOSRunRejectsUnresolvableArgv0 pins that a binary LookPath cannot find is an error
// rather than a deferred exec attempt.
func TestOSRunRejectsUnresolvableArgv0(t *testing.T) {
	_, err := Combined(context.Background(), Exec{}, Cmd{Argv: []string{"go-deploy-utility-test-nonexistent-binary"}})
	if err == nil {
		t.Fatal("a run must fail when argv[0] cannot be resolved on PATH")
	}
}

// TestOSStreamDeliversOutputBeforeExitAndCancelIsReported: the child prints one line
// and then blocks for far longer than this test waits, so seeing that line proves the
// output was not buffered until exit. Cancelling then ends the run, and the run reports
// the cancellation: whether that was a clean end is the caller's call.
// (Was ...AndCancelIsCleanEnd in solmq-conn-util, where Stream returned nil.)
func TestOSStreamDeliversOutputBeforeExitAndCancelIsReported(t *testing.T) {
	helperOn(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w, arrived := waitFor("drip-line")

	errc := make(chan error, 1)
	go func() { errc <- (Exec{}).Run(ctx, Cmd{Argv: helperArgv(t, "drip")}, w, io.Discard) }()

	select {
	case <-arrived:
	case err := <-errc:
		t.Fatalf("the child exited before its first line was seen (err=%v); the run is buffering", err)
	case <-time.After(30 * time.Second):
		t.Fatal("no output arrived while the child was still running; the run is buffering")
	}

	cancel()
	select {
	case err := <-errc:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("a cancelled run must report context.Canceled, got %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the run did not return after the context was cancelled")
	}
}

// TestOSStreamKeepsStdoutAndStderrApart pins why Run takes two writers: `logs > app.log`
// must land log lines in the file and leave the platform's diagnostics on the terminal.
func TestOSStreamKeepsStdoutAndStderrApart(t *testing.T) {
	helperOn(t)
	var out, errOut strings.Builder
	if err := (Exec{}).Run(context.Background(), Cmd{Argv: helperArgv(t, "both")}, &out, &errOut); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "stdout-line") || strings.Contains(out.String(), "stderr-line") {
		t.Errorf("stdout writer should carry only stdout, got %q", out.String())
	}
	if !strings.Contains(errOut.String(), "stderr-line") || strings.Contains(errOut.String(), "stdout-line") {
		t.Errorf("stderr writer should carry only stderr, got %q", errOut.String())
	}
}

// TestOSStreamReportsAFailureThatWasNotCancelled: an uncancelled non-zero exit is a real
// failure, and output written before it still reaches the writer.
func TestOSStreamReportsAFailureThatWasNotCancelled(t *testing.T) {
	helperOn(t)
	var out strings.Builder
	err := (Exec{}).Run(context.Background(), Cmd{Argv: helperArgv(t, "fail")}, &out, io.Discard)
	if err == nil {
		t.Fatal("a non-zero exit with no cancellation must be an error")
	}
	if errors.Is(err, context.Canceled) {
		t.Errorf("an uncancelled failure must not read as a cancellation: %v", err)
	}
	if !strings.Contains(out.String(), "before-exit") {
		t.Errorf("output written before the failure should still reach the writer, got %q", out.String())
	}
}

// TestOSStreamRejectsEmptyAndUnresolvableArgv pins that every path refuses the same
// things: they share Check and Resolve.
func TestOSStreamRejectsEmptyAndUnresolvableArgv(t *testing.T) {
	if err := (Exec{}).Run(context.Background(), Cmd{}, io.Discard, io.Discard); err == nil {
		t.Error("an empty argv must be refused")
	}
	err := (Exec{}).Run(context.Background(), Cmd{Argv: []string{"go-deploy-utility-no-such-binary"}}, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "not found on PATH") {
		t.Errorf("an unresolvable binary must be refused by name, got %v", err)
	}
}

// TestOSAttachHandsTheChildTheCallersFilesNotPipes: os/exec answers any writer that is
// not an *os.File with a pipe, so an Attach built on io.Writer would take the terminal
// away from the engine. The child reports what it was actually given.
func TestOSAttachHandsTheChildTheCallersFilesNotPipes(t *testing.T) {
	helperOn(t)
	stdin, stdout, stderr := attachFiles(t)
	code, err := (Exec{}).Attach(Cmd{Argv: helperArgv(t, "stdiokind")}, stdin, stdout, stderr)
	if err != nil || code != 0 {
		t.Fatalf("Attach = (%d, %v), want (0, nil); stderr %q", code, err, readFile(t, stderr))
	}
	if got, want := readFile(t, stdout), "in=true out=true err=true\n"; !strings.Contains(got, want) {
		t.Errorf("the child saw %q, want %q -- os/exec interposed a pipe", got, want)
	}
}

// TestOSAttachStdinIsTheHandedFile: what the caller gives as stdin is what the child
// reads, with nothing copied through this process.
func TestOSAttachStdinIsTheHandedFile(t *testing.T) {
	helperOn(t)
	stdin, stdout, stderr := attachFiles(t)
	if _, err := stdin.WriteString("hello-attach\n"); err != nil {
		t.Fatalf("writing stdin: %v", err)
	}
	if _, err := stdin.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("rewinding stdin: %v", err)
	}
	code, err := (Exec{}).Attach(Cmd{Argv: helperArgv(t, "stdin")}, stdin, stdout, stderr)
	if err != nil || code != 0 {
		t.Fatalf("Attach = (%d, %v), want (0, nil); stderr %q", code, err, readFile(t, stderr))
	}
	if got := readFile(t, stdout); got != "hello-attach\n" {
		t.Errorf("child echoed %q, want the handed stdin", got)
	}
}

// TestOSAttachReportsTheChildExitStatus: a session that ended non-zero is reported as
// that status, not as an error, because only the caller knows what it means.
func TestOSAttachReportsTheChildExitStatus(t *testing.T) {
	helperOn(t)
	stdin, stdout, stderr := attachFiles(t)
	code, err := (Exec{}).Attach(Cmd{Argv: helperArgv(t, "fail")}, stdin, stdout, stderr)
	if err != nil {
		t.Fatalf("a child that ran and exited non-zero is not an error here: %v", err)
	}
	if code != 3 {
		t.Errorf("code = %d, want 3", code)
	}
	if got := readFile(t, stdout); !strings.Contains(got, "before-exit") {
		t.Errorf("the child's output should still have reached the caller's file, got %q", got)
	}
}

// TestOSAttachEnvReachesChild: the environment behaves exactly as under Run.
func TestOSAttachEnvReachesChild(t *testing.T) {
	helperOn(t)
	t.Setenv("RUNNER_TEST_AMBIENT_ONLY", "ambient")
	t.Setenv("RUNNER_TEST_OVERRIDE", "ambient")
	stdin, stdout, stderr := attachFiles(t)
	cmd := Cmd{
		Argv: helperArgv(t, "env"),
		Env:  []string{"RUNNER_TEST_OVERRIDE=supplied", "RUNNER_TEST_EXTRA_ONLY=extra"},
	}
	code, err := (Exec{}).Attach(cmd, stdin, stdout, stderr)
	if err != nil || code != 0 {
		t.Fatalf("Attach = (%d, %v), want (0, nil); stderr %q", code, err, readFile(t, stderr))
	}
	want := "AMBIENT_ONLY=ambient OVERRIDE=supplied EXTRA_ONLY=extra\n"
	if got := readFile(t, stdout); got != want {
		t.Errorf("child env = %q, want %q", got, want)
	}
}

// TestOSAttachRefusesACmdCarryingStdinText: a Cmd that wants bytes written to the child
// cannot also be a session reading the operator's keyboard.
func TestOSAttachRefusesACmdCarryingStdinText(t *testing.T) {
	stdin, stdout, stderr := attachFiles(t)
	_, err := (Exec{}).Attach(Cmd{Argv: []string{"go"}, Stdin: []byte("payload")}, stdin, stdout, stderr)
	if err == nil {
		t.Fatal("a Cmd carrying stdin bytes must be refused")
	}
	if !strings.Contains(err.Error(), "Cmd.Stdin") {
		t.Errorf("the error should name the field, got %v", err)
	}
}

// TestOSAttachRefusesANilFile fails here rather than at the child's first write.
func TestOSAttachRefusesANilFile(t *testing.T) {
	stdin, stdout, stderr := attachFiles(t)
	for _, c := range []struct {
		name          string
		in, out, errf *os.File
	}{
		{"no stdin", nil, stdout, stderr},
		{"no stdout", stdin, nil, stderr},
		{"no stderr", stdin, stdout, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := (Exec{}).Attach(Cmd{Argv: []string{"go"}}, c.in, c.out, c.errf); err == nil {
				t.Fatal("a nil file must be refused")
			}
		})
	}
}

// TestOSAttachRejectsEmptyAndUnresolvableArgv mirrors the Run refusals.
func TestOSAttachRejectsEmptyAndUnresolvableArgv(t *testing.T) {
	stdin, stdout, stderr := attachFiles(t)
	if _, err := (Exec{}).Attach(Cmd{}, stdin, stdout, stderr); err == nil {
		t.Error("an empty argv must be refused")
	}
	if _, err := (Exec{}).Attach(Cmd{Argv: []string{"go-deploy-utility-no-such-binary"}}, stdin, stdout, stderr); err == nil {
		t.Error("an unresolvable argv[0] must be refused")
	}
}

// TestOSRunSplitKeepsTheStreamsApart: `oc get -o json` on OpenShift prints a deprecation
// warning on stderr, and merged into stdout it breaks the JSON on its first character.
func TestOSRunSplitKeepsTheStreamsApart(t *testing.T) {
	helperOn(t)
	out, errOut, err := Split(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "both")})
	if err != nil {
		t.Fatalf("Split returned error: %v", err)
	}
	if out != "stdout-line\n" {
		t.Errorf("stdout = %q, want only the stdout line", out)
	}
	if errOut != "stderr-line\n" {
		t.Errorf("stderr = %q, want only the stderr line", errOut)
	}
}

// TestOSRunSplitWiresStdinAndEnv: the split path is otherwise the same run.
func TestOSRunSplitWiresStdinAndEnv(t *testing.T) {
	helperOn(t)
	out, _, err := Split(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "stdin"), Stdin: []byte("hello-split\n")})
	if err != nil || out != "hello-split\n" {
		t.Fatalf("stdin: out=%q err=%v", out, err)
	}

	t.Setenv("RUNNER_TEST_AMBIENT_ONLY", "ambient")
	t.Setenv("RUNNER_TEST_OVERRIDE", "ambient")
	out, _, err = Split(context.Background(), Exec{}, Cmd{
		Argv: helperArgv(t, "env"),
		Env:  []string{"RUNNER_TEST_OVERRIDE=supplied", "RUNNER_TEST_EXTRA_ONLY=extra"},
	})
	if err != nil {
		t.Fatalf("env: %v", err)
	}
	if want := "AMBIENT_ONLY=ambient OVERRIDE=supplied EXTRA_ONLY=extra\n"; out != want {
		t.Errorf("env: out = %q, want %q", out, want)
	}
}

// TestOSRunSplitRejectsEmptyAndUnresolvableArgv mirrors Run and Attach.
func TestOSRunSplitRejectsEmptyAndUnresolvableArgv(t *testing.T) {
	if _, _, err := Split(context.Background(), Exec{}, Cmd{}); err == nil {
		t.Error("an empty argv must be refused")
	}
	if _, _, err := Split(context.Background(), Exec{}, Cmd{Argv: []string{"go-deploy-utility-no-such-binary"}}); err == nil {
		t.Error("an unresolvable argv[0] must be refused")
	}
}

// ---- lifted from solace-util internal/engine/runner_test.go and resolve_test.go ------

func TestExecOutput(t *testing.T) {
	helperOn(t)
	out, err := Output(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "echo", "hello")}, nil)
	if err != nil {
		t.Fatalf("Output(echo): unexpected err %v", err)
	}
	if !strings.Contains(string(out), "hello") {
		t.Errorf("Output(echo) = %q, want it to contain %q", out, "hello")
	}
}

func TestExecOutputFail(t *testing.T) {
	helperOn(t)
	argv := helperArgv(t, "fail")
	out, err := Output(context.Background(), Exec{}, Cmd{Argv: argv}, nil)
	if err == nil {
		t.Fatalf("Output(fail): expected err, got nil (out=%q)", out)
	}
	if !strings.Contains(err.Error(), argv[0]) {
		t.Errorf("Output(fail) err = %q, want it to contain binary name %q", err, argv[0])
	}
}

func TestExecRun(t *testing.T) {
	helperOn(t)

	var errOK error
	_ = captureStdout(t, func() {
		errOK = (Exec{}).Run(context.Background(), Cmd{Argv: helperArgv(t, "echo", "streamed")}, os.Stdout, os.Stderr)
	})
	if errOK != nil {
		t.Fatalf("Run(echo): unexpected err %v", errOK)
	}

	argv := helperArgv(t, "fail")
	var errFail error
	_ = captureStdout(t, func() {
		errFail = (Exec{}).Run(context.Background(), Cmd{Argv: argv}, os.Stdout, os.Stderr)
	})
	if errFail == nil {
		t.Fatalf("Run(fail): expected err, got nil")
	}
	if !strings.Contains(errFail.Error(), argv[0]) {
		t.Errorf("Run(fail) err = %q, want it to contain binary name %q", errFail, argv[0])
	}
}

func TestExecRunInput(t *testing.T) {
	helperOn(t)

	var err error
	got := captureStdout(t, func() {
		err = (Exec{}).Run(context.Background(), Cmd{Argv: helperArgv(t, "cat"), Stdin: []byte("meow")}, os.Stdout, os.Stderr)
	})
	if err != nil {
		t.Fatalf("Run(cat): unexpected err %v", err)
	}
	if !strings.Contains(got, "meow") {
		t.Errorf("Run(cat) streamed %q, want it to contain %q", got, "meow")
	}

	argv := helperArgv(t, "fail")
	var errFail error
	_ = captureStdout(t, func() {
		errFail = (Exec{}).Run(context.Background(), Cmd{Argv: argv, Stdin: []byte("meow")}, os.Stdout, os.Stderr)
	})
	if errFail == nil {
		t.Fatal("Run(fail) with stdin: expected err, got nil")
	}
	if !strings.Contains(errFail.Error(), argv[0]) {
		t.Errorf("Run(fail) with stdin err = %q, want it to contain binary name %q", errFail, argv[0])
	}
}

// TestExecRunEnv: the child receives the extra variable, and still inherits this
// process's environment (the helper only runs because helperEnv was inherited).
func TestExecRunEnv(t *testing.T) {
	helperOn(t)

	var err error
	got := captureStdout(t, func() {
		err = (Exec{}).Run(context.Background(), Cmd{
			Argv: helperArgv(t, "env", "SOLACE_TEST_SECRET"),
			Env:  []string{"SOLACE_TEST_SECRET=from-parent"},
		}, os.Stdout, os.Stderr)
	})
	if err != nil {
		t.Fatalf("Run with env: unexpected err %v", err)
	}
	if !strings.Contains(got, "from-parent") {
		t.Errorf("child printed %q, want it to contain the injected value", got)
	}

	argv := helperArgv(t, "fail")
	var errFail error
	_ = captureStdout(t, func() {
		errFail = (Exec{}).Run(context.Background(), Cmd{Argv: argv, Env: []string{"A=b"}}, os.Stdout, os.Stderr)
	})
	if errFail == nil {
		t.Fatal("Run(fail) with env: expected err, got nil")
	}
	if !strings.Contains(errFail.Error(), argv[0]) {
		t.Errorf("Run(fail) with env err = %q, want it to name the binary %q", errFail, argv[0])
	}
}

// TestExecRunInteractive: an attached session runs, and a failing one reports its exit
// status rather than an error (solace-util's RunInteractive returned the wrapped error;
// Attach returns the status and the caller decides).
func TestExecRunInteractive(t *testing.T) {
	helperOn(t)
	stdin, stdout, stderr := attachFiles(t)

	code, err := (Exec{}).Attach(Cmd{Argv: helperArgv(t)}, stdin, stdout, stderr)
	if err != nil || code != 0 {
		t.Fatalf("Attach = (%d, %v), want (0, nil)", code, err)
	}

	code, err = (Exec{}).Attach(Cmd{Argv: helperArgv(t, "fail")}, stdin, stdout, stderr)
	if err != nil || code != 3 {
		t.Errorf("Attach(fail) = (%d, %v), want (3, nil): a failed session must not read as success", code, err)
	}
}

// TestExecOutputInput: stdin goes in and stdout comes back in the return value, not to
// the real stdout -- the curl -K - path, with no other way to read the response.
func TestExecOutputInput(t *testing.T) {
	helperOn(t)
	out, err := Output(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "cat"), Stdin: []byte("secret-body")}, nil)
	if err != nil {
		t.Fatalf("Output(cat): unexpected err %v", err)
	}
	if !strings.Contains(string(out), "secret-body") {
		t.Errorf("Output(cat) = %q, want it to contain %q", out, "secret-body")
	}
}

// TestExecOutputInputFail mirrors TestExecOutputFail with stdin.
func TestExecOutputInputFail(t *testing.T) {
	helperOn(t)
	argv := helperArgv(t, "fail")
	out, err := Output(context.Background(), Exec{}, Cmd{Argv: argv, Stdin: []byte("secret-body")}, nil)
	if err == nil {
		t.Fatalf("Output(fail): expected err, got nil (out=%q)", out)
	}
	if !strings.Contains(err.Error(), argv[0]) {
		t.Errorf("Output(fail) err = %q, want it to contain binary name %q", err, argv[0])
	}
}

// TestExecEchoesOnEveryMethod: every path resolves and announces, not just Run. A path
// that skipped it would be a hole in exactly the transparency a verbose mode provides --
// and the capture helpers are the ones that read cluster state.
func TestExecEchoesOnEveryMethod(t *testing.T) {
	helperOn(t)
	argv := helperArgv(t, "echo", "x")
	stdin, stdout, stderr := attachFiles(t)
	cases := []struct {
		method string
		call   func(Exec) error
	}{
		{"Run", func(e Exec) error { return e.Run(context.Background(), Cmd{Argv: argv}, nil, nil) }},
		{"Attach", func(e Exec) error { _, err := e.Attach(Cmd{Argv: argv}, stdin, stdout, stderr); return err }},
		{"Combined", func(e Exec) error { _, err := Combined(context.Background(), e, Cmd{Argv: argv}); return err }},
		{"Split", func(e Exec) error { _, _, err := Split(context.Background(), e, Cmd{Argv: argv}); return err }},
		{"Output", func(e Exec) error { _, err := Output(context.Background(), e, Cmd{Argv: argv}, nil); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.method, func(t *testing.T) {
			var paths []string
			var gotArgs [][]string
			e := Exec{Announce: func(path string, args []string) {
				paths = append(paths, path)
				gotArgs = append(gotArgs, args)
			}}
			if err := tc.call(e); err != nil {
				t.Fatalf("%s: %v", tc.method, err)
			}
			if len(paths) != 1 || !filepath.IsAbs(paths[0]) {
				t.Fatalf("%s announced %q, want one absolute path", tc.method, paths)
			}
			if strings.Join(gotArgs[0], " ") != strings.Join(argv[1:], " ") {
				t.Errorf("%s announced args %q, want %q", tc.method, gotArgs[0], argv[1:])
			}
		})
	}
}

// TestResolveMissingBinaryIsActionable: a name that resolves nowhere fails before any
// process starts, naming what was not found.
func TestResolveMissingBinaryIsActionable(t *testing.T) {
	const missing = "go-deploy-utility-no-such-binary-exists"
	_, err := Resolve(missing)
	if err == nil {
		t.Fatal("resolving a binary that is not on PATH must fail")
	}
	runErr := (Exec{}).Run(context.Background(), Cmd{Argv: []string{missing, "x"}}, nil, nil)
	if runErr == nil {
		t.Fatal("running a binary that is not on PATH must fail")
	}
	for _, got := range []error{err, runErr} {
		for _, want := range []string{missing, "not found on PATH"} {
			if !strings.Contains(got.Error(), want) {
				t.Errorf("error = %v, want it to contain %q", got, want)
			}
		}
	}
}

// writeProbe writes an executable file into dir that LookPath would accept, named so it
// cannot collide with anything installed, and returns its base name without extension.
func writeProbe(t *testing.T, dir string) string {
	t.Helper()
	base := "go-deploy-utility-cwd-probe"
	file, body, mode := base, "#!/bin/sh\nexit 0\n", os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		// Windows resolves by PATHEXT; .bat is in the default list.
		file, body, mode = base+".bat", "@echo off\r\nexit /b 0\r\n", 0o666
	}
	if err := os.WriteFile(filepath.Join(dir, file), []byte(body), mode); err != nil {
		t.Fatalf("writing the probe: %v", err)
	}
	return base
}

// TestResolveRefusesCurrentDirectory: even a bare name must not resolve to a file
// sitting in the working directory -- the "binary shipped alongside the env file" case.
// Driven through Exec.Run, so what is pinned is that nothing executes.
func TestResolveRefusesCurrentDirectory(t *testing.T) {
	dir := t.TempDir()
	base := writeProbe(t, dir)
	t.Chdir(dir)

	// Which layer stops it depends on the host: Windows resolves the current directory
	// implicitly and Go reports exec.ErrDot; unix does not resolve it at all unless "."
	// is on PATH (TestResolveRefusesARelativeResult forces that case).
	_, lookErr := exec.LookPath(base)
	err := (Exec{}).Run(context.Background(), Cmd{Argv: []string{base, "arg"}}, nil, nil)
	if err == nil {
		t.Fatal("a binary in the current directory was executed; it must never resolve from cwd")
	}
	if errors.Is(lookErr, exec.ErrDot) {
		if !strings.Contains(err.Error(), "current directory") {
			t.Errorf("error = %v, want it to say the current directory was refused", err)
		}
		return
	}
	t.Logf("host does not resolve from the current directory (LookPath: %v); "+
		"the ErrDot branch is exercised where it does", lookErr)
	if !strings.Contains(err.Error(), "not found on PATH") {
		t.Errorf("error = %v, want the not-found message", err)
	}
}

// ---- new ----------------------------------------------------------------------------

func TestCmdCheck(t *testing.T) {
	ok := []Cmd{
		{Argv: []string{"kubectl"}},
		{Argv: []string{"kubectl", ""}},
		{Argv: []string{"docker"}, Env: []string{"A=", "B=c=d"}},
	}
	for _, c := range ok {
		if err := c.Check(); err != nil {
			t.Errorf("%q: Check = %v, want nil", c.Argv, err)
		}
	}
	bad := []struct {
		name string
		c    Cmd
		want string
	}{
		{"nil argv", Cmd{}, "Cmd.Argv is empty"},
		{"empty argv", Cmd{Argv: []string{}}, "Cmd.Argv is empty"},
		{"empty program", Cmd{Argv: []string{"", "x"}}, "Cmd.Argv[0] is empty"},
		{"no equals", Cmd{Argv: []string{"docker"}, Env: []string{"A=b", "hunter2"}}, "Cmd.Env[1] is not KEY=VALUE"},
		{"empty key", Cmd{Argv: []string{"docker"}, Env: []string{"=hunter2"}}, "Cmd.Env[0] is not KEY=VALUE"},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.c.Check()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Check = %v, want it to contain %q", err, tc.want)
			}
			if strings.Contains(err.Error(), "hunter2") {
				t.Errorf("Check printed the refused entry: %v", err)
			}
		})
	}
}

// TestRunnerImplementations pins, at compile time, that both production runners
// satisfy the interface. Held inside a test so the package keeps no package-level var.
func TestRunnerImplementations(t *testing.T) {
	var r Runner = Exec{}
	r = Echo{W: io.Discard}
	if err := r.Run(context.Background(), Cmd{Argv: []string{"x"}}, nil, nil); err != nil {
		t.Errorf("Echo.Run through the interface: %v", err)
	}
}

// TestRunStdinAndEnvTogether is the capability solace-util lacked: one call that feeds
// stdin and adds to the environment.
func TestRunStdinAndEnvTogether(t *testing.T) {
	helperOn(t)
	out, err := Combined(context.Background(), Exec{}, Cmd{
		Argv:  helperArgv(t, "stdin"),
		Stdin: []byte("payload"),
		Env:   []string{"RUNNER_TEST_EXTRA_ONLY=x"},
	})
	if err != nil || out != "payload" {
		t.Fatalf("stdin with env: out=%q err=%v", out, err)
	}
	out, err = Combined(context.Background(), Exec{}, Cmd{
		Argv:  helperArgv(t, "env", "RUNNER_TEST_EXTRA_ONLY"),
		Stdin: []byte("ignored"),
		Env:   []string{"RUNNER_TEST_EXTRA_ONLY=both-arrived"},
	})
	if err != nil || out != "both-arrived" {
		t.Fatalf("env with stdin: out=%q err=%v", out, err)
	}
}

// TestErrorsNeverCarryArgsStdinOrEnv: a failure names the program and nothing it was
// handed. An argument can carry a token, stdin a password, the environment a secret.
func TestErrorsNeverCarryArgsStdinOrEnv(t *testing.T) {
	helperOn(t)
	const secret = "s3cr3t-value"
	argv := append(helperArgv(t, "fail"), "--password="+secret)
	stdin, stdout, stderr := attachFiles(t)
	runErr := (Exec{}).Run(context.Background(), Cmd{
		Argv:  argv,
		Stdin: []byte(secret),
		Env:   []string{"SECRET=" + secret},
	}, nil, nil)
	_, attachErr := (Exec{}).Attach(Cmd{Argv: []string{"go-deploy-utility-no-such-binary", secret}, Env: []string{"SECRET=" + secret}},
		stdin, stdout, stderr)
	_, missErr := Combined(context.Background(), Exec{}, Cmd{Argv: []string{"go-deploy-utility-no-such-binary", secret}})
	for _, err := range []error{runErr, attachErr, missErr} {
		if err == nil {
			t.Fatal("expected a failure")
		}
		if strings.Contains(err.Error(), secret) {
			t.Errorf("error carries the secret: %v", err)
		}
	}
}

func TestTypedNilFileWriterIsRefused(t *testing.T) {
	var nilFile *os.File
	for _, tc := range []struct {
		name           string
		stdout, stderr io.Writer
	}{
		{"stdout", nilFile, nil},
		{"stderr", nil, nilFile},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, r := range []Runner{Exec{}, Echo{W: io.Discard}} {
				err := r.Run(context.Background(), Cmd{Argv: []string{"go"}}, tc.stdout, tc.stderr)
				if err == nil || !strings.Contains(err.Error(), "the "+tc.name+" writer is a nil *os.File") {
					t.Errorf("%T: err = %v, want the typed-nil refusal for %s", r, err, tc.name)
				}
			}
		})
	}
}

// TestResolveRefusesARelativeResult covers both ways a resolution can end relative to
// the working directory: an explicit ./name, and "." on PATH (exec.ErrDot on every OS).
func TestResolveRefusesARelativeResult(t *testing.T) {
	dir := t.TempDir()
	base := writeProbe(t, dir)
	t.Chdir(dir)

	_, err := Resolve("./" + base)
	if err == nil || !strings.Contains(err.Error(), "relative path") {
		t.Errorf("explicit ./%s: err = %v, want the relative-path refusal", base, err)
	}

	t.Setenv("PATH", "."+string(os.PathListSeparator)+os.Getenv("PATH"))
	_, err = Resolve(base)
	if err == nil || !strings.Contains(err.Error(), "current directory") {
		t.Errorf("%s via . on PATH: err = %v, want the current-directory refusal", base, err)
	}
}

// TestCancelInterruptsBeforeKill: a cancelled child is interrupted first, so it can stop
// cleanly. Windows cannot send an interrupt; there the kill at the end of the grace
// period is the only stop, which TestCancelKillsAfterGrace covers.
func TestCancelInterruptsBeforeKill(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("an interrupt cannot be sent to a child process on Windows")
	}
	helperOn(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready, isReady := waitFor("ready")
	var out bytes.Buffer
	var mu sync.Mutex
	w := writerFunc(func(p []byte) (int, error) {
		mu.Lock()
		out.Write(p)
		mu.Unlock()
		return ready.Write(p)
	})

	errc := make(chan error, 1)
	go func() {
		errc <- (Exec{Grace: time.Minute}).Run(ctx, Cmd{Argv: helperArgv(t, "trapint")}, w, io.Discard)
	}()
	select {
	case <-isReady:
	case <-time.After(30 * time.Second):
		t.Fatal("the child never became ready")
	}
	cancel()
	select {
	case err := <-errc:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want it to wrap context.Canceled", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the run did not end on the interrupt; the grace period is a minute, so it was not killed")
	}
	mu.Lock()
	defer mu.Unlock()
	if !strings.Contains(out.String(), "interrupted") {
		t.Errorf("the child never saw the interrupt; output %q", out.String())
	}
}

// TestCancelKillsAfterGrace: a child that ignores the interrupt is killed when the grace
// period ends.
func TestCancelKillsAfterGrace(t *testing.T) {
	helperOn(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w, isReady := waitFor("ready")
	errc := make(chan error, 1)
	go func() {
		errc <- (Exec{Grace: 100 * time.Millisecond}).Run(ctx, Cmd{Argv: helperArgv(t, "ignoreint")}, w, io.Discard)
	}()
	select {
	case <-isReady:
	case <-time.After(30 * time.Second):
		t.Fatal("the child never became ready")
	}
	cancel()
	select {
	case err := <-errc:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v, want it to wrap context.Canceled", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("a child that ignores the interrupt was not killed after the grace period")
	}
}

func TestDeadlineIsReportedAsDeadline(t *testing.T) {
	helperOn(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := (Exec{Grace: 100 * time.Millisecond}).Run(ctx, Cmd{Argv: helperArgv(t, "drip")}, io.Discard, io.Discard)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want it to wrap context.DeadlineExceeded", err)
	}
	if !strings.Contains(err.Error(), "stopped") {
		t.Errorf("err = %v, want it to say the command was stopped", err)
	}
}

// TestCancelledBeforeStartRunsAndAnnouncesNothing: a context that is already done
// starts nothing, and a verbose trail does not claim a command ran.
func TestCancelledBeforeStartRunsAndAnnouncesNothing(t *testing.T) {
	helperOn(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	announced := false
	var out bytes.Buffer
	err := (Exec{Announce: func(string, []string) { announced = true }}).Run(ctx,
		Cmd{Argv: helperArgv(t, "echo", "ran")}, &out, nil)
	if !errors.Is(err, context.Canceled) || !strings.Contains(err.Error(), "not started") {
		t.Errorf("err = %v, want a not-started error wrapping context.Canceled", err)
	}
	if announced || out.Len() != 0 {
		t.Errorf("announced=%v output=%q, want nothing", announced, out.String())
	}
}

// TestSuccessfulExitWithAHeldPipeSucceeds: a command that exits 0 while a process it
// started still holds its output open succeeds once the grace period ends, instead of
// waiting for that process forever.
func TestSuccessfulExitWithAHeldPipeSucceeds(t *testing.T) {
	helperOn(t)
	start := time.Now()
	out, err := Combined(context.Background(), Exec{Grace: 200 * time.Millisecond}, Cmd{Argv: helperArgv(t, "orphan")})
	// The grandchild would otherwise outlive the test by its sleep, and on Windows a
	// running copy of the test binary keeps it locked against go test's cleanup.
	var pid int
	if _, scanErr := fmt.Sscanf(out, "grandchild=%d", &pid); scanErr == nil {
		if p, findErr := os.FindProcess(pid); findErr == nil {
			t.Cleanup(func() { _ = p.Kill() })
		}
	}
	if err != nil {
		t.Fatalf("err = %v, want success: the command itself exited 0", err)
	}
	if !strings.Contains(out, "parent-done") {
		t.Errorf("output = %q, want the command's own output", out)
	}
	if d := time.Since(start); d > 6*time.Second {
		t.Errorf("the run took %v; it waited for the grandchild instead of the grace period", d)
	}
}

// TestSplitAndOutputKeepBytesOnFailure: callers scan a failing command's output for the
// reason, so the capture helpers return it with the error.
func TestSplitAndOutputKeepBytesOnFailure(t *testing.T) {
	helperOn(t)
	out, _, err := Split(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "fail")})
	if err == nil || out != "before-exit\n" {
		t.Errorf("Split(fail) = %q, %v; want the output and an error", out, err)
	}
	b, err := Output(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "fail")}, nil)
	if err == nil || string(b) != "before-exit\n" {
		t.Errorf("Output(fail) = %q, %v; want the output and an error", b, err)
	}
}

// TestOutputStreamsStderrLive: Output captures stdout only; stderr goes to the writer
// it was given.
func TestOutputStreamsStderrLive(t *testing.T) {
	helperOn(t)
	var errOut bytes.Buffer
	b, err := Output(context.Background(), Exec{}, Cmd{Argv: helperArgv(t, "both")}, &errOut)
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if string(b) != "stdout-line\n" || errOut.String() != "stderr-line\n" {
		t.Errorf("Output = %q, stderr writer = %q; want each stream in its own place", b, errOut.String())
	}
}

// TestOSAttachSignalledChildIsFailure: a session killed by a signal reports -1 from
// os/exec, which no caller can exit with; Attach reports exitcode.Failure. On Windows the
// kill sets a real status of 1, which is the same answer.
func TestOSAttachSignalledChildIsFailure(t *testing.T) {
	helperOn(t)
	stdin, stdout, stderr := attachFiles(t)
	code, err := (Exec{}).Attach(Cmd{Argv: helperArgv(t, "killself")}, stdin, stdout, stderr)
	if err != nil || code != 1 {
		t.Errorf("Attach(killself) = (%d, %v), want (1, nil)", code, err)
	}
}

// TestAttachStartFailureIsAnError: a file that resolves but cannot start never ran, so
// it is an error rather than an exit status.
func TestAttachStartFailureIsAnError(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "not-a-program")
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(name, []byte("not a program\n"), 0o755); err != nil {
		t.Fatalf("writing the file: %v", err)
	}
	stdin, stdout, stderr := attachFiles(t)
	code, err := (Exec{}).Attach(Cmd{Argv: []string{name}}, stdin, stdout, stderr)
	if err == nil || code != 0 {
		t.Fatalf("Attach = (%d, %v), want (0, an error)", code, err)
	}
	if !strings.Contains(err.Error(), name) {
		t.Errorf("err = %v, want it to name the program", err)
	}
}
