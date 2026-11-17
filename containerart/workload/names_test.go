package workload

import (
	"strings"
	"testing"
)

// TestValidDNSLabel: the label grammar, with the 63 and 64 character boundary.
func TestValidDNSLabel(t *testing.T) {
	for in, want := range map[string]bool{
		"a": true, "0": true, "solace-primary": true, "a-b-c": true, "0a": true, "a1": true,
		strings.Repeat("a", 63): true, strings.Repeat("a", 64): false,
		"": false, "-a": false, "a-": false, "A": false, "a_b": false, "a.b": false, "a b": false, "caf\u00e9": false,
	} {
		if got := ValidDNSLabel(in); got != want {
			t.Errorf("ValidDNSLabel(%q) = %v, want %v", in, got, want)
		}
	}
}

// TestCheckName: the container name is a DNS label, and the error does not echo it.
func TestCheckName(t *testing.T) {
	rows(t, "CheckName", CheckName,
		[]string{"a", "solace-primary", strings.Repeat("a", 63)},
		[]string{"", strings.Repeat("a", 64), "-a", "a-", "Solace", "a_" + marker, marker + ".x", marker + " x"})
}

// TestValidHostName: RFC 1123, with the 63 and 253 character boundaries and the
// underscore that B's old rule allowed.
func TestValidHostName(t *testing.T) {
	label63 := strings.Repeat("a", 63)
	rows(t, "ValidHostName", ValidHostName,
		[]string{
			"solace", "a.b.c", "node-1.example.com", "1abc", label63,
			strings.Join([]string{label63, label63, label63, strings.Repeat("a", 61)}, "."),
		},
		[]string{
			"", "solace_primary", "-a", "a-", "a..b", ".a", "a.", strings.Repeat("a", 64), "a." + strings.Repeat("b", 64),
			strings.Join([]string{label63, label63, label63, strings.Repeat("a", 62)}, "."),
			"a b " + marker, "caf\u00e9-" + marker, marker + "_x", "a.-b", "a-.b",
		})
}

// TestValidTimezone: S13, the narrow IANA name grammar.
func TestValidTimezone(t *testing.T) {
	rows(t, "ValidTimezone", ValidTimezone,
		[]string{
			"UTC", "Europe/Paris", "America/Argentina/Buenos_Aires", "Etc/GMT+5", "Etc/GMT-14", "America/Port-au-Prince",
			strings.Repeat("a", 14), "a/" + strings.Repeat("b", 14),
		},
		[]string{
			"", "EST5EDT,M3.2.0,M11.1.0", "EST5EDT" + marker + ",M3", strings.Repeat("a", 15), "a/" + strings.Repeat("b", 15),
			"-Europe/Paris", "Europe/-Paris", "Europe//Paris", "/UTC", "UTC/", "Europe/Par is " + marker, "Europe/Paris:",
			"Asia/Ho.Chi", "caf\u00e9/" + marker,
		})
}

// TestCheckComposeVar: C6. Valid names pass; the six names and three prefixes are
// refused in any case and echoed (they are names); an invalid name is never echoed.
func TestCheckComposeVar(t *testing.T) {
	for _, in := range []string{"DB_PASSWORD", "MQ_USER", "_X", "A1", "DOCKERFILE", "LDAP_PASS", "COMPOSER_HOME", "PATHS", "HOMEDIR"} {
		if f := CheckComposeVar(in); f != nil {
			t.Errorf("CheckComposeVar(%q) refused: %v", in, f)
		}
	}
	for _, in := range []string{
		"PATH", "path", "HOME", "Home", "IFS", "SHELL", "BASH_ENV", "LD_PRELOAD", "ld_library_path",
		"DOCKER_HOST", "docker_config", "COMPOSE_FILE", "COMPOSE_PROJECT_NAME",
	} {
		f := CheckComposeVar(in)
		if f == nil || f.Value != in {
			t.Errorf("CheckComposeVar(%q) = %v, want a refusal that names the variable", in, f)
		}
	}
	rows(t, "CheckComposeVar", CheckComposeVar, nil,
		[]string{"", "1A", "A-B", "A.B", "my secret " + marker, "${" + marker + "}", marker + "=1", "caf\u00e9"})
}

// TestCheckMountSource: C8 and Q7 per sink, including the drive-letter rows.
func TestCheckMountSource(t *testing.T) {
	tests := []struct {
		sink Sink
		in   string
		ok   bool
	}{
		{SinkCompose, "/opt/libs", true},
		{SinkCompose, "/", true},
		{SinkCompose, "/opt/a b", true},
		{SinkCompose, "/opt/a%b", true},
		{SinkCompose, `C:\x`, true},
		{SinkCompose, `C:\certs\store.jks`, true},
		{SinkCompose, "C:/certs/store.jks", true},
		{SinkCompose, "c:/x", true},
		{SinkCompose, "", false},
		{SinkCompose, "relative/x", false},
		{SinkCompose, "./x", false},
		{SinkCompose, "C:", false},
		{SinkCompose, "C:x", false},
		{SinkCompose, `C:\x:y`, false},
		{SinkCompose, `D:\a\b:C:\c`, false},
		{SinkCompose, "/a:b", false},
		{SinkCompose, "1:/x", false},
		{SinkCompose, `\\server\share`, false},
		{SinkCompose, "/a\x00b " + marker, false},
		{SinkCompose, "/a\u200bb " + marker, false},
		{SinkCompose, "/a\nb " + marker, false},
		{SinkQuadlet, "/opt/libs", true},
		{SinkQuadlet, "/data", true},
		{SinkQuadlet, "/opt/a%b", true},
		{SinkQuadlet, "C:/x", false},
		{SinkQuadlet, `C:\x`, false},
		{SinkQuadlet, "/a:b", false},
		{SinkQuadlet, "relative", false},
		{SinkQuadlet, "", false},
		{SinkQuadlet, "/opt/a b " + marker, false},
		{SinkQuadlet, "/opt/a\tb " + marker, false},
		{SinkQuadlet, "/opt/$" + marker, false},
		{SinkQuadlet, `/opt/"` + marker + `"`, false},
		{SinkQuadlet, "/opt/'" + marker + "'", false},
		{SinkQuadlet, "/opt/`" + marker + "`", false},
		{SinkQuadlet, `/opt/a\` + marker, false},
		{SinkKubernetes, "/opt/libs", false},
		{SinkKubernetes, marker, false},
	}
	for _, tc := range tests {
		f := CheckMountSource(tc.sink, tc.in)
		if tc.ok {
			if f != nil {
				t.Errorf("CheckMountSource(%v, %q) refused: %v", tc.sink, tc.in, f)
			}
			continue
		}
		expectRefusal(t, "CheckMountSource", tc.in, f, true)
	}
}
