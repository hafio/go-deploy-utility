package inspect

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestContainersArgv(t *testing.T) {
	c, ok := Containers([]string{"docker", "--context", "prod"}, "a", "b")
	if !ok {
		t.Fatal("names given, so there is a command to run")
	}
	if want := []string{"docker", "--context", "prod", "container", "inspect", "a", "b"}; !reflect.DeepEqual(c.Argv, want) {
		t.Errorf("argv = %v, want %v", c.Argv, want)
	}
	if c, ok := Containers([]string{"docker"}); ok || c.Argv != nil {
		t.Errorf("no names = (%+v, %v), want nothing to run", c, ok)
	}
	base := make([]string, 1, 8)
	base[0] = "podman"
	Containers(base, "x")
	if base[:2][1] != "" {
		t.Errorf("the base was written through: %v", base[:2])
	}
}

// dockerInspect is a docker answer (solmq-conn-util's dockerInspectJSON, with a
// mount mode, a second network and an environment added): the leading slash on the
// name, and the health block spelled Health.
const dockerInspect = `[
  {
    "Name": "/solmq-connector",
    "Created": "2026-08-18T04:10:00Z",
    "State": {
      "Status": "running", "Running": true, "ExitCode": 0,
      "StartedAt": "2026-08-18T04:12:07Z", "FinishedAt": "0001-01-01T00:00:00Z",
      "Health": {"Status": "healthy", "FailingStreak": 0}
    },
    "RestartCount": 2,
    "Image": "sha256:localimageid",
    "Config": {
      "Image": "solace/solace-pubsub-connector-ibmmq:2.14.1",
      "Labels": {"com.docker.compose.project": "eg"},
      "Env": ["SOLACE_ADMIN_PASSWORD=hunter2-secret"]
    },
    "HostConfig": {"NanoCpus": 1000000000, "Memory": 1073741824},
    "Mounts": [
      {"Type": "bind", "Source": "/srv/solmq/application.yml", "Destination": "/app/external/spring/config/application.yml", "RW": false},
      {"Type": "volume", "Name": "solmq-libs", "Destination": "/app/external/libs", "RW": true}
    ],
    "NetworkSettings": {"Networks": {"eg_default": {}, "backend": {}}}
  }
]`

// podmanInspect is podman's answer (solace-util's podmanInspect): no slash, the older
// Healthcheck spelling, and the secrets podman lists.
const podmanInspect = `[{
  "Name": "solbroker",
  "RestartCount": 0,
  "State": {
    "Status": "exited", "Running": false, "ExitCode": 137, "OOMKilled": true,
    "Error": "container killed",
    "StartedAt": "2026-09-08T10:11:12.13Z", "FinishedAt": "2026-09-08T11:00:00Z",
    "Healthcheck": {"Status": "unhealthy", "FailingStreak": 4}
  },
  "Config": {"Image": "solace/solace-pubsub-standard:10.10.1.128"},
  "Mounts": [{"Source": "/opt/solace/data", "Destination": "/var/lib/solace/jail", "RW": true}],
  "Secrets": [{"Name": "solace-admin-password"}, {"Name": "solace-tls-servercertificate"}]
}]`

func TestDecodeDocker(t *testing.T) {
	cs, err := Decode(dockerInspect)
	if err != nil || len(cs) != 1 {
		t.Fatalf("Decode = %+v, %v", cs, err)
	}
	c := cs[0]
	want := Container{
		Name:         "solmq-connector",
		Created:      "2026-08-18T04:10:00Z",
		Image:        "solace/solace-pubsub-connector-ibmmq:2.14.1",
		Labels:       map[string]string{"com.docker.compose.project": "eg"},
		RestartCount: 2,
		NanoCPUs:     1000000000,
		Memory:       1073741824,
		Mounts: []Mount{
			{Type: "bind", Source: "/srv/solmq/application.yml", Destination: "/app/external/spring/config/application.yml"},
			{Type: "volume", Name: "solmq-libs", Destination: "/app/external/libs", RW: true},
		},
		Networks: []string{"backend", "eg_default"},
		State: State{
			Status: "running", Running: true, StartedAt: "2026-08-18T04:12:07Z",
			FinishedAt: "0001-01-01T00:00:00Z", Health: Health{Status: "healthy"},
		},
	}
	if !reflect.DeepEqual(c, want) {
		t.Errorf("Decode =\n%+v\nwant\n%+v", c, want)
	}
}

func TestDecodePodman(t *testing.T) {
	cs, err := Decode(podmanInspect)
	if err != nil || len(cs) != 1 {
		t.Fatalf("Decode = %+v, %v", cs, err)
	}
	c := cs[0]
	if c.Name != "solbroker" || c.Networks != nil || c.Labels != nil {
		t.Errorf("Decode = %+v", c)
	}
	if want := (State{Status: "exited", ExitCode: 137, OOMKilled: true, Error: "container killed",
		StartedAt: "2026-09-08T10:11:12.13Z", FinishedAt: "2026-09-08T11:00:00Z",
		Health: Health{Status: "unhealthy", FailingStreak: 4}}); c.State != want {
		t.Errorf("State = %+v, want %+v", c.State, want)
	}
	if want := []string{"solace-admin-password", "solace-tls-servercertificate"}; !reflect.DeepEqual(c.Secrets, want) {
		t.Errorf("Secrets = %v, want %v", c.Secrets, want)
	}
}

// TestDecodeHealthSpellings carries solace-util's TestInspectHealthReadsBothSpellings
// and TestInspectDistinguishesNoHealthcheckFromUnknown: the first spelling with a
// status wins, an empty Health block does not hide a populated Healthcheck, and no
// healthcheck is the zero Health.
func TestDecodeHealthSpellings(t *testing.T) {
	for _, c := range []struct {
		name  string
		state string
		want  Health
	}{
		{"Health", `"Health": {"Status": "starting", "FailingStreak": 1}`, Health{"starting", 1}},
		{"Healthcheck", `"Healthcheck": {"Status": "healthy"}`, Health{Status: "healthy"}},
		{"both, Health first", `"Health": {"Status": "healthy"}, "Healthcheck": {"Status": "unhealthy"}`,
			Health{Status: "healthy"}},
		{"empty Health beside Healthcheck", `"Health": {"Status": ""}, "Healthcheck": {"Status": "unhealthy", "FailingStreak": 2}`,
			Health{"unhealthy", 2}},
		{"null Health", `"Health": null`, Health{}},
		{"no healthcheck", `"Status": "running"`, Health{}},
	} {
		cs, err := Decode(fmt.Sprintf(`[{"Name": "x", "State": {%s}}]`, c.state))
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if cs[0].State.Health != c.want {
			t.Errorf("%s: Health = %+v, want %+v", c.name, cs[0].State.Health, c.want)
		}
	}
}

// TestDecodeNeverReadsEnv: the environment holds secrets, so it has no field to land
// in, and nothing decoded carries one.
func TestDecodeNeverReadsEnv(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeFor[Container](), reflect.TypeFor[State](), reflect.TypeFor[Mount]()} {
		for i := range typ.NumField() {
			if strings.Contains(strings.ToLower(typ.Field(i).Name), "env") {
				t.Errorf("%s has a field %s; the environment must never be decoded", typ.Name(), typ.Field(i).Name)
			}
		}
	}
	cs, err := Decode(dockerInspect)
	if err != nil {
		t.Fatal(err)
	}
	if s := fmt.Sprintf("%+v", cs); strings.Contains(s, "hunter2") {
		t.Errorf("a secret from Config.Env reached the decoded value: %s", s)
	}
}

// TestDecodeRefuses: no answer, a broken one and a single object are errors, so a
// failed probe never reads as "no containers" (solace-util's
// TestInspectRejectsAnEmptyReply, solmq-conn-util's TestParseInspectFilterAndErrors).
func TestDecodeRefuses(t *testing.T) {
	for _, out := range []string{"", "  \n", "not json", `{"Name": "x"}`, `[{"Name": 3}]`} {
		if _, err := Decode(out); err == nil {
			t.Errorf("Decode(%q) must fail", out)
		}
	}
	for _, out := range []string{"[]", " [] \n", "null"} {
		if cs, err := Decode(out); err != nil || len(cs) != 0 {
			t.Errorf("Decode(%q) = %+v, %v; want no containers", out, cs, err)
		}
	}
}

// TestDecodePartial carries solace-util's TestInspectDecodesPartialOutput: an
// answer without most fields still decodes, with zero values.
func TestDecodePartial(t *testing.T) {
	cs, err := Decode(`[{"Name": "/x"}, {"Name": "y", "State": {"Status": "created"}}]`)
	if err != nil || len(cs) != 2 {
		t.Fatalf("Decode = %+v, %v", cs, err)
	}
	if cs[0].Name != "x" || cs[0].State != (State{}) || cs[1].State.Status != "created" {
		t.Errorf("Decode = %+v", cs)
	}
}

// TestPrintable carries solace-util's TestPrintableStripsControlCharacters, and adds
// the C1 range a terminal can also act on.
func TestPrintable(t *testing.T) {
	for in, want := range map[string]string{
		"solace/pubsub:10.10": "solace/pubsub:10.10",
		"exited\x1b[2J":       "exited[2J",
		"a\x00b\x7fc":         "abc",
		"line\nbreak":         "linebreak",
		"csi\xc2\x9b31m":      "csi31m",
		"caf\xc3\xa9":         "caf\xc3\xa9",
		"":                    "",
	} {
		if got := Printable(in); got != want {
			t.Errorf("Printable(%q) = %q, want %q", in, got, want)
		}
	}
}
