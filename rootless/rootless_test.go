package rootless

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestDecide is the rule: rootless unless rootful is asked for, rootful only for
// root, root never rootful unasked, and no euid leaves the setting standing (the rule
// rows began as solace-util's TestGuardPodmanEUID).
func TestDecide(t *testing.T) {
	for _, c := range []struct {
		name    string
		rootful bool
		euid    int
		want    bool
		wantErr bool
	}{
		{"a user is rootless by default", false, 1000, true, false},
		{"a user asking for rootful fails", true, 1000, false, true},
		{"root asking for rootful is rootful", true, 0, false, false},
		{"root without asking fails", false, 0, false, true},
		{"non-posix euid is rootless by default", false, -1, true, false},
		{"non-posix euid keeps rootful", true, -1, false, false},
	} {
		got, err := Decide(c.euid, c.rootful, "podman.rootful")
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", c.name, err, c.wantErr)
			continue
		}
		if err == nil && got != c.want {
			t.Errorf("%s: rootless = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestDecideMessagesNameTheKeyAndTheFix: each refusal names the key, the account it
// found and both ways out (solace-util's TestGuardPodmanEUIDMessagesNameTheKeyAndTheFix).
// The key is the caller's, so two different spellings are passed and each must come
// back: the package knows no tool's config.
func TestDecideMessagesNameTheKeyAndTheFix(t *testing.T) {
	for _, key := range []string{"podman.rootful", "engine.rootful"} {
		_, err := Decide(0, false, key)
		for _, want := range []string{"running as root", "only when asked", key + "=true", "without sudo"} {
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("%s: the root-without-asking refusal should name %q, got: %v", key, want, err)
			}
		}
		_, err = Decide(1000, true, key)
		for _, want := range []string{key + "=true", "requires root", "euid 1000", "sudo", "remove " + key} {
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("%s: the rootful-as-non-root refusal should name %q, got: %v", key, want, err)
			}
		}
	}
}

// env is a fixed getenv for the seams.
func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

// dirs is an isDir seam that knows which directories exist and records what it was
// asked.
func dirs(asked *[]string, exist ...string) func(string) (bool, error) {
	return func(d string) (bool, error) {
		*asked = append(*asked, d)
		for _, e := range exist {
			if d == e {
				return true, nil
			}
		}
		return false, nil
	}
}

// TestUserSession carries solace-util's TestEnsureUserSession* cases: both values
// derived when unset, an inherited value kept as it is, and each half independent,
// with the runtime dir proven to exist before anything is exported.
func TestUserSession(t *testing.T) {
	for _, c := range []struct {
		name  string
		env   map[string]string
		exist string
		want  Session
	}{
		{"derives both when unset", nil, "/run/user/1000",
			Session{"/run/user/1000", true, "unix:path=/run/user/1000/bus", true}},
		{"never overwrites inherited",
			map[string]string{"XDG_RUNTIME_DIR": "/custom/run", "DBUS_SESSION_BUS_ADDRESS": "unix:path=/custom/bus"},
			"/custom/run", Session{"/custom/run", false, "unix:path=/custom/bus", false}},
		{"derives the bus from the inherited dir", map[string]string{"XDG_RUNTIME_DIR": "/custom/run"},
			"/custom/run", Session{"/custom/run", false, "unix:path=/custom/run/bus", true}},
		{"keeps an inherited bus with a derived dir",
			map[string]string{"DBUS_SESSION_BUS_ADDRESS": "unix:path=/custom/bus"},
			"/run/user/1000", Session{"/run/user/1000", true, "unix:path=/custom/bus", false}},
	} {
		var asked []string
		got, err := UserSession(1000, env(c.env), dirs(&asked, c.exist))
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: Session = %+v, want %+v", c.name, got, c.want)
		}
		if !reflect.DeepEqual(asked, []string{c.exist}) {
			t.Errorf("%s: probed %v, want only the runtime dir %s", c.name, asked, c.exist)
		}
	}
}

// TestUserSessionMissingDir: a runtime dir that is not there exports nothing and
// says why (solace-util's TestEnsureUserSessionExportsNothingWhenTheDirIsMissing).
func TestUserSessionMissingDir(t *testing.T) {
	var asked []string
	got, err := UserSession(1000, env(nil), dirs(&asked))
	if err == nil {
		t.Fatal("a missing runtime dir must fail")
	}
	if got != (Session{}) || got.Env() != nil {
		t.Errorf("nothing may be exported, got %+v", got)
	}
	for _, want := range []string{"/run/user/1000", "does not exist", "sudo", "enable-linger"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to name %q", err, want)
		}
	}
}

// TestUserSessionProbeError: a probe that could not run is not a missing directory,
// so it must not send the operator after sudo or linger.
func TestUserSessionProbeError(t *testing.T) {
	cause := errors.New("sh: not found")
	_, err := UserSession(1000, env(nil), func(string) (bool, error) { return false, cause })
	if !errors.Is(err, cause) {
		t.Fatalf("err = %v, want the probe's error wrapped", err)
	}
	if strings.Contains(err.Error(), "sudo") || !strings.Contains(err.Error(), "/run/user/1000") {
		t.Errorf("err = %v, want the dir named and no sudo diagnosis", err)
	}
}

func TestUserSessionWithoutAnEUID(t *testing.T) {
	got, err := UserSession(-1, env(nil), func(string) (bool, error) {
		t.Error("no euid: nothing to probe")
		return false, nil
	})
	if err != nil || got != (Session{}) {
		t.Errorf("UserSession(-1) = %+v, %v; want an empty Session", got, err)
	}
}

func TestSessionEnvAndString(t *testing.T) {
	s := Session{"/run/user/1000", true, "unix:path=/custom/bus", false}
	if want := []string{"XDG_RUNTIME_DIR=/run/user/1000"}; !reflect.DeepEqual(s.Env(), want) {
		t.Errorf("Env = %v, want only the derived entry %v", s.Env(), want)
	}
	if want := "XDG_RUNTIME_DIR=/run/user/1000 (derived) DBUS_SESSION_BUS_ADDRESS=unix:path=/custom/bus (inherited)"; s.String() != want {
		t.Errorf("String = %q, want %q", s.String(), want)
	}
	both := Session{"/run/user/1000", true, "unix:path=/run/user/1000/bus", true}
	want := []string{"XDG_RUNTIME_DIR=/run/user/1000", "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus"}
	if !reflect.DeepEqual(both.Env(), want) {
		t.Errorf("Env = %v, want %v", both.Env(), want)
	}
}

// TestCheckLinger keeps the words solace-util's TestCheckLingerRefusesWhenDisabled
// pins, and names the uid and what logind said.
func TestCheckLinger(t *testing.T) {
	if err := CheckLinger(1000, "yes"); err != nil {
		t.Errorf("linger on must pass, got %v", err)
	}
	for _, state := range []string{"no", "", "odd"} {
		err := CheckLinger(1000, state)
		if err == nil {
			t.Errorf("Linger=%s must fail", state)
			continue
		}
		for _, want := range []string{"linger is not enabled", "uid 1000", "Linger=" + state, "logout", "loginctl enable-linger"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("Linger=%s: err = %v, want it to name %q", state, err, want)
			}
		}
	}
}

func TestParseLimit(t *testing.T) {
	for _, tc := range []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"1048576", 1048576, false},
		{"0", 0, false},
		{"infinity", Unbounded, false},  // systemd's spelling
		{"unlimited", Unbounded, false}, // the shell's
		{"", 0, true},
		{"not-a-number", 0, true},
	} {
		got, err := ParseLimit(tc.in, "fixture")
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseLimit(%q) err = %v, wantErr %v", tc.in, err, tc.wantErr)
			continue
		}
		if err == nil && got != tc.want {
			t.Errorf("ParseLimit(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestLimitText(t *testing.T) {
	if LimitText(Unbounded) != "infinity" || LimitText(0) != "0" || LimitText(1048576) != "1048576" {
		t.Errorf("LimitText = %q, %q, %q", LimitText(Unbounded), LimitText(0), LimitText(1048576))
	}
}

// procLimits is /proc/self/limits as Linux writes it, with a hardened core limit.
// Two lines carry no units.
func procLimits() string {
	return "Limit                     Soft Limit           Hard Limit           Units     \n" +
		"Max cpu time              unlimited            unlimited            seconds   \n" +
		"Max core file size        0                    0                    bytes     \n" +
		"Max open files            1024                 524288               files     \n" +
		"Max nice priority         0                    0                    \n" +
		"Max realtime timeout      unlimited            unlimited            us        \n"
}

func TestHardLimit(t *testing.T) {
	if got := ProcLimits().Argv; !reflect.DeepEqual(got, []string{"cat", "/proc/self/limits"}) {
		t.Errorf("ProcLimits argv = %v", got)
	}
	for name, want := range map[string]int{
		"Max core file size": 0,
		"Max open files":     524288,
		"Max nice priority":  0,
		"Max cpu time":       Unbounded,
	} {
		if got, err := HardLimit(procLimits(), name); err != nil || got != want {
			t.Errorf("HardLimit(%q) = %d, %v; want %d", name, got, err, want)
		}
	}
	// A name that is only a prefix of a line's name is not that line.
	if _, err := HardLimit(procLimits(), "Max open"); err == nil || !strings.Contains(err.Error(), `"Max open"`) {
		t.Errorf("a partial name must not match, got %v", err)
	}
	if _, err := HardLimit("", "Max core file size"); err == nil || !strings.Contains(err.Error(), "no") {
		t.Errorf("an empty answer must fail, got %v", err)
	}
	if _, err := HardLimit("Max core file size        0\n", "Max core file size"); err == nil ||
		!strings.Contains(err.Error(), "no hard limit") {
		t.Errorf("a line with no hard limit must fail, got %v", err)
	}
	if _, err := HardLimit("Max core file size        0    lots   bytes\n", "Max core file size"); err == nil ||
		!strings.Contains(err.Error(), `"lots"`) {
		t.Errorf("an unparsable hard limit must fail naming it, got %v", err)
	}
}

func TestCheckCore(t *testing.T) {
	for _, c := range []struct {
		hard, need int
		ok         bool
	}{
		{Unbounded, Unbounded, true},
		{0, Unbounded, false},
		{4096, 4096, true},
		{4095, 4096, false},
		{0, 0, true},
	} {
		if err := CheckCore(c.hard, c.need); (err == nil) != c.ok {
			t.Errorf("CheckCore(%d, %d) = %v, want ok %v", c.hard, c.need, err, c.ok)
		}
	}
	err := CheckCore(0, Unbounded)
	for _, want := range []string{"session core is 0, needs infinity", "setrlimit RLIMIT_CORE", "Raise",
		"core ulimit to 0 or less"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to name %q", err, want)
		}
	}
}

func TestCheckIDMap(t *testing.T) {
	const stock = "0:1 1:65536 \n"
	for _, c := range []struct {
		name string
		out  string
		id   int
		want string // "" for no error
	}{
		{"inside the range", stock, 1000, ""},
		{"the last id of the range", stock, 65536, ""},
		{"one past the range", stock, 65537,
			`container.runUser subuid 65537 is outside this user's subuid ranges (0:1 1:65536)`},
		{"no allocation", "\n", 1000, `this user has no subuid allocation, so container.runUser "1000:0" cannot work`},
		{"id 0 is the invoker, even with no allocation", "", 0, ""},
		{"malformed pairs are skipped", "junk 0:x y:1 1:65536", 1000, ""},
		{"a second range counts", "0:1 1:10 11:65536", 5000, ""},
	} {
		err := CheckIDMap(c.out, "subuid", c.id, "container.runUser", "1000:0")
		switch {
		case c.want == "" && err != nil:
			t.Errorf("%s: unexpected %v", c.name, err)
		case c.want != "" && (err == nil || err.Error() != c.want):
			t.Errorf("%s: err = %v, want %q", c.name, err, c.want)
		}
	}
}

// TestNextFreeRange carries solace-util's TestSuggestSubIDRangeAvoidsExistingAllocations
// cases: start past every allocation, whatever the order, and never narrower than
// the need.
func TestNextFreeRange(t *testing.T) {
	for _, tc := range []struct {
		name       string
		etc        string
		need       int
		start, end int
	}{
		{"empty host starts at the floor", "", 1001, 100000, 165535},
		{"next block after one account", "alice:100000:65536\nalice:100000:65536", 1001, 165536, 231071},
		{
			"highest wins regardless of order",
			"bob:200000:65536\nalice:100000:65536\ncarol:165536:65536",
			1001, 265536, 331071,
		},
		{"junk lines are skipped", "not-a-range\nalice:abc:65536\nalice:100000:65536", 1001, 165536, 231071},
		{"a large need widens the block", "", 1000002, 100000, 1100001},
		{"a range below the floor is passed over", "old:1000:500", 1001, 100000, 165535},
	} {
		start, end := NextFreeRange(tc.etc, tc.need)
		if start != tc.start || end != tc.end {
			t.Errorf("%s: range = %d-%d, want %d-%d", tc.name, start, end, tc.start, tc.end)
		}
		if end-start+1 < tc.need {
			t.Errorf("%s: range covers %d ids, need %d", tc.name, end-start+1, tc.need)
		}
	}
}

// TestSubIDFix is the remedy an operator forwards: the next free block for this
// account, named literally, and a caveat only when the files could not be read
// (solace-util's TestIDMappingErrProposesTheNextFreeBlock and
// TestSuggestSubIDRangeSaysWhenItIsGuessing).
func TestSubIDFix(t *testing.T) {
	start, end := NextFreeRange("alice:100000:65536\n", 1001)
	got := SubIDFix("solace-svc", start, end, true)
	want := "  Ask an administrator to run:\n" +
		"    sudo usermod --add-subuids 165536-231071 --add-subgids 165536-231071 solace-svc\n" +
		"  then, as this user:\n" +
		"    podman system migrate"
	if got != want {
		t.Errorf("SubIDFix =\n%s\nwant\n%s", got, want)
	}
	guess := SubIDFix("solace-svc", 100000, 165535, false)
	if !strings.Contains(guess, "unreadable") || !strings.Contains(guess, "overlap") {
		t.Errorf("an uncomputed range must be flagged as one to check, got:\n%s", guess)
	}
}
