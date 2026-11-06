package containercli

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestInfoArgv(t *testing.T) {
	if got, want := Info([]string{"docker", "--context", "foo"}).Argv, []string{"docker", "--context", "foo", "info"}; !reflect.DeepEqual(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
}

// TestInfoErrorPerEngine carries both tools' preflight cases (solace-util's
// preflight_test.go, solmq-conn-util's TestPreflightFailureWrapsPlatformHint): the
// cause is wrapped, the engine and the command are named, and the hint is the next
// step for that engine.
func TestInfoErrorPerEngine(t *testing.T) {
	cause := errors.New("exit status 1")
	for _, c := range []struct {
		engine   Engine
		rootless bool
		want     []string
	}{
		{Docker, false, []string{"cannot talk to the Docker engine", `("docker" info failed)`, "start the daemon",
			"docker context use", "`docker` group"}},
		{Podman, false, []string{"cannot talk to the Podman engine", "sudo systemctl start podman.socket",
			"podman machine start"}},
		{Podman, true, []string{"systemctl --user start podman.socket", "loginctl enable-linger",
			"do NOT use sudo for podman itself"}},
		{"nerdctl", false, []string{"cannot talk to the nerdctl engine", "check that nerdctl is installed"}},
		{"", false, []string{"cannot talk to the container engine"}},
	} {
		base := []string{string(c.engine)}
		if c.engine == "" {
			base = []string{"docker"}
		}
		err := InfoError(c.engine, base, c.rootless, "", cause)
		if !errors.Is(err, cause) {
			t.Errorf("%s: the cause must be wrapped, got %v", c.engine, err)
		}
		for _, w := range c.want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("%s rootless=%v: err = %v, want it to name %q", c.engine, c.rootless, err, w)
			}
		}
	}
	// The rootful podman hint must never send a rootless operator to sudo, and the
	// rootless one must say not to.
	if strings.Contains(EngineHint(Podman, true), "sudo systemctl") {
		t.Errorf("the rootless hint must not start the rootful engine: %s", EngineHint(Podman, true))
	}
}

// TestInfoErrorCarriesTheEngineOutput: a caller that captured the engine's answer
// passes it on, and the hint still comes last.
func TestInfoErrorCarriesTheEngineOutput(t *testing.T) {
	err := InfoError(Docker, []string{"docker"}, false, "  Cannot connect to the Docker daemon\n", errors.New("exit status 1"))
	want := "cannot talk to the Docker engine (\"docker\" info failed): exit status 1\n" +
		"Cannot connect to the Docker daemon\n  start the daemon"
	if !strings.HasPrefix(err.Error(), want) {
		t.Errorf("err =\n%s\nwant it to start\n%s", err, want)
	}
}

// TestComposeUpAndDown carries solmq-conn-util's TestDockerUpAndDown: the file on
// stdin, the secrets in the child's env, nothing on disk.
func TestComposeUpAndDown(t *testing.T) {
	compose := []string{"docker", "compose"}
	file := []byte("name: \"eg\"\nservices: {}\n")
	env := []string{"SOLACE_ADMIN_PASSWORD=s3cr3t"}
	head := []string{"docker", "compose", "-f", "-", "--project-directory", "/srv/spec"}
	for _, c := range []struct {
		name string
		got  []string
		tail []string
	}{
		{"up", ComposeUp(compose, file, "/srv/spec", env, false).Argv, []string{"up", "-d"}},
		{"up recreate", ComposeUp(compose, file, "/srv/spec", env, true).Argv, []string{"up", "-d", "--force-recreate"}},
		{"down", ComposeDown(compose, file, "/srv/spec", env).Argv, []string{"down"}},
	} {
		if want := append(append([]string(nil), head...), c.tail...); !reflect.DeepEqual(c.got, want) {
			t.Errorf("%s argv = %v, want %v", c.name, c.got, want)
		}
	}
	for _, c := range []struct {
		name  string
		stdin []byte
		env   []string
	}{
		{"up", ComposeUp(compose, file, "/srv/spec", env, true).Stdin, ComposeUp(compose, file, "/srv/spec", env, true).Env},
		{"down", ComposeDown(compose, file, "/srv/spec", env).Stdin, ComposeDown(compose, file, "/srv/spec", env).Env},
	} {
		if string(c.stdin) != string(file) {
			t.Errorf("%s stdin = %q, want the file", c.name, c.stdin)
		}
		if !reflect.DeepEqual(c.env, env) {
			t.Errorf("%s env = %v, want %v", c.name, c.env, env)
		}
	}
	if c := ComposeUp([]string{"docker-compose"}, file, "/d", nil, false); c.Argv[0] != "docker-compose" || c.Env != nil {
		t.Errorf("a standalone compose leads the argv and no env stays nil: %+v", c)
	}
}

// TestLifecycleArgv carries solmq-conn-util's TestLifecycleArgv: plain engine verbs
// on the container, the base's own flags first.
func TestLifecycleArgv(t *testing.T) {
	base := []string{"docker", "--context", "prod"}
	for _, c := range []struct {
		got  []string
		verb string
	}{
		{Start(base, "connector").Argv, "start"},
		{Stop(base, "connector").Argv, "stop"},
		{Restart(base, "connector").Argv, "restart"},
		{Remove(base, "connector").Argv, "rm"},
	} {
		if want := []string{"docker", "--context", "prod", c.verb, "connector"}; !reflect.DeepEqual(c.got, want) {
			t.Errorf("argv = %v, want %v", c.got, want)
		}
	}
}

// TestListArgvAndTabs pins real tab bytes in the template, not a backslash and a t.
func TestListArgvAndTabs(t *testing.T) {
	all := List([]string{"podman"}, false).Argv
	if want := []string{"podman", "ps", "--all", "--no-trunc", "--format", "{{.Names}}\x09{{.Image}}\x09{{.Status}}"}; !reflect.DeepEqual(all, want) {
		t.Errorf("argv = %q, want %q", all, want)
	}
	running := List([]string{"docker"}, true).Argv
	if want := []string{"docker", "ps", "--all", "--no-trunc", "--filter", "status=running", "--format",
		"{{.Names}}\x09{{.Image}}\x09{{.Status}}"}; !reflect.DeepEqual(running, want) {
		t.Errorf("running argv = %q, want %q", running, want)
	}
}

func TestParseList(t *testing.T) {
	out := "solace\tsolace/solace-pubsub-standard:10.10\tUp 5 minutes\r\n" +
		"\n" +
		"short\tonly-two\n" +
		"  edge \t solace/solace-pubsub-standard:10.9 \t Exited (0) 2 days ago \n"
	want := []Row{
		{"solace", "solace/solace-pubsub-standard:10.10", "Up 5 minutes"},
		{"edge", "solace/solace-pubsub-standard:10.9", "Exited (0) 2 days ago"},
	}
	if got := ParseList(out); !reflect.DeepEqual(got, want) {
		t.Errorf("ParseList = %+v, want %+v", got, want)
	}
	if got := ParseList(""); got != nil {
		t.Errorf("ParseList(\"\") = %+v, want none", got)
	}
}

// TestHasNameIsExact carries solace-util's TestContainerRunningMatchesNameExactly
// listings: a sibling or a prefix is not the container.
func TestHasNameIsExact(t *testing.T) {
	row := func(name string) string { return name + "\timg\tUp" }
	for _, c := range []struct {
		name    string
		listing string
		want    bool
	}{
		{"exact", row("solace"), true},
		{"among others", row("other") + "\n" + row("solace") + "\n" + row("x"), true},
		{"sibling only", row("solace-edge"), false},
		{"prefix only", row("sol"), false},
		{"empty", "", false},
		{"padded", row("  solace  "), true},
	} {
		if got := HasName(ParseList(c.listing), "solace"); got != c.want {
			t.Errorf("%s: HasName = %v, want %v", c.name, got, c.want)
		}
	}
	if HasName(ParseList("\timg\tUp"), "") {
		t.Error("an empty name matches nothing")
	}
}

// TestByImage carries solmq-conn-util's TestEngineNamesByImage cases, now over
// three-column rows.
func TestByImage(t *testing.T) {
	rows := ParseList("conn-a\tsolace/solace-pubsub-connector-ibmmq:2.14\tUp\n" +
		"web\tnginx:1.27\tUp\n" +
		"conn-b\tregistry.local/solace-pubsub-connector-ibmmq@sha256:ab\tExited (1)\n")
	got := ByImage(rows, "solace-pubsub-connector-ibmmq")
	if len(got) != 2 || got[0].Name != "conn-a" || got[1].Name != "conn-b" {
		t.Errorf("ByImage = %+v, want conn-a then conn-b", got)
	}
	if got := ByImage(rows, "no-such-image"); got != nil {
		t.Errorf("no match = %+v, want none", got)
	}
}

func TestValidName(t *testing.T) {
	for name, want := range map[string]bool{
		"solace": true, "s": true, "sol_pod.1-a": true, "9lives": true,
		"": false, "-x": false, "_x": false, ".x": false, "a b": false, "a/b": false, "a;b": false,
		"a\x1b": false, "caf\xc3\xa9": false,
	} {
		if got := ValidName(name); got != want {
			t.Errorf("ValidName(%q) = %v, want %v", name, got, want)
		}
	}
}

// TestExec carries solmq-conn-util's TestExecArgvPerPlatform engine rows: the
// engine's flags first, then the container, and no `--` or -c.
func TestExec(t *testing.T) {
	for _, c := range []struct {
		o    ExecOptions
		want []string
	}{
		{ExecOptions{}, []string{"docker", "exec", "pod-0", "show", "version"}},
		{ExecOptions{Stdin: true}, []string{"docker", "exec", "-i", "pod-0", "show", "version"}},
		{ExecOptions{Stdin: true, TTY: true}, []string{"docker", "exec", "-i", "-t", "pod-0", "show", "version"}},
	} {
		got, err := Exec([]string{"docker"}, "pod-0", c.o, "show", "version")
		if err != nil || !reflect.DeepEqual(got.Argv, c.want) || got.Stdin != nil {
			t.Errorf("Exec(%+v) = %+v, %v; want argv %v and no stdin", c.o, got, err, c.want)
		}
	}
}

func TestExecRefuses(t *testing.T) {
	for _, c := range []struct {
		name string
		argv []string
		o    ExecOptions
		want string
	}{
		{"", []string{"sh"}, ExecOptions{}, "container name"},
		{"pod-0", nil, ExecOptions{}, "command to run"},
		{"pod-0", []string{"sh"}, ExecOptions{TTY: true}, "a tty needs stdin"},
	} {
		if _, err := Exec([]string{"docker"}, c.name, c.o, c.argv...); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("Exec(%q, %v, %+v) err = %v, want %q", c.name, c.argv, c.o, err, c.want)
		}
	}
}

// TestUpload carries solace-util's TestTransportUploadQuotesDest and
// solmq-conn-util's TestInstallScriptArgv: the destination is one quoted word, the
// directory comes from it, and the body travels on stdin.
func TestUpload(t *testing.T) {
	c, err := Upload([]string{"podman"}, "sol-pod", "/tmp/a'b", []byte("body"), false)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"podman", "exec", "-i", "sol-pod", "sh", "-c", `cat > '/tmp/a'\''b'`}; !reflect.DeepEqual(c.Argv, want) {
		t.Errorf("argv = %q, want %q", c.Argv, want)
	}
	if string(c.Stdin) != "body" {
		t.Errorf("stdin = %q, want the body", c.Stdin)
	}
	c, err = Upload([]string{"docker"}, "solmq-connector", "/tmp/solmq/status.sh", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := c.Argv[len(c.Argv)-1], `mkdir -p '/tmp/solmq' && cat > '/tmp/solmq/status.sh'`; got != want {
		t.Errorf("script = %q, want %q", got, want)
	}
	if c.Stdin == nil {
		t.Error("an empty body is still stdin, so the file is created empty")
	}
}

func TestUploadRefuses(t *testing.T) {
	for _, c := range [][2]string{{"", "/tmp/x"}, {"sol-pod", ""}} {
		if _, err := Upload([]string{"podman"}, c[0], c[1], []byte("x"), false); err == nil {
			t.Errorf("Upload(%q, %q) must be refused", c[0], c[1])
		}
	}
}

// TestLogsFlagOrder carries solmq-conn-util's TestLogsArgvPerPlatform engine rows:
// flags in a fixed order before the container, and an empty value leaves its flag
// out.
func TestLogsFlagOrder(t *testing.T) {
	got := Logs([]string{"podman"}, "solmq-connector", LogOptions{Follow: true, Timestamps: true, Tail: "0", Since: "10m0s"}).Argv
	if want := []string{"podman", "logs", "-f", "--timestamps", "--tail", "0", "--since", "10m0s", "solmq-connector"}; !reflect.DeepEqual(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
	if got := Logs([]string{"docker"}, "solace", LogOptions{}).Argv; !reflect.DeepEqual(got, []string{"docker", "logs", "solace"}) {
		t.Errorf("no options argv = %v", got)
	}
}

func TestNoPrevious(t *testing.T) {
	err := NoPrevious(Podman)
	for _, want := range []string{"--previous", "kubernetes", "Podman", "leave it off"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to name %q", err, want)
		}
	}
}

// TestBuildersCopyInputs: no builder writes through a slice with spare capacity, and
// a caller's later change to the file or env does not reach the command.
func TestBuildersCopyInputs(t *testing.T) {
	base := make([]string, 1, 8)
	base[0] = "docker"
	a, b := Start(base, "x"), Logs(base, "y", LogOptions{})
	if a.Argv[1] != "start" || b.Argv[1] != "logs" || base[:2][1] != "" {
		t.Errorf("argvs share the base: %v / %v / %v", a.Argv, b.Argv, base[:2])
	}
	file, env := []byte("services: {}"), []string{"K=v"}
	c := ComposeUp([]string{"docker", "compose"}, file, "/d", env, false)
	file[0], env[0] = 'X', "K=changed"
	if string(c.Stdin) != "services: {}" || c.Env[0] != "K=v" {
		t.Errorf("the command must hold copies: stdin %q env %v", c.Stdin, c.Env)
	}
	if c := ComposeDown([]string{"docker", "compose"}, nil, "/d", nil); c.Stdin == nil {
		t.Error("an empty file is still stdin")
	}
}

// TestComposeDigest: the digest is a property of the deploy, not of how the tool
// happened to order its environment; any change to the file or to one value moves
// it; entries cannot run into each other or into the file; and the caller's slice
// keeps its order.
func TestComposeDigest(t *testing.T) {
	file := []byte("name: \"solace\"\n")
	env := []string{"B=2", "A=1"}
	d := ComposeDigest(file, env)
	if len(d) != 64 || strings.Trim(d, "0123456789abcdef") != "" {
		t.Fatalf("digest = %q, want 64 lower-case hex digits", d)
	}
	if env[0] != "B=2" {
		t.Errorf("env = %v; the caller's slice must not be sorted", env)
	}
	if got := ComposeDigest(file, []string{"A=1", "B=2"}); got != d {
		t.Error("the digest must not depend on the environment's order")
	}
	for _, c := range []struct {
		name string
		file []byte
		env  []string
	}{
		{"file", []byte("name: \"solace2\"\n"), env},
		{"value", file, []string{"A=1", "B=3"}},
		{"extra entry", file, []string{"A=1", "B=2", "C="}},
		{"entries merged", file, []string{"A=1B=2"}},
		{"entry moved into the file", append(append([]byte(nil), file...), "A=1"...), []string{"B=2"}},
		{"no env", file, nil},
	} {
		if ComposeDigest(c.file, c.env) == d {
			t.Errorf("%s: the digest did not change", c.name)
		}
	}
	if ComposeDigest(nil, []string{"a=b", "c"}) == ComposeDigest(nil, []string{"a=bc"}) {
		t.Error("entry boundaries must be part of the digest")
	}
}
