// Package quadlet holds the pieces of a podman quadlet .container unit that
// solmq-conn-util and solace-util write the same way: the escaping of values, the
// Environment= and Secret= lines, and the container's privilege posture.
//
// Each tool keeps its own renderer. A line here has no newline; Posture is a block
// and ends in one.
package quadlet

import (
	"fmt"
	"strings"
)

// EscapePercent doubles every '%'. systemd expands % specifiers in every
// assignment of a unit, including what quadlet passes on to the service it
// generates, so a literal '%' (the %p of a JVM error-file path, or a
// percent-encoded byte in a probe URL) would otherwise be replaced, or the line
// dropped. It is for a value written unquoted: quotes and backslashes are left
// alone there, because podman splits an unquoted command line itself.
func EscapePercent(s string) string { return strings.ReplaceAll(s, "%", "%%") }

// Escape prepares s to sit inside a double-quoted value. A backslash and a
// double quote are backslash-escaped, '%' and '$' are doubled, and a control
// character is written as a C escape that podman decodes when it reads the unit:
// \n, \r, or \xNN for the rest. A raw newline would end the line and let the rest
// of the value be read as further keys of the unit. A tab is kept, since a quoted
// value holds it as it is.
//
// Quadlet copies the value into the ExecStart of the service it generates and
// escapes neither '%' nor '$' there, and systemd expands both in a command line:
// a % specifier, and $NAME or ${NAME} from the unit's environment, an unknown name
// becoming empty. Doubled, each reaches the container as written.
//
// NUL is written as \x00, which podman refuses, dropping the whole assignment. A
// tool refuses control characters when the config is loaded, so this is the
// backstop, not the rule.
func Escape(s string) string {
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

// Environment renders one Environment= line with the whole assignment quoted, so
// a value holding a space stays one assignment instead of several. Both tools
// quote every assignment, whether or not it needs it.
func Environment(key, value string) string {
	return `Environment="` + Escape(key+"="+value) + `"`
}

// Secret renders one Secret= line mounting the podman secret name as a file at
// target, which must be an absolute path: a bare name lands under podman's own
// /run/secrets instead. '%' is doubled in both. A comma in name is not supported,
// since it would start another option.
//
// No mode= is set. Podman's default, 0444 owned by root, is what both tools rely
// on, and it matches what compose gives a secret.
func Secret(name, target string) string {
	return "Secret=" + EscapePercent(name) + ",type=mount,target=" + EscapePercent(target)
}

// Posture is the container's privilege posture, stated rather than inherited:
// no-new-privileges on and every capability dropped, as two lines that belong in
// [Container]. In [Service], NoNewPrivileges is systemd's own key. It would bind
// the podman process instead of the container, and rootless podman needs its
// setuid helpers. Quadlet has no Privileged= key, since a container is
// unprivileged unless asked, so there is nothing to state for it.
//
// User= and Group= are not part of the block: a tool may leave the image's own
// user in place, so it writes those lines itself.
const Posture = "NoNewPrivileges=true\n" +
	"DropCapability=all\n"
