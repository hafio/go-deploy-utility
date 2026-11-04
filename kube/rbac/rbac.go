// Package rbac asks kubectl whether this identity may do something before a tool
// does it (`auth can-i`), and reads the answer the way solmq-conn-util and
// solace-util both do: the last line is the verdict, a refusal is told apart from no
// answer at all, and a refusal says what to ask for.
//
// A tool probes only what it is about to write, and only what that write needs: the
// lowest permission that does the job. A step taken only for convenience is not
// probed as a requirement, and a refusal of one is a warning, never a failure.
package rbac

import (
	"fmt"
	"strings"

	"github.com/hafio/go-deploy-utility/runner"
)

// Probe is one permission question: may this identity Verb the Resource (a plural
// resource name, such as deployments or namespaces), optionally the one object Name
// or its Subresource (log, exec), in Namespace. AllNamespaces asks cluster-wide, as
// does any cluster-scoped resource.
type Probe struct {
	Verb, Resource, Subresource, Name, Namespace string
	AllNamespaces                                bool
}

// what is the resource as a message names it: resource[/name][/subresource].
func (p Probe) what() string {
	s := p.Resource
	if p.Name != "" {
		s += "/" + p.Name
	}
	if p.Subresource != "" {
		s += "/" + p.Subresource
	}
	return s
}

// String is the permission as a message names it, such as "get pods/log" or
// "bind clusterroles/operator-role".
func (p Probe) String() string { return p.Verb + " " + p.what() }

// clusterWide reports whether the question is asked of the whole cluster.
func (p Probe) clusterWide() bool { return p.AllNamespaces || ClusterScoped(p.Resource) }

// where is where the permission applies, and the kind of binding that grants it
// there. A role binding cannot grant anything outside its own namespace, so a
// cluster-wide question names a cluster role binding.
func (p Probe) where() (where, binding string) {
	switch {
	case p.clusterWide():
		return "the cluster", "cluster role binding"
	case p.Namespace != "":
		return fmt.Sprintf("namespace %q", p.Namespace), "role binding"
	}
	return "the current namespace", "role binding"
}

// ClusterScoped reports whether resource (a plural resource name, in any case) lives
// outside every namespace. A namespace on such a question is at best ignored and at
// worst fails with a NotFound that has nothing to do with the permission.
func ClusterScoped(resource string) bool {
	switch strings.ToLower(resource) {
	case "customresourcedefinitions", "namespaces", "nodes", "clusterroles", "clusterrolebindings",
		"storageclasses", "persistentvolumes":
		return true
	}
	return false
}

// CanI asks the question: `auth can-i <verb> <resource>[/<name>] [--subresource
// <s>]`, then --all-namespaces for a cluster-wide question, else -n <namespace>, or
// nothing for the context's namespace. The subresource is a flag because
// `pods/log` would read as a pod named log. base is the guarded kubectl argv, and
// it is copied.
func CanI(base []string, p Probe) runner.Cmd {
	target := p.Resource
	if p.Name != "" {
		target += "/" + p.Name
	}
	argv := make([]string, 0, len(base)+8)
	argv = append(append(argv, base...), "auth", "can-i", p.Verb, target)
	if p.Subresource != "" {
		argv = append(argv, "--subresource", p.Subresource)
	}
	switch {
	case p.clusterWide():
		argv = append(argv, "--all-namespaces")
	case p.Namespace != "":
		argv = append(argv, "-n", p.Namespace)
	}
	return runner.Cmd{Argv: argv}
}

// Answer is the verdict in can-i's stdout: its last non-empty line, trimmed, so a
// warning printed above it, or a CRLF, does not hide it. "" when there is none.
func Answer(out string) string {
	lines := strings.Split(out, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}

// Refused reports whether answer is a refusal, in any of kubectl's three spellings:
// "no", "no - <reason>" and "no (<evaluation error>)". Something that merely starts
// with "no", such as "no such host", is not one.
func Refused(answer string) bool {
	return answer == "no" || strings.HasPrefix(answer, "no - ") || strings.HasPrefix(answer, "no (")
}

// Verdict reads a probe without failing on it, for a caller choosing between two
// paths: allowed when the probe exited 0 and answered yes; answered when the API
// server gave a verdict at all. kubectl exits 1 on a refusal, so a refusal counts as
// answered whatever the exit status.
func Verdict(stdout string, runErr error) (allowed, answered bool) {
	a := Answer(stdout)
	switch {
	case runErr == nil && a == "yes":
		return true, true
	case Refused(a):
		return false, true
	}
	return false, false
}

// RefusedError is a refusal the API server actually answered: this is RBAC, not a
// login problem. Find it with errors.As.
type RefusedError struct {
	Probe  Probe
	Answer string
}

func (e *RefusedError) Error() string {
	where, binding := e.Probe.where()
	return fmt.Sprintf("not allowed to %s in %s (auth can-i answered %q): ask a cluster admin for a %s "+
		"granting %s on %s there, or select a context that has it, then re-run",
		e.Probe, where, e.Answer, binding, e.Probe.Verb, e.Probe.what())
}

// Check turns a probe's result into nil, or an error that says what to do. Pass
// stdout alone (runner.Split or runner.Output), since a warning on stderr after the
// verdict would otherwise hide it, and stderr when it was captured ("" when it already
// reached the terminal).
//
// A refusal is a *RefusedError. No verdict at all (an expired login, no context, an
// unreachable server) wraps runErr with kubectl's own words and the login hint. An
// exit 0 without "yes" means a wrapper changed the output, and is refused rather than
// taken for a permission nobody confirmed.
func Check(p Probe, stdout, stderr string, runErr error) error {
	a := Answer(stdout)
	switch {
	case runErr == nil && a == "yes":
		return nil
	case Refused(a):
		return &RefusedError{Probe: p, Answer: a}
	case runErr != nil:
		where, _ := p.where()
		said := ""
		if s := strings.TrimSpace(stderr); s != "" {
			said = "\n" + s
		}
		return fmt.Errorf("cannot check permission to %s in %s: %w%s\n"+
			"  log in or select a working context first (kubectl: `kubectl config use-context <context>`; "+
			"OpenShift: `oc login <server>`), or point the configured command at the right profile",
			p, where, runErr, said)
	}
	return fmt.Errorf("could not read the answer to `auth can-i %s` (got %q); if the configured command "+
		"wraps kubectl, make sure it passes stdout through unchanged", p, a)
}
