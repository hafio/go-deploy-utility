// Package build holds the argv and stdin helpers the builder packages share:
// containercli, kube and podman. It is internal, so neither tool can import it; the
// builders' own signatures stay the public surface, and their output does not
// depend on which package holds these lines.
package build

import (
	"path"
	"strings"

	"github.com/hafio/go-deploy-utility/runner"
)

// Cmd is base followed by args, in a new slice, so the caller's base is never
// written through.
func Cmd(base []string, args ...string) runner.Cmd {
	argv := make([]string, 0, len(base)+len(args))
	argv = append(argv, base...)
	return runner.Cmd{Argv: append(argv, args...)}
}

// Stdin is a copy of b that is never nil, so the child always reads exactly b and
// then EOF.
func Stdin(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// ShQuote single-quotes s for sh, so it is one word with nothing expanded.
func ShQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// UploadScript is the sh script that writes its stdin to dest: `cat > '<dest>'`,
// with `mkdir -p '<dir>' && ` first when mkdir is set. dest is a path inside a
// container, so it is POSIX whatever the host, and path, not filepath, takes its
// directory. Both are quoted, so each is one word and a leading ~ is not expanded.
func UploadScript(dest string, mkdir bool) string {
	script := "cat > " + ShQuote(dest)
	if mkdir {
		script = "mkdir -p " + ShQuote(path.Dir(dest)) + " && " + script
	}
	return script
}
