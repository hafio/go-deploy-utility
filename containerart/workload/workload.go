package workload

import (
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/hafio/go-deploy-utility/execguard"
)

// Sink names the renderer a check runs for. The zero value is invalid, so a caller
// that did not choose one is refused.
type Sink int

const (
	// SinkCompose is the docker compose file.
	SinkCompose Sink = iota + 1

	// SinkQuadlet is the podman quadlet unit.
	SinkQuadlet

	// SinkKubernetes is the Kubernetes Deployment.
	SinkKubernetes
)

// String returns the sink's name as the error messages spell it.
func (s Sink) String() string {
	switch s {
	case SinkCompose:
		return "compose"
	case SinkQuadlet:
		return "quadlet"
	case SinkKubernetes:
		return "Kubernetes"
	}
	return "sink " + strconv.Itoa(int(s))
}

func (s Sink) valid() bool { return s >= SinkCompose && s <= SinkKubernetes }

// FieldError is one refused value: the field, the rule and, only where S8 allows,
// the value. It carries no free text, and the package doc says what counts as
// free text. Every Rule is a fixed sentence chosen in this package that says what
// the rule is and what to write instead.
type FieldError struct {
	// Field names what was checked, such as "image" or "health.interval". A tool
	// renames it to its own key with At.
	Field string

	// Rule says what failed, the rule and what to write instead.
	Rule string

	// Value is the offending value, set only for a name, a word from a fixed list
	// or a number. It is empty for everything else.
	Value string
}

// Error reads `field: rule`, or `field "value": rule` when a value is echoed.
func (e *FieldError) Error() string {
	if e.Value == "" {
		return e.Field + ": " + e.Rule
	}
	return e.Field + " " + strconv.Quote(e.Value) + ": " + e.Rule
}

// At returns a copy of e that names field instead, so a tool can report its own
// key. A nil receiver stays nil.
func (e *FieldError) At(field string) *FieldError {
	if e == nil {
		return nil
	}
	c := *e
	c.Field = field
	return &c
}

func fail(field, rule string) *FieldError {
	return &FieldError{Field: field, Rule: rule}
}

func failf(field, format string, args ...any) *FieldError {
	return &FieldError{Field: field, Rule: fmt.Sprintf(format, args...)}
}

func failValue(field, value, rule string) *FieldError {
	return &FieldError{Field: field, Rule: rule, Value: value}
}

// checkSink refuses a Sink that is not one of the three, which is a wiring
// mistake in the caller.
func checkSink(field string, s Sink) *FieldError {
	if s.valid() {
		return nil
	}
	return failf(field, "the sink is not set (got %d): pass SinkCompose, SinkQuadlet or SinkKubernetes", int(s))
}

// decimal parses a plain decimal number: digits only, no sign and no leading zero
// ("0" itself is fine), at most ten digits so that the result cannot overflow.
func decimal(s string) (uint64, bool) {
	if s == "" || len(s) > 10 || (len(s) > 1 && s[0] == '0') {
		return 0, false
	}
	var n uint64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
		n = n*10 + uint64(s[i]-'0')
	}
	return n, true
}

// isDigits reports whether s is one or more ASCII digits.
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// scan returns the byte offset of the first rune of s that is invalid UTF-8 or
// that refuse reports, or -1.
func scan(s string, refuse func(rune) bool) int {
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		if (r == utf8.RuneError && n == 1) || refuse(r) {
			return i
		}
		i += n
	}
	return -1
}

// CheckText refuses text that is not valid UTF-8 or that holds a control character
// (execguard.IsControl) or an invisible one (execguard.IsInvisible). It is the
// floor for every string a tool supplies. The error gives a byte offset and never
// the text. An empty string passes: whether a field is mandatory is its own rule.
func CheckText(s string) *FieldError {
	// U+0085 (NEL) ends a line for a YAML reader, and execguard.IsControl lets it
	// through today (N13), so it is refused here by name.
	at := scan(s, func(r rune) bool { return execguard.IsControl(r) || execguard.IsInvisible(r) || r == 0x85 })
	if at < 0 {
		return nil
	}
	return failf("text", "byte offset %d holds invalid UTF-8, a control character or an invisible character; "+
		"the text is not shown, remove the character", at)
}

// inlineRefused reports the runes CheckInlineText refuses: every line break a YAML
// reader honours (CR, NEL, LS and PS), every other control character except newline
// and tab, DEL, the C1 range and U+FEFF.
func inlineRefused(r rune) bool {
	switch r {
	case '\n', '\t':
		return false
	case '\r', 0x7f, 0x85, 0x2028, 0x2029, 0xfeff:
		return true
	}
	return r < 0x20 || (r >= 0x80 && r <= 0x9f)
}

// CheckInlineText is the check for text that is written inline into a structured
// file, such as a config's content in a compose block scalar (C7). It refuses CR,
// U+0085, U+2028, U+2029, every other control character but newline and tab, DEL,
// the C1 range, U+FEFF and invalid UTF-8. A YAML reader ends a line at the first
// four, so text carrying one could end its own block and be read as keys. Tab is
// allowed: shell and XML text may hold one. The error gives a byte offset and never
// the text.
func CheckInlineText(s string) *FieldError {
	at := scan(s, inlineRefused)
	if at < 0 {
		return nil
	}
	return failf("inline text", "byte offset %d holds invalid UTF-8, a line break other than newline or a control "+
		"character; the text is not shown, remove the character", at)
}
