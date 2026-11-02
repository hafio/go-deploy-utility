package confirm

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func terminal() bool    { return true }
func notTerminal() bool { return false }

// countingReader records whether anything read from it.
type countingReader struct {
	r     io.Reader
	reads int
}

func (c *countingReader) Read(p []byte) (int, error) {
	c.reads++
	return c.r.Read(p)
}

// TestPromptKeepsBufferedAnswersForLaterQuestions: two questions answered by one
// write ("y\nyes\n") both get their answer. A fresh reader per question would
// buffer the second answer away and report the second question unanswered.
func TestPromptKeepsBufferedAnswersForLaterQuestions(t *testing.T) {
	var out strings.Builder
	p := New(Options{IsTerminal: terminal, In: strings.NewReader("y\nyes\n"), Out: &out})
	first, err := p.Gate(false, Lenient, "Delete broker x? [y/N] ", "the delete")
	if err != nil || !first {
		t.Errorf(`first question answered "y" = %v, %v; want true, nil`, first, err)
	}
	second, err := p.Gate(false, Strict, "Purge the data? type yes: ", "the purge")
	if err != nil || !second {
		t.Errorf(`second question answered "yes" = %v, %v; want true, nil -- the answer was buffered away by the first`, second, err)
	}
}

// TestPromptSeamsBeatStdin: a Prompter given IsTerminal and In seams never
// touches os.Stdin. Here os.Stdin is a closed pipe, so any read of it would come
// back unanswered and the gate would fail.
func TestPromptSeamsBeatStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	defer r.Close()
	swapStdin(t, r)

	var out strings.Builder
	p := New(Options{IsTerminal: terminal, In: strings.NewReader("y\n"), Out: &out})
	ok, err := p.Gate(false, Lenient, "Overwrite out.cli? [y/N] ", "the overwrite")
	if err != nil || !ok {
		t.Errorf("Gate through the seams = %v, %v; want true, nil", ok, err)
	}
}

// TestPrompterDefaultsFromZeroOptions: an empty Options and a nil *Prompter both
// fall back to stdin, stderr and the default flag without panicking.
func TestPrompterDefaultsFromZeroOptions(t *testing.T) {
	for name, p := range map[string]*Prompter{"zero options": New(Options{}), "nil prompter": nil} {
		_ = p.Terminal() // the real stdin's answer is host-shaped; only "no panic" is asserted
		if p.out() != os.Stderr {
			t.Errorf("%s: questions go to %v, want os.Stderr", name, p.out())
		}
		if p.flag() != DefaultNoPromptFlag {
			t.Errorf("%s: flag = %q, want %q", name, p.flag(), DefaultNoPromptFlag)
		}
		if p.in() == nil {
			t.Errorf("%s: no reader", name)
		}
	}
	p := New(Options{})
	if p.in() != p.in() {
		t.Error("a Prompter built a second reader; it must keep one per run")
	}
	var calls int
	q := New(Options{IsTerminal: func() bool { calls++; return false }})
	if q.Terminal() || calls != 1 {
		t.Errorf("Terminal() with a seam = %v after %d calls, want false after 1", q.Terminal(), calls)
	}
}

// TestGateNoPromptSkipsEverything: --no-prompt confirms without asking, without
// probing the terminal, and without reading.
func TestGateNoPromptSkipsEverything(t *testing.T) {
	in := &countingReader{r: strings.NewReader("n\n")}
	var out strings.Builder
	p := New(Options{
		IsTerminal: func() bool { t.Error("the terminal was probed under --no-prompt"); return true },
		In:         in,
		Out:        &out,
	})
	ok, err := p.Gate(true, Strict, "Purge? ", "the purge")
	if err != nil || !ok {
		t.Errorf("Gate(noPrompt) = %v, %v; want true, nil", ok, err)
	}
	if in.reads != 0 || out.Len() != 0 {
		t.Errorf("Gate(noPrompt) read %d times and wrote %q, want nothing", in.reads, out.String())
	}
}

// TestGateNonTerminalFailsNamingTheFlag is the settled rule: no terminal and no
// --no-prompt fails at once, naming the flag, and the piped answer is never read.
func TestGateNonTerminalFailsNamingTheFlag(t *testing.T) {
	in := &countingReader{r: strings.NewReader("y\n")}
	var out strings.Builder
	p := New(Options{IsTerminal: notTerminal, In: in, Out: &out})
	ok, err := p.Gate(false, Lenient, "remove docker: this tears down x -- continue? [y/N] ", "the teardown")
	if ok {
		t.Error("Gate on a non-terminal confirmed")
	}
	const want = "confirming the teardown interactively: stdin is not a terminal, so this prompt " +
		"cannot be answered interactively; pass --no-prompt instead"
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
	if !IsNotTerminal(err) {
		t.Error("the error does not carry *NotTerminalError")
	}
	if in.reads != 0 || out.Len() != 0 {
		t.Errorf("Gate read %d times and wrote %q on a non-terminal, want nothing", in.reads, out.String())
	}
}

// TestGateCustomNoPromptFlagAndDoing: the flag and the action text are the tool's.
func TestGateCustomNoPromptFlagAndDoing(t *testing.T) {
	p := New(Options{IsTerminal: notTerminal, NoPromptFlag: "--yes", Out: io.Discard})
	_, err := p.Gate(false, Lenient, "? ", "the namespace removal")
	if err == nil || !strings.HasPrefix(err.Error(), "confirming the namespace removal interactively: ") ||
		!strings.HasSuffix(err.Error(), "; pass --yes instead") {
		t.Errorf("err = %v, want the tool's action and flag", err)
	}
}

// TestGateTerminalMatrix: on a terminal each reply gives its answer, with no
// error, and the question is written exactly once.
func TestGateTerminalMatrix(t *testing.T) {
	for _, tc := range []struct {
		form  Form
		reply string
		want  bool
	}{
		{Lenient, "y\n", true}, {Lenient, "yes\n", true}, {Lenient, "Y\n", true},
		{Lenient, "n\n", false}, {Lenient, "\n", false}, {Lenient, "maybe\n", false},
		{Strict, "yes\n", true}, {Strict, "YES\n", true}, {Strict, "y\n", false},
	} {
		var out strings.Builder
		p := New(Options{IsTerminal: terminal, In: strings.NewReader(tc.reply), Out: &out})
		ok, err := p.Gate(false, tc.form, "Q? ", "x")
		if err != nil || ok != tc.want {
			t.Errorf("form %d reply %q: Gate = %v, %v; want %v, nil", tc.form, tc.reply, ok, err, tc.want)
		}
		if out.String() != "Q? " {
			t.Errorf("form %d reply %q: wrote %q, want the question once", tc.form, tc.reply, out.String())
		}
	}
}

// TestGateUnansweredOnTerminalErrors: Ctrl-D at the question is an error with
// the flag named, not a silent decline.
func TestGateUnansweredOnTerminalErrors(t *testing.T) {
	var out strings.Builder
	p := New(Options{IsTerminal: terminal, In: strings.NewReader(""), Out: &out})
	ok, err := p.Gate(false, Lenient, "Q? ", "the teardown")
	if ok {
		t.Error("an unanswered gate confirmed")
	}
	const want = "confirming the teardown interactively: reading input: EOF; pass --no-prompt instead"
	if err == nil || err.Error() != want || !errors.Is(err, io.EOF) {
		t.Errorf("err = %v, want %q wrapping io.EOF", err, want)
	}
	if out.String() != "Q? \n" {
		t.Errorf("wrote %q, want the question and a newline", out.String())
	}
}

// TestAskNonTerminalNeverWrites: Ask and ReadLine refuse a non-terminal before
// writing the question or reading the answer.
func TestAskNonTerminalNeverWrites(t *testing.T) {
	in := &countingReader{r: strings.NewReader("yes\n")}
	var out strings.Builder
	p := New(Options{IsTerminal: notTerminal, In: in, Out: &out})
	for _, f := range []Form{Lenient, Strict} {
		if yes, answered, err := p.Ask("Q? ", f); yes || answered || !IsNotTerminal(err) {
			t.Errorf("Ask form %d = %v, %v, %v; want false, false, *NotTerminalError", f, yes, answered, err)
		}
	}
	if text, answered, err := p.ReadLine("Q? "); text != "" || answered || !IsNotTerminal(err) {
		t.Errorf("ReadLine = %q, %v, %v; want empty, false, *NotTerminalError", text, answered, err)
	}
	if in.reads != 0 || out.Len() != 0 {
		t.Errorf("read %d times and wrote %q, want nothing", in.reads, out.String())
	}
}

// TestAskUnansweredIsNotAnError: Ask reports end of input as unanswered and
// leaves the wording to the caller; a last line without a newline still answers.
func TestAskUnansweredIsNotAnError(t *testing.T) {
	p := New(Options{IsTerminal: terminal, In: strings.NewReader(""), Out: io.Discard})
	if yes, answered, err := p.Ask("Q? ", Lenient); yes || answered || err != nil {
		t.Errorf("Ask at end of input = %v, %v, %v; want false, false, nil", yes, answered, err)
	}
	p = New(Options{IsTerminal: terminal, In: strings.NewReader("y"), Out: io.Discard})
	if yes, answered, err := p.Ask("Q? ", Lenient); !yes || !answered || err != nil {
		t.Errorf("Ask with \"y\" and no newline = %v, %v, %v; want true, true, nil", yes, answered, err)
	}
}

// TestReadLineReturnsTrimmedText: a free-text answer, such as a menu choice,
// comes back trimmed.
func TestReadLineReturnsTrimmedText(t *testing.T) {
	var out strings.Builder
	p := New(Options{IsTerminal: terminal, In: strings.NewReader("  2 \n"), Out: &out})
	text, answered, err := p.ReadLine("Choose [1-3]: ")
	if text != "2" || !answered || err != nil {
		t.Errorf("ReadLine = %q, %v, %v; want \"2\", true, nil", text, answered, err)
	}
	if out.String() != "Choose [1-3]: " {
		t.Errorf("wrote %q, want the prompt", out.String())
	}
}

// TestNotTerminalErrorMessage pins the cause text the tools' error lines embed.
func TestNotTerminalErrorMessage(t *testing.T) {
	const want = "stdin is not a terminal, so this prompt cannot be answered interactively"
	if got := (&NotTerminalError{}).Error(); got != want {
		t.Errorf("NotTerminalError = %q, want %q", got, want)
	}
}

// TestIsNotTerminalThroughWrapping: the check sees through wrapping and rejects
// unrelated errors.
func TestIsNotTerminalThroughWrapping(t *testing.T) {
	if !IsNotTerminal(fmt.Errorf("remove: %w", &NotTerminalError{})) {
		t.Error("IsNotTerminal missed a wrapped *NotTerminalError")
	}
	for _, err := range []error{nil, errors.New("other"), Unanswered()} {
		if IsNotTerminal(err) {
			t.Errorf("IsNotTerminal(%v) = true", err)
		}
	}
}

// TestUnansweredWrapsEOF: the end-of-input error keeps io.EOF for errors.Is.
func TestUnansweredWrapsEOF(t *testing.T) {
	err := Unanswered()
	if !errors.Is(err, io.EOF) || err.Error() != "reading input: EOF" {
		t.Errorf("Unanswered() = %v, want \"reading input: EOF\" wrapping io.EOF", err)
	}
}
