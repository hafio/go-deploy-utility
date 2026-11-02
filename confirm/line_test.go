package confirm

import (
	"bufio"
	"strings"
	"testing"
)

// TestPromptYesNo pins the lenient confirmation: y/yes (any case) accept,
// everything else declines.
//
// The second return is the distinction the whole prompt layer rests on. An empty
// stream is UNANSWERED, not a "no": a reply of "n" is the operator's decision,
// while EOF means nobody was there to make one. A bare "\n" is a real answer --
// someone pressed Enter -- and declines. A final line with no trailing newline
// still counts.
func TestPromptYesNo(t *testing.T) {
	cases := []struct {
		in           string
		want         bool
		wantAnswered bool
	}{
		{"y\n", true, true}, {"yes\n", true, true}, {"Y\n", true, true}, {"YES\n", true, true},
		{"n\n", false, true}, {"no\n", false, true}, {"\n", false, true}, {"maybe\n", false, true},
		{"y", true, true},  // no trailing newline: still an answer
		{"", false, false}, // EOF with no input: nobody answered
	}
	for _, tc := range cases {
		var out strings.Builder
		got, answered := YesNo(strings.NewReader(tc.in), &out, "? ")
		if got != tc.want || answered != tc.wantAnswered {
			t.Errorf("YesNo(%q) = (%v, %v), want (%v, %v)", tc.in, got, answered, tc.want, tc.wantAnswered)
		}
	}
}

// TestPromptYes pins the strict confirmation: only an exact (trimmed,
// case-insensitive) "yes" accepts; a bare "y" is not enough. Same
// answered-vs-declined split as TestPromptYesNo.
func TestPromptYes(t *testing.T) {
	cases := []struct {
		in           string
		want         bool
		wantAnswered bool
	}{
		{"yes\n", true, true}, {"YES\n", true, true}, {"  yes  \n", true, true},
		{"y\n", false, true}, {"no\n", false, true}, {"\n", false, true}, {"yess\n", false, true},
		{"yes", true, true},
		{"", false, false},
	}
	for _, tc := range cases {
		var out strings.Builder
		got, answered := Yes(strings.NewReader(tc.in), &out, "? ")
		if got != tc.want || answered != tc.wantAnswered {
			t.Errorf("Yes(%q) = (%v, %v), want (%v, %v)", tc.in, got, answered, tc.want, tc.wantAnswered)
		}
	}
}

// TestLineReusesBufferedReader: a *bufio.Reader passed in is used as-is, so bytes
// it buffered past the first line are still there for the next question.
func TestLineReusesBufferedReader(t *testing.T) {
	br := bufio.NewReader(strings.NewReader("first\nsecond\n"))
	var out strings.Builder
	if got, ok := Line(br, &out, "1? "); !ok || got != "first" {
		t.Errorf("first Line = %q, %v; want first, true", got, ok)
	}
	if got, ok := Line(br, &out, "2? "); !ok || got != "second" {
		t.Errorf("second Line = %q, %v; want second, true -- the buffered answer was lost", got, ok)
	}
	if out.String() != "1? 2? " {
		t.Errorf("prompts written = %q, want both prompts and nothing else", out.String())
	}
	// End of input: the prompt gets its own newline so the next message starts clean.
	out.Reset()
	if _, ok := Line(br, &out, "3? "); ok {
		t.Error("Line at end of input reported an answer")
	}
	if out.String() != "3? \n" {
		t.Errorf("unanswered prompt output = %q, want the prompt and a newline", out.String())
	}
}
