// Package kube builds the kubectl (or oc) commands that solmq-conn-util and
// solace-util run the same way, and reads their answers: manifests on stdin, the
// namespace read, create and delete, scaling and rollouts, reads, exec, upload and
// logs.
//
// Every builder takes base, the kubectl argv the tool has already built and checked
// with execguard, including any --context or --kubeconfig the operator configured,
// and copies it. A namespace goes after the subcommand's own arguments as -n <ns>;
// an empty one leaves the flag out, so the context's namespace applies.
//
// Builders trust the names they are given: a tool checks every namespace and object
// name with ValidLabel when it loads its configuration, since a name starting with
// '-' would be read as a flag. Nothing here executes: a tool runs the returned Cmd
// through its runner.Runner, and the flows that string commands together (create a
// namespace only when absent, wait for a rollout) stay in the tool.
package kube

import (
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/hafio/go-deploy-utility/runner"
)

// cmd is base followed by args, in a new slice.
func cmd(base []string, args ...string) runner.Cmd {
	argv := make([]string, 0, len(base)+len(args))
	argv = append(argv, base...)
	return runner.Cmd{Argv: append(argv, args...)}
}

// stdin is a copy of b that is never nil, so the child always reads exactly b and
// then EOF.
func stdin(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// nsArgs is -n <ns>, or nothing for the context's own namespace.
func nsArgs(ns string) []string {
	if ns == "" {
		return nil
	}
	return []string{"-n", ns}
}

// Apply creates or updates what manifest describes: `apply -f -`, the manifest on
// stdin, so it is never written to disk.
func Apply(base []string, manifest []byte) runner.Cmd {
	c := cmd(base, "apply", "-f", "-")
	c.Stdin = stdin(manifest)
	return c
}

// Delete removes what manifest describes, and is not an error for what is already
// gone: `delete -f - --ignore-not-found`, the manifest on stdin.
func Delete(base []string, manifest []byte) runner.Cmd {
	c := cmd(base, "delete", "-f", "-", "--ignore-not-found")
	c.Stdin = stdin(manifest)
	return c
}

// NamespaceManifest is the Namespace document both tools write. Apply it on its own,
// before anything that goes into it. ns must be a DNS-1123 label, since it is written
// into the YAML as it is.
func NamespaceManifest(ns string) ([]byte, error) {
	if !ValidLabel(ns) {
		return nil, fmt.Errorf("kube: the namespace %q is not a DNS-1123 label: use lower-case letters, digits "+
			"and '-', start and end with a letter or digit, at most 63 characters", ns)
	}
	return []byte("apiVersion: v1\nkind: Namespace\nmetadata:\n  name: " + ns + "\n"), nil
}

// NamespaceState is what GetNamespace found. The zero value is NamespaceUnknown, so a
// state nobody checked is never taken for an answer.
type NamespaceState int

const (
	// NamespaceUnknown: the read failed or answered something else, typically because
	// the identity may not get namespaces. Decide by the create probe instead.
	NamespaceUnknown NamespaceState = iota

	// NamespaceAbsent: the namespace does not exist.
	NamespaceAbsent

	// NamespaceExists: the namespace exists, so nothing needs creating and no
	// cluster-wide permission is asked for.
	NamespaceExists
)

// GetNamespace reads whether ns exists: `get namespace <ns> --ignore-not-found -o
// name`.
func GetNamespace(base []string, ns string) runner.Cmd {
	return Get(base, GetOptions{IgnoreNotFound: true, Output: "name"}, "namespace", ns)
}

// ParseNamespace reads GetNamespace's answer: exactly namespace/<ns> exists, nothing
// is absent, and anything else is unknown. A read that failed is unknown too, so do
// not call this with its output.
func ParseNamespace(out, ns string) NamespaceState {
	switch strings.TrimSpace(out) {
	case "namespace/" + ns:
		return NamespaceExists
	case "":
		return NamespaceAbsent
	}
	return NamespaceUnknown
}

// MissingNamespace is the error for a namespace known to be absent that this identity
// may not create (err is the refused create probe, wrapped). The namespace has to
// exist before anything is deployed into it, so it also says who can create it.
func MissingNamespace(base []string, ns string, err error) error {
	return fmt.Errorf("namespace %q does not exist, and this identity may not create it (%w).\n"+
		"  Or have a cluster admin create it: %s create namespace %s", ns, err, strings.Join(base, " "), ns)
}

// DeleteNamespace deletes ns and everything left in it: `delete namespace <ns>
// --ignore-not-found`. Run it only through the gate in kube/nsgate.
func DeleteNamespace(base []string, ns string) runner.Cmd {
	return cmd(base, "delete", "namespace", ns, "--ignore-not-found")
}

// Scale sets the replica count of kind/name: `scale <kind>/<name> -n <ns>
// --replicas=<n>`.
func Scale(base []string, kind, name, ns string, replicas int) runner.Cmd {
	args := append([]string{"scale", kind + "/" + name}, nsArgs(ns)...)
	return cmd(base, append(args, "--replicas="+strconv.Itoa(replicas))...)
}

// RolloutRestart restarts the pods of kind/name in a rolling fashion: `rollout restart
// <kind>/<name> -n <ns>`.
func RolloutRestart(base []string, kind, name, ns string) runner.Cmd {
	return cmd(base, append([]string{"rollout", "restart", kind + "/" + name}, nsArgs(ns)...)...)
}

// RolloutTimeout is how long RolloutStatus waits when given no timeout.
const RolloutTimeout = "300s"

// RolloutStatus waits until kind/name has rolled out, or the timeout passes:
// `rollout status <kind>/<name> -n <ns> --timeout=<t>`, RolloutTimeout when timeout
// is "". The wait is a convenience: when it is refused (Forbidden), the change is
// still applied, so a tool warns rather than failing. A timeout is a real result.
func RolloutStatus(base []string, kind, name, ns, timeout string) runner.Cmd {
	if timeout == "" {
		timeout = RolloutTimeout
	}
	args := append([]string{"rollout", "status", kind + "/" + name}, nsArgs(ns)...)
	return cmd(base, append(args, "--timeout="+timeout)...)
}

// Forbidden reports whether kubectl's output is an RBAC refusal, as opposed to a
// missing object, a timeout or a connection failure. A tool uses it to turn a step it
// runs only for convenience into a warning when the identity may not take it.
func Forbidden(out string) bool {
	return strings.Contains(out, "Error from server (Forbidden)") || strings.Contains(out, " is forbidden: ")
}

// GetOptions are get's flags. AllNamespaces outranks Namespace.
type GetOptions struct {
	Namespace      string
	AllNamespaces  bool
	Selector       string // -l
	IgnoreNotFound bool   // a named object that is absent is no error; never use it on a list
	Output         string // -o: json, name, wide, jsonpath=...
}

// Get reads objects: `get <what...> [--all-namespaces | -n <ns>] [-l <selector>]
// [--ignore-not-found] [-o <output>]`.
func Get(base []string, o GetOptions, what ...string) runner.Cmd {
	args := append([]string{"get"}, what...)
	if o.AllNamespaces {
		args = append(args, "--all-namespaces")
	} else {
		args = append(args, nsArgs(o.Namespace)...)
	}
	if o.Selector != "" {
		args = append(args, "-l", o.Selector)
	}
	if o.IgnoreNotFound {
		args = append(args, "--ignore-not-found")
	}
	if o.Output != "" {
		args = append(args, "-o", o.Output)
	}
	return cmd(base, args...)
}

// ExecOptions are exec's flags. Stdin keeps the child's stdin open (-i); TTY
// allocates a terminal (-t), which needs Stdin too. Container picks one container of
// a multi-container pod (-c).
type ExecOptions struct {
	Namespace, Container string
	Stdin, TTY           bool
}

// Exec runs argv in pod: `exec [-i] [-t] <pod> [-n <ns>] [-c <container>] --
// <argv...>`. It refuses an empty pod, an empty argv, and a terminal without stdin.
func Exec(base []string, pod string, o ExecOptions, argv ...string) (runner.Cmd, error) {
	switch {
	case pod == "":
		return runner.Cmd{}, errors.New("kube: exec needs a pod name")
	case len(argv) == 0:
		return runner.Cmd{}, fmt.Errorf("kube: exec in %s needs a command to run", pod)
	case o.TTY && !o.Stdin:
		return runner.Cmd{}, errors.New("kube: a tty needs stdin: ask for both or neither")
	}
	args := []string{"exec"}
	if o.Stdin {
		args = append(args, "-i")
	}
	if o.TTY {
		args = append(args, "-t")
	}
	args = append(append(args, pod), nsArgs(o.Namespace)...)
	if o.Container != "" {
		args = append(args, "-c", o.Container)
	}
	return cmd(base, append(append(args, "--"), argv...)...), nil
}

// shQuote single-quotes s for sh, so it is one word with nothing expanded.
func shQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// Upload writes body to dest inside pod, through Exec with stdin: `sh -c "cat >
// '<dest>'"`, with `mkdir -p '<dir>' && ` first when mkdir is set. dest is a POSIX
// path inside the container, quoted so it is one word. An empty pod or destination
// is refused.
func Upload(base []string, pod string, o ExecOptions, dest string, body []byte, mkdir bool) (runner.Cmd, error) {
	if dest == "" {
		return runner.Cmd{}, fmt.Errorf("kube: an upload to %q needs a destination path", pod)
	}
	script := "cat > " + shQuote(dest)
	if mkdir {
		script = "mkdir -p " + shQuote(path.Dir(dest)) + " && " + script
	}
	o.Stdin, o.TTY = true, false
	c, err := Exec(base, pod, o, "sh", "-c", script)
	if err != nil {
		return runner.Cmd{}, err
	}
	c.Stdin = stdin(body)
	return c, nil
}

// LogOptions are logs' flags. Tail and Since are passed as given, so the tool
// validates them first; "" leaves the flag out, and "0" is zero lines.
type LogOptions struct {
	Namespace, Container         string
	Previous, Follow, Timestamps bool
	Tail, Since                  string
}

// Logs reads target's log, a pod or kind/name: `logs <target> [-n <ns>] [-c
// <container>] [-p] [-f] [--timestamps] [--tail <t>] [--since <s>]`.
func Logs(base []string, target string, o LogOptions) runner.Cmd {
	args := append([]string{"logs", target}, nsArgs(o.Namespace)...)
	if o.Container != "" {
		args = append(args, "-c", o.Container)
	}
	for _, f := range []struct {
		on   bool
		flag string
	}{{o.Previous, "-p"}, {o.Follow, "-f"}, {o.Timestamps, "--timestamps"}} {
		if f.on {
			args = append(args, f.flag)
		}
	}
	if o.Tail != "" {
		args = append(args, "--tail", o.Tail)
	}
	if o.Since != "" {
		args = append(args, "--since", o.Since)
	}
	return cmd(base, args...)
}

// ValidLabel reports whether s is a DNS-1123 label: 1 to 63 lower-case letters,
// digits and '-', starting and ending with a letter or digit. Namespaces are labels,
// and both tools hold their object names to the same rule.
func ValidLabel(s string) bool {
	if len(s) == 0 || len(s) > 63 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		alnum := c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
		if !alnum && (c != '-' || i == 0 || i == len(s)-1) {
			return false
		}
	}
	return true
}
