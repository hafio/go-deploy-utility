package kube

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestApplyAndDelete carries both tools' stdin rules (solmq-conn-util's
// TestKubernetesDeployApplyOnStdin and TestKubernetesRemoveUsesDeleteVerb,
// solace-util's TestApplyOnStdin and TestDeleteStdin): the manifest on stdin, never
// on argv or disk.
func TestApplyAndDelete(t *testing.T) {
	base := []string{"kubectl", "--context", "prod"}
	manifest := []byte("kind: ConfigMap\n")
	a := Apply(base, manifest)
	if want := []string{"kubectl", "--context", "prod", "apply", "-f", "-"}; !reflect.DeepEqual(a.Argv, want) {
		t.Errorf("Apply argv = %v, want %v", a.Argv, want)
	}
	d := Delete([]string{"oc"}, manifest)
	if want := []string{"oc", "delete", "-f", "-", "--ignore-not-found"}; !reflect.DeepEqual(d.Argv, want) {
		t.Errorf("Delete argv = %v, want %v", d.Argv, want)
	}
	manifest[0] = 'X'
	if string(a.Stdin) != "kind: ConfigMap\n" || string(d.Stdin) != "kind: ConfigMap\n" {
		t.Errorf("stdin must be a copy: %q / %q", a.Stdin, d.Stdin)
	}
	if Apply(base, nil).Stdin == nil || Delete(base, nil).Stdin == nil {
		t.Error("a nil manifest is still stdin, never the terminal's")
	}
}

// TestNamespaceManifest pins the four lines both tools write, and refuses a name the
// YAML could not carry as it is.
func TestNamespaceManifest(t *testing.T) {
	got, err := NamespaceManifest("solace")
	if err != nil || string(got) != "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: solace\n" {
		t.Errorf("NamespaceManifest = %q, %v", got, err)
	}
	for _, bad := range []string{"", "Solace", "a: b", "x\nkind: Secret", "-x"} {
		if _, err := NamespaceManifest(bad); err == nil || !strings.Contains(err.Error(), "DNS-1123") {
			t.Errorf("NamespaceManifest(%q) err = %v, want a refusal naming the rule", bad, err)
		}
	}
}

// TestParseNamespace carries solace-util's TestCreateNamespace* read cases: exists,
// absent, and anything else unknown; the zero value is unknown.
func TestParseNamespace(t *testing.T) {
	if want := []string{"kubectl", "get", "namespace", "solace", "--ignore-not-found", "-o", "name"}; !reflect.DeepEqual(GetNamespace([]string{"kubectl"}, "solace").Argv, want) {
		t.Errorf("GetNamespace argv = %v, want %v", GetNamespace([]string{"kubectl"}, "solace").Argv, want)
	}
	for _, c := range []struct {
		out  string
		want NamespaceState
	}{
		{"namespace/solace\n", NamespaceExists},
		{"", NamespaceAbsent},
		{"  \n", NamespaceAbsent},
		{"namespace/other", NamespaceUnknown},
		{"Error from server", NamespaceUnknown},
	} {
		if got := ParseNamespace(c.out, "solace"); got != c.want {
			t.Errorf("ParseNamespace(%q) = %d, want %d", c.out, got, c.want)
		}
	}
	var zero NamespaceState
	if zero != NamespaceUnknown {
		t.Error("the zero NamespaceState must be unknown")
	}
}

// TestMissingNamespace carries solace-util's TestCreateNamespaceStopsOnPreflightFailure
// wording: what is missing, the refusal wrapped, and who can fix it with which
// command.
func TestMissingNamespace(t *testing.T) {
	cause := errors.New("not allowed to create namespaces in the cluster")
	err := MissingNamespace([]string{"oc"}, "solace", cause)
	if !errors.Is(err, cause) {
		t.Errorf("the refusal must be wrapped: %v", err)
	}
	for _, want := range []string{`namespace "solace" does not exist`, "may not create it", "oc create namespace solace"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to name %q", err, want)
		}
	}
}

func TestDeleteNamespace(t *testing.T) {
	if want := []string{"kubectl", "delete", "namespace", "solace", "--ignore-not-found"}; !reflect.DeepEqual(DeleteNamespace([]string{"kubectl"}, "solace").Argv, want) {
		t.Errorf("argv = %v, want %v", DeleteNamespace([]string{"kubectl"}, "solace").Argv, want)
	}
}

// TestScaleAndRollout carries solmq-conn-util's TestLifecycleArgv and solace-util's
// TestRestartPod and TestReplicasStart: the namespace after the target, and the
// default wait.
func TestScaleAndRollout(t *testing.T) {
	k := []string{"kubectl"}
	for _, c := range []struct {
		got, want []string
	}{
		{Scale(k, "deployment", "c", "ns", 0).Argv, []string{"kubectl", "scale", "deployment/c", "-n", "ns", "--replicas=0"}},
		{Scale([]string{"oc"}, "statefulset", "b", "", 3).Argv, []string{"oc", "scale", "statefulset/b", "--replicas=3"}},
		{RolloutRestart(k, "deployment", "c", "ns").Argv, []string{"kubectl", "rollout", "restart", "deployment/c", "-n", "ns"}},
		{RolloutStatus(k, "statefulset", "b", "solace", "").Argv,
			[]string{"kubectl", "rollout", "status", "statefulset/b", "-n", "solace", "--timeout=300s"}},
		{RolloutStatus(k, "deployment", "op", "", "60s").Argv, []string{"kubectl", "rollout", "status", "deployment/op", "--timeout=60s"}},
	} {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("argv = %v, want %v", c.got, c.want)
		}
	}
}

// TestForbidden tells an RBAC refusal from every other failure, so only a refused
// convenience step is turned into a warning.
func TestForbidden(t *testing.T) {
	for _, c := range []struct {
		out  string
		want bool
	}{
		{`Error from server (Forbidden): deployments.apps "c" is forbidden: User "u" cannot get resource`, true},
		{`error: deployments.apps "c" is forbidden: User "u" cannot watch resource "deployments"`, true},
		{`Error from server (NotFound): deployments.apps "c" not found`, false},
		{`error: timed out waiting for the condition`, false},
		{`The connection to the server localhost:8080 was refused`, false},
		{``, false},
	} {
		if got := Forbidden(c.out); got != c.want {
			t.Errorf("Forbidden(%q) = %v, want %v", c.out, got, c.want)
		}
	}
}

// TestGet carries solmq-conn-util's TestKubernetesPodsJSONArgv and
// TestKubernetesGetJSONArgv shapes and solace-util's name and jsonpath reads.
func TestGet(t *testing.T) {
	k := []string{"kubectl"}
	for _, c := range []struct {
		o    GetOptions
		what []string
		want []string
	}{
		{GetOptions{Namespace: "prod", Selector: "app=c", Output: "json"}, []string{"pods"},
			[]string{"kubectl", "get", "pods", "-n", "prod", "-l", "app=c", "-o", "json"}},
		{GetOptions{Namespace: "prod", AllNamespaces: true, Output: "json"}, []string{"pods", "a", "b"},
			[]string{"kubectl", "get", "pods", "a", "b", "--all-namespaces", "-o", "json"}},
		{GetOptions{Namespace: "prod", Output: "json"}, []string{"deployment", "c"},
			[]string{"kubectl", "get", "deployment", "c", "-n", "prod", "-o", "json"}},
		{GetOptions{Namespace: "solace", IgnoreNotFound: true, Output: "name"}, []string{"pod", "p-0"},
			[]string{"kubectl", "get", "pod", "p-0", "-n", "solace", "--ignore-not-found", "-o", "name"}},
		{GetOptions{Output: "jsonpath={.metadata.name}"}, []string{"secret", "s"},
			[]string{"kubectl", "get", "secret", "s", "-o", "jsonpath={.metadata.name}"}},
		{GetOptions{}, []string{"svc"}, []string{"kubectl", "get", "svc"}},
	} {
		if got := Get(k, c.o, c.what...).Argv; !reflect.DeepEqual(got, c.want) {
			t.Errorf("Get(%+v, %v) = %v, want %v", c.o, c.what, got, c.want)
		}
	}
}

// TestExec carries solmq-conn-util's TestExecArgvPerPlatform kubernetes rows and
// solace-util's TestTransportExecArgs: the pod, then its namespace and container,
// then `--`.
func TestExec(t *testing.T) {
	for _, c := range []struct {
		o    ExecOptions
		want []string
	}{
		{ExecOptions{Namespace: "prod", Container: "connector", Stdin: true, TTY: true},
			[]string{"kubectl", "exec", "-i", "-t", "pod-0", "-n", "prod", "-c", "connector", "--", "sh"}},
		{ExecOptions{Namespace: "solace", Stdin: true}, []string{"kubectl", "exec", "-i", "pod-0", "-n", "solace", "--", "sh"}},
		{ExecOptions{}, []string{"kubectl", "exec", "pod-0", "--", "sh"}},
	} {
		got, err := Exec([]string{"kubectl"}, "pod-0", c.o, "sh")
		if err != nil || !reflect.DeepEqual(got.Argv, c.want) || got.Stdin != nil {
			t.Errorf("Exec(%+v) = %+v, %v; want argv %v and no stdin", c.o, got, err, c.want)
		}
	}
}

func TestExecRefuses(t *testing.T) {
	for _, c := range []struct {
		pod  string
		argv []string
		o    ExecOptions
		want string
	}{
		{"", []string{"sh"}, ExecOptions{}, "pod name"},
		{"p", nil, ExecOptions{}, "command to run"},
		{"p", []string{"sh"}, ExecOptions{TTY: true}, "a tty needs stdin"},
	} {
		if _, err := Exec([]string{"kubectl"}, c.pod, c.o, c.argv...); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("Exec(%q, %v, %+v) err = %v, want %q", c.pod, c.argv, c.o, err, c.want)
		}
	}
}

// TestUpload carries solace-util's TestTransportUpload and TestTransportUploadQuotesDest:
// the body on stdin only, and the destination one quoted word.
func TestUpload(t *testing.T) {
	c, err := Upload([]string{"kubectl"}, "p-0", ExecOptions{Namespace: "solace", TTY: true}, "/tmp/a'b", []byte("secret"), false)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"kubectl", "exec", "-i", "p-0", "-n", "solace", "--", "sh", "-c", `cat > '/tmp/a'\''b'`}
	if !reflect.DeepEqual(c.Argv, want) || string(c.Stdin) != "secret" {
		t.Errorf("Upload = %q stdin %q; want %q with the body on stdin", c.Argv, c.Stdin, want)
	}
	c, err = Upload([]string{"kubectl"}, "p-0", ExecOptions{Container: "connector"}, "/tmp/solmq/s.sh", nil, true)
	if err != nil || c.Argv[len(c.Argv)-1] != `mkdir -p '/tmp/solmq' && cat > '/tmp/solmq/s.sh'` || c.Stdin == nil {
		t.Errorf("mkdir Upload = %q, %v", c.Argv, err)
	}
	if _, err := Upload([]string{"kubectl"}, "p-0", ExecOptions{}, "", nil, false); err == nil {
		t.Error("an empty destination must be refused")
	}
	if _, err := Upload([]string{"kubectl"}, "", ExecOptions{}, "/tmp/x", nil, false); err == nil {
		t.Error("an empty pod must be refused")
	}
}

// TestLogs carries solmq-conn-util's TestLogsArgvPerPlatform kubernetes rows and
// solace-util's TestLogsPassthrough: the target first, its namespace and container,
// then the flags in a fixed order; "0" lines survives.
func TestLogs(t *testing.T) {
	got := Logs([]string{"kubectl"}, "pod-a", LogOptions{Namespace: "prod", Container: "connector", Previous: true,
		Follow: true, Timestamps: true, Tail: "50", Since: "10m0s"}).Argv
	want := []string{"kubectl", "logs", "pod-a", "-n", "prod", "-c", "connector", "-p", "-f", "--timestamps",
		"--tail", "50", "--since", "10m0s"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
	if got := Logs([]string{"kubectl"}, "pod/p-0", LogOptions{Namespace: "solace", Tail: "0"}).Argv; !reflect.DeepEqual(got,
		[]string{"kubectl", "logs", "pod/p-0", "-n", "solace", "--tail", "0"}) {
		t.Errorf("argv = %v", got)
	}
}

func TestValidLabel(t *testing.T) {
	for s, want := range map[string]bool{
		"solace": true, "a": true, "a-1": true, "0x": true, strings.Repeat("a", 63): true,
		"": false, strings.Repeat("a", 64): false, "-a": false, "a-": false, "A": false, "a.b": false,
		"a_b": false, "a b": false,
	} {
		if got := ValidLabel(s); got != want {
			t.Errorf("ValidLabel(%q) = %v, want %v", s, got, want)
		}
	}
}

// TestBuildersCopyInputs: no builder writes through a base with spare capacity.
func TestBuildersCopyInputs(t *testing.T) {
	base := make([]string, 1, 16)
	base[0] = "kubectl"
	a := Scale(base, "deployment", "c", "", 1)
	b := Get(base, GetOptions{}, "pods")
	if a.Argv[1] != "scale" || b.Argv[1] != "get" || base[:2][1] != "" {
		t.Errorf("argvs share the base: %v / %v / %v", a.Argv, b.Argv, base[:2])
	}
}
