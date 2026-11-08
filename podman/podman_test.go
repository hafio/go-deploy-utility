package podman

import (
	"errors"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// TestSupportsReplace holds the version to 4.7, where `secret create --replace`
// exists, whatever the distribution suffix.
func TestSupportsReplace(t *testing.T) {
	for v, want := range map[string]bool{
		"5.8.2": true, "4.7.0": true, "4.9.4-rhel": true, "v4.7.1": true, "5.8.2\n": true,
		"4.6.2": false, "4.5": false, "3.4.4": false, "": false, "out": false, "4": false, "4.x": false,
	} {
		if got := SupportsReplace(v); got != want {
			t.Errorf("SupportsReplace(%q) = %v, want %v", v, got, want)
		}
	}
}

func TestVersionArgv(t *testing.T) {
	c := Version([]string{"podman"})
	if want := []string{"podman", "version", "--format", "{{.Client.Version}}"}; !reflect.DeepEqual(c.Argv, want) {
		t.Errorf("argv = %v, want %v", c.Argv, want)
	}
	if c.Stdin != nil || c.Env != nil {
		t.Errorf("Version carries no stdin or env, got %+v", c)
	}
}

// TestCreateSecretValueOnlyOnStdin: below 4.7 the secret is removed and then
// created, and the value reaches podman on stdin of the create alone.
func TestCreateSecretValueOnlyOnStdin(t *testing.T) {
	const name, value = "mq-conn-1-password", "s3cr3t"
	rm, ok := RemoveSecrets([]string{"podman"}, name)
	if !ok {
		t.Fatal("RemoveSecrets with a name must have a command to run")
	}
	if want := []string{"podman", "secret", "rm", "--ignore", name}; !reflect.DeepEqual(rm.Argv, want) {
		t.Errorf("rm argv = %v, want %v", rm.Argv, want)
	}
	if rm.Stdin != nil {
		t.Errorf("rm must not carry stdin, got %q", rm.Stdin)
	}
	create := CreateSecret([]string{"podman"}, name, []byte(value), false)
	if want := []string{"podman", "secret", "create", name, "-"}; !reflect.DeepEqual(create.Argv, want) {
		t.Errorf("create argv = %v, want %v", create.Argv, want)
	}
	if string(create.Stdin) != value {
		t.Errorf("create stdin = %q, want %q", create.Stdin, value)
	}
	for _, a := range append(rm.Argv, create.Argv...) {
		if strings.Contains(a, value) {
			t.Errorf("the secret value must never appear in argv, found in %q", a)
		}
	}
}

// TestCreateSecretReplaceIsOneCall: on 4.7 and later one create --replace does it.
func TestCreateSecretReplaceIsOneCall(t *testing.T) {
	c := CreateSecret([]string{"podman"}, "c-tls-ts", []byte("bytes"), true)
	if want := []string{"podman", "secret", "create", "--replace", "c-tls-ts", "-"}; !reflect.DeepEqual(c.Argv, want) {
		t.Errorf("argv = %v, want %v", c.Argv, want)
	}
	if string(c.Stdin) != "bytes" {
		t.Errorf("stdin = %q, want the value", c.Stdin)
	}
}

// TestCreateSecretLabelOrder pins --replace before the labels, and the labels in
// the order given, so a tool's golden argv does not depend on map order.
func TestCreateSecretLabelOrder(t *testing.T) {
	c := CreateSecret([]string{"podman"}, "sol-tls", []byte("pem"), true, "solace-util.sha256=ab12", "x=y")
	want := []string{"podman", "secret", "create", "--replace", "--label", "solace-util.sha256=ab12",
		"--label", "x=y", "sol-tls", "-"}
	if !reflect.DeepEqual(c.Argv, want) {
		t.Errorf("argv = %v, want %v", c.Argv, want)
	}
	c = CreateSecret([]string{"podman"}, "sol-tls", []byte("pem"), false, "solace-util.sha256=ab12")
	want = []string{"podman", "secret", "create", "--label", "solace-util.sha256=ab12", "sol-tls", "-"}
	if !reflect.DeepEqual(c.Argv, want) {
		t.Errorf("argv = %v, want %v", c.Argv, want)
	}
}

// TestCreateSecretEmptyValueIsNonNil: nil Stdin means "no input" to a runner, so an
// empty value must still be an empty, non-nil slice, and a copy of the caller's.
func TestCreateSecretEmptyValueIsNonNil(t *testing.T) {
	for _, v := range [][]byte{nil, {}} {
		if c := CreateSecret([]string{"podman"}, "n", v, false); c.Stdin == nil || len(c.Stdin) != 0 {
			t.Errorf("CreateSecret(%v) stdin = %#v, want an empty non-nil slice", v, c.Stdin)
		}
	}
	value := []byte("abc")
	c := CreateSecret([]string{"podman"}, "n", value, false)
	value[0] = 'X'
	if string(c.Stdin) != "abc" {
		t.Errorf("stdin = %q; it must be a copy, unaffected by the caller's later writes", c.Stdin)
	}
}

func TestRemoveSecrets(t *testing.T) {
	names := []string{"mq-conn-1-user", "mq-conn-1-password", "truststore-password"}
	c, ok := RemoveSecrets([]string{"podman"}, names...)
	if !ok {
		t.Fatal("names given, so there is a command to run")
	}
	if want := append([]string{"podman", "secret", "rm", "--ignore"}, names...); !reflect.DeepEqual(c.Argv, want) {
		t.Errorf("argv = %v, want one batched %v", c.Argv, want)
	}
	if c, ok := RemoveSecrets([]string{"podman"}); ok || c.Argv != nil {
		t.Errorf("no names = (%+v, %v), want nothing to run", c, ok)
	}
}

// TestBuildersCopyBase: a base with spare capacity must not be written through, or
// one builder's argv would change another's.
func TestBuildersCopyBase(t *testing.T) {
	base := make([]string, 1, 8)
	base[0] = "podman"
	a := Version(base)
	b := IDMap(base, false)
	c := CreateSecret(base, "n", []byte("v"), false)
	if a.Argv[1] != "version" || b.Argv[1] != "info" || c.Argv[1] != "secret" {
		t.Errorf("argvs share the base's backing array: %v / %v / %v", a.Argv, b.Argv, c.Argv)
	}
	if len(base) != 1 || base[:2][1] != "" {
		t.Errorf("base was written through: %v", base[:2])
	}
	custom := Version([]string{"podman", "--remote"})
	if want := []string{"podman", "--remote", "version", "--format", "{{.Client.Version}}"}; !reflect.DeepEqual(custom.Argv, want) {
		t.Errorf("a multi-token base must lead the argv: %v", custom.Argv)
	}
}

// TestIDMapArgv pins solace-util's template, one call per map.
func TestIDMapArgv(t *testing.T) {
	for gid, want := range map[bool]string{
		false: "{{range .Host.IDMappings.UIDMap}}{{.ContainerID}}:{{.Size}} {{end}}",
		true:  "{{range .Host.IDMappings.GIDMap}}{{.ContainerID}}:{{.Size}} {{end}}",
	} {
		c := IDMap([]string{"podman"}, gid)
		if w := []string{"podman", "info", "--format", want}; !reflect.DeepEqual(c.Argv, w) {
			t.Errorf("IDMap(gid=%v) argv = %v, want %v", gid, c.Argv, w)
		}
	}
}

// env is a fixed getenv for the seams.
func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

// homeAt is a home seam that always answers dir.
func homeAt(dir string) func() (string, error) {
	return func() (string, error) { return dir, nil }
}

// abs is the slash path p made absolute on the OS running the test. Windows calls
// a path absolute only with a drive, so it gains "C:" there, as an
// XDG_CONFIG_HOME set on that host would carry.
func abs(p string) string {
	if runtime.GOOS == "windows" {
		return "C:" + filepath.FromSlash(p)
	}
	return p
}

// TestQuadletDir: rootful is fixed; rootless follows XDG_CONFIG_HOME and falls back
// to <home>/.config (cases from solmq-conn-util's TestResolveQuadletScope and
// solace-util's TestApplyDefaultsPodmanRootlessXDG and ...HomeDir).
func TestQuadletDir(t *testing.T) {
	for _, c := range []struct {
		name     string
		rootless bool
		env      map[string]string
		want     string
	}{
		{"rootful", false, nil, "/etc/containers/systemd"},
		{"rootful ignores XDG", false, map[string]string{"XDG_CONFIG_HOME": "/x"}, "/etc/containers/systemd"},
		{"rootless under home", true, nil, "/home/deployer/.config/containers/systemd"},
		{"rootless under XDG", true, map[string]string{"XDG_CONFIG_HOME": abs("/srv/cfg")}, filepath.ToSlash(abs("/srv/cfg")) + "/containers/systemd"},
		{"an empty XDG is unset", true, map[string]string{"XDG_CONFIG_HOME": ""}, "/home/deployer/.config/containers/systemd"},
		{"XDG is cleaned", true, map[string]string{"XDG_CONFIG_HOME": abs("/srv/cfg/")}, filepath.ToSlash(abs("/srv/cfg")) + "/containers/systemd"},
	} {
		got, err := QuadletDir(c.rootless, env(c.env), homeAt("/home/deployer"))
		if err != nil || got != c.want {
			t.Errorf("%s: QuadletDir = %q, %v; want %q", c.name, got, err, c.want)
		}
	}
}

func TestQuadletDirRefusesRelativeXDG(t *testing.T) {
	_, err := QuadletDir(true, env(map[string]string{"XDG_CONFIG_HOME": "cfg"}), homeAt("/home/deployer"))
	if err == nil {
		t.Fatal("a relative XDG_CONFIG_HOME must be refused, since podman cannot use it")
	}
	for _, want := range []string{`"cfg"`, "absolute", "unset it"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to name %q", err, want)
		}
	}
}

func TestQuadletDirHomeErrors(t *testing.T) {
	cause := errors.New("$HOME is not defined")
	_, err := QuadletDir(true, env(nil), func() (string, error) { return "", cause })
	if !errors.Is(err, cause) || !strings.Contains(err.Error(), "set HOME") {
		t.Errorf("err = %v, want the cause wrapped and the fix named", err)
	}
	_, err = QuadletDir(true, env(nil), homeAt(""))
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("an empty home must be refused, got %v", err)
	}
	if _, err := QuadletDir(true, env(map[string]string{"XDG_CONFIG_HOME": abs("/x")}), func() (string, error) {
		return "", cause
	}); err != nil {
		t.Errorf("with XDG set the home is never asked for, got %v", err)
	}
}

// TestSecretData: the stored value is asked for on stdout through a fixed template,
// with nothing about it on argv, and an empty name is refused.
func TestSecretData(t *testing.T) {
	base := make([]string, 1, 4)
	base[0] = "podman"
	c, err := SecretData(base, "sol-tls")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"podman", "secret", "inspect", "--showsecret", "--format", "{{.SecretData}}", "sol-tls"}
	if !reflect.DeepEqual(c.Argv, want) || c.Stdin != nil || c.Env != nil {
		t.Errorf("cmd = %+v, want argv %v and no stdin or env", c, want)
	}
	if base[:2][1] != "" {
		t.Errorf("the builder wrote into the caller's base: %v", base[:2])
	}
	if _, err := SecretData([]string{"podman"}, ""); err == nil || !strings.Contains(err.Error(), "podman:") {
		t.Errorf("an empty name: err = %v, want a refusal naming the package", err)
	}
}

// TestSecretMatches: only the value followed by podman's one newline matches, so a
// trailing newline in the value or in the store is never lost to a trim.
func TestSecretMatches(t *testing.T) {
	for _, c := range []struct {
		out, value string
		want       bool
	}{
		{"abc\n", "abc", true},
		{"-----PEM-----\n\n", "-----PEM-----\n", true},
		{"\n", "", true},
		{"abc", "abc", false},
		{"abc\n", "abc\n", false},
		{"abc\n\n", "abc", false},
		{"abd\n", "abc", false},
		{"", "", false},
	} {
		if got := SecretMatches(c.out, []byte(c.value)); got != c.want {
			t.Errorf("SecretMatches(%q, %q) = %v, want %v", c.out, c.value, got, c.want)
		}
	}
}
