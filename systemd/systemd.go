// Package systemd builds the systemctl and loginctl commands solmq-conn-util and
// solace-util run to drive a podman quadlet unit, and reads their answers.
//
// user selects the per-user manager (`systemctl --user`), which a rootless
// deployment uses; the system manager is used otherwise. Nothing here executes: a
// tool runs the returned Cmd through its runner.Runner.
package systemd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hafio/go-deploy-utility/runner"
)

// Systemctl is `systemctl [--user] <args...>`.
func Systemctl(user bool, args ...string) runner.Cmd {
	argv := make([]string, 0, 2+len(args))
	argv = append(argv, "systemctl")
	if user {
		argv = append(argv, "--user")
	}
	return runner.Cmd{Argv: append(argv, args...)}
}

// DaemonReload makes the manager reread its units, which runs podman's quadlet
// generator over a unit just written or removed.
func DaemonReload(user bool) runner.Cmd { return Systemctl(user, "daemon-reload") }

// Start starts unit. On a unit that is already active it does nothing, so a changed
// unit needs Restart.
func Start(user bool, unit string) runner.Cmd { return Systemctl(user, "start", unit) }

// Stop stops unit.
func Stop(user bool, unit string) runner.Cmd { return Systemctl(user, "stop", unit) }

// Restart restarts unit, or starts it if it is not running.
func Restart(user bool, unit string) runner.Cmd { return Systemctl(user, "restart", unit) }

// IsActive asks for unit's state, which systemctl prints on stdout. It exits
// non-zero for every state but an active one, so read stdout alone (runner.Output
// or runner.Split), ignore the exit error, and pass what was printed to State. A
// combined capture would mix a warning on stderr into the state.
func IsActive(user bool, unit string) runner.Cmd { return Systemctl(user, "is-active", unit) }

// State is the state word IsActive printed, trimmed. answered is false when nothing
// was printed, which means the probe itself failed and says nothing about the unit.
func State(out string) (state string, answered bool) {
	state = strings.TrimSpace(out)
	return state, state != ""
}

// Active reports whether state is one `systemctl is-active` exits 0 for: active, or
// reloading. Deploy restarts such a unit to load a changed one and starts any other.
func Active(state string) bool {
	switch state {
	case "active", "reloading":
		return true
	}
	return false
}

// Stopped reports whether state confirms the unit is not running: inactive, failed
// or unknown. Any other answer, no answer included, leaves it possibly running, so
// a teardown that must know refuses rather than guessing.
func Stopped(state string) bool {
	switch state {
	case "inactive", "failed", "unknown":
		return true
	}
	return false
}

// NRestarts asks how many times systemd has restarted unit:
// `systemctl [--user] show <unit> -p NRestarts --value`. On podman this count, not
// the engine's, is the restart count, because systemd restarts the container.
func NRestarts(user bool, unit string) runner.Cmd {
	return Systemctl(user, "show", unit, "-p", "NRestarts", "--value")
}

// ParseNRestarts reads NRestarts's answer. Anything but a non-negative decimal is an
// error. systemd answers a unit it does not know with an empty value rather than a
// failure, which is reported here so the caller can fall back to the engine's count.
func ParseNRestarts(out string) (int, error) {
	s := strings.TrimSpace(out)
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("systemd: NRestarts answered %q, which is not a restart count; systemd answers "+
			"nothing for a unit it does not know, so check the unit name and that the unit is loaded", s)
	}
	return n, nil
}

// ShowLinger asks logind whether uid lingers: `loginctl show-user <uid>
// --property=Linger`, which prints Linger=yes or Linger=no. It exits non-zero for a
// user with no session and no linger, which a caller reports as "cannot read the
// linger state", not as "no".
func ShowLinger(uid int) runner.Cmd {
	return runner.Cmd{Argv: []string{"loginctl", "show-user", strconv.Itoa(uid), "--property=Linger"}}
}

// EnableLinger turns linger on for the account running it. It names no user on
// purpose: naming one is the form that needs an administrator, while a user's own
// account is normally allowed to enable its own.
func EnableLinger() runner.Cmd { return runner.Cmd{Argv: []string{"loginctl", "enable-linger"}} }

// ParseLinger is the value ShowLinger printed after Linger=: "yes", "no", or
// whatever else logind said.
func ParseLinger(out string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(out), "Linger="))
}
