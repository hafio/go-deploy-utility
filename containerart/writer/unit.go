package writer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Unit writes the lines of a podman quadlet .container unit (Q3, Q4). The
// renderer chooses the order of the lines inside a section; the writer owns the
// order of the sections and the blank line between them. A Unit collects every
// problem it finds and Bytes returns them together, so nothing partial is written.
//
// Section names are Unit, Container, Service and Install, written in that order
// whatever the order of the Section calls. A section with no line is left out,
// except [Service], which is always written.
type Unit struct {
	comment string
	cur     int
	lines   [4][]string
	errs    []error
}

// NewUnit returns an empty Unit. Its first line call must follow a Section call.
func NewUnit() *Unit { return &Unit{cur: -1} }

// sectionIndex returns the position of a section name in the unit's order, or -1.
func sectionIndex(name string) int {
	switch name {
	case "Unit":
		return 0
	case "Container":
		return 1
	case "Service":
		return 2
	case "Install":
		return 3
	}
	return -1
}

// sectionName is the inverse of sectionIndex.
func sectionName(i int) string {
	return [...]string{"Unit", "Container", "Service", "Install"}[i]
}

func (u *Unit) fail(path, rule string) {
	u.errs = append(u.errs, &Error{Path: path, Rule: rule})
}

// Comment sets the one header comment, written as "# text" before the first
// section (Q10); a second call replaces it and an empty text writes none. A comment
// holding a line break or any other control character is refused, since the rest of
// it would be read as a key. The text is not echoed.
func (u *Unit) Comment(text string) {
	if strings.ContainsFunc(text, unicode.IsControl) {
		u.fail("Comment", "the text holds a line break or a control character; the text is not shown")
		return
	}
	u.comment = text
}

// Section makes name the section that the following lines go to: Unit, Container,
// Service or Install. Naming a section again continues it. Any other name is
// refused and not echoed, and the current section stays as it was.
func (u *Unit) Section(name string) {
	i := sectionIndex(name)
	if i < 0 {
		u.fail("Section", "is not a unit section (the name is not shown): use Unit, Container, Service or Install")
		return
	}
	u.cur = i
}

// target checks key and the current section, and returns the path its errors use
// and whether the line may be written. A key is echoed in the path only when it is
// valid, since a refused one may be a secret.
func (u *Unit) target(key string) (string, bool) {
	prefix := ""
	if u.cur >= 0 {
		prefix = sectionName(u.cur) + "."
	}
	if key == "" || strings.ContainsFunc(key, func(r rune) bool {
		return r == '=' || unicode.IsSpace(r) || unicode.IsControl(r)
	}) {
		u.fail(prefix+"?", "the key is empty or holds '=', whitespace or a control character; the key is not shown")
		return "", false
	}
	if u.cur < 0 {
		u.fail(key, "the line is written before any Section call: call Section first")
		return "", false
	}
	return prefix + key, true
}

// Bare writes `Key=value` with every '%' doubled, since systemd expands a % specifier
// in every line of a unit. Quotes and backslashes are left alone: podman splits an
// unquoted command line itself. A value holding a line break or another control
// character is refused, since the rest would be read as further keys, and so is a
// value holding '$': quadlet copies [Container] values into the ExecStart it
// generates, where systemd expands $NAME. A value that needs '$' is written with
// Quoted. The value is not echoed.
func (u *Unit) Bare(key, value string) {
	path, ok := u.target(key)
	if !ok {
		return
	}
	if strings.ContainsFunc(value, func(r rune) bool { return unicode.IsControl(r) || r == '$' }) {
		u.fail(path, "the value holds a line break, another control character or '$', which would start another key "+
			"or be expanded by systemd; the value is not shown")
		return
	}
	u.lines[u.cur] = append(u.lines[u.cur], key+"="+percent(value))
}

// Quoted writes `Key="value"` with the whole value quoted and escaped, the form of
// Environment= and Label= (Q4). Inside the quotes `\`, `"`, `%` and `$` are escaped,
// and a control character is written as a C escape podman decodes: \n, \r, or \xNN
// for the rest; a tab is kept. A raw newline would end the line and let the rest of
// the value be read as further keys of the unit. The caller passes the whole
// assignment as the value, such as "TZ=Europe/Paris". Invalid UTF-8 passes through
// as it is, so a caller checks its text first. These are the bytes of
// quadlet.Environment.
func (u *Unit) Quoted(key, value string) {
	if _, ok := u.target(key); !ok {
		return
	}
	u.lines[u.cur] = append(u.lines[u.cur], key+`="`+escapeQuoted(value)+`"`)
}

// Extra writes a line a tool passes through, as `Key=value` with '%' doubled. M does
// not know which extra keys unquote a value, so a value holding whitespace, a quote,
// a backslash or '$' is refused (Q4), and the error names the key and not the value.
// Whether the key is allowed at all is internal/posture's question, asked by the
// renderer first.
func (u *Unit) Extra(key, value string) {
	path, ok := u.target(key)
	if !ok {
		return
	}
	if strings.ContainsFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || r == '"' || r == '\\' || r == '$'
	}) {
		u.fail(path, "the value holds whitespace, a quote, a backslash or '$', which an unquoted extra cannot carry "+
			"(the value is not shown): write a value without them")
		return
	}
	u.lines[u.cur] = append(u.lines[u.cur], key+"="+percent(value))
}

// Bytes returns the unit: the header comment if there is one, then each section in
// order with one blank line between them and none after the last, every line ending
// in a newline. Every problem found by the calls so far is returned joined with
// errors.Join, each an *Error, and nothing partial comes back with it.
func (u *Unit) Bytes() ([]byte, error) {
	if len(u.errs) > 0 {
		return nil, errors.Join(u.errs...)
	}
	var sections []string
	for i, lines := range u.lines {
		if len(lines) == 0 && sectionName(i) != "Service" {
			continue
		}
		text := "[" + sectionName(i) + "]\n"
		for _, l := range lines {
			text += l + "\n"
		}
		sections = append(sections, text)
	}
	out := strings.Join(sections, "\n")
	if u.comment != "" {
		out = "# " + u.comment + "\n" + out
	}
	return []byte(out), nil
}

// percent doubles every '%'.
func percent(s string) string { return strings.ReplaceAll(s, "%", "%%") }

// escapeQuoted prepares s to sit inside a double-quoted value; see Quoted.
func escapeQuoted(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '\\':
			b.WriteString(`\\`)
		case c == '"':
			b.WriteString(`\"`)
		case c == '%':
			b.WriteString("%%")
		case c == '$':
			b.WriteString("$$")
		case c == '\n':
			b.WriteString(`\n`)
		case c == '\r':
			b.WriteString(`\r`)
		case c == '\t':
			b.WriteByte(c)
		case c < 0x20 || c == 0x7f:
			fmt.Fprintf(&b, `\x%02x`, c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
