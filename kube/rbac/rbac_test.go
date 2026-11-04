package rbac

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestCanIArgv carries solace-util's TestCanIArgsScopesByResourceKind and
// solmq-conn-util's TestPreflightReadVerbsAskOnlyWhatTheyUse shapes: a cluster-scoped
// question is asked cluster-wide, a namespaced one in its namespace, and a
// subresource is a flag.
func TestCanIArgv(t *testing.T) {
	k := []string{"kubectl", "--context", "prod"}
	for _, c := range []struct {
		p    Probe
		want []string
	}{
		{Probe{Verb: "create", Resource: "namespaces", Namespace: "solace"}, []string{"auth", "can-i", "create", "namespaces", "--all-namespaces"}},
		{Probe{Verb: "create", Resource: "customresourcedefinitions"}, []string{"auth", "can-i", "create", "customresourcedefinitions", "--all-namespaces"}},
		{Probe{Verb: "create", Resource: "secrets", Namespace: "solace"}, []string{"auth", "can-i", "create", "secrets", "-n", "solace"}},
		{Probe{Verb: "patch", Resource: "deployments", Namespace: "solace-system"}, []string{"auth", "can-i", "patch", "deployments", "-n", "solace-system"}},
		{Probe{Verb: "bind", Resource: "clusterroles", Name: "op-role"}, []string{"auth", "can-i", "bind", "clusterroles/op-role", "--all-namespaces"}},
		{Probe{Verb: "list", Resource: "pubsubpluseventbrokers", AllNamespaces: true, Namespace: "x"}, []string{"auth", "can-i", "list", "pubsubpluseventbrokers", "--all-namespaces"}},
		{Probe{Verb: "get", Resource: "pods", Subresource: "log", Namespace: "prod"}, []string{"auth", "can-i", "get", "pods", "--subresource", "log", "-n", "prod"}},
		{Probe{Verb: "create", Resource: "deployments"}, []string{"auth", "can-i", "create", "deployments"}},
		{Probe{Verb: "delete", Resource: "Namespaces"}, []string{"auth", "can-i", "delete", "Namespaces", "--all-namespaces"}},
	} {
		if got := CanI(k, c.p).Argv; !reflect.DeepEqual(got, append(append([]string(nil), k...), c.want...)) {
			t.Errorf("CanI(%+v) = %v, want %v after the base", c.p, got, c.want)
		}
	}
}

func TestClusterScoped(t *testing.T) {
	for r, want := range map[string]bool{
		"namespaces": true, "NODES": true, "clusterroles": true, "clusterrolebindings": true, "storageclasses": true,
		"persistentvolumes": true, "customresourcedefinitions": true,
		"persistentvolumeclaims": false, "deployments": false, "secrets": false, "": false,
	} {
		if got := ClusterScoped(r); got != want {
			t.Errorf("ClusterScoped(%q) = %v, want %v", r, got, want)
		}
	}
}

// TestAnswer carries solace-util's TestCanIAnswerReadsTheLastLine cases.
func TestAnswer(t *testing.T) {
	for _, c := range []struct{ out, want string }{
		{"yes\n", "yes"},
		{"no", "no"},
		{"Warning: resource 'namespaces' is not namespace scoped\nyes\n", "yes"},
		{"Warning: a\nWarning: b\n\nno - RBAC: access denied\n\n", "no - RBAC: access denied"},
		{"yes\r\n", "yes"},
		{"", ""},
		{"  \n \n", ""},
		// A warning printed after the verdict hides it, which is why Check takes stdout alone.
		{"no\nWarning: printed after the answer\n", "Warning: printed after the answer"},
	} {
		if got := Answer(c.out); got != c.want {
			t.Errorf("Answer(%q) = %q, want %q", c.out, got, c.want)
		}
	}
}

// TestRefused carries solace-util's TestCanIRefusedReadsEveryRefusalSpelling.
func TestRefused(t *testing.T) {
	for a, want := range map[string]bool{
		"no": true, "no - RBAC: access denied": true, "no (no RBAC policy matched)": true,
		"nope": false, "no such host": false, "": false, "yes": false,
	} {
		if got := Refused(a); got != want {
			t.Errorf("Refused(%q) = %v, want %v", a, got, want)
		}
	}
}

// TestVerdict carries solace-util's TestCanIAnswersWithoutFailing: kubectl exits 1
// on "no", which is still an answer; a yes with an error is not trusted.
func TestVerdict(t *testing.T) {
	exit1 := errors.New("exit status 1")
	for _, c := range []struct {
		out               string
		err               error
		allowed, answered bool
	}{
		{"yes\n", nil, true, true},
		{"Warning: x\nyes\n", nil, true, true},
		{"no\n", exit1, false, true},
		{"no - RBAC: access denied\n", exit1, false, true},
		{"", exit1, false, false},
		{"yes\n", exit1, false, false},
		{"maybe\n", nil, false, false},
	} {
		allowed, answered := Verdict(c.out, c.err)
		if allowed != c.allowed || answered != c.answered {
			t.Errorf("Verdict(%q, %v) = %v, %v; want %v, %v", c.out, c.err, allowed, answered, c.allowed, c.answered)
		}
	}
}

// TestCheck carries solace-util's TestPreflightRefusalIsNotAllowed,
// TestPreflightAcceptsAWarnedYes and TestPreflightRefusesAnUnreadableAnswer, and
// solmq-conn-util's TestPreflightRefusalNamesThePermission.
func TestCheck(t *testing.T) {
	exit1 := errors.New("exit status 1")
	if err := Check(Probe{Verb: "create", Resource: "secrets"}, "Warning: x\nyes\n", "", nil); err != nil {
		t.Errorf("a warned yes must pass: %v", err)
	}
	for _, c := range []struct {
		p    Probe
		want []string
	}{
		{Probe{Verb: "create", Resource: "secrets", Namespace: "solace"},
			[]string{`not allowed to create secrets in namespace "solace"`, `answered "no"`, "role binding granting create on secrets"}},
		{Probe{Verb: "create", Resource: "deployments"},
			[]string{"in the current namespace", "a role binding"}},
		{Probe{Verb: "delete", Resource: "namespaces", Namespace: "solace"},
			[]string{"delete namespaces in the cluster", "cluster role binding"}},
		{Probe{Verb: "get", Resource: "pods", Subresource: "log", Namespace: "prod"},
			[]string{"get pods/log", "granting get on pods/log"}},
	} {
		err := Check(c.p, "no\n", "", exit1)
		var refused *RefusedError
		if !errors.As(err, &refused) || refused.Probe != c.p || refused.Answer != "no" {
			t.Errorf("%v: err = %v, want a *RefusedError", c.p, err)
			continue
		}
		for _, w := range c.want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("%v: err = %v, want it to name %q", c.p, err, w)
			}
		}
	}
	// No verdict: the cause is wrapped, kubectl's own words follow, and the hint is to log in.
	err := Check(Probe{Verb: "create", Resource: "secrets", Namespace: "solace"}, "",
		"error: You must be logged in to the server (Unauthorized)\n", exit1)
	var refused *RefusedError
	if errors.As(err, &refused) || !errors.Is(err, exit1) {
		t.Errorf("an unanswered probe is not a refusal and wraps the cause: %v", err)
	}
	for _, w := range []string{"cannot check permission to create secrets", "Unauthorized", "oc login", "configured command"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("err = %v, want it to name %q", err, w)
		}
	}
	if err := Check(Probe{Verb: "get", Resource: "pods"}, "", "", exit1); strings.Contains(err.Error(), "\n\n") {
		t.Errorf("no stderr adds no blank line: %q", err)
	}
	// Exit 0 without a yes is refused rather than taken for a permission.
	err = Check(Probe{Verb: "get", Resource: "pods"}, "maybe\n", "", nil)
	if err == nil || !strings.Contains(err.Error(), `could not read the answer to `+"`auth can-i get pods`"+` (got "maybe")`) {
		t.Errorf("err = %v, want the unreadable-answer error", err)
	}
}

func TestProbeString(t *testing.T) {
	for _, c := range []struct {
		p    Probe
		want string
	}{
		{Probe{Verb: "get", Resource: "pods", Subresource: "log"}, "get pods/log"},
		{Probe{Verb: "bind", Resource: "clusterroles", Name: "op-role"}, "bind clusterroles/op-role"},
		{Probe{Verb: "create", Resource: "deployments", Namespace: "x"}, "create deployments"},
	} {
		if got := c.p.String(); got != c.want {
			t.Errorf("String() = %q, want %q", got, c.want)
		}
	}
}
