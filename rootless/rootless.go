// Package rootless holds the rules and host checks for running podman as an ordinary
// user that solmq-conn-util and solace-util share: rootless unless rootful is asked
// for, the user session systemctl --user needs, linger, the session's core limit, and
// whether the container's ids fit the user's subordinate ranges.
//
// The checks take what a tool has already read (environment values, command
// output, file contents) and return an error that says what is wrong and how to
// fix it. Reading stays with the tool, through its runner where it has one, so a
// test fake sees every probe and a preview runs none. A tool may append its own
// remedy, such as a config key or a drop-in file it writes, to an error from here.
package rootless

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hafio/go-deploy-utility/runner"
)

const (
	runtimeDirKey  = "XDG_RUNTIME_DIR"
	busKey         = "DBUS_SESSION_BUS_ADDRESS"
	runtimeDirBase = "/run/user/"

	// subIDFloor and subIDBlock are shadow-utils' own SUB_UID_MIN and SUB_UID_COUNT
	// defaults, where a range would start and how wide it would be had useradd
	// allocated one.
	subIDFloor = 100000
	subIDBlock = 65536
)

// Decide applies the rootless rule both tools follow and reports whether to deploy
// rootless. Rootless is the default: rootful podman is deployed only when asked for.
// rootful is the tool's setting that asks for it, false when unset, and key is how
// the tool names that setting in its config, such as "podman.rootful", for the
// refusals. euid is the effective uid, or -1 where there is none (Windows).
//
// An ordinary user deploys rootless as itself, and asking for rootful as one is
// refused, since rootful needs root. Root deploys rootful only when asked; root
// without it is refused, because deploying on behalf of a rootless user is not built
// yet and rootful is never assumed. With no euid to compare against, the setting
// stands, so an artifact still renders. The bool is meaningless when the error is not
// nil.
func Decide(euid int, rootful bool, key string) (bool, error) {
	switch {
	case euid < 0:
		return !rootful, nil
	case rootful && euid != 0:
		return false, fmt.Errorf("%s=true requires root, but running as euid %d; re-run with sudo -- or "+
			"remove %s to deploy rootless as this user", key, euid, key)
	case !rootful && euid == 0:
		return false, fmt.Errorf("running as root (euid 0), and rootful podman is deployed only when asked "+
			"for: set %s=true to deploy the rootful engine -- or run as the rootless user itself, without sudo",
			key)
	}
	return !rootful, nil
}

// Session is the user session `systemctl --user` and loginctl talk through: the
// runtime directory and the D-Bus address in it. A Derived field marks a value
// UserSession made up because the environment did not carry it; the others were
// inherited and are left as they are.
type Session struct {
	// RuntimeDir is XDG_RUNTIME_DIR, /run/user/<euid> when derived.
	RuntimeDir        string
	RuntimeDirDerived bool

	// Bus is DBUS_SESSION_BUS_ADDRESS, unix:path=<RuntimeDir>/bus when derived.
	Bus        string
	BusDerived bool
}

// UserSession repairs the session sudo, su and cron strip: each of
// XDG_RUNTIME_DIR and DBUS_SESSION_BUS_ADDRESS is kept when set and derived when
// not. The runtime directory, inherited or derived, must exist before anything is
// exported, or the failure would surface later as an obscure systemctl error;
// isDir reports whether it does, and an error from it is a probe that could not
// run, not a missing directory. getenv is os.Getenv.
//
// Export the result with Env before the first `systemctl --user`, and memoise it per
// run. With no POSIX euid (-1) there is nothing to repair, and the Session is empty.
func UserSession(euid int, getenv func(string) string, isDir func(string) (bool, error)) (Session, error) {
	if euid < 0 {
		return Session{}, nil
	}
	s := Session{RuntimeDir: getenv(runtimeDirKey)}
	if s.RuntimeDir == "" {
		s.RuntimeDir, s.RuntimeDirDerived = runtimeDirBase+strconv.Itoa(euid), true
	}
	ok, err := isDir(s.RuntimeDir)
	if err != nil {
		return Session{}, fmt.Errorf("rootless: cannot tell whether the user runtime directory %s exists: %w; "+
			"fix what stopped the probe, then run again", s.RuntimeDir, err)
	}
	if !ok {
		return Session{}, fmt.Errorf("the user runtime directory %s does not exist, so `systemctl --user` and "+
			"`loginctl` have no bus to talk to.\n"+
			"  Either this process is running under sudo/su/cron (which strip %s -- run as the rootless "+
			"user's own login session instead), or that user has no session at all "+
			"(`loginctl enable-linger <user>` is what makes %s%d persist without one)",
			s.RuntimeDir, runtimeDirKey, runtimeDirBase, euid)
	}
	s.Bus = getenv(busKey)
	if s.Bus == "" {
		s.Bus, s.BusDerived = "unix:path="+s.RuntimeDir+"/bus", true
	}
	return s, nil
}

// Env is the derived values as KEY=VALUE entries, for os.Setenv or runner.Cmd.Env.
// An inherited value is not repeated, so exporting these never overwrites one.
func (s Session) Env() []string {
	var env []string
	if s.RuntimeDirDerived {
		env = append(env, runtimeDirKey+"="+s.RuntimeDir)
	}
	if s.BusDerived {
		env = append(env, busKey+"="+s.Bus)
	}
	return env
}

// String reports both values and where each came from, such as
// "XDG_RUNTIME_DIR=/run/user/1000 (derived) DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus (derived)".
func (s Session) String() string {
	return fmt.Sprintf("%s=%s (%s) %s=%s (%s)",
		runtimeDirKey, s.RuntimeDir, origin(s.RuntimeDirDerived), busKey, s.Bus, origin(s.BusDerived))
}

func origin(derived bool) string {
	if derived {
		return "derived"
	}
	return "inherited"
}

// CheckLinger refuses a user whose linger state, as systemd.ParseLinger read it, is
// not "yes". Without linger, systemd stops the user's manager at logout, taking a
// rootless unit with it, and starts nothing at boot, whatever the unit's WantedBy
// says.
func CheckLinger(uid int, state string) error {
	if state == "yes" {
		return nil
	}
	return fmt.Errorf("linger is not enabled for uid %d (Linger=%s), so its user units stop at logout and "+
		"do not start at boot.\n"+
		"  Run:\n"+
		"    loginctl enable-linger", uid, state)
}

// Unbounded is a limit with no ceiling: systemd's "infinity", the shell's
// "unlimited".
const Unbounded = math.MaxInt

// ParseLimit reads one limit value: a decimal, or "infinity" or "unlimited" for
// Unbounded. source names where v came from, for the error.
func ParseLimit(v, source string) (int, error) {
	switch v {
	case "infinity", "unlimited":
		return Unbounded, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("cannot parse the limit %q from %s: %w", v, source, err)
	}
	return n, nil
}

// LimitText writes a limit for a message, with Unbounded as "infinity" rather than
// an implausible number.
func LimitText(v int) string {
	if v == Unbounded {
		return "infinity"
	}
	return strconv.Itoa(v)
}

// ProcLimits reads the limits the child inherits from this session, in bytes where
// the limit is a size: `cat /proc/self/limits`. It is a command rather than a file
// read so that it runs on the host the runner reaches. Pass its output to
// HardLimit; an empty answer means the host has no /proc to ask.
func ProcLimits() runner.Cmd {
	return runner.Cmd{Argv: []string{"cat", "/proc/self/limits"}}
}

// HardLimit is the hard limit on the line of procLimits named name, such as
// "Max core file size". The kernel pads each name to a wide column, so the name is
// everything before the first run of two spaces, and is matched whole. After it
// come the soft limit, the hard limit and, on most lines, the units.
func HardLimit(procLimits, name string) (int, error) {
	for _, line := range strings.Split(procLimits, "\n") {
		label, rest, ok := strings.Cut(line, "  ")
		if !ok || strings.TrimSpace(label) != name {
			continue
		}
		f := strings.Fields(rest)
		if len(f) < 2 {
			return 0, fmt.Errorf("rootless: /proc/self/limits line %q has no hard limit; read it on the Linux "+
				"host that runs podman", line)
		}
		return ParseLimit(f[1], "/proc/self/limits "+name)
	}
	return 0, fmt.Errorf("rootless: /proc/self/limits has no %q line; read it on the Linux host that runs podman",
		name)
}

// CheckCore refuses a session whose hard core limit is below need, the core ulimit
// the container asks for. `podman exec` gives the process it starts the container's
// limits from this session, and an unprivileged runtime cannot raise a hard limit,
// so the container starts but every command run inside it fails.
func CheckCore(hard, need int) error {
	if hard >= need {
		return nil
	}
	return fmt.Errorf("session core is %s, needs %s: podman exec, and so every command run inside the "+
		"container, would fail with `setrlimit RLIMIT_CORE: Operation not permitted`. Raise this user's hard "+
		"core limit to %s, or lower the container's core ulimit to %s or less",
		LimitText(hard), LimitText(need), LimitText(need), LimitText(hard))
}

// CheckIDMap refuses a container id that the user's mapping does not reach. out is
// podman.IDMap's output for file's kind, file is "subuid" or "subgid", and key and
// value name the setting that asked for id, as the operator wrote it. Call it only
// when deploying rootless, since rootful podman reports no mappings.
//
// Id 0 always passes, since it is the invoking user itself. A refusal says what is
// wrong; append SubIDFix for how to fix it, then the tool's own alternatives.
func CheckIDMap(out, file string, id int, key, value string) error {
	if id == 0 {
		return nil
	}
	ranges := strings.TrimSpace(out)
	if ranges == "" {
		return fmt.Errorf("this user has no %s allocation, so %s %q cannot work", file, key, value)
	}
	if covers(ranges, id) {
		return nil
	}
	return fmt.Errorf("%s %s %d is outside this user's %s ranges (%s)", key, file, id, file, ranges)
}

// covers reports whether id falls inside any containerID:size range of ranges. A
// pair that does not parse is skipped rather than refused: podman owns the format.
func covers(ranges string, id int) bool {
	for _, pair := range strings.Fields(ranges) {
		first, size, found := strings.Cut(pair, ":")
		if !found {
			continue
		}
		start, err := strconv.Atoi(first)
		if err != nil {
			continue
		}
		n, err := strconv.Atoi(size)
		if err != nil {
			continue
		}
		if id >= start && id < start+n {
			return true
		}
	}
	return false
}

// NextFreeRange proposes a subordinate range for a user that needs need ids (a
// container id N needs N+1): max(65536, need) ids, starting at 100000 or past the
// end of every range in subids, the contents of /etc/subuid and /etc/subgid
// together. Two users sharing a range map to the same host ids, which is why it is
// computed rather than fixed. A line that is not name:start:count is skipped. With
// the files unreadable, pass "" and report the result as not computed to SubIDFix.
// Ranges an NSS source (such as SSSD) grants are not in the files, so the proposal
// is still worth checking on such a host.
func NextFreeRange(subids string, need int) (start, end int) {
	size := max(subIDBlock, need)
	highest := subIDFloor
	for line := range strings.SplitSeq(strings.TrimSpace(subids), "\n") {
		f := strings.Split(strings.TrimSpace(line), ":")
		if len(f) != 3 {
			continue
		}
		first, err1 := strconv.Atoi(f[1])
		count, err2 := strconv.Atoi(f[2])
		if err1 != nil || err2 != nil {
			continue
		}
		if first+count > highest {
			highest = first + count
		}
	}
	return highest, highest + size - 1
}

// SubIDFix is how to fix a CheckIDMap refusal: the usermod line an administrator
// runs for user, granting start-end as both subuids and subgids, then the
// `podman system migrate` the user runs so podman picks the range up. user is
// named literally, because the line is meant to be forwarded and run by someone
// else. computed is false when the range was not computed from the files, and the
// line then says to check it does not overlap.
//
// A refusal reads: the error, a period, a newline, then this.
func SubIDFix(user string, start, end int, computed bool) string {
	caveat := ""
	if !computed {
		caveat = "   # /etc/subuid unreadable: check this does not overlap"
	}
	return fmt.Sprintf("  Ask an administrator to run:\n"+
		"    sudo usermod --add-subuids %d-%d --add-subgids %d-%d %s%s\n"+
		"  then, as this user:\n"+
		"    podman system migrate", start, end, start, end, user, caveat)
}
