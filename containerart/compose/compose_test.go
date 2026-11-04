package compose

import (
	"fmt"
	"strings"
	"testing"
)

// TestComposeQuoteEscapes pins the render-side half of the java-options gate:
// validate refuses these characters, but a value that reached the renderer
// anyway must still come out as the string the operator wrote -- '$' doubled
// against compose interpolation, quotes and backslashes escaped for YAML.
func TestComposeQuoteEscapes(t *testing.T) {
	for in, want := range map[string]string{
		"-Xmx1g":        `"-Xmx1g"`,
		"-Dx=${Y}":      `"-Dx=$${Y}"`,
		`-Dx="a"`:       `"-Dx=\"a\""`,
		`-Dpath=C:\tmp`: `"-Dpath=C:\\tmp"`,
	} {
		if got := Quote(in); got != want {
			t.Errorf("Quote(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestEscape(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"", ""},
		{"Europe/London", "Europe/London"},
		{"a$b${c}d", "a$$b$${c}d"},
		{"$$", "$$$$"},
		{"echo $(id -u)", "echo $$(id -u)"},
	} {
		if got := Escape(c.in); got != c.want {
			t.Errorf("Escape(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestQuoteMatchesBothTools: Quote writes the bytes each tool wrote before it
// existed, for every printable value either tool routes to it. The two old forms
// are inline copies, since this module imports neither tool: solace-util's
// fmt %q of composeEscape (internal/render/render.go), and solmq-conn-util's
// composeQuote (internal/dockergen/dockergen.go).
func TestQuoteMatchesBothTools(t *testing.T) {
	solaceUtil := func(s string) string { return fmt.Sprintf("%q", strings.ReplaceAll(s, "$", "$$")) }
	solmqConnUtil := func(s string) string {
		escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s)
		return `"` + strings.ReplaceAll(escaped, "$", "$$") + `"`
	}
	for _, in := range []string{
		"", "solace", "My.Broker", "1000001:0", "8080:8080", "/var/lib/solace:/var/lib/solace:Z",
		"Europe/London", `-Dx="a" -Dpath=C:\tmp`, "a$b${c}d",
		"curl -fs http://localhost:5550/health-check/readiness",
	} {
		got := Quote(in)
		if want := solaceUtil(in); got != want {
			t.Errorf("Quote(%q) = %s, solace-util wrote %s", in, got, want)
		}
		if want := solmqConnUtil(in); got != want {
			t.Errorf("Quote(%q) = %s, solmq-conn-util wrote %s", in, got, want)
		}
	}
	// YAML 1.1 reads these bare as a boolean or a number; quoted, they stay names.
	for _, in := range []string{"yes", "off", "0123", "1.5"} {
		if got := Quote(in); got != `"`+in+`"` {
			t.Errorf("Quote(%q) = %s, want it quoted", in, got)
		}
	}
}

// TestQuoteEscapesUnprintable: a control character or an invisible rune comes out
// as an escape YAML reads back as the same character, never raw.
func TestQuoteEscapesUnprintable(t *testing.T) {
	for in, want := range map[string]string{
		"a\nb":             `"a\nb"`,
		"a\tb":             `"a\tb"`,
		"\x1b[31m":         `"\x1b[31m"`,
		"a\xe2\x80\xa8b":   `"a\u2028b"`,
		"caf\xc3\xa9":      "\"caf\xc3\xa9\"",
		"\xef\xbb\xbfname": `"\ufeffname"`,
	} {
		if got := Quote(in); got != want {
			t.Errorf("Quote(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestFlowList(t *testing.T) {
	for _, c := range []struct {
		in   []string
		want string
	}{
		{nil, "[]"},
		{[]string{"CMD", "curl", "-fs", "http://localhost:5550/health-check/readiness"},
			`["CMD", "curl", "-fs", "http://localhost:5550/health-check/readiness"]`},
		{[]string{"CMD", "sh", "-c", `test "$PORT"`}, `["CMD", "sh", "-c", "test \"$$PORT\""]`},
	} {
		if got := FlowList(c.in); got != c.want {
			t.Errorf("FlowList(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

// TestPostureBytes pins the block to the bytes both tools write today, so a
// change to it is a deliberate change to both tools' artifacts.
func TestPostureBytes(t *testing.T) {
	want := "    privileged: false\n    cap_drop:\n      - ALL\n    security_opt:\n      - no-new-privileges=true\n"
	if Posture != want {
		t.Errorf("Posture = %q, want %q", Posture, want)
	}
}

// goldenSecrets are two of the secrets in solace-util's
// internal/render/testdata/docker_compose_primary.golden.
func goldenSecrets() []Secret {
	return []Secret{
		{Name: "solace-admin-password", Target: "/mnt/secrets/username_admin_password", EnvVar: "SOLACE_ADMIN_PASSWORD"},
		{Name: "solace-tls-servercertificate", Target: "/mnt/certs/server/tls.pem", EnvVar: "SOLACE_TLS_SERVERCERTIFICATE"},
	}
}

// TestServiceSecrets reproduces solace-util's golden lines, and writes nothing
// for no secrets rather than a key holding a null.
func TestServiceSecrets(t *testing.T) {
	want := "    secrets:\n" +
		"      - source: \"solace-admin-password\"\n" +
		"        target: /mnt/secrets/username_admin_password\n" +
		"      - source: \"solace-tls-servercertificate\"\n" +
		"        target: /mnt/certs/server/tls.pem\n"
	if got := ServiceSecrets(goldenSecrets()); got != want {
		t.Errorf("ServiceSecrets =\n%s\nwant\n%s", got, want)
	}
	if got := ServiceSecrets(nil); got != "" {
		t.Errorf("ServiceSecrets(nil) = %q, want empty", got)
	}
}

// TestTopSecrets reproduces solace-util's golden lines: each value comes from the
// environment compose runs with, never from the file.
func TestTopSecrets(t *testing.T) {
	want := "secrets:\n" +
		"  \"solace-admin-password\":\n" +
		"    environment: SOLACE_ADMIN_PASSWORD\n" +
		"  \"solace-tls-servercertificate\":\n" +
		"    environment: SOLACE_TLS_SERVERCERTIFICATE\n"
	if got := TopSecrets(goldenSecrets()); got != want {
		t.Errorf("TopSecrets =\n%s\nwant\n%s", got, want)
	}
	if got := TopSecrets(nil); got != "" {
		t.Errorf("TopSecrets(nil) = %q, want empty", got)
	}
}
