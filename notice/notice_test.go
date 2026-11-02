package notice

import (
	"strings"
	"testing"
)

// The three committed notices, copied byte for byte from each repo's README.md
// (lines 3-9). If the template or the wrap rule changes, these fail first.
var committed = []struct {
	name, kind, covers string
	lines              []string
}{
	{
		name: "solmq-conn-util", kind: "tool",
		covers: "the Solace PubSub+ Connector for IBM MQ that it configures and deploys",
		lines: []string{
			"> [!WARNING]",
			"> **Not a supported Solace product.** `solmq-conn-util` was created by Solace",
			"> Professional Services and is supported only by Solace Professional Services --",
			"> not by Solace Support. For help with this tool, contact your Solace",
			"> Professional Services representative rather than opening a Solace Support",
			"> case. This notice covers this tool only, not the Solace PubSub+ Connector for",
			"> IBM MQ that it configures and deploys.",
		},
	},
	{
		name: "solace-util", kind: "tool",
		covers: "the Solace PubSub+ Event Broker or the EventBroker Operator that it deploys and operates",
		lines: []string{
			"> [!WARNING]",
			"> **Not a supported Solace product.** `solace-util` was created by Solace",
			"> Professional Services and is supported only by Solace Professional Services --",
			"> not by Solace Support. For help with this tool, contact your Solace",
			"> Professional Services representative rather than opening a Solace Support",
			"> case. This notice covers this tool only, not the Solace PubSub+ Event Broker",
			"> or the EventBroker Operator that it deploys and operates.",
		},
	},
	{
		name: "go-deploy-utility", kind: "module",
		covers: "the Solace products that the tools built on it deploy",
		lines: []string{
			"> [!WARNING]",
			"> **Not a supported Solace product.** `go-deploy-utility` was created by Solace",
			"> Professional Services and is supported only by Solace Professional Services --",
			"> not by Solace Support. For help with this module, contact your Solace",
			"> Professional Services representative rather than opening a Solace Support",
			"> case. This notice covers this module only, not the Solace products that the",
			"> tools built on it deploy.",
		},
	},
}

func mustNew(t *testing.T, name, kind, covers string) Notice {
	t.Helper()
	n, err := New(name, kind, covers)
	if err != nil {
		t.Fatalf("New(%q, %q, %q): %v", name, kind, covers, err)
	}
	return n
}

// TestNoticeMatchesCommittedBlocks is the reason the renderer exists: one rule
// must reproduce every committed copy exactly, or adopting it would reflow docs.
func TestNoticeMatchesCommittedBlocks(t *testing.T) {
	for _, c := range committed {
		t.Run(c.name, func(t *testing.T) {
			n := mustNew(t, c.name, c.kind, c.covers)
			got := n.Lines()
			if strings.Join(got, "\n") != strings.Join(c.lines, "\n") {
				t.Errorf("Lines() for %s:\n got:\n%s\nwant:\n%s", c.name, strings.Join(got, "\n"), strings.Join(c.lines, "\n"))
			}
			if want := strings.Join(c.lines, "\n") + "\n"; n.Markdown() != want {
				t.Errorf("Markdown() for %s = %q, want %q", c.name, n.Markdown(), want)
			}
		})
	}
}

// TestNoticeLineWidth: every line fits Width, the alert tag leads, every other
// line is quoted, and a word longer than a line gets its own line instead of
// being dropped or looping.
func TestNoticeLineWidth(t *testing.T) {
	for _, c := range committed {
		n := mustNew(t, c.name, c.kind, c.covers)
		lines := n.Lines()
		if lines[0] != "> [!WARNING]" {
			t.Errorf("%s: first line = %q, want the alert tag", c.name, lines[0])
		}
		for i, ln := range lines {
			if len(ln) > Width {
				t.Errorf("%s: line %d is %d bytes, over %d: %q", c.name, i+1, len(ln), Width, ln)
			}
			if !strings.HasPrefix(ln, "> ") {
				t.Errorf("%s: line %d is not quoted: %q", c.name, i+1, ln)
			}
		}
	}
	long := strings.Repeat("x", Width+5)
	got := wrap("a "+long+" b", "> ")
	want := []string{"> a", "> " + long, "> b"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("wrap with an over-long word = %q, want %q", got, want)
	}
	if got := wrap("", "> "); got != nil {
		t.Errorf("wrap of nothing = %q, want nil", got)
	}
}

// TestNewRejectsBadInput: every rejected value comes back named, with the field,
// and New never returns a usable notice alongside an error.
func TestNewRejectsBadInput(t *testing.T) {
	const okName, okKind, okCovers = "solace-util", "tool", "the broker it deploys"
	for _, tc := range []struct {
		name, kind, covers string
		field, value       string
	}{
		{"", okKind, okCovers, "name", `""`},
		{"solace util", okKind, okCovers, "name", `"solace util"`},
		{"Solace-util", okKind, okCovers, "name", `"Solace-util"`},
		{"-solace", okKind, okCovers, "name", `"-solace"`},
		{"solace`util", okKind, okCovers, "name", "\"solace`util\""},
		{okName, "", okCovers, "kind", `""`},
		{okName, "two words", okCovers, "kind", `"two words"`},
		{okName, "Tool", okCovers, "kind", `"Tool"`},
		{okName, okKind, "", "covers", "covers is empty"},
		{okName, okKind, " leading", "covers", `" leading"`},
		{okName, okKind, "trailing ", "covers", `"trailing "`},
		{okName, okKind, "double  space", "covers", `"double  space"`},
		{okName, okKind, "a period.", "covers", `"a period."`},
		{okName, okKind, "a\ttab", "covers", `"a\ttab"`},
		{okName, okKind, "a\nnewline", "covers", `"a\nnewline"`},
		{okName, okKind, "a\rreturn", "covers", `"a\rreturn"`},
		{okName, okKind, "a\x7fdel", "covers", `"a\x7fdel"`},
		{okName, okKind, "an \u2014 dash", "covers", "0xe2"},
	} {
		n, err := New(tc.name, tc.kind, tc.covers)
		if err == nil {
			t.Errorf("New(%q, %q, %q) accepted bad input", tc.name, tc.kind, tc.covers)
			continue
		}
		if n != (Notice{}) {
			t.Errorf("New(%q, %q, %q) returned a notice alongside its error", tc.name, tc.kind, tc.covers)
		}
		msg := err.Error()
		if !strings.Contains(msg, tc.field) || !strings.Contains(msg, tc.value) {
			t.Errorf("New(%q, %q, %q) error = %q, want it to name the %s and %s", tc.name, tc.kind, tc.covers, msg, tc.field, tc.value)
		}
	}
}

// TestRenderedFormsAreASCII: the notice reaches terminals, logs and PowerShell
// 5.1, so every rendered form is printable ASCII, with newlines only where the
// Markdown form joins its lines.
func TestRenderedFormsAreASCII(t *testing.T) {
	check := func(what, s string, allowNewline bool) {
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c == '\n' && allowNewline {
				continue
			}
			if c < 0x20 || c > 0x7e {
				t.Errorf("%s has byte 0x%02x at %d: %q", what, c, i, s)
			}
		}
	}
	for _, c := range committed {
		n := mustNew(t, c.name, c.kind, c.covers)
		for _, ln := range n.Lines() {
			check(c.name+" Lines", ln, false)
		}
		check(c.name+" Markdown", n.Markdown(), true)
		check(c.name+" Plain", n.Plain(), false)
	}
	for _, ln := range Short() {
		check("Short", ln, false)
	}
}

// TestPlainStripsMarkup pins the words-only form solmq-conn-util compares its
// generator page against.
func TestPlainStripsMarkup(t *testing.T) {
	c := committed[0]
	n := mustNew(t, c.name, c.kind, c.covers)
	want := "Not a supported Solace product. solmq-conn-util was created by Solace Professional Services " +
		"and is supported only by Solace Professional Services -- not by Solace Support. For help with " +
		"this tool, contact your Solace Professional Services representative rather than opening a " +
		"Solace Support case. This notice covers this tool only, not the Solace PubSub+ Connector for " +
		"IBM MQ that it configures and deploys."
	if got := n.Plain(); got != want {
		t.Errorf("Plain() =\n%q\nwant\n%q", got, want)
	}
}

// TestShortMatchesVersionForm pins the two lines solmq-conn-util's `version`
// prints today, byte for byte.
func TestShortMatchesVersionForm(t *testing.T) {
	want := []string{
		"Not a supported Solace product. Created by Solace Professional Services and",
		"supported only by Solace Professional Services -- not by Solace Support.",
	}
	if got := Short(); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("Short() = %q, want %q", got, want)
	}
}

// TestLinesReturnsFreshSlice: a caller that edits the returned slices cannot
// change what the next caller gets.
func TestLinesReturnsFreshSlice(t *testing.T) {
	c := committed[1]
	n := mustNew(t, c.name, c.kind, c.covers)
	n.Lines()[1] = "hijacked"
	if got := n.Lines()[1]; got != c.lines[1] {
		t.Errorf("Lines()[1] after mutating an earlier result = %q, want %q", got, c.lines[1])
	}
	Short()[0] = "hijacked"
	if got := Short()[0]; got == "hijacked" {
		t.Error("Short() shares its slice between calls")
	}
}

// TestZeroNoticeRendersNothing: a Notice built without New must not render an
// empty-named notice; it renders nothing, which a doc gate then reports.
func TestZeroNoticeRendersNothing(t *testing.T) {
	var n Notice
	if got := n.Lines(); got != nil {
		t.Errorf("zero Notice Lines() = %q, want nil", got)
	}
	if got := n.Markdown(); got != "" {
		t.Errorf("zero Notice Markdown() = %q, want empty", got)
	}
	if got := n.Plain(); got != "" {
		t.Errorf("zero Notice Plain() = %q, want empty", got)
	}
}
