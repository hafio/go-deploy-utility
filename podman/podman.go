// Package podman builds the podman commands solmq-conn-util and solace-util run the
// same way: the client version, secrets on stdin, the user's id mappings, and where
// a quadlet unit is written.
//
// Every builder takes base, the engine argv the tool has already built and checked
// with execguard (["podman"], or the operator's configured command), and copies it,
// so appending never writes into the caller's slice. Nothing here executes: a tool
// runs the returned Cmd through its runner.Runner.
package podman

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hafio/go-deploy-utility/runner"
)

// SystemQuadletDir is where rootful podman's quadlet generator reads units from.
const SystemQuadletDir = "/etc/containers/systemd"

// cmd is base followed by args, in a new slice.
func cmd(base []string, args ...string) runner.Cmd {
	argv := make([]string, 0, len(base)+len(args))
	argv = append(argv, base...)
	return runner.Cmd{Argv: append(argv, args...)}
}

// Version asks for the client version, such as 5.8.2. Read stdout alone
// (runner.Output or runner.Split): podman can print warnings on stderr first.
func Version(base []string) runner.Cmd {
	return cmd(base, "version", "--format", "{{.Client.Version}}")
}

// SupportsReplace reports whether version, as Version prints it, is podman 4.7 or
// newer, where `secret create --replace` exists. Surrounding space and a leading v
// are ignored. A version that does not parse is not, so the caller keeps to
// removing the secret and then creating it.
func SupportsReplace(version string) bool {
	num := func(s string) (int, bool) {
		end := 0
		for end < len(s) && s[end] >= '0' && s[end] <= '9' {
			end++
		}
		n, err := strconv.Atoi(s[:end])
		return n, err == nil
	}
	parts := strings.SplitN(strings.TrimPrefix(strings.TrimSpace(version), "v"), ".", 3)
	if len(parts) < 2 {
		return false
	}
	major, ok1 := num(parts[0])
	minor, ok2 := num(parts[1])
	return ok1 && ok2 && (major > 4 || major == 4 && minor >= 7)
}

// CreateSecret stores value in podman's secret store under name:
// `secret create [--replace] [--label l]... <name> -`. The value goes to the child
// on stdin and never onto argv; Stdin is a copy, and non-nil even for an empty
// value, so the child always reads exactly the value and then EOF. Each label is
// written as given, as key=value.
//
// To replace a secret of the same name, pass replace when SupportsReplace holds for
// the host's version. Otherwise run RemoveSecrets for the name first, then this
// with replace false. Each step's error is the caller's to word.
//
// name is not checked here: both tools derive it from names validated at load, and
// podman refuses one it cannot store.
func CreateSecret(base []string, name string, value []byte, replace bool, labels ...string) runner.Cmd {
	args := []string{"secret", "create"}
	if replace {
		args = append(args, "--replace")
	}
	for _, l := range labels {
		args = append(args, "--label", l)
	}
	c := cmd(base, append(args, name, "-")...)
	c.Stdin = make([]byte, len(value))
	copy(c.Stdin, value)
	return c
}

// SecretLabel asks for one label of the named secret: `secret inspect --format
// {{index .Spec.Labels "<key>"}} <name>`. Read stdout alone and trim it: an absent
// label prints an empty line, and a missing secret is an error. Both tools write a
// digest of the value as a label with CreateSecret and compare it here, so an
// unchanged secret is neither replaced nor counted as a change.
//
// key is placed inside a Go template, so an empty key, or one holding '"', '\',
// '{', '}' or a control character, is refused; so is an empty name.
func SecretLabel(base []string, name, key string) (runner.Cmd, error) {
	if name == "" || key == "" {
		return runner.Cmd{}, fmt.Errorf("podman: reading a secret label needs a secret name and a label key, "+
			"got %q and %q", name, key)
	}
	for _, r := range key {
		if r == '"' || r == '\\' || r == '{' || r == '}' || r < 0x20 || r == 0x7f {
			return runner.Cmd{}, fmt.Errorf("podman: the label key %q holds %q, which cannot be placed in a "+
				"--format template; use letters, digits, '.', '-' and '_'", key, r)
		}
	}
	return cmd(base, "secret", "inspect", "--format", `{{index .Spec.Labels "`+key+`"}}`, name), nil
}

// RemoveSecrets removes every named secret in one call, `secret rm --ignore
// <names...>`, so a name that is already gone is not an error. With no names there
// is nothing to run, and ok is false.
func RemoveSecrets(base []string, names ...string) (c runner.Cmd, ok bool) {
	if len(names) == 0 {
		return runner.Cmd{}, false
	}
	return cmd(base, append([]string{"secret", "rm", "--ignore"}, names...)...), true
}

// IDMap asks podman for the invoking user's uid map, or its gid map when gid is
// set, as space-separated containerID:size pairs ("0:1 1:65536 "). This is what
// podman will actually map, whichever backend granted the ranges (the files, or an
// NSS source such as SSSD), and rootless.CheckIDMap reads it. Rootful podman
// reports no mappings, so ask only when deploying rootless.
func IDMap(base []string, gid bool) runner.Cmd {
	field := "UIDMap"
	if gid {
		field = "GIDMap"
	}
	return cmd(base, "info", "--format", "{{range .Host.IDMappings."+field+"}}{{.ContainerID}}:{{.Size}} {{end}}")
}

// QuadletDir is the directory a tool writes a quadlet unit to: SystemQuadletDir for
// rootful podman, and for rootless the user's config directory followed by
// containers/systemd. The config directory is $XDG_CONFIG_HOME when it is set,
// else <home>/.config, which is how os.UserConfigDir, and so podman's own generator,
// resolves it on Linux. getenv is os.Getenv and home is os.UserHomeDir. The result
// uses forward slashes, since the unit is for a Linux host.
//
// It is the directory written, not every directory podman reads. The generator
// runs under the user's systemd manager and reads that manager's environment, so an
// XDG_CONFIG_HOME set only in a shell profile is not seen there.
//
// A relative XDG_CONFIG_HOME is an error, since podman cannot use it either, and so
// are a home directory that cannot be resolved and an empty one.
func QuadletDir(rootless bool, getenv func(string) string, home func() (string, error)) (string, error) {
	if !rootless {
		return SystemQuadletDir, nil
	}
	dir := getenv("XDG_CONFIG_HOME")
	if dir != "" && !filepath.IsAbs(dir) {
		return "", fmt.Errorf("podman: XDG_CONFIG_HOME is %q, which is not an absolute path, and podman's "+
			"quadlet generator cannot use it either; set it to an absolute path, or unset it to use "+
			"$HOME/.config", dir)
	}
	if dir == "" {
		h, err := home()
		if err != nil {
			return "", fmt.Errorf("podman: the rootless quadlet directory needs this user's home directory, "+
				"which cannot be resolved: %w; set HOME, or set XDG_CONFIG_HOME to an absolute path", err)
		}
		if h == "" {
			return "", errors.New("podman: the rootless quadlet directory needs this user's home directory, " +
				"which is empty; set HOME, or set XDG_CONFIG_HOME to an absolute path")
		}
		dir = filepath.Join(h, ".config")
	}
	return filepath.ToSlash(filepath.Join(dir, "containers", "systemd")), nil
}
