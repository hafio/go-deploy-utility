package writer_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/hafio/go-deploy-utility/containerart/workload"
	"github.com/hafio/go-deploy-utility/containerart/writer"
)

// block writes text as the value of the key content in dialect d.
func block(t *testing.T, d writer.Dialect, text string) string {
	t.Helper()
	return render(t, writer.Map(writer.KV("content", writer.Block(text))), d)
}

// TestBlockComposeMarkers: C10. Compose writes the bytes as given under the marker
// the text needs: |- for no final newline, | for one, |+ for several. A text of
// newlines only is kept with |+, and an empty text is |- with no line.
func TestBlockComposeMarkers(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"a", "content: |-\n  a\n"},
		{"a\nb", "content: |-\n  a\n  b\n"},
		{"a\nb\n", "content: |\n  a\n  b\n"},
		{"a\n\n", "content: |+\n  a\n\n"},
		{"a\n\n\n", "content: |+\n  a\n\n\n"},
		{"", "content: |-\n"},
		{"\n", "content: |+\n\n"},
		{"\n\n", "content: |+\n\n\n"},
		{"echo $HOME\n", "content: |\n  echo $$HOME\n"},
		{"tab\there\n", "content: |\n  tab\there\n"},
		{"caf\u00e9 \U0001F600\n", "content: |\n  caf\u00e9 \U0001F600\n"},
	} {
		if got := block(t, writer.Compose, c.in); got != c.want {
			t.Errorf("Block(%q) on compose = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestBlockIndentIndicator: a first non-empty line that starts with a space adds the
// indent indicator 2, ahead of the chomping marker.
func TestBlockIndentIndicator(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{" a\nb", "content: |2-\n   a\n  b\n"},
		{" a\n", "content: |2\n   a\n"},
		{" a\n\n", "content: |2+\n   a\n\n"},
		{"\n a\n", "content: |2\n\n   a\n"},
		{"a\n b\n", "content: |\n  a\n   b\n"},
		{"\ta", "content: |-\n  \ta\n"},
	} {
		if got := block(t, writer.Compose, c.in); got != c.want {
			t.Errorf("Block(%q) on compose = %q, want %q", c.in, got, c.want)
		}
	}
	if got, want := block(t, writer.Kubernetes, " a\n"), "content: |2\n   a\n"; got != want {
		t.Errorf("Block(\" a\\n\") on Kubernetes = %q, want %q", got, want)
	}
}

// TestBlockEmptyLinesCarryNoIndent: an empty line inside the text is an empty line
// in the file, at any depth.
func TestBlockEmptyLinesCarryNoIndent(t *testing.T) {
	if got, want := block(t, writer.Compose, "a\n\nb\n"), "content: |\n  a\n\n  b\n"; got != want {
		t.Errorf("Block with an empty line = %q, want %q", got, want)
	}
	nested := writer.Map(writer.KV("configs", writer.Seq(writer.Map(
		writer.KV("source", writer.Str("c")),
		writer.KV("content", writer.Block("x\n\ny\n")),
	))))
	want := "configs:\n  - source: \"c\"\n    content: |\n      x\n\n      y\n"
	got := render(t, nested, writer.Compose)
	if got != want {
		t.Errorf("nested Block = %q, want %q", got, want)
	}
	for _, line := range strings.Split(got, "\n") {
		if line != "" && strings.TrimSpace(line) == "" {
			t.Errorf("a line holds only spaces: %q in %q", line, got)
		}
	}
}

// TestBlockKubernetesNoKeep: P8. Kubernetes never writes |+: text with two or more
// final newlines, or only newlines, is a quoted string; the rest is | or |- as on
// compose, and '$' is left alone.
func TestBlockKubernetesNoKeep(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"a\n\n", "content: \"a\\n\\n\"\n"},
		{"a\n\n\n", "content: \"a\\n\\n\\n\"\n"},
		{"\n", "content: \"\\n\"\n"},
		{"\n\n", "content: \"\\n\\n\"\n"},
		{"$X\n\n", "content: \"$X\\n\\n\"\n"},
		{"a\nb\n", "content: |\n  a\n  b\n"},
		{"a\nb", "content: |-\n  a\n  b\n"},
		{"", "content: |-\n"},
		{"echo $HOME\n", "content: |\n  echo $HOME\n"},
	} {
		got := block(t, writer.Kubernetes, c.in)
		if got != c.want {
			t.Errorf("Block(%q) on Kubernetes = %q, want %q", c.in, got, c.want)
		}
		if strings.Contains(got, "|+") {
			t.Errorf("Block(%q) on Kubernetes wrote |+: %q", c.in, got)
		}
	}
}

// TestBlockRefusals: CR, U+0085, U+2028, U+2029, the other control characters, DEL,
// C1, U+FEFF and invalid UTF-8 are refused on both dialects without echoing the text
// or naming an offset; newline, tab and U+FFFD pass.
func TestBlockRefusals(t *testing.T) {
	for _, bad := range []string{
		"\r", "\u0085", "\u2028", "\u2029", "\x00", "\x01", "\x1b", "\x0b", "\x0c", "\x7f", "\u009b", "\ufeff",
		"\xff", "\xc3",
	} {
		in := "a" + bad + marker
		for _, d := range []writer.Dialect{writer.Compose, writer.Kubernetes} {
			_, err := writer.Marshal(writer.Map(writer.KV("content", writer.Block(in))), d)
			expectRefusal(t, "Block", in, err)
			if err != nil && strings.Contains(err.Error(), "offset") {
				t.Errorf("Block(%q) names an offset: %v", in, err)
			}
		}
	}
	for _, ok := range []string{"a\nb", "a\tb", "\ufffd", "caf\u00e9"} {
		if _, err := writer.Marshal(writer.Map(writer.KV("content", writer.Block(ok))), writer.Compose); err != nil {
			t.Errorf("Block(%q) refused: %v", ok, err)
		}
	}
}

// TestBlockRefusalsMatchCheckInlineText: the writer holds its own copy of the set
// workload.CheckInlineText refuses, since it imports only the standard library. This
// walks every rune up to U+3000, the edges above it and two invalid byte strings, and
// requires the two to agree on each.
func TestBlockRefusalsMatchCheckInlineText(t *testing.T) {
	var texts []string
	for r := rune(0); r <= 0x3000; r++ {
		texts = append(texts, "a"+string(r)+"b")
	}
	for _, r := range []rune{0xfeff, 0xfffd, 0xffff, 0x10000, utf8.MaxRune} {
		texts = append(texts, "a"+string(r)+"b")
	}
	texts = append(texts, "a\xffb", "a\xc3", "a\xe2\x80", "\xed\xa0\x80")
	for _, text := range texts {
		_, err := writer.Marshal(writer.Map(writer.KV("content", writer.Block(text))), writer.Compose)
		if refused := workload.CheckInlineText(text) != nil; refused != (err != nil) {
			t.Errorf("text %q: workload.CheckInlineText refuses = %v, writer.Block refuses = %v", text, refused, err != nil)
		}
	}
}
