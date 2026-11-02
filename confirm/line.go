package confirm

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Line writes prompt to out and returns one trimmed line read from in.
//
// answered is false when the stream ended before any input arrived. That is NOT
// the same as a reply of "no": it means nobody was in a position to reply at all,
// and callers report the two differently. A final line with no trailing newline
// still counts as an answer.
func Line(in io.Reader, out io.Writer, prompt string) (text string, answered bool) {
	fmt.Fprint(out, prompt)
	// Reuse the caller's reader when it already is one (a Prompter hands over the
	// run's single memoised reader) so nothing buffered past this line is lost; wrap
	// only a bare reader.
	br, ok := in.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(in)
	}
	line, err := br.ReadString('\n')
	if line == "" && err != nil {
		// The prompt deliberately ends mid-line and EOF supplies no newline of its
		// own, so without this the next message would be appended to the question.
		fmt.Fprintln(out)
		return "", false
	}
	return strings.TrimSpace(line), true
}

// YesNo returns true for y/yes (case-insensitive) -- the lenient form used to
// confirm a reversible action. answered carries Line's meaning.
func YesNo(in io.Reader, out io.Writer, prompt string) (yes, answered bool) {
	line, ok := Line(in, out, prompt)
	if !ok {
		return false, false
	}
	switch strings.ToLower(line) {
	case "y", "yes":
		return true, true
	}
	return false, true
}

// Yes returns true only for an exact "yes" (case-insensitive) -- the strict form
// required before an irreversible purge.
func Yes(in io.Reader, out io.Writer, prompt string) (yes, answered bool) {
	line, ok := Line(in, out, prompt)
	if !ok {
		return false, false
	}
	return strings.ToLower(line) == "yes", true
}
