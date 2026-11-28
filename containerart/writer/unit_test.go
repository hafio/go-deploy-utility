package writer_test

import (
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/containerart/writer"
)

// containerLines writes the lines fn adds to [Container] and returns them without
// the section header and the [Service] section the writer always adds.
func containerLines(t *testing.T, fn func(u *writer.Unit)) string {
	t.Helper()
	u := writer.NewUnit()
	u.Section("Container")
	fn(u)
	out, err := u.Bytes()
	if err != nil {
		t.Fatalf("Bytes refused: %v", err)
	}
	body, ok := strings.CutPrefix(string(out), "[Container]\n")
	if !ok {
		t.Fatalf("unit does not start with [Container]: %q", out)
	}
	body, ok = strings.CutSuffix(body, "\n[Service]\n")
	if !ok {
		t.Fatalf("unit does not end with an empty [Service]: %q", out)
	}
	return body
}

// TestUnitQuoted: Quoted writes the bytes of quadlet.Environment and quadlet.Escape.
// The vectors are copied from quadlet_test.go (TestSystemdEnvEscapes,
// TestEscapeWritesControlsAsCEscapes, TestEnvironmentIsAlwaysQuoted), not imported.
func TestUnitQuoted(t *testing.T) {
	for _, c := range []struct{ key, value, want string }{
		{"Environment", "V=-Xmx1g", `Environment="V=-Xmx1g"` + "\n"},
		{"Environment", "V=-XX:ErrorFile=/tmp/hs_err_%p.log", `Environment="V=-XX:ErrorFile=/tmp/hs_err_%%p.log"` + "\n"},
		{"Environment", "V=-Xms512m -Xmx1g", `Environment="V=-Xms512m -Xmx1g"` + "\n"},
		{"Environment", "V=-Dx=%t -Dy=z", `Environment="V=-Dx=%%t -Dy=z"` + "\n"},
		{"Environment", `V=-Dx="a"`, `Environment="V=-Dx=\"a\""` + "\n"},
		{"Environment", `V=-Dpath=C:\tmp`, `Environment="V=-Dpath=C:\\tmp"` + "\n"},
		{"Environment", "V=-Xmx${HEAP}", `Environment="V=-Xmx$${HEAP}"` + "\n"},
		{"Environment", "V=cost$5", `Environment="V=cost$$5"` + "\n"},
		{"Environment", "routername=solace-primary", `Environment="routername=solace-primary"` + "\n"},
		{"Environment", "EMPTY=", `Environment="EMPTY="` + "\n"},
		{"Environment", "V=a\nb", `Environment="V=a\nb"` + "\n"},
		{"Environment", "V=a\r\nb", `Environment="V=a\r\nb"` + "\n"},
		{"Environment", "V=a\tb", "Environment=\"V=a\tb\"\n"},
		{"Environment", "V=\x00", `Environment="V=\x00"` + "\n"},
		{"Environment", "V=\x1b[31m", `Environment="V=\x1b[31m"` + "\n"},
		{"Environment", "V=\x7f", `Environment="V=\x7f"` + "\n"},
		{"Environment", "V=caf\xc3\xa9", "Environment=\"V=caf\xc3\xa9\"\n"},
		{"Label", "deploy.solace.com/instance=prod 1", `Label="deploy.solace.com/instance=prod 1"` + "\n"},
	} {
		got := containerLines(t, func(u *writer.Unit) { u.Quoted(c.key, c.value) })
		if got != c.want {
			t.Errorf("Quoted(%q, %q) = %q, want %q", c.key, c.value, got, c.want)
		}
	}
	got := containerLines(t, func(u *writer.Unit) { u.Quoted("Environment", "TZ=UTC\n[Service]\nExecStartPre=/bin/false") })
	if strings.Count(got, "\n") != 1 {
		t.Errorf("Quoted wrote a raw newline, which would add keys to the unit: %q", got)
	}
}

// TestUnitBare: '%' is doubled, quotes and backslashes are left alone, and a line
// break, another control character or '$' is refused without echoing the value.
func TestUnitBare(t *testing.T) {
	for _, c := range []struct{ key, value, want string }{
		{"Image", "registry.example/app:1.0", "Image=registry.example/app:1.0\n"},
		{"HealthTimeout", "30s", "HealthTimeout=30s\n"},
		{"Description", "a%b", "Description=a%%b\n"},
		{"HealthCmd", "http://localhost/a%20b", "HealthCmd=http://localhost/a%%20b\n"},
		{"Secret", "a%b,type=mount,target=/run/app/%t", "Secret=a%%b,type=mount,target=/run/app/%%t\n"},
		{"HealthCmd", `sh -c "test \"x\""`, `HealthCmd=sh -c "test \"x\""` + "\n"},
		{"Label", "", "Label=\n"},
	} {
		got := containerLines(t, func(u *writer.Unit) { u.Bare(c.key, c.value) })
		if got != c.want {
			t.Errorf("Bare(%q, %q) = %q, want %q", c.key, c.value, got, c.want)
		}
	}
	for _, value := range []string{
		"a\nb " + marker, "a\rb " + marker, marker + "\n[Service]", "a\tb " + marker, "a\x00b " + marker,
		"a\x7fb " + marker, "$" + marker, "${" + marker + "}",
	} {
		u := writer.NewUnit()
		u.Section("Container")
		u.Bare("Description", value)
		_, err := u.Bytes()
		expectRefusal(t, "Bare", value, err)
	}
	for _, key := range []string{"", "a=b" + marker, "a b " + marker, "a\nb" + marker, "a\x00" + marker} {
		u := writer.NewUnit()
		u.Section("Container")
		u.Bare(key, "v")
		_, err := u.Bytes()
		expectRefusal(t, "Bare key", key, err)
	}
}

// TestUnitExtra: Q4. An extra is bare with '%' doubled; whitespace, a quote, a
// backslash or '$' in the value is refused, the key is named and the value is not;
// a bad key is refused and not named.
func TestUnitExtra(t *testing.T) {
	for _, c := range []struct{ key, value, want string }{
		{"ShmSize", "2g", "ShmSize=2g\n"},
		{"Ulimit", "nofile=2448:1048576", "Ulimit=nofile=2448:1048576\n"},
		{"Tmpfs", "/run", "Tmpfs=/run\n"},
		{"Exec", "a%b", "Exec=a%%b\n"},
		{"Notify", "", "Notify=\n"},
	} {
		got := containerLines(t, func(u *writer.Unit) { u.Extra(c.key, c.value) })
		if got != c.want {
			t.Errorf("Extra(%q, %q) = %q, want %q", c.key, c.value, got, c.want)
		}
	}
	for _, value := range []string{
		"a b " + marker, "a\tb " + marker, "a\nb " + marker, "a\u00a0b " + marker, "a\"b " + marker, `a\b ` + marker,
		"a$b " + marker, "${" + marker + "}",
	} {
		u := writer.NewUnit()
		u.Section("Container")
		u.Extra("ShmSize", value)
		_, err := u.Bytes()
		expectRefusal(t, "Extra", value, err)
		if err != nil && !strings.Contains(err.Error(), "Container.ShmSize") {
			t.Errorf("Extra(%q) = %v, want the key named", value, err)
		}
	}
	for _, key := range []string{"", "a=b" + marker, "a b " + marker, "a\nb" + marker, "a\rb" + marker} {
		u := writer.NewUnit()
		u.Section("Container")
		u.Extra(key, "v")
		_, err := u.Bytes()
		expectRefusal(t, "Extra key", key, err)
	}
}

// TestUnitSections: Q3. The sections come out as Unit, Container, Service, Install
// whatever the order of the calls, with one blank line between and none after the
// last; [Service] is always written; the header comment comes first; lines keep the
// renderer's order within a section.
func TestUnitSections(t *testing.T) {
	u := writer.NewUnit()
	u.Comment("Generated by the tool")
	u.Section("Install")
	u.Bare("WantedBy", "default.target")
	u.Section("Container")
	u.Bare("Image", "x")
	u.Bare("ContainerName", "y")
	u.Section("Unit")
	u.Bare("Description", "d")
	u.Section("Container")
	u.Bare("User", "1000")
	out, err := u.Bytes()
	if err != nil {
		t.Fatalf("Bytes refused: %v", err)
	}
	want := "# Generated by the tool\n" +
		"[Unit]\nDescription=d\n" +
		"\n" +
		"[Container]\nImage=x\nContainerName=y\nUser=1000\n" +
		"\n" +
		"[Service]\n" +
		"\n" +
		"[Install]\nWantedBy=default.target\n"
	if string(out) != want {
		t.Errorf("unit =\n%s\nwant\n%s", out, want)
	}

	// Service last, so nothing follows it; no comment writes none; Comment replaces.
	u = writer.NewUnit()
	u.Comment("first")
	u.Comment("")
	u.Section("Service")
	u.Bare("Restart", "always")
	out, err = u.Bytes()
	if err != nil || string(out) != "[Service]\nRestart=always\n" {
		t.Errorf("service-only unit = %q, %v; want [Service] and one line", out, err)
	}
	u.Comment("second")
	out, _ = u.Bytes()
	if !strings.HasPrefix(string(out), "# second\n[Service]\n") || strings.HasSuffix(string(out), "\n\n") {
		t.Errorf("unit with a replaced comment = %q", out)
	}

	// An empty unit is still a valid one: [Service] is always written.
	out, err = writer.NewUnit().Bytes()
	if err != nil || string(out) != "[Service]\n" {
		t.Errorf("empty unit = %q, %v; want [Service] alone", out, err)
	}
}

// TestUnitCollectsEveryProblem: a bad section, a line before any section, a bad key,
// a line break, a bad extra and a bad comment are all reported together, as
// *writer.Error values, with nothing partial and no value echoed.
func TestUnitCollectsEveryProblem(t *testing.T) {
	u := writer.NewUnit()
	u.Bare("Orphan", "v")
	u.Section("Bogus " + marker)
	u.Section("Container")
	u.Bare("a b "+marker, "v")
	u.Bare("K", "a\n"+marker)
	u.Extra("E", "a b "+marker)
	u.Comment("x\n" + marker)
	u.Quoted("Environment", "A=b")
	out, err := u.Bytes()
	if out != nil || err == nil {
		t.Fatalf("Bytes = %q, %v; want no output and an error", out, err)
	}
	msg := err.Error()
	if strings.Contains(msg, marker) {
		t.Errorf("the joined error echoed the marker: %q", msg)
	}
	if got := strings.Count(msg, "\n") + 1; got != 6 {
		t.Errorf("the joined error holds %d problems, want 6: %q", got, msg)
	}
	for _, want := range []string{"Orphan: ", "Section: ", "Container.?: ", "Container.K: ", "Container.E: ", "Comment: "} {
		if !strings.Contains(msg, want) {
			t.Errorf("the joined error lacks %q: %q", want, msg)
		}
	}
}
