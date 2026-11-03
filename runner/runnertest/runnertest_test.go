package runnertest

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hafio/go-deploy-utility/runner"
)

// files opens three temp files to stand in for a terminal.
func files(t *testing.T) (stdin, stdout, stderr *os.File) {
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

func cmd(argv ...string) runner.Cmd { return runner.Cmd{Argv: argv} }

// TestFakeZeroValue: a zero Fake answers every call with the zero Reply and still
// records it, and it satisfies runner.Runner.
func TestFakeZeroValue(t *testing.T) {
	var r runner.Runner = &Fake{}
	out, err := runner.Combined(context.Background(), r, cmd("kubectl", "get", "pods"))
	if out != "" || err != nil {
		t.Errorf("Combined = %q, %v; want the zero reply", out, err)
	}
	stdin, stdout, stderr := files(t)
	code, err := r.Attach(cmd("bash"), stdin, stdout, stderr)
	if code != 0 || err != nil {
		t.Errorf("Attach = %d, %v; want the zero reply", code, err)
	}
	if n := len(r.(*Fake).Calls()); n != 2 {
		t.Errorf("recorded %d calls, want 2", n)
	}
}

func TestFakeRecordsCallsInOrder(t *testing.T) {
	f := &Fake{}
	var out bytes.Buffer
	stdin, stdout, stderr := files(t)
	_ = f.Run(context.Background(), runner.Cmd{Argv: []string{"kubectl", "apply", "-f", "-"}, Stdin: []byte("kind: Pod")},
		&out, os.Stderr)
	_ = f.Run(context.Background(), runner.Cmd{Argv: []string{"docker", "compose", "up"}, Env: []string{"A=b"}}, nil, nil)
	_, _ = f.Attach(runner.Cmd{Argv: []string{"bash"}}, stdin, stdout, stderr)

	calls := f.Calls()
	if len(calls) != 3 {
		t.Fatalf("recorded %d calls, want 3", len(calls))
	}
	if got := calls[0].String(); got != "kubectl apply -f -" {
		t.Errorf("call 0 = %q", got)
	}
	if string(calls[0].Stdin) != "kind: Pod" || calls[0].Stdout != &out || calls[0].Stderr != os.Stderr || calls[0].Attach {
		t.Errorf("call 0 recorded %+v", calls[0])
	}
	if calls[1].Stdin != nil || len(calls[1].Env) != 1 || calls[1].Env[0] != "A=b" || calls[1].Stdout != nil {
		t.Errorf("call 1 recorded %+v", calls[1])
	}
	if !calls[2].Attach || calls[2].Stdout != nil || calls[2].String() != "bash" {
		t.Errorf("call 2 recorded %+v", calls[2])
	}
}

// TestFakeCopiesItsInputs: a caller that reuses its slices, or a test that edits what
// Calls returned, cannot rewrite the record. A non-nil empty stdin stays non-nil.
func TestFakeCopiesItsInputs(t *testing.T) {
	f := &Fake{}
	argv := []string{"kubectl", "get"}
	stdin := []byte("abc")
	env := []string{"A=b"}
	_ = f.Run(context.Background(), runner.Cmd{Argv: argv, Stdin: stdin, Env: env}, nil, nil)
	_ = f.Run(context.Background(), runner.Cmd{Argv: argv, Stdin: []byte{}}, nil, nil)
	argv[1], stdin[0], env[0] = "delete", 'X', "A=changed"

	got := f.Calls()
	got[0].Argv[0] = "edited"
	again := f.Calls()
	if again[0].String() != "kubectl get" || string(again[0].Stdin) != "abc" || again[0].Env[0] != "A=b" {
		t.Errorf("the record changed: %+v", again[0])
	}
	if again[1].Stdin == nil || len(again[1].Stdin) != 0 {
		t.Errorf("a non-nil empty stdin must stay non-nil and empty, got %#v", again[1].Stdin)
	}
}

// TestFakeWritesScriptedStreams: the scripted output goes to the writers the caller
// passed, so every capture helper works against the Fake.
func TestFakeWritesScriptedStreams(t *testing.T) {
	boom := errors.New("boom")
	f := &Fake{Script: func(int, Call) Reply { return Reply{Stdout: "out\n", Stderr: "err\n", Err: boom} }}
	ctx := context.Background()

	combined, err := runner.Combined(ctx, f, cmd("oc", "get"))
	if combined != "out\nerr\n" || !errors.Is(err, boom) {
		t.Errorf("Combined = %q, %v", combined, err)
	}
	out, errOut, err := runner.Split(ctx, f, cmd("oc", "get"))
	if out != "out\n" || errOut != "err\n" || !errors.Is(err, boom) {
		t.Errorf("Split = %q, %q, %v", out, errOut, err)
	}
	var live bytes.Buffer
	b, err := runner.Output(ctx, f, cmd("oc", "get"), &live)
	if string(b) != "out\n" || live.String() != "err\n" || !errors.Is(err, boom) {
		t.Errorf("Output = %q, stderr %q, %v", b, live.String(), err)
	}
	// Nil writers discard, as they do for Exec.
	if err := f.Run(ctx, cmd("oc", "get"), nil, nil); !errors.Is(err, boom) {
		t.Errorf("Run with nil writers = %v", err)
	}
}

// TestFakeSeq answers by position, the way solmq-conn-util's fakes do, and gives the
// zero Reply past the end.
func TestFakeSeq(t *testing.T) {
	boom := errors.New("second fails")
	replies := []Reply{{Stdout: "first"}, {Err: boom}}
	f := &Fake{Script: Seq(replies...)}
	replies[0].Stdout = "edited after Seq"
	ctx := context.Background()

	if out, err := runner.Combined(ctx, f, cmd("a")); out != "first" || err != nil {
		t.Errorf("call 0 = %q, %v", out, err)
	}
	if _, err := runner.Combined(ctx, f, cmd("b")); !errors.Is(err, boom) {
		t.Errorf("call 1 = %v", err)
	}
	if out, err := runner.Combined(ctx, f, cmd("c")); out != "" || err != nil {
		t.Errorf("call 2, past the end = %q, %v; want the zero reply", out, err)
	}
}

// TestFakeFailOn answers by argv content, the way solace-util's fakes do: a token
// containing the text fails, everything else succeeds.
func TestFakeFailOn(t *testing.T) {
	boom := errors.New("refused")
	f := &Fake{Script: FailOn("secret", boom)}
	ctx := context.Background()
	if err := f.Run(ctx, cmd("podman", "secret", "create", "x", "-"), nil, nil); !errors.Is(err, boom) {
		t.Errorf("matching call = %v, want the error", err)
	}
	if err := f.Run(ctx, cmd("podman", "ps"), nil, nil); err != nil {
		t.Errorf("other call = %v, want nil", err)
	}
	if err := f.Run(ctx, runner.Cmd{Argv: []string{"podman", "login"}, Stdin: []byte("secret")}, nil, nil); err != nil {
		t.Errorf("stdin is not argv: %v", err)
	}
}

func TestFakeAttach(t *testing.T) {
	boom := errors.New("no tty")
	f := &Fake{Script: Seq(Reply{Code: 7}, Reply{Err: boom})}
	stdin, stdout, stderr := files(t)
	if code, err := f.Attach(cmd("bash"), stdin, stdout, stderr); code != 7 || err != nil {
		t.Errorf("Attach 0 = %d, %v; want 7, nil", code, err)
	}
	if _, err := f.Attach(cmd("bash"), stdin, stdout, stderr); !errors.Is(err, boom) {
		t.Errorf("Attach 1 = %v", err)
	}
}

// TestFakeRefusesWhatExecRefuses: a builder bug fails under the Fake as it would in
// production, and the refused call is neither recorded nor shown to the Script.
func TestFakeRefusesWhatExecRefuses(t *testing.T) {
	scripted := 0
	f := &Fake{Script: func(int, Call) Reply { scripted++; return Reply{} }}
	stdin, stdout, stderr := files(t)
	var nilFile *os.File
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	for name, err := range map[string]error{
		"empty argv":       f.Run(context.Background(), runner.Cmd{}, nil, nil),
		"malformed env":    f.Run(context.Background(), runner.Cmd{Argv: []string{"x"}, Env: []string{"hunter2"}}, nil, nil),
		"typed-nil writer": f.Run(context.Background(), cmd("x"), nilFile, nil),
		"done context":     f.Run(cancelled, cmd("x"), nil, nil),
	} {
		if err == nil {
			t.Errorf("Run with %s must be refused", name)
		}
	}
	if _, err := f.Attach(runner.Cmd{Argv: []string{"bash"}, Stdin: []byte("x")}, stdin, stdout, stderr); err == nil {
		t.Error("Attach with stdin bytes must be refused")
	}
	if _, err := f.Attach(cmd("bash"), stdin, nil, stderr); err == nil {
		t.Error("Attach with a nil file must be refused")
	}
	if n := len(f.Calls()); n != 0 || scripted != 0 {
		t.Errorf("refused calls reached the record (%d) or the Script (%d)", n, scripted)
	}
}

// TestFakeScriptMayReadCalls: the Script runs outside the lock, so a Script that looks
// at the record does not deadlock.
func TestFakeScriptMayReadCalls(t *testing.T) {
	f := &Fake{}
	f.Script = func(n int, c Call) Reply {
		calls := f.Calls()
		if len(calls) != n+1 || calls[n].String() != c.String() {
			return Reply{Err: errors.New("the Script saw a record that does not end with its own call")}
		}
		return Reply{Stdout: c.String()}
	}
	out, err := runner.Combined(context.Background(), f, cmd("kubectl", "version"))
	if err != nil || out != "kubectl version" {
		t.Errorf("Combined = %q, %v", out, err)
	}
}

// TestFakeConcurrentUse: a Fake shared by goroutines records every call once. Run with
// DEV_RACE=1 to have the race detector check the locking.
func TestFakeConcurrentUse(t *testing.T) {
	f := &Fake{Script: func(int, Call) Reply { return Reply{Stdout: "x"} }}
	const workers, each = 8, 50
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < each; i++ {
				_, _ = runner.Combined(context.Background(), f, cmd("kubectl", "get"))
			}
		}()
	}
	wg.Wait()
	if n := len(f.Calls()); n != workers*each {
		t.Errorf("recorded %d calls, want %d", n, workers*each)
	}
}

// TestCallHasAndString: Has matches inside any argv token, and String quotes as Echo
// does; a hand-built empty Call renders as nothing rather than panicking.
func TestCallHasAndString(t *testing.T) {
	c := Call{Argv: []string{"kubectl", "--context=prod", "get", "pods"}}
	if !c.Has("prod") || !c.Has("kubectl") || c.Has("staging") {
		t.Errorf("Has misreports on %v", c.Argv)
	}
	if got := (Call{Argv: []string{"echo", "hello world"}}).String(); got != "echo 'hello world'" {
		t.Errorf("String = %q", got)
	}
	if got := (Call{}).String(); got != "" {
		t.Errorf("empty Call String = %q", got)
	}
}
