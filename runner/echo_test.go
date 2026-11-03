package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// ---- lifted from solace-util internal/engine/runner_test.go -------------------------

func TestQuoteTok(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", "''"},
		{"plain", "kubectl", "kubectl"},
		{"plainDashes", "--dry-run=client", "--dry-run=client"},
		{"spaces", "hello world", "'hello world'"},
		{"tab", "a\tb", "'a\tb'"},
		{"newline", "a\nb", "'a\nb'"},
		{"singleQuote", "it's", `'it'\''s'`},
		{"dollar", "$HOME", "'$HOME'"},
		{"pipe", "a|b", "'a|b'"},
		{"semicolon", "a;b", "'a;b'"},
		{"redirect", "a>b", "'a>b'"},
		{"glob", "*.yaml", "'*.yaml'"},
		{"brackets", "a[0]", "'a[0]'"},
		{"braces", "a{b}", "'a{b}'"},
		{"parens", "a(b)", "'a(b)'"},
		{"doubleQuote", `a"b`, `'a"b'`},
		{"ampersand", "a&b", "'a&b'"},
		{"question", "a?b", "'a?b'"},
		// Added with the shared charset: every character the command guard treats as
		// shell syntax now quotes, and so does a leading '~', which a shell expands.
		{"backtick", "a`b", "'a`b'"},
		{"backslash", `a\b`, `'a\b'`},
		{"hash", "a#b", "'a#b'"},
		{"bang", "a!b", "'a!b'"},
		{"closeOnly", "a)", "'a)'"},
		{"leadingTilde", "~/kube", "'~/kube'"},
		{"embeddedTilde", "a~b", "a~b"},
		{"nonASCIIPrintable", "caf\u00e9", "caf\u00e9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quoteTok(tt.in); got != tt.want {
				t.Errorf("quoteTok(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestQuote(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		args []string
		want string
	}{
		{"nameOnly", "kubectl", nil, "kubectl"},
		{"nameAndArgs", "kubectl", []string{"get", "pods"}, "kubectl get pods"},
		{"quotedArg", "echo", []string{"hello world"}, "echo 'hello world'"},
		{"emptyArg", "echo", []string{""}, "echo ''"},
		{"quoteInArg", "echo", []string{"it's"}, `echo 'it'\''s'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Quote(tt.cmd, tt.args...); got != tt.want {
				t.Errorf("Quote(%q, %q) = %q, want %q", tt.cmd, tt.args, got, tt.want)
			}
		})
	}
}

func TestEchoRun(t *testing.T) {
	var buf bytes.Buffer
	if err := (Echo{W: &buf}).Run(context.Background(), Cmd{Argv: []string{"kubectl", "get", "pods"}}, nil, nil); err != nil {
		t.Fatalf("Run: unexpected err %v", err)
	}
	want := "+ kubectl get pods\n"
	if got := buf.String(); got != want {
		t.Errorf("Run wrote %q, want %q", got, want)
	}
}

func TestEchoRunInteractive(t *testing.T) {
	var buf bytes.Buffer
	stdin, stdout, stderr := attachFiles(t)
	code, err := (Echo{W: &buf}).Attach(Cmd{Argv: []string{"bash", "-c", "echo hi"}}, stdin, stdout, stderr)
	if err != nil || code != 0 {
		t.Fatalf("Attach = (%d, %v), want (0, nil)", code, err)
	}
	want := "+ bash -c 'echo hi'\n"
	if got := buf.String(); got != want {
		t.Errorf("Attach wrote %q, want %q", got, want)
	}
}

func TestEchoRunInput(t *testing.T) {
	var buf bytes.Buffer
	in := []byte("apiVersion: v1")
	if err := (Echo{W: &buf}).Run(context.Background(), Cmd{Argv: []string{"kubectl", "apply", "-f", "-"}, Stdin: in}, nil, nil); err != nil {
		t.Fatalf("Run: unexpected err %v", err)
	}
	want := fmt.Sprintf("+ kubectl apply -f -  <<< (%d bytes on stdin)\n", len(in))
	if got := buf.String(); got != want {
		t.Errorf("Run wrote %q, want %q", got, want)
	}
}

// TestEchoRunEnv: the variable names must be visible (they are what an operator checks)
// and no value may be. The command still leads the line, so "+ <cmd>" stays greppable.
func TestEchoRunEnv(t *testing.T) {
	var buf bytes.Buffer
	env := []string{"SOLACE_ADMIN_PASSWORD=hunter2", "SOLACE_REDUNDANCY_PSK=psk-value"}
	if err := (Echo{W: &buf}).Run(context.Background(), Cmd{Argv: []string{"docker", "compose", "up", "-d"}, Env: env}, nil, nil); err != nil {
		t.Fatalf("Run: unexpected err %v", err)
	}
	want := "+ docker compose up -d  <<< (env: SOLACE_ADMIN_PASSWORD=*** SOLACE_REDUNDANCY_PSK=***)\n"
	if got := buf.String(); got != want {
		t.Errorf("Run wrote %q, want %q", got, want)
	}
}

// TestEchoRunEnvNoEnv: with nothing to annotate the line is exactly the plain one,
// rather than a dangling "(env: )".
func TestEchoRunEnvNoEnv(t *testing.T) {
	var buf bytes.Buffer
	if err := (Echo{W: &buf}).Run(context.Background(), Cmd{Argv: []string{"docker", "compose", "ps"}, Env: nil}, nil, nil); err != nil {
		t.Fatalf("Run: unexpected err %v", err)
	}
	want := "+ docker compose ps\n"
	if got := buf.String(); got != want {
		t.Errorf("Run wrote %q, want %q", got, want)
	}
}

func TestMaskEnv(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want string
	}{
		{"none", nil, ""},
		{"one", []string{"A=b"}, "A=***"},
		{"two", []string{"A=b", "C=d"}, "A=*** C=***"},
		{"valueWithEquals", []string{"A=b=c"}, "A=***"},
		{"emptyValue", []string{"A="}, "A=***"},
		{"oddName", []string{"A B=c"}, "'A B'=***"},
		// No '=' at all: the entry must not be echoed as if it were a key, since the
		// whole entry may be the value this function exists to hide.
		{"noEquals", []string{"hunter2"}, "'<malformed>'=***"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskEnv(tt.in); got != tt.want {
				t.Errorf("MaskEnv(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestEchoOutput(t *testing.T) {
	var buf bytes.Buffer
	out, err := Output(context.Background(), Echo{W: &buf}, Cmd{Argv: []string{"helm", "version"}}, nil)
	if err != nil {
		t.Fatalf("Output: unexpected err %v", err)
	}
	if out != nil {
		t.Errorf("Output returned %v, want nil", out)
	}
	want := "+ helm version\n"
	if got := buf.String(); got != want {
		t.Errorf("Output wrote %q, want %q", got, want)
	}
}

// TestEchoDefaultWriter: a zero-value Echo (nil W) writes to os.Stdout.
func TestEchoDefaultWriter(t *testing.T) {
	got := captureStdout(t, func() {
		if err := (Echo{}).Run(context.Background(), Cmd{Argv: []string{"kubectl", "get", "pods"}}, nil, nil); err != nil {
			t.Errorf("Run: unexpected err %v", err)
		}
	})
	want := "+ kubectl get pods\n"
	if got != want {
		t.Errorf("Echo{}.Run wrote %q, want %q", got, want)
	}
}

// TestEchoOutputInput: the credential-bearing call (curl -K - with a password on stdin).
// The byte count is visible; the stdin body never reaches echoed output, which is
// printed, logged and pasted into tickets.
func TestEchoOutputInput(t *testing.T) {
	var buf bytes.Buffer
	in := []byte("password=hunter2")
	out, err := Output(context.Background(), Echo{W: &buf}, Cmd{Argv: []string{"curl", "-K", "-"}, Stdin: in}, nil)
	if err != nil {
		t.Fatalf("Output: unexpected err %v", err)
	}
	if out != nil {
		t.Errorf("Output returned %v, want nil", out)
	}
	want := fmt.Sprintf("+ curl -K -  <<< (%d bytes on stdin)\n", len(in))
	if got := buf.String(); got != want {
		t.Errorf("Output wrote %q, want %q", got, want)
	}
	if strings.Contains(buf.String(), "hunter2") {
		t.Errorf("Output leaked the stdin body into echoed output: %q", buf.String())
	}
}

// ---- new ----------------------------------------------------------------------------

// TestEchoStdinAndEnvTogether pins the one-line form for a Cmd carrying both, and that
// a non-nil empty stdin is still annotated, as solace-util's RunInput did.
func TestEchoStdinAndEnvTogether(t *testing.T) {
	var buf bytes.Buffer
	e := Echo{W: &buf}
	if err := e.Run(context.Background(), Cmd{Argv: []string{"docker", "login", "--password-stdin"},
		Stdin: []byte("hunter2"), Env: []string{"DOCKER_CONFIG=/tmp/x"}}, nil, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := e.Run(context.Background(), Cmd{Argv: []string{"kubectl", "apply", "-f", "-"}, Stdin: []byte{}}, nil, nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	want := "+ docker login --password-stdin  <<< (7 bytes on stdin)  <<< (env: DOCKER_CONFIG=***)\n" +
		"+ kubectl apply -f -  <<< (0 bytes on stdin)\n"
	if got := buf.String(); got != want {
		t.Errorf("Echo wrote %q, want %q", got, want)
	}
}

// TestEchoRefusesWhatExecRefuses: a builder bug fails under Echo exactly as it would
// under Exec, and prints nothing.
func TestEchoRefusesWhatExecRefuses(t *testing.T) {
	var buf bytes.Buffer
	e := Echo{W: &buf}
	stdin, stdout, stderr := attachFiles(t)
	if err := e.Run(context.Background(), Cmd{}, nil, nil); err == nil {
		t.Error("Run: an empty argv must be refused")
	}
	if err := e.Run(context.Background(), Cmd{Argv: []string{"docker"}, Env: []string{"hunter2"}}, nil, nil); err == nil {
		t.Error("Run: a malformed env entry must be refused")
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := e.Run(cancelled, Cmd{Argv: []string{"kubectl", "apply"}}, nil, nil); !errors.Is(err, context.Canceled) {
		t.Errorf("Run: a done context must be refused as Exec refuses it, got %v", err)
	}
	if _, err := e.Attach(Cmd{Argv: []string{"bash"}, Stdin: []byte("x")}, stdin, stdout, stderr); err == nil {
		t.Error("Attach: stdin bytes must be refused")
	}
	if _, err := e.Attach(Cmd{Argv: []string{"bash"}}, nil, stdout, stderr); err == nil {
		t.Error("Attach: a nil file must be refused")
	}
	if buf.Len() != 0 {
		t.Errorf("a refused call printed %q", buf.String())
	}
}

// TestQuoteEscapesUnprintable: a display line must never carry a raw terminal control,
// so a token holding one is written Go-escaped. Tab and newline stay inside single
// quotes, as they always have.
func TestQuoteEscapesUnprintable(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"carriageReturn", "a\rb", `"a\rb"`},
		{"escape", "a\x1b[2Jb", `"a\x1b[2Jb"`},
		{"del", "a\x7fb", `"a\x7fb"`},
		{"c1csi", "a\u009bb", `"a\u009bb"`},
		{"nextLine", "a\u0085b", `"a\u0085b"`},
		{"lineSeparator", "a\u2028b", `"a\u2028b"`},
		{"nbsp", "a\u00a0b", `"a\u00a0b"`},
		{"zeroWidthSpace", "a\u200bb", `"a\u200bb"`},
		{"bidiOverride", "a\u202eb", `"a\u202eb"`},
		{"invalidUTF8", "a\xffb", `"a\xffb"`},
		{"tabStaysQuoted", "a\tb", "'a\tb'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quoteTok(tt.in); got != tt.want {
				t.Errorf("quoteTok(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
