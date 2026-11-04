package systemd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/runner"
)

// TestUserFlagOnEveryBuilder: --user follows systemctl and leads the verb, and the
// system manager gets none.
func TestUserFlagOnEveryBuilder(t *testing.T) {
	const unit = "solace.service"
	for _, c := range []struct {
		name  string
		build func(user bool) runner.Cmd
		tail  []string
	}{
		{"Systemctl", func(u bool) runner.Cmd { return Systemctl(u, "cat", unit) }, []string{"cat", unit}},
		{"DaemonReload", DaemonReload, []string{"daemon-reload"}},
		{"Start", func(u bool) runner.Cmd { return Start(u, unit) }, []string{"start", unit}},
		{"Stop", func(u bool) runner.Cmd { return Stop(u, unit) }, []string{"stop", unit}},
		{"Restart", func(u bool) runner.Cmd { return Restart(u, unit) }, []string{"restart", unit}},
		{"IsActive", func(u bool) runner.Cmd { return IsActive(u, unit) }, []string{"is-active", unit}},
	} {
		if got, want := c.build(false).Argv, append([]string{"systemctl"}, c.tail...); !reflect.DeepEqual(got, want) {
			t.Errorf("%s system argv = %v, want %v", c.name, got, want)
		}
		if got, want := c.build(true).Argv, append([]string{"systemctl", "--user"}, c.tail...); !reflect.DeepEqual(got, want) {
			t.Errorf("%s user argv = %v, want %v", c.name, got, want)
		}
	}
}

// TestStateWords: Active is what is-active exits 0 for, Stopped is what confirms
// the unit is down, and no answer is neither.
func TestStateWords(t *testing.T) {
	for _, c := range []struct {
		out             string
		state           string
		answered        bool
		active, stopped bool
	}{
		{"active\n", "active", true, true, false},
		{"reloading\n", "reloading", true, true, false},
		{"activating\n", "activating", true, false, false},
		{"deactivating\n", "deactivating", true, false, false},
		{"inactive\n", "inactive", true, false, true},
		{"failed\n", "failed", true, false, true},
		{"unknown\n", "unknown", true, false, true},
		{"", "", false, false, false},
		{"  \n", "", false, false, false},
	} {
		state, answered := State(c.out)
		if state != c.state || answered != c.answered {
			t.Errorf("State(%q) = %q, %v; want %q, %v", c.out, state, answered, c.state, c.answered)
		}
		if got := Active(state); got != c.active {
			t.Errorf("Active(%q) = %v, want %v", state, got, c.active)
		}
		if got := Stopped(state); got != c.stopped {
			t.Errorf("Stopped(%q) = %v, want %v", state, got, c.stopped)
		}
	}
}

// TestNRestarts carries solmq-conn-util's TestSystemctlNRestarts cases: the argv for
// each manager, and the answers systemd gives.
func TestNRestarts(t *testing.T) {
	const unit = "solmq-connector.service"
	if want := []string{"systemctl", "--user", "show", unit, "-p", "NRestarts", "--value"}; !reflect.DeepEqual(NRestarts(true, unit).Argv, want) {
		t.Errorf("user argv = %v, want %v", NRestarts(true, unit).Argv, want)
	}
	if want := []string{"systemctl", "show", unit, "-p", "NRestarts", "--value"}; !reflect.DeepEqual(NRestarts(false, unit).Argv, want) {
		t.Errorf("system argv = %v, want %v", NRestarts(false, unit).Argv, want)
	}
	for out, want := range map[string]int{"7\n": 7, "0\n": 0, " 12 ": 12} {
		if got, err := ParseNRestarts(out); err != nil || got != want {
			t.Errorf("ParseNRestarts(%q) = %d, %v; want %d", out, got, err, want)
		}
	}
}

// TestParseNRestartsRefuses: an unknown unit answers nothing, and anything that is
// not a count is refused rather than read as zero.
func TestParseNRestartsRefuses(t *testing.T) {
	for _, out := range []string{"", "\n", "-1", "x", "7 restarts"} {
		_, err := ParseNRestarts(out)
		if err == nil {
			t.Errorf("ParseNRestarts(%q) must fail", out)
			continue
		}
		if !strings.Contains(err.Error(), "unit name") {
			t.Errorf("ParseNRestarts(%q) err = %v, want it to say what to check", out, err)
		}
	}
}

func TestLingerArgv(t *testing.T) {
	if want := []string{"loginctl", "show-user", "1000", "--property=Linger"}; !reflect.DeepEqual(ShowLinger(1000).Argv, want) {
		t.Errorf("ShowLinger argv = %v, want %v", ShowLinger(1000).Argv, want)
	}
	// No user named: the form that needs no administrator.
	if want := []string{"loginctl", "enable-linger"}; !reflect.DeepEqual(EnableLinger().Argv, want) {
		t.Errorf("EnableLinger argv = %v, want %v", EnableLinger().Argv, want)
	}
}

func TestParseLinger(t *testing.T) {
	for out, want := range map[string]string{
		"Linger=yes\n": "yes",
		"Linger=no\n":  "no",
		" Linger=yes ": "yes",
		"Linger=\n":    "",
		"odd answer":   "odd answer",
	} {
		if got := ParseLinger(out); got != want {
			t.Errorf("ParseLinger(%q) = %q, want %q", out, got, want)
		}
	}
}
