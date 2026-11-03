package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hafio/go-deploy-utility/execguard"
)

// Echo is a Runner that prints the command it would run instead of running it, one line
// per call, such as "+ kubectl apply -f -  <<< (412 bytes on stdin)" or
// "+ docker compose up -d  <<< (env: ADMIN_PASSWORD=***)". The example lines are inline
// because a doc comment's indented line that starts with '+' is a list item, not code.
//
// It applies the same checks as Exec and writes nothing to the command's own stdout or
// stderr, so a caller that parses output reads nothing and degrades rather than
// mistaking the echoed line for an answer. A tool may type-assert a Runner against Echo
// (a value type) to skip work that only makes sense against a real engine.
type Echo struct {
	// W receives the echoed lines; nil means os.Stdout.
	W io.Writer
}

func (e Echo) w() io.Writer {
	if e.W != nil {
		return e.W
	}
	return os.Stdout
}

// echo prints the line. The stdin bytes are never shown, only their count, and every
// environment value is masked: stdin and the environment are how a secret reaches a
// child without passing through an argv, and echoed output is logged and pasted into
// tickets.
func (e Echo) echo(c Cmd) {
	line := "+ " + Quote(c.Argv[0], c.Argv[1:]...)
	if c.Stdin != nil {
		line += fmt.Sprintf("  <<< (%d bytes on stdin)", len(c.Stdin))
	}
	if len(c.Env) > 0 {
		line += "  <<< (env: " + MaskEnv(c.Env) + ")"
	}
	fmt.Fprintln(e.w(), line)
}

// Run implements Runner. A done ctx is refused as Exec refuses it, and nothing is
// printed.
func (e Echo) Run(ctx context.Context, c Cmd, stdout, stderr io.Writer) error {
	if err := CheckRun(c, stdout, stderr); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%s: not started: %w", c.Argv[0], err)
	}
	e.echo(c)
	return nil
}

// Attach implements Runner. It reports exit code 0.
func (e Echo) Attach(c Cmd, stdin, stdout, stderr *os.File) (int, error) {
	if err := CheckAttach(c, stdin, stdout, stderr); err != nil {
		return 0, err
	}
	e.echo(c)
	return 0, nil
}

// MaskEnv renders KEY=VALUE entries as "KEY=*** KEY2=***": the keys are what a reader
// needs to check the wiring, and every value is dropped. An entry with no '=' shows as
// the quoted key '<malformed>' rather than as itself, since the whole entry may be the
// value.
func MaskEnv(env []string) string {
	masked := make([]string, 0, len(env))
	for _, pair := range env {
		key, _, ok := strings.Cut(pair, "=")
		if !ok {
			key = "<malformed>"
		}
		masked = append(masked, quoteTok(key)+"=***")
	}
	return strings.Join(masked, " ")
}

// Quote renders a command line for display, one token at a time. Display only: nothing
// that runs ever goes through it.
func Quote(name string, args ...string) string {
	var b strings.Builder
	b.WriteString(quoteTok(name))
	for _, a := range args {
		b.WriteByte(' ')
		b.WriteString(quoteTok(a))
	}
	return b.String()
}

// quoteTok renders one token. A token with a rune that does not print -- a control
// character other than tab and newline, an invisible or non-ASCII space rune, invalid
// UTF-8 -- is written Go-escaped, so a display line can never carry a raw terminal
// control. Otherwise a token that is empty, holds whitespace or any character the
// command guard treats as shell syntax, or starts with '~', is single-quoted so it pastes
// back into a shell as one word.
func quoteTok(s string) string {
	if s == "" {
		return "''"
	}
	if !printable(s) {
		return strconv.QuoteToASCII(s)
	}
	if strings.HasPrefix(s, "~") || strings.ContainsAny(s, " \t\n"+execguard.UnsafeChars) {
		return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
	}
	return s
}

func printable(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if r != '\t' && r != '\n' && !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
