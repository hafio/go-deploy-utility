package workload

import (
	"strings"
	"testing"
)

// marker stands for a secret pasted into a free-text field. No check may echo it.
const marker = "hunter2"

// expectRefusal fails unless f refuses in without repeating it. whole also
// requires that the error not hold the whole input; the checks that echo a number
// turn it off, and the marker alone is still checked.
func expectRefusal(t *testing.T, what, in string, f *FieldError, whole bool) {
	t.Helper()
	if f == nil {
		t.Errorf("%s(%q) accepted, want a refusal", what, in)
		return
	}
	msg := f.Error()
	if f.Field == "" || f.Rule == "" {
		t.Errorf("%s(%q): field or rule is empty in %+v", what, in, f)
	}
	if strings.Contains(msg, marker) {
		t.Errorf("%s(%q) echoed the value: %q", what, in, msg)
	}
	if whole && len(in) >= 6 && strings.Contains(msg, in) {
		t.Errorf("%s(%q) echoed the value: %q", what, in, msg)
	}
}

// rows runs a one-argument check over its accepted and refused inputs. Every
// refused input is also held to S8: the error must not repeat it.
func rows(t *testing.T, what string, check func(string) *FieldError, accept, reject []string) {
	t.Helper()
	for _, in := range accept {
		if f := check(in); f != nil {
			t.Errorf("%s(%q) refused: %v", what, in, f)
		}
	}
	for _, in := range reject {
		expectRefusal(t, what, in, check(in), true)
	}
}

// TestFieldErrorText: the field, an optional quoted value, and the rule.
func TestFieldErrorText(t *testing.T) {
	f := &FieldError{Field: "ulimit", Rule: "is refused"}
	if got, want := f.Error(), "ulimit: is refused"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	f.Value = "nice"
	if got, want := f.Error(), `ulimit "nice": is refused`; got != want {
		t.Errorf("Error() with a value = %q, want %q", got, want)
	}
}

// TestFieldErrorAt: At renames on a copy and leaves nil alone.
func TestFieldErrorAt(t *testing.T) {
	f := fail("image", "bad")
	g := f.At("image.tag")
	if g.Field != "image.tag" || g.Rule != "bad" {
		t.Errorf("At = %+v, want the field renamed and the rule kept", g)
	}
	if f.Field != "image" {
		t.Errorf("At changed the receiver: %+v", f)
	}
	if (*FieldError)(nil).At("x") != nil {
		t.Error("At on nil returned non-nil")
	}
}

// TestSinkString: the names the messages use.
func TestSinkString(t *testing.T) {
	for sink, want := range map[Sink]string{SinkCompose: "compose", SinkQuadlet: "quadlet", SinkKubernetes: "Kubernetes", Sink(0): "sink 0"} {
		if got := sink.String(); got != want {
			t.Errorf("Sink(%d).String() = %q, want %q", int(sink), got, want)
		}
	}
}

// TestInvalidSinkIsRefused: the zero Sink and an unknown one fail every check that
// takes a sink, rather than choosing a grammar.
func TestInvalidSinkIsRefused(t *testing.T) {
	health := Health{Test: []string{"curl"}, Interval: 1e9, Timeout: 1e9, Retries: 1}
	for _, sink := range []Sink{0, 4, -1} {
		checks := []*FieldError{
			ValidRestart(sink, "always"),
			ValidNetwork(sink, "host"),
			ValidMemory(sink, "8g"),
			ValidShmSize(sink, "1g"),
			CheckMountSource(sink, "/opt"),
			health.Check(sink),
			Ulimit{Name: "nofile", Soft: 1, Hard: 1}.Check(sink),
		}
		for i, f := range checks {
			expectRefusal(t, "check "+string(rune('0'+i)), "sink", f, true)
		}
	}
}

// TestCheckText: text is valid UTF-8 with no control or invisible character, and
// the error gives an offset, not the text.
func TestCheckText(t *testing.T) {
	rows(t, "CheckText", CheckText,
		[]string{"", "plain", "-Xmx1g -Dkey=value", "caf\u00e9 \U0001F600", "a b"},
		[]string{
			marker + "\x00", marker + "\n", marker + "\t", "\x7f" + marker, marker + "\u009b",
			marker + "\u200b", marker + "\u202e", marker + "\u00ad", "\xff" + marker, marker + "\xc3",
			marker + "\u0085",
		})
	f := CheckText("ab\x00")
	if f == nil || !strings.Contains(f.Rule, "byte offset 2") {
		t.Errorf("CheckText(\"ab\\x00\") = %v, want an error naming byte offset 2", f)
	}
}

// TestCheckInlineText: C7. The four line breaks a YAML reader honours and the other
// controls are refused; newline, tab and ordinary Unicode pass.
func TestCheckInlineText(t *testing.T) {
	rows(t, "CheckInlineText", CheckInlineText,
		[]string{"", "line one\nline two\n", "tab\there", "caf\u00e9 \U0001F600", "\ufffd", "port: 8080\n  host: x\n"},
		[]string{
			"a\rb " + marker, "a\u0085b " + marker, "port: 8080\u2028privileged: true " + marker, "a\u2029b " + marker,
			"\x00" + marker, "\x1b[0m" + marker, "\x7f" + marker, "\u009b" + marker, "\ufeff" + marker,
			"\xff\xfe" + marker, "x\vy " + marker, "x\fy " + marker, "line\r\nbreak " + marker,
		})
	f := CheckInlineText("ab\u2028cd")
	if f == nil || !strings.Contains(f.Rule, "byte offset 2") {
		t.Errorf("CheckInlineText with U+2028 at 2 = %v, want an error naming byte offset 2", f)
	}
}
