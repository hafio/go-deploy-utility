package confirm

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

// DefaultNoPromptFlag is the flag an error tells the operator to pass.
const DefaultNoPromptFlag = "--no-prompt"

// Form is how strict an answer must be.
type Form int

const (
	// Lenient accepts y or yes, in any case.
	Lenient Form = iota

	// Strict accepts only yes, in any case.
	Strict
)

// Options are the Prompter's seams. Every field is optional.
type Options struct {
	// IsTerminal reports whether the operator can answer; nil means StdinIsTerminal.
	IsTerminal func() bool

	// In is where answers are read from; nil means os.Stdin.
	In io.Reader

	// Out is where questions are written; nil means os.Stderr.
	Out io.Writer

	// NoPromptFlag is named in errors; empty means DefaultNoPromptFlag.
	NoPromptFlag string
}

// Prompter asks questions for one run. Methods are safe on a nil *Prompter,
// which behaves like New(Options{}) without the shared reader.
type Prompter struct {
	o      Options
	reader *bufio.Reader
}

// New returns a Prompter with the given seams.
func New(o Options) *Prompter { return &Prompter{o: o} }

// Terminal reports whether the operator can answer, through the IsTerminal seam.
// A tool that must agree with the gate (for example before opening a shell) asks
// this rather than probing stdin itself.
func (p *Prompter) Terminal() bool {
	if p == nil || p.o.IsTerminal == nil {
		return StdinIsTerminal()
	}
	return p.o.IsTerminal()
}

// in returns the run's single buffered reader, creating it on first use.
func (p *Prompter) in() *bufio.Reader {
	if p == nil {
		return bufio.NewReader(os.Stdin)
	}
	if p.reader == nil {
		src := io.Reader(os.Stdin)
		if p.o.In != nil {
			src = p.o.In
		}
		p.reader = bufio.NewReader(src)
	}
	return p.reader
}

func (p *Prompter) out() io.Writer {
	if p == nil || p.o.Out == nil {
		return os.Stderr
	}
	return p.o.Out
}

func (p *Prompter) flag() string {
	if p == nil || p.o.NoPromptFlag == "" {
		return DefaultNoPromptFlag
	}
	return p.o.NoPromptFlag
}

// ReadLine asks prompt and returns the trimmed reply. On a non-terminal it
// returns a *NotTerminalError without writing the prompt or reading anything.
// answered is false when the input ended before a reply.
func (p *Prompter) ReadLine(prompt string) (text string, answered bool, err error) {
	if !p.Terminal() {
		return "", false, &NotTerminalError{}
	}
	text, answered = Line(p.in(), p.out(), prompt)
	return text, answered, nil
}

// Ask asks a yes/no question in the given Form. On a non-terminal it returns a
// *NotTerminalError without writing the prompt or reading anything. End of input
// on a terminal is (false, false, nil): unanswered, which the caller phrases.
func (p *Prompter) Ask(prompt string, f Form) (yes, answered bool, err error) {
	if !p.Terminal() {
		return false, false, &NotTerminalError{}
	}
	if f == Strict {
		yes, answered = Yes(p.in(), p.out(), prompt)
	} else {
		yes, answered = YesNo(p.in(), p.out(), prompt)
	}
	return yes, answered, nil
}

// Gate is the destructive-command confirmation. With noPrompt it confirms without
// asking or reading. Otherwise it asks; a non-terminal or an unanswered prompt is
// an error worded
//
//	confirming <doing> interactively: <cause>; pass --no-prompt instead
//
// which the tool reports with exit status 1. A reply that declines is
// (false, nil), so the tool prints its own "cancelled" and exits 0.
func (p *Prompter) Gate(noPrompt bool, f Form, prompt, doing string) (bool, error) {
	if noPrompt {
		return true, nil
	}
	yes, answered, err := p.Ask(prompt, f)
	if err == nil && !answered {
		err = Unanswered()
	}
	if err != nil {
		return false, fmt.Errorf("confirming %s interactively: %w; pass %s instead", doing, err, p.flag())
	}
	return yes, nil
}

// NotTerminalError is returned when stdin cannot be answered interactively.
// Test for it with IsNotTerminal.
type NotTerminalError struct{}

func (*NotTerminalError) Error() string {
	return "stdin is not a terminal, so this prompt cannot be answered interactively"
}

// IsNotTerminal reports whether err carries a *NotTerminalError anywhere in its chain.
func IsNotTerminal(err error) bool {
	var nt *NotTerminalError
	return errors.As(err, &nt)
}

// Unanswered is the error for input that ended before a reply, worded the way
// Gate reports it, for callers of Ask and ReadLine that need the same message.
func Unanswered() error { return fmt.Errorf("reading input: %w", io.EOF) }
