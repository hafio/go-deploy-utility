package abbrev

import (
	"strings"
	"testing"
)

// sample is the shape every real set has: canonical words in a meaningful order,
// one short form each, one word deliberately carrying none.
func sample() *Set {
	return New("thing", []Entry{
		{Canonical: "kubernetes", Short: []string{"kube"}, Note: "the section key"},
		{Canonical: "docker", Short: []string{"dk"}},
		{Canonical: "podman", Short: []string{"pm"}},
		{Canonical: "start", Note: "no short form: ambiguous with stop"},
	})
}

// TestExpandResolvesBothSpellings pins the one lookup callers get: a canonical
// word resolves to itself, so a parser needs a membership test and a lookup no
// longer -- a value parser relies on that to stay a single branch.
func TestExpandResolvesBothSpellings(t *testing.T) {
	s := sample()
	for _, tc := range []struct {
		in   string
		want string
		ok   bool
	}{
		{in: "kubernetes", want: "kubernetes", ok: true},
		{in: "kube", want: "kubernetes", ok: true},
		{in: "dk", want: "docker", ok: true},
		{in: "start", want: "start", ok: true},
		{in: "k8s"},
		{in: "KUBE"},
		{in: " docker"},
		{in: ""},
	} {
		got, ok := s.Expand(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("Expand(%q) = %q, %v; want %q, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

// TestOrderIsDeclarationOrder pins that nothing sorts the entries. The order is
// meaningful in the sets the tools declare -- redundancy order for roles, the
// order platforms are offered in -- and it is the order errors and the reference
// print, so an alphabetical Names() would silently reword both.
func TestOrderIsDeclarationOrder(t *testing.T) {
	s := sample()
	wantNames := []string{"kubernetes", "docker", "podman", "start"}
	if got := s.Names(); strings.Join(got, ",") != strings.Join(wantNames, ",") {
		t.Errorf("Names() = %v, want %v", got, wantNames)
	}
	wantShorts := []string{"kube", "dk", "pm"}
	if got := s.Shorts(); strings.Join(got, ",") != strings.Join(wantShorts, ",") {
		t.Errorf("Shorts() = %v, want %v", got, wantShorts)
	}
	entries := s.Entries()
	if len(entries) != 4 {
		t.Fatalf("Entries() returned %d entries, want 4", len(entries))
	}
	if entries[0].Note != "the section key" {
		t.Errorf("Entries()[0].Note = %q, want the declared note", entries[0].Note)
	}
	if len(entries[3].Short) != 0 {
		t.Errorf("Entries()[3].Short = %v, want none: a word may be declared with no short form", entries[3].Short)
	}
}

// TestAccessorsReturnCopies stops a caller reaching into the set through a slice
// it was handed. A tool shares one set between its parsers, its CLI wiring and
// its docs generator, so a mutation here would reach all of them.
func TestAccessorsReturnCopies(t *testing.T) {
	s := sample()
	s.Entries()[0].Canonical = "hijacked"
	s.Entries()[0].Short[0] = "hijacked"
	s.Short("kubernetes")[0] = "hijacked"
	s.Names()[0] = "hijacked"
	if got, ok := s.Expand("kube"); !ok || got != "kubernetes" {
		t.Errorf("Expand(\"kube\") = %q, %v after mutating returned slices; want %q, true", got, ok, "kubernetes")
	}
	if got := s.Short("kubernetes"); len(got) != 1 || got[0] != "kube" {
		t.Errorf("Short(\"kubernetes\") = %v, want [kube]", got)
	}
}

// TestShortOfUnknownWord is what alias wiring relies on: it asks the set about
// every command in a tree, most of which have no entry, and must get nothing
// back rather than a panic.
func TestShortOfUnknownWord(t *testing.T) {
	if got := sample().Short("nothing-like-this"); got != nil {
		t.Errorf("Short of an undeclared word = %v, want nil", got)
	}
}

// TestList pins the error sentence exactly, because a tool's "must be one of"
// message is this rendering and its own tests read it back.
func TestList(t *testing.T) {
	for _, tc := range []struct {
		name    string
		entries []Entry
		want    string
	}{
		{
			name:    "three with short forms",
			entries: sample().Entries()[:3],
			want:    "kubernetes (kube), docker (dk) or podman (pm)",
		},
		{
			name:    "a word with no short form prints bare",
			entries: []Entry{{Canonical: "start"}, {Canonical: "stop"}},
			want:    "start or stop",
		},
		{
			name:    "one entry needs no conjunction",
			entries: []Entry{{Canonical: "docker", Short: []string{"dk"}}},
			want:    "docker (dk)",
		},
		{
			name:    "empty set renders nothing",
			entries: nil,
			want:    "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := New("thing", tc.entries).List(); got != tc.want {
				t.Errorf("List() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestKind(t *testing.T) {
	if got := sample().Kind(); got != "thing" {
		t.Errorf("Kind() = %q, want %q", got, "thing")
	}
}

// TestNewPanicsOnAWiringBug is the whole reason this package exists: every rule
// that makes a short form safe is checked once, here, at construction. A panic
// rather than an error is deliberate -- these are declarations in source, so the
// failure belongs at the moment the set is built, which makes every test that
// builds one a validation test.
func TestNewPanicsOnAWiringBug(t *testing.T) {
	for _, tc := range []struct {
		name    string
		kind    string
		entries []Entry
		want    string
	}{
		{
			name: "no kind",
			want: "kind is empty",
		},
		{
			name:    "empty canonical",
			kind:    "thing",
			entries: []Entry{{Canonical: ""}},
			want:    "no canonical name",
		},
		{
			name:    "canonical declared twice",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker"}, {Canonical: "docker"}},
			want:    "declared twice",
		},
		{
			name: "one short form claimed by two words",
			kind: "thing",
			entries: []Entry{
				{Canonical: "docker", Short: []string{"dk"}},
				{Canonical: "darkroom", Short: []string{"dk"}},
			},
			want: "claimed by both",
		},
		{
			name: "short form shadows a canonical name declared later",
			kind: "thing",
			entries: []Entry{
				{Canonical: "docker", Short: []string{"pm"}},
				{Canonical: "pm"},
			},
			want: "in its own right",
		},
		{
			name:    "empty short form",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{""}}},
			want:    "empty short form",
		},
		{
			name:    "short form is not shorter",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"dockers"}}},
			want:    "not shorter",
		},
		{
			name:    "short form the same length is a synonym",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"moby42"}}},
			want:    "not shorter",
		},
		{
			name:    "uppercase short form",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"DK"}}},
			want:    "lowercase letters or digits",
		},
		{
			name:    "short form carrying a shell metacharacter",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"d;k"}}},
			want:    "lowercase letters or digits",
		},
		{
			name:    "short form with a space",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"d k"}}},
			want:    "lowercase letters or digits",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("New(%q, %v) did not panic, want a panic naming %q", tc.kind, tc.entries, tc.want)
				}
				if msg, _ := r.(string); !strings.Contains(msg, tc.want) {
					t.Errorf("panic = %q, want it to name %q", r, tc.want)
				}
			}()
			New(tc.kind, tc.entries)
		})
	}
}

// TestDigitsAreAllowed keeps the charset from being letters-only by accident: a
// product word that ends in a number is exactly the case an abbreviation is worth
// having for, and refusing it would be a rule nobody chose.
func TestDigitsAreAllowed(t *testing.T) {
	s := New("thing", []Entry{{Canonical: "kubernetes", Short: []string{"k8s"}}})
	if got, ok := s.Expand("k8s"); !ok || got != "kubernetes" {
		t.Errorf("Expand(\"k8s\") = %q, %v; want %q, true", got, ok, "kubernetes")
	}
}

// mustPanic runs fn and returns the panic message, failing the test when fn does
// not panic or panics with something other than a string.
func mustPanic(t *testing.T, fn func()) (msg string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("did not panic")
		}
		s, ok := r.(string)
		if !ok {
			t.Fatalf("panicked with %T %v, want a string message", r, r)
		}
		msg = s
	}()
	fn()
	return ""
}

// TestNewPanicsOnShadowedCanonicalDeclaredFirst covers the other declaration
// order: TestNewPanicsOnAWiringBug declares the shadowed canonical AFTER the short
// form, and the separate canonical pass exists so that order does not matter.
func TestNewPanicsOnShadowedCanonicalDeclaredFirst(t *testing.T) {
	msg := mustPanic(t, func() {
		New("thing", []Entry{
			{Canonical: "pm"},
			{Canonical: "docker", Short: []string{"pm"}},
		})
	})
	if !strings.Contains(msg, "in its own right") {
		t.Errorf("panic = %q, want it to name the shadowed canonical", msg)
	}
}

// TestNewPanicsOnShortFormRepeatedWithinOneEntry: a word listing the same short
// form twice is a duplicate claim too, even though both claims name one owner.
func TestNewPanicsOnShortFormRepeatedWithinOneEntry(t *testing.T) {
	msg := mustPanic(t, func() {
		New("thing", []Entry{{Canonical: "docker", Short: []string{"dk", "dk"}}})
	})
	if !strings.Contains(msg, "claimed by both") {
		t.Errorf("panic = %q, want it to report the repeated claim", msg)
	}
}

// TestNewPanicMessagesSayWhatToDo: every panic ends with the fix, so the message
// is actionable without reading this package.
func TestNewPanicMessagesSayWhatToDo(t *testing.T) {
	for _, fn := range []func(){
		func() { New("", nil) },
		func() { New("thing", []Entry{{Canonical: ""}}) },
		func() { New("thing", []Entry{{Canonical: "docker", Short: []string{"DK"}}}) },
	} {
		msg := mustPanic(t, fn)
		if !strings.Contains(msg, "abbrev.New") {
			t.Errorf("panic = %q, want it to say what to fix", msg)
		}
	}
}

// TestEntriesCloneKeepsNilShort pins clone's nil guard both ways: a word declared
// with no short form comes back with a nil slice, not an allocated empty one.
func TestEntriesCloneKeepsNilShort(t *testing.T) {
	s := sample()
	if got := s.Entries()[3].Short; got != nil {
		t.Errorf("Entries()[3].Short = %#v, want nil for a word declared with no short form", got)
	}
	if got := s.Short("start"); len(got) != 0 {
		t.Errorf("Short(\"start\") = %v, want none", got)
	}
	if got := s.Entries()[0].Short; len(got) != 1 || got[0] != "kube" {
		t.Errorf("Entries()[0].Short = %v, want [kube]", got)
	}
}

// TestSetIsIndependentOfInputSlice is the input half of TestAccessorsReturnCopies:
// changing the slices a caller passed to New must not reach into the built Set.
func TestSetIsIndependentOfInputSlice(t *testing.T) {
	short := []string{"dk"}
	entries := []Entry{{Canonical: "docker", Short: short}}
	s := New("thing", entries)
	short[0] = "zz"
	entries[0].Canonical = "hijacked"
	if got := s.Short("docker"); len(got) != 1 || got[0] != "dk" {
		t.Errorf("Short(\"docker\") = %v after mutating the input, want [dk]", got)
	}
	if got := s.Names(); len(got) != 1 || got[0] != "docker" {
		t.Errorf("Names() = %v after mutating the input, want [docker]", got)
	}
}

// TestToolAliasDeclarationsConstruct guards the rules against drifting away from a
// real consumer: solmq-conn-util's verb, platform and per-verb target aliases,
// copied here as literals, must build without a panic and resolve. solace-util's
// sets are this package's own origin and are covered by its own tests.
func TestToolAliasDeclarationsConstruct(t *testing.T) {
	sets := []struct {
		kind    string
		entries []Entry
	}{
		{"command", []Entry{
			{Canonical: "generate", Short: []string{"gen"}},
			{Canonical: "deploy", Short: []string{"dp"}},
			{Canonical: "remove", Short: []string{"rm"}},
			{Canonical: "start"},
			{Canonical: "stop"},
			{Canonical: "restart"},
			{Canonical: "status", Short: []string{"sts"}},
			{Canonical: "logs", Short: []string{"lg"}},
			{Canonical: "cli"},
			{Canonical: "version", Short: []string{"ver"}},
			{Canonical: "validate", Short: []string{"vld"}},
			{Canonical: "examples", Short: []string{"eg"}},
			{Canonical: "download", Short: []string{"dl"}},
			{Canonical: "auto-complete"},
			{Canonical: "help"},
		}},
		{"platform", []Entry{
			{Canonical: "kubernetes", Short: []string{"kube"}},
			{Canonical: "docker", Short: []string{"dk"}},
			{Canonical: "podman", Short: []string{"pm"}},
		}},
		{"generate target", []Entry{{Canonical: "config", Short: []string{"cfg"}}}},
		{"status target", []Entry{
			{Canonical: "container", Short: []string{"cnt"}},
			{Canonical: "application", Short: []string{"app"}},
			{Canonical: "all"},
		}},
	}
	for _, set := range sets {
		t.Run(set.kind, func(t *testing.T) {
			s := New(set.kind, set.entries)
			for _, e := range set.entries {
				for _, short := range e.Short {
					if got, ok := s.Expand(short); !ok || got != e.Canonical {
						t.Errorf("Expand(%q) = %q, %v; want %q, true", short, got, ok, e.Canonical)
					}
				}
			}
		})
	}
}
