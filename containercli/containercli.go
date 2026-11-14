// Package containercli builds the docker and podman engine commands that
// solmq-conn-util and solace-util run the same way, and reads their answers: the
// engine preflight, compose with its file on stdin, the lifecycle verbs, the
// container listing, exec, upload and logs.
//
// Every builder takes base, the engine argv the tool has already built and checked
// with execguard (["docker"], ["podman"], or a wrapper such as ["lima", "nerdctl"]),
// or compose, the compose argv the tool derived the same way (["docker", "compose"]
// or ["docker-compose"]). Builders copy what they are given, so appending never
// writes into the caller's slices. Nothing here executes: a tool runs the returned
// Cmd through its runner.Runner.
package containercli

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/hafio/go-deploy-utility/internal/build"
	"github.com/hafio/go-deploy-utility/runner"
)

// Engine names a container engine, for messages and hints.
type Engine string

const (
	Docker Engine = "docker"
	Podman Engine = "podman"
)

// title is the engine's name as a message spells it.
func (e Engine) title() string {
	switch e {
	case Docker:
		return "Docker"
	case Podman:
		return "Podman"
	case "":
		return "container"
	}
	return string(e)
}

// Info is `<engine> info`, the preflight that runs before any other engine command:
// it fails fast, with InfoError's hint, when the engine cannot be reached.
func Info(base []string) runner.Cmd { return build.Cmd(base, "info") }

// InfoError is the error for a failed Info: which engine, the command as run, the
// cause (wrapped), the engine's own output when the caller captured it (out, or ""
// when it already reached the terminal), and EngineHint's next step.
func InfoError(engine Engine, base []string, rootless bool, out string, err error) error {
	said := ""
	if o := strings.TrimSpace(out); o != "" {
		said = "\n" + o
	}
	return fmt.Errorf("cannot talk to the %s engine (%q info failed): %w%s\n  %s",
		engine.title(), strings.Join(base, " "), err, said, EngineHint(engine, rootless))
}

// EngineHint is the one thing to try next when the engine cannot be reached. A
// rootless podman engine gets its own hint: its usual failure is a missing user
// session, and starting the system service with sudo would start the rootful engine
// the deployment is not using.
func EngineHint(engine Engine, rootless bool) string {
	switch {
	case engine == Docker:
		return "start the daemon (`sudo systemctl start docker`) or select a working context " +
			"(`docker context use <name>`), and check this user is in the `docker` group " +
			"(`newgrp docker` applies a fresh membership without logging out)"
	case engine == Podman && rootless:
		return "start the rootless user service: `systemctl --user start podman.socket` " +
			"(and `loginctl enable-linger` so it survives logout); if podman named /etc/subuid, " +
			"this account has no subuid range and an administrator has to allocate one " +
			"(`sudo usermod --add-subuids 100000-165535 --add-subgids 100000-165535 $(id -un)`, " +
			"then `podman system migrate`) -- do NOT use sudo for podman itself; " +
			"a rootless deployment runs as this user"
	case engine == Podman:
		return "start the engine: `sudo systemctl start podman.socket`, and re-run with the privileges " +
			"rootful podman needs; on macOS or Windows, start the podman machine instead " +
			"(`podman machine start`)"
	}
	return fmt.Sprintf("check that %s is installed and its service is running, then re-run", engine.title())
}

// composeCmd is `<compose> -f - --project-directory <dir> <args...>`, with file on
// stdin and env for the child alone.
func composeCmd(compose []string, file []byte, projectDir string, env []string, args ...string) runner.Cmd {
	c := build.Cmd(compose, append([]string{"-f", "-", "--project-directory", projectDir}, args...)...)
	c.Stdin = build.Stdin(file)
	c.Env = append([]string(nil), env...)
	return c
}

// ComposeUp brings the project in file up: `<compose> -f - --project-directory
// <dir> up -d`, plus --force-recreate when recreate is set. The file is read from
// stdin and never written to disk; env carries the secrets the file's
// environment-provider secrets name, for the compose child alone. projectDir
// anchors the file's relative paths; the project name is the file's own name: key.
func ComposeUp(compose []string, file []byte, projectDir string, env []string, recreate bool) runner.Cmd {
	args := []string{"up", "-d"}
	if recreate {
		args = append(args, "--force-recreate")
	}
	return composeCmd(compose, file, projectDir, env, args...)
}

// ComposeDown takes the project in file down, read from stdin the same way.
func ComposeDown(compose []string, file []byte, projectDir string, env []string) runner.Cmd {
	return composeCmd(compose, file, projectDir, env, "down")
}

// ComposeDigest is the hex sha256 of a compose deploy: the file, prefixed with its
// length, then every entry of env (the compose child's environment, secret values
// included), sorted, each preceded by a NUL. An environment entry cannot hold a
// NUL, so no two different deploys hash the same bytes. env is not modified.
//
// Compose recreates a container when its service configuration changes, but a
// secret's content is not part of that configuration. A tool therefore writes a
// label whose value the file interpolates from an environment variable ("${VAR}"),
// sets that variable to this digest, and lets compose decide: any change to the
// file or to a value changes the label, so compose recreates, and an unchanged
// deploy changes nothing. Before `up`, the tool compares the digest with the
// running container's label (inspect.Container.Labels) to know whether compose is
// about to recreate it, and so whether to ask first. The digest is computed with
// the variable itself absent from env.
func ComposeDigest(file []byte, env []string) string {
	sorted := append([]string(nil), env...)
	sort.Strings(sorted)
	h := sha256.New()
	fmt.Fprintf(h, "%d\n", len(file))
	h.Write(file)
	for _, e := range sorted {
		h.Write([]byte{0})
		h.Write([]byte(e))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Start starts the named container. The lifecycle verbs act on the container
// itself rather than through compose, so they work on every compose implementation
// and need no file.
func Start(base []string, name string) runner.Cmd { return build.Cmd(base, "start", name) }

// Stop stops the named container.
func Stop(base []string, name string) runner.Cmd { return build.Cmd(base, "stop", name) }

// Restart restarts the named container.
func Restart(base []string, name string) runner.Cmd { return build.Cmd(base, "restart", name) }

// Remove removes the named, stopped container.
func Remove(base []string, name string) runner.Cmd { return build.Cmd(base, "rm", name) }

// listFormat is the ps template: name, image and status, separated by tabs.
const listFormat = "{{.Names}}\t{{.Image}}\t{{.Status}}"

// List lists every container, or only running ones when running is set:
// `ps --all --no-trunc [--filter status=running] --format <names, image, status>`.
// Read the answer with ParseList and match in Go: both engines treat
// `--filter name=` as an unanchored regular expression, so it cannot decide whether
// a container exists.
func List(base []string, running bool) runner.Cmd {
	args := []string{"ps", "--all", "--no-trunc"}
	if running {
		args = append(args, "--filter", "status=running")
	}
	return build.Cmd(base, append(args, "--format", listFormat)...)
}

// Row is one container in List's answer.
type Row struct{ Name, Image, Status string }

// ParseList reads List's answer, in listing order. Each field is trimmed, and a
// blank line or one with fewer than three fields is skipped.
func ParseList(out string) []Row {
	var rows []Row
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) < 3 {
			continue
		}
		rows = append(rows, Row{
			Name:   strings.TrimSpace(f[0]),
			Image:  strings.TrimSpace(f[1]),
			Status: strings.TrimSpace(f[2]),
		})
	}
	return rows
}

// HasName reports whether a row is named exactly name: a sibling such as
// solace-edge is not solace.
func HasName(rows []Row, name string) bool {
	for _, r := range rows {
		if name != "" && r.Name == name {
			return true
		}
	}
	return false
}

// ByImage is the rows whose image contains substr, in listing order.
func ByImage(rows []Row, substr string) []Row {
	var out []Row
	for _, r := range rows {
		if strings.Contains(r.Image, substr) {
			out = append(out, r)
		}
	}
	return out
}

// ValidName reports whether name is a container name both engines accept,
// [A-Za-z0-9][A-Za-z0-9_.-]*. A name read from the engine that fails it is reported
// rather than passed back to the engine on argv.
func ValidName(name string) bool {
	for i := 0; i < len(name); i++ {
		c := name[i]
		alnum := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
		if !alnum && (i == 0 || c != '_' && c != '.' && c != '-') {
			return false
		}
	}
	return name != ""
}

// ExecOptions are exec's flags. Stdin keeps the child's stdin open (-i), for input
// on Cmd.Stdin or an attached terminal. TTY allocates a terminal (-t), which needs
// Stdin too.
type ExecOptions struct{ Stdin, TTY bool }

// Exec runs argv in the named container: `exec [-i] [-t] <name> <argv...>`, with no
// `--`. It refuses an empty name, an empty argv, and a terminal without stdin.
func Exec(base []string, name string, o ExecOptions, argv ...string) (runner.Cmd, error) {
	switch {
	case name == "":
		return runner.Cmd{}, errors.New("containercli: exec needs a container name")
	case len(argv) == 0:
		return runner.Cmd{}, fmt.Errorf("containercli: exec in %s needs a command to run", name)
	case o.TTY && !o.Stdin:
		return runner.Cmd{}, errors.New("containercli: a tty needs stdin: ask for both or neither")
	}
	args := []string{"exec"}
	if o.Stdin {
		args = append(args, "-i")
	}
	if o.TTY {
		args = append(args, "-t")
	}
	return build.Cmd(base, append(append(args, name), argv...)...), nil
}

// Upload writes body to dest inside the named container: `exec -i <name> sh -c
// "cat > '<dest>'"`, with body on stdin, and with `mkdir -p '<dir>' && ` first when
// mkdir is set. dest is a path inside the container, so it is POSIX whatever the
// host; it is quoted, so it is one word and a leading ~ is not expanded. An empty
// name or destination is refused.
func Upload(base []string, name, dest string, body []byte, mkdir bool) (runner.Cmd, error) {
	if name == "" || dest == "" {
		return runner.Cmd{}, fmt.Errorf("containercli: an upload needs a container name and a destination, "+
			"got %q and %q", name, dest)
	}
	c, err := Exec(base, name, ExecOptions{Stdin: true}, "sh", "-c", build.UploadScript(dest, mkdir))
	c.Stdin = build.Stdin(body)
	return c, err
}

// LogOptions are the logs flags both engines take. Tail and Since are passed as
// given, so the tool validates them first; "" leaves the flag out.
type LogOptions struct {
	Follow, Timestamps bool
	Tail, Since        string
}

// Logs reads the named container's log: `logs [-f] [--timestamps] [--tail T]
// [--since S] <name>`.
func Logs(base []string, name string, o LogOptions) runner.Cmd {
	args := []string{"logs"}
	if o.Follow {
		args = append(args, "-f")
	}
	if o.Timestamps {
		args = append(args, "--timestamps")
	}
	if o.Tail != "" {
		args = append(args, "--tail", o.Tail)
	}
	if o.Since != "" {
		args = append(args, "--since", o.Since)
	}
	return build.Cmd(base, append(args, name)...)
}

// NoPrevious is the refusal for --previous on an engine: it is a Kubernetes option,
// and neither engine keeps a previous container's log.
func NoPrevious(engine Engine) error {
	return fmt.Errorf("--previous is a kubernetes option: %s keeps no previous container's log, so there "+
		"is nothing to read; leave it off", engine.title())
}
