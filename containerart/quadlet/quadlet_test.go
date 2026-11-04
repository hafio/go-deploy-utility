package quadlet

import (
	"strings"
	"testing"
)

// TestSystemdEnvEscapes pins the rules the Environment= line exists for: '%'
// doubled so systemd does not expand it as a specifier (the %p of a JVM error-file
// path would otherwise become the unit's name), the whole assignment quoted so a
// value with a space stays one assignment, and quotes and backslashes escaped
// inside those quotes.
func TestSystemdEnvEscapes(t *testing.T) {
	for _, c := range []struct{ value, want string }{
		{"-Xmx1g", `Environment="V=-Xmx1g"`},
		{"-XX:ErrorFile=/tmp/hs_err_%p.log", `Environment="V=-XX:ErrorFile=/tmp/hs_err_%%p.log"`},
		{"-Xms512m -Xmx1g", `Environment="V=-Xms512m -Xmx1g"`},
		{"-Dx=%t -Dy=z", `Environment="V=-Dx=%%t -Dy=z"`},
		{`-Dx="a"`, `Environment="V=-Dx=\"a\""`},
		{`-Dpath=C:\tmp`, `Environment="V=-Dpath=C:\\tmp"`},
	} {
		if got := Environment("V", c.value); got != c.want {
			t.Errorf("Environment(%q) = %s, want %s", c.value, got, c.want)
		}
	}
}

// TestQuadletEscape guards the systemd Environment=" " escaping.
func TestQuadletEscape(t *testing.T) {
	if got := Escape(`a%b"c\d`); got != `a%%b\"c\\d` {
		t.Errorf("Escape = %q", got)
	}
}

// TestEscapePercent doubles '%' and nothing else: an unquoted value is split by
// podman itself, so a quote or backslash there must reach it as written.
func TestEscapePercent(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"", ""},
		{"30s", "30s"},
		{"http://localhost/a%20b", "http://localhost/a%%20b"},
		{"%%", "%%%%"},
		{`sh -c "test \"$X\""`, `sh -c "test \"$X\""`},
	} {
		if got := EscapePercent(c.in); got != c.want {
			t.Errorf("EscapePercent(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestEscapeWritesControlsAsCEscapes: a newline in a value cannot end the line and
// start new keys in the unit; every control character but tab is written as a C
// escape podman decodes, and other bytes pass through.
func TestEscapeWritesControlsAsCEscapes(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"a\nb", `a\nb`},
		{"a\r\nb", `a\r\nb`},
		{"a\tb", "a\tb"},
		{"\x00", `\x00`},
		{"\x1b[31m", `\x1b[31m`},
		{"\x7f", `\x7f`},
		{"caf\xc3\xa9", "caf\xc3\xa9"},
	} {
		if got := Escape(c.in); got != c.want {
			t.Errorf("Escape(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	line := Environment("TZ", "UTC\n[Service]\nExecStartPre=/bin/false")
	if strings.Contains(line, "\n") {
		t.Errorf("Environment wrote a raw newline, which would add keys to the unit: %q", line)
	}
}

// TestEnvironmentIsAlwaysQuoted reproduces solace-util's golden lines
// (internal/render/testdata/podman_quadlet_primary.golden): every assignment is
// quoted, whether or not it needs it, and an empty value stays an assignment.
func TestEnvironmentIsAlwaysQuoted(t *testing.T) {
	for _, c := range []struct{ key, value, want string }{
		{"routername", "solace-primary", `Environment="routername=solace-primary"`},
		{"nodetype", "message_routing", `Environment="nodetype=message_routing"`},
		{"EMPTY", "", `Environment="EMPTY="`},
	} {
		if got := Environment(c.key, c.value); got != c.want {
			t.Errorf("Environment(%q, %q) = %s, want %s", c.key, c.value, got, c.want)
		}
	}
}

// TestSecretLine writes the mount both tools use, with '%' doubled in the name and
// the target so systemd does not read a specifier.
func TestSecretLine(t *testing.T) {
	for _, c := range []struct{ name, target, want string }{
		{"solace-admin-password", "/mnt/secrets/username_admin_password",
			"Secret=solace-admin-password,type=mount,target=/mnt/secrets/username_admin_password"},
		{"a%b", "/run/app/%t", "Secret=a%%b,type=mount,target=/run/app/%%t"},
	} {
		if got := Secret(c.name, c.target); got != c.want {
			t.Errorf("Secret(%q, %q) = %s, want %s", c.name, c.target, got, c.want)
		}
	}
}

// TestPostureBytes pins the block to the lines both tools write today.
func TestPostureBytes(t *testing.T) {
	if want := "NoNewPrivileges=true\nDropCapability=all\n"; Posture != want {
		t.Errorf("Posture = %q, want %q", Posture, want)
	}
}
