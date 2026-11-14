package execguard

import (
	"strings"
	"testing"
)

// testTool is the binary name the escalation remedy names in these tests.
const testTool = "example-tool"

// The rules fixtures mirror solace-util's four field rules, so the lifted matrices
// keep their meaning: one Kubernetes CLI field, one runtime field per container
// platform, and the compose field with its one subword.
func clusterRules() Rules {
	return Rules{Field: "kubernetes.command", Scope: "k8s", Allowed: []string{"kubectl", "oc"}}
}

func runtimeRules(platform string) Rules {
	if platform == "podman" {
		return Rules{Field: "podman.command", Scope: "podman", Allowed: []string{"podman"}}
	}
	return Rules{Field: "docker.command", Scope: "docker", Allowed: []string{"docker", "docker-compose", "nerdctl"}}
}

func composeRules() Rules {
	r := runtimeRules("docker")
	r.Field = "docker.compose"
	r.Subword = "compose"
	return r
}

// forced builds Approvals straight from a set, past Approve, the way a caller
// that populated approvals some other way would arrive.
func forced(extra map[string]bool) Approvals { return Approvals{names: extra} }

// TestCheckCommandAccepts is the accept half of the guard matrix: the shapes an
// operator legitimately writes. Each case names the field it stands for, because
// the same rules apply to every command field from one implementation.
func TestCheckCommandAccepts(t *testing.T) {
	cases := []struct {
		name  string
		rules Rules
		cmd   []string
		extra map[string]bool
	}{
		// Every allowlisted binary, bare.
		{"kubectl bare", clusterRules(), []string{"kubectl"}, nil},
		{"oc bare", clusterRules(), []string{"oc"}, nil},
		{"docker bare", runtimeRules("docker"), []string{"docker"}, nil},
		{"nerdctl bare", runtimeRules("docker"), []string{"nerdctl"}, nil},
		{"podman bare", runtimeRules("podman"), []string{"podman"}, nil},

		// A profile: flags and their values are the normal way to carry a context.
		{"flag args", clusterRules(), []string{"kubectl", "--context", "prod", "-n", "solace"}, nil},
		{"kubeconfig profile", clusterRules(), []string{"kubectl", "--kubeconfig", "/tmp/kc"}, nil},
		// --flag=value is self-contained, so the NEXT token is not its value and
		// must stand on its own -- here another flag.
		{"joined flag then flag", clusterRules(), []string{"kubectl", "--context=prod", "--namespace=solace"}, nil},
		// A bare "-" is a flag by shape; kubectl and docker both use it for stdin.
		{"lone dash", clusterRules(), []string{"kubectl", "-"}, nil},

		// The subword is accepted LAST after any prefix that already validated --
		// the tokens before it do not have to be an allowed binary. This is a tool's
		// own derived default: from `docker --context prod` it appends `compose`.
		{"compose after a flag value", composeRules(), []string{"docker", "--context", "prod", "compose"}, nil},
		{"compose after a joined flag", composeRules(), []string{"docker", "--log-level=error", "compose"}, nil},

		// An EMBEDDED tilde is an 8.3 short name, which is what a real Windows home
		// directory routinely is. Only a LEADING one is refused.
		{"8.3 short name in a flag value", clusterRules(),
			[]string{"kubectl", "--kubeconfig", "C:/Users/RUNNER~1/.kube/config"}, nil},

		// Windows portability: one .exe suffix is stripped before the allowlist.
		{"exe suffix", clusterRules(), []string{"kubectl.exe"}, nil},
		{"EXE suffix", runtimeRules("docker"), []string{"docker.EXE"}, nil},

		// Chained runners, only WITH the operator's approval. All non-escalating:
		// an approval can never carry sudo and its relatives.
		{"lima podman with hatch", runtimeRules("podman"), []string{"lima", "podman"}, map[string]bool{"lima": true}},
		{"microk8s kubectl with hatch", clusterRules(), []string{"microk8s", "kubectl"}, map[string]bool{"microk8s": true}},
		{"lima flag then nerdctl", runtimeRules("docker"), []string{"lima", "--tty=false", "nerdctl"}, map[string]bool{"lima": true}},

		// docker.compose is the one field allowed a bare subcommand, and only this one.
		{"compose plugin", composeRules(), []string{"docker", "compose"}, nil},
		{"compose standalone", composeRules(), []string{"docker-compose"}, nil},
		// Derived from a wrapped runtime: `compose` still lands last.
		{"compose behind a wrapper", composeRules(), []string{"lima", "nerdctl", "compose"}, map[string]bool{"lima": true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := Check(tc.rules, tc.cmd, forced(tc.extra)); err != nil {
				t.Errorf("Check(%q) = %v, want accepted", tc.cmd, err)
			}
		})
	}
}

// TestCheckCommandRejects is the reject half. Every case carries the substring the
// message must contain, because an error that does not name the offending token and
// the way out is a failed error, not a passed test.
func TestCheckCommandRejects(t *testing.T) {
	cases := []struct {
		name  string
		rules Rules
		cmd   []string
		extra map[string]bool
		want  string
	}{
		// Empty.
		{"empty command", clusterRules(), []string{}, nil, "kubernetes.command is empty"},
		{"nil command", clusterRules(), nil, nil, "kubernetes.command is empty"},
		{"empty argument", clusterRules(), []string{"kubectl", ""}, nil, "kubernetes.command[1] is an empty argument"},

		// argv[0]: an unlisted binary, whatever it is.
		{"curl", clusterRules(), []string{"curl"}, nil, `"curl" is not a binary this tool runs`},
		{"bash", runtimeRules("docker"), []string{"bash"}, nil, `"bash" is not a binary this tool runs`},
		// Right binary, wrong platform: podman is not a Kubernetes CLI.
		{"cross-platform binary", clusterRules(), []string{"podman"}, nil, `"podman" is not a binary this tool runs`},
		// The allowlist and the escape hatch are both named in the message.
		{"names the way out", clusterRules(), []string{"curl"}, nil, "--allow-command curl"},
		{"names the allowlist", clusterRules(), []string{"curl"}, nil, "kubectl, oc"},

		// argv[0]: any path form. This is the attack the bare-name rule exists for --
		// a kubectl shipped in the same archive as the env file.
		{"relative path", clusterRules(), []string{"./kubectl"}, nil, "must be a bare binary name, not a path"},
		{"absolute path", clusterRules(), []string{"/usr/bin/kubectl"}, nil, "must be a bare binary name, not a path"},
		{"windows path", clusterRules(), []string{"C:/tools/kubectl.exe"}, nil, "must be a bare binary name, not a path"},
		{"parent dir", clusterRules(), []string{"../kubectl"}, nil, "must be a bare binary name, not a path"},
		// A backslash path trips the charset first, which is also a refusal.
		{"backslash path", clusterRules(), []string{`.\kubectl.exe`}, nil, `is not allowed in a command token`},

		// Flag shape: a bare word in subcommand position.
		{"bare subcommand", clusterRules(), []string{"kubectl", "delete"}, nil, "not allowed in subcommand position"},
		{"smuggled delete", clusterRules(), []string{"kubectl", "delete", "ns", "prod"}, nil, "not allowed in subcommand position"},
		{"compose smuggling rm", composeRules(), []string{"docker", "rm"}, nil, "not allowed in subcommand position"},
		// After --flag=value the next token is NOT a flag value, so a bare word there
		// is still smuggling.
		{"joined flag then word", clusterRules(), []string{"kubectl", "--context=prod", "delete"}, nil, "not allowed in subcommand position"},
		// The subword exception is field-scoped.
		{"compose subword on command", runtimeRules("docker"), []string{"docker", "compose"}, nil, "not allowed in subcommand position"},
		// ...and only as the LAST token.
		{"compose then extra word", composeRules(), []string{"docker", "compose", "up"}, nil, "not allowed in subcommand position"},

		// A chained runner WITHOUT the escape hatch.
		{"lima no hatch", runtimeRules("podman"), []string{"lima", "podman"}, nil, `"lima" is not a binary this tool runs`},
		{"microk8s no hatch", clusterRules(), []string{"microk8s", "kubectl"}, nil, `"microk8s" is not a binary this tool runs`},
		// The hatch approves the name it was given, not any other.
		{"wrong hatch", runtimeRules("podman"), []string{"lima", "podman"}, map[string]bool{"colima": true}, `"lima" is not a binary this tool runs`},
		// An escalation wrapper is refused even when something managed to put it in
		// the approvals: allowed() strips the category back out, and the message says
		// it is never permitted rather than offering an approval Approve would refuse.
		{"sudo cannot be approved", runtimeRules("podman"), []string{"sudo", "podman"}, map[string]bool{"sudo": true}, `"sudo" is a privilege-escalation wrapper (sudo), which is never permitted`},
		{"doas cannot be approved", runtimeRules("podman"), []string{"doas", "podman"}, map[string]bool{"doas": true}, `"doas" is a privilege-escalation wrapper (doas), which is never permitted`},
		{"pkexec cannot be approved", clusterRules(), []string{"pkexec", "kubectl"}, map[string]bool{"pkexec": true}, `"pkexec" is a privilege-escalation wrapper (pkexec), which is never permitted`},

		// The literal end-of-flags token.
		{"double dash", clusterRules(), []string{"kubectl", "--"}, nil, `"--" is not allowed`},
		{"double dash then word", clusterRules(), []string{"kubectl", "--", "delete"}, nil, `"--" is not allowed`},

		// The charset, one case per class.
		{"semicolon", clusterRules(), []string{"kubectl;rm"}, nil, `contains ";"`},
		{"pipe", clusterRules(), []string{"kubectl|tee"}, nil, `contains "|"`},
		{"ampersand", clusterRules(), []string{"kubectl&"}, nil, `contains "&"`},
		{"redirect out", clusterRules(), []string{"kubectl>out"}, nil, `contains ">"`},
		{"redirect in", clusterRules(), []string{"kubectl<in"}, nil, `contains "<"`},
		{"subshell open", clusterRules(), []string{"kubectl("}, nil, `contains "("`},
		{"subshell close", clusterRules(), []string{"kubectl)"}, nil, `contains ")"`},
		{"glob star", clusterRules(), []string{"kube*"}, nil, `contains "*"`},
		{"glob question", clusterRules(), []string{"kubectl?"}, nil, `contains "?"`},
		{"bracket open", clusterRules(), []string{"kubectl["}, nil, `contains "["`},
		{"bracket close", clusterRules(), []string{"kubectl]"}, nil, `contains "]"`},
		{"brace open", clusterRules(), []string{"kubectl{"}, nil, `contains "{"`},
		{"brace close", clusterRules(), []string{"kubectl}"}, nil, `contains "}"`},
		{"hash", clusterRules(), []string{"kubectl#x"}, nil, `contains "#"`},
		{"bang", clusterRules(), []string{"kubectl!"}, nil, `contains "!"`},
		{"dollar", clusterRules(), []string{"$KUBECTL"}, nil, `contains "$"`},
		{"backtick", clusterRules(), []string{"`kubectl`"}, nil, "contains \"`\""},
		{"single quote", clusterRules(), []string{"'kubectl'"}, nil, `contains "'"`},
		{"double quote", clusterRules(), []string{`"kubectl"`}, nil, `contains "\""`},
		{"backslash", clusterRules(), []string{`kube\ctl`}, nil, `contains "\\"`},
		{"control char", clusterRules(), []string{"kubectl\n--all"}, nil, "contains a control character"},
		{"NUL", clusterRules(), []string{"kube\x00ctl"}, nil, "contains a control character"},
		{"DEL", clusterRules(), []string{"kubectl\x7f"}, nil, "contains a control character"},
		// Whitespace inside ONE argument -- only reachable through the list form,
		// since the string form splits on whitespace before it gets here.
		{"embedded space", clusterRules(), []string{"C:/Program Files/kubectl.exe"}, nil, "contains whitespace inside one argument"},
		{"space in flag value", clusterRules(), []string{"kubectl", "--kubeconfig", "/tmp/my kc"}, nil, "contains whitespace inside one argument"},
		// ...and whitespace means the whole Unicode property, not just ASCII space.
		{"nbsp", clusterRules(), []string{"kubectl", "--context", "pr\u00a0od"}, nil, "contains whitespace inside one argument"},
		{"ideographic space", clusterRules(), []string{"kubectl", "--context", "pr\u3000od"}, nil, "contains whitespace inside one argument"},
		{"figure space", clusterRules(), []string{"kubectl", "--context", "pr\u2007od"}, nil, "contains whitespace inside one argument"},
		{"line separator", clusterRules(), []string{"kubectl", "--context", "pr\u2028od"}, nil, "contains whitespace inside one argument"},
		{"paragraph separator", clusterRules(), []string{"kubectl", "--context", "pr\u2029od"}, nil, "contains whitespace inside one argument"},
		{"ogham space mark", clusterRules(), []string{"kubectl", "--context", "pr\u1680od"}, nil, "contains whitespace inside one argument"},
		{"NEL", clusterRules(), []string{"kubectl", "--context", "pr\u0085od"}, nil, "contains whitespace inside one argument"},
		// Invisible formatting characters get their own refusal.
		{"zero-width space", clusterRules(), []string{"kubectl", "--context", "pr\u200bod"}, nil, "invisible formatting character"},
		{"zero-width joiner", clusterRules(), []string{"kubectl", "--context", "pr\u200dod"}, nil, "invisible formatting character"},
		{"RTL override", clusterRules(), []string{"kubectl", "--context", "pr\u202eod"}, nil, "invisible formatting character"},
		{"soft hyphen", clusterRules(), []string{"kubectl", "--context", "pr\u00adod"}, nil, "invisible formatting character"},
		// The message names the code point, since the character cannot be seen.
		{"names the code point", clusterRules(), []string{"kubectl", "--context", "pr\u200bod"}, nil, "U+200B"},
		// The charset applies to every token, not just argv[0].
		{"metachar in later token", clusterRules(), []string{"kubectl", "--context", "prod;rm"}, nil, `kubernetes.command[2]`},
		// A LEADING tilde, refused at every position but with two different
		// messages. At index 0 it is a path where a bare binary name belongs;
		// anywhere else it is a tilde nothing will expand.
		{"tilde as argv[0]", clusterRules(), []string{"~/bin/kubectl"}, nil, "argv[0] names the binary to run"},
		{"bare tilde as argv[0]", clusterRules(), []string{"~"}, nil, "a home directory cannot be part of it"},
		{"unexpanded tilde in a flag value", clusterRules(),
			[]string{"kubectl", "--kubeconfig", "~/solace/kubecontext"}, nil, "nothing downstream expands"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Check(tc.rules, tc.cmd, forced(tc.extra))
			if err == nil {
				t.Fatalf("Check(%q) accepted a value it must reject", tc.cmd)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v\n  want it to contain %q", err, tc.want)
			}
		})
	}
}

// TestEscalationFloorFoldsCaseBothWays: the deny list folds case, at both
// enforcement points, because the filesystems that resolve the name do -- Windows
// (which has shipped sudo.exe since 2024) and macOS by default, where `Sudo`
// resolves to exactly the binary `sudo` names. The allowlist stays exact, which
// the last subtest pins: folding it too would let an env file naming `KUBECTL` be
// approved by `kubectl`'s entry and then execute a different file on any
// case-sensitive filesystem.
func TestEscalationFloorFoldsCaseBothWays(t *testing.T) {
	t.Run("Approve refuses any casing", func(t *testing.T) {
		for _, name := range []string{"Sudo", "SUDO", "sUdO", "Sudo.exe", "RunAs", "PkExec"} {
			if _, err := Approve(testTool, []string{name}); err == nil {
				t.Errorf("Approve(%q) was accepted; the escalation floor must refuse every casing", name)
			}
		}
	})
	t.Run("the allowed() purge is the second belt", func(t *testing.T) {
		// Straight past Approve: the purge alone must still strip it.
		set := runtimeRules("docker").allowed(forced(map[string]bool{"Sudo": true, "SUDO": true, "docker": true}))
		for _, gone := range []string{"Sudo", "SUDO"} {
			if set[gone] {
				t.Errorf("allowed() kept %q; the purge must strip an escalation wrapper whatever its casing", gone)
			}
		}
		if !set["docker"] {
			t.Error("allowed() dropped a legitimate binary")
		}
	})
	t.Run("the allowlist itself stays case-sensitive", func(t *testing.T) {
		if err := Check(clusterRules(), []string{"KUBECTL"}, Approvals{}); err == nil {
			t.Error("Check accepted KUBECTL: the allowlist must match exactly, unlike the deny list")
		}
	})
}

// TestFlagValuePositionIsNotGuaranteed documents the guard's one acknowledged
// limit, so it is a stated property rather than a surprise found later. A token
// following a value-shaped flag is accepted as that flag's value, because argument
// arity is unknowable without modelling every flag of every allowed binary. If a
// future change narrows this, this test should fail and be rewritten as a
// rejection.
func TestFlagValuePositionIsNotGuaranteed(t *testing.T) {
	// `--tls` and `--insecure-skip-tls-verify` take NO value, so the word after
	// each reaches subcommand position in the real argv. The guard cannot tell
	// this from a genuine value.
	for _, tc := range []struct {
		rules Rules
		cmd   []string
	}{
		{runtimeRules("docker"), []string{"docker", "--tls", "rm"}},
		{clusterRules(), []string{"kubectl", "--insecure-skip-tls-verify", "delete"}},
	} {
		if err := Check(tc.rules, tc.cmd, Approvals{}); err != nil {
			t.Fatalf("the flag-value limit changed: Check(%q) now returns %v -- "+
				"if that is deliberate, replace this test with a rejection case", tc.cmd, err)
		}
	}
	// What the guard DOES still hold in that position: the whole charset.
	for _, dirty := range [][]string{
		{"docker", "--tls", "rm;curl"},
		{"docker", "--tls", "rm\u3000curl"},
		{"docker", "--tls", "r\u200bm"},
	} {
		if err := Check(runtimeRules("docker"), dirty, Approvals{}); err == nil {
			t.Errorf("Check(%q) accepted a charset violation in a flag value", dirty)
		}
	}
}

// TestFlagValueCanNameAFileToLoad: a flag value that is a relative path passes,
// although the flag makes the allowed binary load and run what the path holds --
// a docker config directory's cli-plugins, or a kubeconfig's exec credential
// command. The package doc accepts this; the test keeps the acceptance visible.
func TestFlagValueCanNameAFileToLoad(t *testing.T) {
	for _, tc := range []struct {
		rules Rules
		cmd   []string
	}{
		{runtimeRules("docker"), []string{"docker", "--config", "./d"}},
		{composeRules(), []string{"docker", "--config", "./d", "compose"}},
		{clusterRules(), []string{"kubectl", "--kubeconfig", "./kc"}},
	} {
		if err := Check(tc.rules, tc.cmd, Approvals{}); err != nil {
			t.Fatalf("the flag-value limit changed: Check(%q) now returns %v -- "+
				"if that is deliberate, replace this test with a rejection case and update the package doc", tc.cmd, err)
		}
	}
}

// TestCharsetAgreesAcrossBothYAMLForms: a command may be written as one string
// (split with strings.Fields, which is Unicode-aware) or as an explicit list
// (preserved token for token). If the guard's whitespace test were narrower than
// Fields', the list form would accept exactly what the string form splits apart.
func TestCharsetAgreesAcrossBothYAMLForms(t *testing.T) {
	spaces := []rune{' ', '\u00a0', '\u1680', '\u2000', '\u2007', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000'}
	for _, r := range spaces {
		if scalar := strings.Fields("kubectl --context a" + string(r) + "b"); len(scalar) != 4 {
			t.Errorf("U+%04X: the string form produced %d tokens (%q), want 4", r, len(scalar), scalar)
		}
		seq := []string{"kubectl", "--context", "a" + string(r) + "b"}
		if err := Check(clusterRules(), seq, Approvals{}); err == nil {
			t.Errorf("U+%04X: the list form accepted a character the string form splits on", r)
		}
	}
}

// TestGuardErrorsAreActionable holds every guard message to the one-line,
// name-the-remedy bar: the offending token quoted, and either the allowlist or the
// escape hatch named.
func TestGuardErrorsAreActionable(t *testing.T) {
	cases := [][]string{
		{"curl"},
		{"./kubectl"},
		{"kubectl", "delete"},
		{"kubectl", "--"},
		{"kubectl;rm"},
		{"sudo", "kubectl"},
	}
	for _, cmd := range cases {
		err := Check(clusterRules(), cmd, Approvals{})
		if err == nil {
			t.Fatalf("Check(%q) must fail", cmd)
		}
		msg := err.Error()
		if strings.Contains(msg, "\n") {
			t.Errorf("Check(%q) message spans lines; it must be one line: %q", cmd, msg)
		}
		if !strings.Contains(msg, "kubernetes.command") {
			t.Errorf("Check(%q) message does not name the field: %q", cmd, msg)
		}
		if !strings.Contains(msg, "--allow-command") && !strings.Contains(msg, "kubectl, oc") &&
			!strings.Contains(msg, "remove it") && !strings.Contains(msg, "let PATH resolve it") &&
			!strings.Contains(msg, "forward slashes") {
			t.Errorf("Check(%q) message names no remedy: %q", cmd, msg)
		}
	}
}

func TestAllowCommandsAccepts(t *testing.T) {
	a, err := Approve(testTool, []string{"lima", "microk8s", "colima.exe"})
	if err != nil {
		t.Fatalf("Approve = %v, want accepted", err)
	}
	// Repeatable: each value extends the same set, and .exe folds to one entry.
	for _, want := range []string{"lima", "microk8s", "colima"} {
		if !a.names[want] {
			t.Errorf("Approve did not record %q (got %v)", want, a.names)
		}
	}
}

// TestAllowCommandsRejects: a bad value is a usage error, and in particular the
// hatch cannot be used to smuggle back the path form argv[0] refuses.
func TestAllowCommandsRejects(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  string
	}{
		{"relative path", "./evil", "must be a bare binary name, not a path"},
		{"absolute path", "/usr/bin/evil", "must be a bare binary name, not a path"},
		{"backslash path", `C:\evil.exe`, "not allowed in a command token"},
		{"metacharacter", "lima;rm", "not allowed in a command token"},
		{"whitespace", "lima nerdctl", "whitespace inside one argument"},
		{"control character", "lima\n", "control character"},
		{"empty", "", "empty argument"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Approve(testTool, []string{tc.value})
			if err == nil {
				t.Fatalf("Approve(%q) accepted a value it must reject", tc.value)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to contain %q", err, tc.want)
			}
			if !strings.Contains(err.Error(), "--allow-command") {
				t.Errorf("error = %v, want it to name the flag the value came from", err)
			}
		})
	}
}

// TestAllowCommandsRejectsEscalation: the escape hatch has a floor. Granting an
// escalation wrapper once on the command line would elevate every command the tool
// issues for the whole life of that env file, while `sudo <tool> ...` elevates one
// invocation the operator chose, at the moment they chose it.
func TestAllowCommandsRejectsEscalation(t *testing.T) {
	for _, name := range []string{"sudo", "sudoedit", "doas", "su", "pkexec", "run0", "systemd-run", "machinectl",
		"setpriv", "capsh", "unshare", "nsenter", "runas", "gsudo", "sudo.exe", "gsudo.EXE"} {
		t.Run(name, func(t *testing.T) {
			a, err := Approve(testTool, []string{name})
			if err == nil {
				t.Fatalf("Approve(%q) approved a privilege-escalation wrapper", name)
			}
			if !strings.Contains(err.Error(), "is never permitted") {
				t.Errorf("error = %v, want it to say the value is never permitted", err)
			}
			// The message must point at the supported alternative, or the rule reads
			// as "you cannot do this" rather than "do it the other way".
			want := escalator(name) + " " + testTool + " ..."
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error = %v, want it to name the %q alternative", err, want)
			}
			if a.names[base(name)] {
				t.Errorf("Approve(%q) recorded the name despite failing", name)
			}
		})
	}
}

// TestEscalationCannotBeAllowedByAnyRoute is the structural backstop: an escalation
// wrapper listed in Rules.Allowed, or put into Approvals some other way, is
// stripped back out. The rule does not depend on Approve being the only door.
func TestEscalationCannotBeAllowedByAnyRoute(t *testing.T) {
	approvals := forced(map[string]bool{"sudo": true, "doas": true, "pkexec": true})
	listed := runtimeRules("podman")
	listed.Allowed = append(listed.Allowed, "sudo", "doas")
	for _, cmd := range [][]string{
		{"sudo", "podman"},
		{"doas", "podman"},
		{"podman", "sudo"},
	} {
		if err := Check(runtimeRules("podman"), cmd, approvals); err == nil {
			t.Errorf("Check(%q) accepted an escalation wrapper from forced approvals", cmd)
		}
		if err := Check(listed, cmd, Approvals{}); err == nil {
			t.Errorf("Check(%q) accepted an escalation wrapper listed in Rules.Allowed", cmd)
		}
	}
	// The floor is specific to escalation: a legitimate wrapper in the same forced
	// set still works, so this is a deny list and not a broken allow set.
	if err := Check(runtimeRules("podman"), []string{"lima", "podman"}, forced(map[string]bool{"lima": true})); err != nil {
		t.Errorf("the deny list broke a legitimate approval: %v", err)
	}
}

// TestAllowedBinaryIsNotGloballyAllowed: an approval widens the allowlist for the
// run that holds it, not for the process -- a second Approvals value in the same
// run does not inherit it.
func TestAllowedBinaryIsNotGloballyAllowed(t *testing.T) {
	approved, err := Approve(testTool, []string{"lima"})
	if err != nil {
		t.Fatalf("Approve = %v", err)
	}
	cmd := []string{"lima", "podman"}
	if err := Check(runtimeRules("podman"), cmd, approved); err != nil {
		t.Fatalf("approved run rejected `lima podman`: %v", err)
	}
	other, err := Approve(testTool, nil)
	if err != nil {
		t.Fatalf("Approve(nil) = %v", err)
	}
	if err := Check(runtimeRules("podman"), cmd, other); err == nil {
		t.Error("an --allow-command approval leaked into an Approvals value that never received it")
	}
}

// TestCheckDeployCommandAcceptReject pins solmq-conn-util's accept/reject matrix,
// written as one string and split the way it splits it. One row changed with the
// shared guard: `sudo podman` with sudo approved is now refused, because no
// approval can carry an escalation wrapper.
func TestCheckDeployCommandAcceptReject(t *testing.T) {
	cases := []struct {
		name     string
		platform string
		cmd      string
		extra    []string
		wantErr  bool
	}{
		{"kubectl", "kubernetes", "kubectl", nil, false},
		{"oc", "kubernetes", "oc", nil, false},
		{"kubectl with context and namespace flags", "kubernetes", "kubectl --context prod -n ns", nil, false},
		{"docker with context flag", "docker", "docker --context foo", nil, false},
		{"podman", "podman", "podman", nil, false},
		{"kubectl.exe (Windows-authored config)", "kubernetes", "kubectl.exe", nil, false},
		{"sudo podman, sudo approval refused", "podman", "sudo podman", []string{"sudo"}, true},

		{"curl is not on any platform allowlist", "kubernetes", "curl", nil, true},
		{"absolute path argv0 rejected", "kubernetes", "/tmp/evil", nil, true},
		{"relative path argv0 rejected", "kubernetes", "./kubectl", nil, true},
		{"bare positional argument after argv0", "kubernetes", "kubectl delete ns prod", nil, true},
		{"sudo podman rejected without extraAllowed", "podman", "sudo podman", nil, true},
		{"bare end-of-flags marker as argv0", "kubernetes", "--", nil, true},
		{"empty command", "kubernetes", "", nil, true},
	}
	rules := map[string]Rules{"kubernetes": clusterRules(), "docker": runtimeRules("docker"), "podman": runtimeRules("podman")}
	for _, c := range cases {
		a, err := Approve(testTool, c.extra)
		if err == nil {
			err = Check(rules[c.platform], strings.Fields(c.cmd), a)
		}
		if (err != nil) != c.wantErr {
			t.Errorf("%s: %q (extra=%v) = %v, want error=%v", c.name, c.cmd, c.extra, err, c.wantErr)
		}
	}
}

// TestCheckDeployCommandEndOfFlagsMarkerMidCommand exercises the "--" rule at a
// non-argv0 position: kubectl is allowed, and "--" as its first argument is still
// refused as an end-of-flags marker.
func TestCheckDeployCommandEndOfFlagsMarkerMidCommand(t *testing.T) {
	err := Check(clusterRules(), strings.Fields("kubectl --"), Approvals{})
	if err == nil || !strings.Contains(err.Error(), `kubernetes.command[1] = "--" is not allowed: end-of-flags`) {
		t.Fatalf("want end-of-flags rejection, got %v", err)
	}
}

// TestTokenCharsetRows carries solmq-conn-util's safe-token rows, each placed in a
// flag-value position, where only the charset applies. One row changed: an
// embedded tilde is now accepted, as an 8.3 short name.
func TestTokenCharsetRows(t *testing.T) {
	for tok, want := range map[string]bool{
		"kubectl": true, "docker": true, "--context=prod": true,
		"/usr/local/bin/kubectl": true, "--namespace=solace-connectors": true, "--server=https://api.k8s.local:6443": true,
		"a~b": true,
		"a b": false, "a;b": false, "a$b": false, "a`b": false, `a\b`: false, "a'b": false, `a"b`: false,
		"a|b": false, "a&b": false, "a>b": false, "a<b": false, "a(b)": false, "a*b": false, "a?b": false, "a#b": false, "a!b": false,
		"a\x00b": false, "a\x7fb": false, "a[b": false, "a]b": false, "a{b": false, "a}b": false,
	} {
		err := Check(clusterRules(), []string{"kubectl", "--value", tok}, Approvals{})
		if (err == nil) != want {
			t.Errorf("token %q: Check = %v, want accepted=%v", tok, err, want)
		}
	}
}

// TestCheckRejectsC1Controls: C1 controls are refused like C0, and the message
// names the code point. NEL stays a whitespace refusal.
func TestCheckRejectsC1Controls(t *testing.T) {
	for _, r := range []rune{0x80, 0x90, 0x9b, 0x9f} {
		err := Check(clusterRules(), []string{"kubectl", "--context", "pr" + string(r) + "od"}, Approvals{})
		if err == nil {
			t.Fatalf("U+%04X: Check accepted a C1 control", r)
		}
		want := "contains a control character (" + hex2(r) + ")"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("U+%04X: error = %v, want it to contain %q", r, err, want)
		}
	}
	err := Check(clusterRules(), []string{"kubectl", "--context", "pr\u0085od"}, Approvals{})
	if err == nil || !strings.Contains(err.Error(), "whitespace") {
		t.Errorf("NEL: error = %v, want the whitespace refusal", err)
	}
}

// hex2 renders a rune below 0x100 as two hex digits, the way the message does.
func hex2(r rune) string {
	const digits = "0123456789abcdef"
	return "0x" + string(digits[r>>4]) + string(digits[r&0xf])
}

// TestCheckRejectsInvalidUTF8: a byte that is not UTF-8 would decode as U+FFFD and
// pass every rune check, yet 0x9b alone is a CSI to an 8-bit terminal. A literal
// U+FFFD is valid text and is accepted.
func TestCheckRejectsInvalidUTF8(t *testing.T) {
	for _, tok := range []string{"pr\x9bod", "\xff", "ok\xc3"} {
		err := Check(clusterRules(), []string{"kubectl", "--context", tok}, Approvals{})
		if err == nil || !strings.Contains(err.Error(), "is not valid UTF-8") {
			t.Errorf("token %q: error = %v, want the UTF-8 refusal", tok, err)
		}
	}
	if err := Check(clusterRules(), []string{"kubectl", "--context", "pr\uFFFDod"}, Approvals{}); err != nil {
		t.Errorf("a literal U+FFFD was refused: %v", err)
	}
}

// TestEscalatorRefusedInEveryBarePosition: an escalation wrapper is named as never
// permitted wherever it would run -- argv[0], or a bare token after an approved
// runner. As a flag's value it runs nothing, and that position is never checked.
func TestEscalatorRefusedInEveryBarePosition(t *testing.T) {
	lima := forced(map[string]bool{"lima": true})
	for _, tc := range []struct {
		cmd  []string
		want string
	}{
		{[]string{"sudo", "podman"}, `podman.command[0] = "sudo" is a privilege-escalation wrapper`},
		{[]string{"lima", "sudo", "podman"}, `podman.command[1] = "sudo" is a privilege-escalation wrapper`},
		{[]string{"lima", "SuDo.exe"}, `podman.command[1] = "SuDo.exe" is a privilege-escalation wrapper (sudo)`},
	} {
		err := Check(runtimeRules("podman"), tc.cmd, lima)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("Check(%q) = %v, want it to contain %q", tc.cmd, err, tc.want)
			continue
		}
		if strings.Contains(err.Error(), "--allow-command sudo") {
			t.Errorf("Check(%q) offers an approval Approve would refuse: %v", tc.cmd, err)
		}
	}
	if err := Check(clusterRules(), []string{"kubectl", "--as", "sudo"}, Approvals{}); err != nil {
		t.Errorf("a flag value named sudo was refused: %v", err)
	}
}

// TestEscalatorListCoversTheCategory pins the whole deny list, through both doors,
// in any case and with .exe. A name outside the category is unaffected.
func TestEscalatorListCoversTheCategory(t *testing.T) {
	names := []string{
		"sudo", "sudoedit", "doas", "su", "pkexec", "run0", "runuser", "sg",
		"systemd-run", "machinectl", "setpriv", "capsh", "unshare", "nsenter", "chroot",
		"gosu", "su-exec", "setuidgid", "s6-setuidgid", "chpst", "runas", "gsudo",
	}
	if len(names) != 22 {
		t.Fatalf("the list has %d names, want 22", len(names))
	}
	for _, name := range names {
		for _, spelling := range []string{name, strings.ToUpper(name), name + ".exe"} {
			if got := escalator(spelling); got != name {
				t.Errorf("escalator(%q) = %q, want %q", spelling, got, name)
			}
			if _, err := Approve(testTool, []string{spelling}); err == nil || !strings.Contains(err.Error(), "never permitted") {
				t.Errorf("Approve(%q) = %v, want never permitted", spelling, err)
			}
			err := Check(clusterRules(), []string{spelling, "kubectl"}, forced(map[string]bool{spelling: true}))
			if err == nil || !strings.Contains(err.Error(), "never permitted") {
				t.Errorf("Check(%q) = %v, want never permitted", spelling, err)
			}
		}
	}
	for _, name := range []string{"sudoku", "lima", "sgdisk"} {
		if escalator(name) != "" {
			t.Errorf("escalator(%q) matched a binary outside the category", name)
		}
	}
}

// TestExeSuffixFoldsAnyCase: the .exe suffix folds in any case; the name before it
// never does, and a bare ".exe" is not a suffix on anything.
func TestExeSuffixFoldsAnyCase(t *testing.T) {
	for _, tok := range []string{"kubectl.exe", "kubectl.EXE", "kubectl.Exe", "kubectl.eXe"} {
		if err := Check(clusterRules(), []string{tok}, Approvals{}); err != nil {
			t.Errorf("Check(%q) = %v, want accepted", tok, err)
		}
	}
	if err := Check(clusterRules(), []string{"KUBECTL.exe"}, Approvals{}); err == nil {
		t.Error("Check accepted KUBECTL.exe: only the suffix folds")
	}
	if got := base(".exe"); got != ".exe" {
		t.Errorf("base(%q) = %q, want it unchanged", ".exe", got)
	}
	a, err := Approve(testTool, []string{"Colima.EXE"})
	if err != nil || !a.names["Colima"] {
		t.Errorf("Approve(Colima.EXE) = %v, %v; want Colima recorded", a.names, err)
	}
}

// TestSubwordOnlyLast: the subword is a bare word everywhere except last place in
// the field that declares it.
func TestSubwordOnlyLast(t *testing.T) {
	for _, cmd := range [][]string{
		{"compose"},
		{"docker", "compose", "compose"},
		{"docker", "compose", "--profile"},
	} {
		if err := Check(composeRules(), cmd, Approvals{}); err == nil {
			t.Errorf("Check(%q) accepted the subword outside last place", cmd)
		}
	}
	if err := Check(composeRules(), []string{"docker", "--debug", "compose"}, Approvals{}); err != nil {
		t.Errorf("Check(docker --debug compose) = %v, want accepted", err)
	}
}

// TestRulesWiringErrors: a Rules value with no field or no allowlist is a wiring
// mistake, reported before anything about the command.
func TestRulesWiringErrors(t *testing.T) {
	for _, tc := range []struct {
		rules Rules
		want  string
	}{
		{Rules{}, "execguard: Rules.Field is empty"},
		{Rules{Field: "x.command"}, "execguard: Rules.Allowed is empty for x.command"},
		{Rules{Field: "x.command", Allowed: []string{"docker"}, Subword: "sudo"}, `Rules.Subword "sudo" for x.command`},
		{Rules{Field: "x.command", Allowed: []string{"docker"}, Subword: "Run0.EXE"}, `Rules.Subword "Run0.EXE"`},
		{Rules{Field: "x.command", Allowed: []string{"docker"}, Subword: "bin/compose"}, `Rules.Subword "bin/compose"`},
		{Rules{Field: "x.command", Allowed: []string{"docker"}, Subword: "com;pose"}, `Rules.Subword "com;pose"`},
	} {
		err := Check(tc.rules, nil, Approvals{})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("Check(%+v) = %v, want it to contain %q", tc.rules, err, tc.want)
		}
	}
}

// TestApproveWiring: an empty tool name is a wiring mistake; no values approve
// nothing; a bad value anywhere leaves nothing recorded; repeats fold.
func TestApproveWiring(t *testing.T) {
	if _, err := Approve("", []string{"lima"}); err == nil || !strings.Contains(err.Error(), "tool is empty") {
		t.Errorf("Approve with no tool = %v, want the wiring error", err)
	}
	a, err := Approve(testTool, nil)
	if err != nil || a.names != nil {
		t.Errorf("Approve(nil) = %v, %v; want the zero value", a.names, err)
	}
	a, err = Approve(testTool, []string{"lima", "sudo"})
	if err == nil || a.names != nil {
		t.Errorf("Approve(lima, sudo) = %v, %v; want an error and nothing recorded", a.names, err)
	}
	a, err = Approve(testTool, []string{"lima", "lima.exe"})
	if err != nil || len(a.names) != 1 || !a.names["lima"] {
		t.Errorf("Approve(lima, lima.exe) = %v, %v; want one entry", a.names, err)
	}
	var zero Approvals
	if err := Check(runtimeRules("podman"), []string{"lima", "podman"}, zero); err == nil {
		t.Error("the zero Approvals approved lima")
	}
}

// TestScopeOmitted: without a Scope the allowlist message drops the "on" part.
func TestScopeOmitted(t *testing.T) {
	err := Check(Rules{Field: "tool.command", Allowed: []string{"kubectl"}}, []string{"curl"}, Approvals{})
	if err == nil || !strings.Contains(err.Error(), "allowed are kubectl --") {
		t.Errorf("error = %v, want %q", err, "allowed are kubectl --")
	}
	err = Check(clusterRules(), []string{"curl"}, Approvals{})
	if err == nil || !strings.Contains(err.Error(), "allowed on k8s are kubectl, oc") {
		t.Errorf("error = %v, want %q", err, "allowed on k8s are kubectl, oc")
	}
}

// TestExportedCharset pins the exported charset a tool's host-path gate derives
// from: every unsafe character is refused in a token, tilde is not in the set,
// and the two predicates draw their lines where the docs say.
func TestExportedCharset(t *testing.T) {
	if strings.Contains(UnsafeChars, "~") {
		t.Error("UnsafeChars contains '~'; the tilde rule is positional")
	}
	for _, c := range UnsafeChars {
		if err := Check(clusterRules(), []string{"kubectl", "--v", "a" + string(c) + "b"}, Approvals{}); err == nil {
			t.Errorf("%q in a token was accepted", c)
		}
	}
	if PathSeparators != `/\` {
		t.Errorf("PathSeparators = %q", PathSeparators)
	}
	for r, want := range map[rune]bool{0x00: true, 0x1f: true, 0x7f: true, 0x80: true, 0x9f: true,
		0x85: false, 0x20: false, 'a': false, 0xa0: false} {
		if got := IsControl(r); got != want {
			t.Errorf("IsControl(U+%04X) = %v, want %v", r, got, want)
		}
	}
	for r, want := range map[rune]bool{0x200b: true, 0xfeff: true, 0x202e: true, 0xad: true,
		'a': false, 0x200a: false, 0x85: false} {
		if got := IsInvisible(r); got != want {
			t.Errorf("IsInvisible(U+%04X) = %v, want %v", r, got, want)
		}
	}
	if ApproveFlag != "--allow-command" {
		t.Errorf("ApproveFlag = %q; both tools register this exact name", ApproveFlag)
	}
}
