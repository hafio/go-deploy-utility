// Package artifacttest checks rendered compose files and quadlet units for the
// container posture both tools promise: privileges stated and never widened,
// quadlet keys within the tool's podman floor, and secrets mounted at absolute
// paths.
//
// The Check functions are pure and return one error listing every problem, so
// their rules are testable without a fake testing.T. The Assert functions wrap
// them for a tool's own tests, which pass in every artifact shape the tool
// renders: a token smuggled in behind a conditional only fails for the shape
// that writes it.
package artifacttest

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/containerart/compose"
	"github.com/hafio/go-deploy-utility/containerart/quadlet"
)

// Widening returns the spellings that would widen a container's privileges in a
// compose file or a quadlet unit: extended privileges, no-new-privileges turned
// off, added capabilities, host devices, kernel tunables, a relaxed SELinux or
// seccomp confinement, and the host's pid, ipc or user namespace. It is the union
// of both tools' lists, keeping the broader spelling where they differed, plus the
// podman flags that can ride a quadlet PodmanArgs= line.
//
// Host networking is not in it (see HostNetwork), and neither are Kubernetes
// spellings. Each call returns a new slice.
func Widening() []string {
	return []string{
		"privileged: true", "--privileged", "Privileged=true",
		"no-new-privileges:false", "no-new-privileges=false", "NoNewPrivileges=false",
		"cap_add", "AddCapability", "SYS_ADMIN",
		"devices:", "AddDevice",
		"sysctls:", "Sysctl=",
		"SecurityLabelDisable", "label=disable", "label:disable",
		"SeccompProfile=", "unconfined", "Unconfined",
		"pid: host", "ipc: host",
		"--cap-add", "--device", "--sysctl",
		"--pid=host", "--pid host", "--ipc=host", "--ipc host", "--userns=host", "--userns host",
	}
}

// HostNetwork returns the compose and quadlet spellings of host networking, then
// the podman flags that can ride a quadlet PodmanArgs= line. The compose spelling
// is always first and the quadlet key second, so a caller can pick either.
// solmq-conn-util forbids host networking in every artifact. solace-util writes it
// only when the operator asks, and checks that no other shape carries it.
func HostNetwork() []string {
	return []string{"network_mode: host", "Network=host",
		"--network=host", "--network host", "--net=host", "--net host"}
}

// CheckAbsent reports every token found in the artifact. Spaces, tabs and quotes
// are dropped from both before matching, and case is folded, so "privileged:true",
// a quoted "privileged" key and "privileged: True" are all caught by
// "privileged: true". A match inside a longer word is reported as well: the check
// errs on the strict side. Two tokens that differ only in case or spacing are
// reported once, under the first spelling.
func CheckAbsent(artifact string, tokens []string) error {
	flat := squeeze(artifact)
	var found []string
	var errs []error
	seen := make(map[string]bool, len(tokens))
	for _, tok := range tokens {
		t := squeeze(tok)
		switch {
		case t == "":
			errs = append(errs, fmt.Errorf("token %q is empty once spaces and quotes are dropped, so it would match anything; fix the token list", tok))
		case seen[t]:
		case strings.Contains(flat, t):
			found = append(found, strconv.Quote(tok))
		}
		seen[t] = true
	}
	if len(found) > 0 {
		errs = append(errs, fmt.Errorf("carries %s, which the container posture forbids; take it out of the renderer", strings.Join(found, ", ")))
	}
	return errors.Join(errs...)
}

// squeeze drops spaces, tabs and quotes, and folds case.
func squeeze(s string) string {
	return strings.ToLower(strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '"', '\'':
			return -1
		}
		return r
	}, s))
}

// CheckComposePosture reports a compose file that does not carry compose.Posture
// exactly once at service indent, or that spells no-new-privileges with the
// deprecated ':' separator. Both tools render one service per file.
func CheckComposePosture(artifact string) error {
	var errs []error
	// Counted by its first line: the block ends in a newline, so counting the
	// whole block would miss a second copy written straight after the first.
	first := compose.Posture[:strings.Index(compose.Posture, "\n")+1]
	switch n := strings.Count(artifact, "\n"+first); {
	case n != 1:
		errs = append(errs, fmt.Errorf("want the posture block (compose.Posture) exactly once at service indent, got %d: privileged: false, cap_drop ALL and no-new-privileges=true are stated, never inherited", n))
	case !strings.Contains(artifact, "\n"+compose.Posture):
		errs = append(errs, errors.New("the posture block is incomplete or out of order; write compose.Posture as it is"))
	}
	if strings.Contains(artifact, "no-new-privileges:") {
		errs = append(errs, errors.New("no-new-privileges uses the deprecated ':' separator, which dockerd warns about on every create; write it as compose.Posture does, with '='"))
	}
	return errors.Join(errs...)
}

// CheckQuadletPosture reports a unit in which a line of quadlet.Posture is
// missing, written more than once, or outside [Container]. Elsewhere quadlet does
// not read it: in [Service], NoNewPrivileges is systemd's own key and binds podman
// instead of the container.
func CheckQuadletPosture(unit string) error {
	var errs []error
	lines := unitLines(unit)
	for _, want := range strings.Split(strings.TrimSuffix(quadlet.Posture, "\n"), "\n") {
		in, out := 0, 0
		for _, l := range lines {
			switch {
			case l.text != want:
			case l.section == "[Container]":
				in++
			default:
				out++
			}
		}
		switch {
		case in+out == 0:
			errs = append(errs, fmt.Errorf("%s is missing; write quadlet.Posture into [Container]", want))
		case out > 0:
			errs = append(errs, fmt.Errorf("%s sits outside [Container], where quadlet does not read it; move it into [Container]", want))
		case in > 1:
			errs = append(errs, fmt.Errorf("%s is written %d times; write it once", want, in))
		}
	}
	return errors.Join(errs...)
}

// CheckQuadletKeys reports every [Container] key that podman at floor (such as
// "4.7") does not accept. Quadlet refuses a unit carrying a key it does not know,
// so each key is a version floor of its own, and a newer key fails the unit on
// every older host. A key missing from the release table is reported too: add it
// with the podman release that introduced it, read from
// pkg/systemd/quadlet/quadlet.go at that tag.
func CheckQuadletKeys(unit, floor string) error {
	limit, err := parseRelease(floor)
	if err != nil {
		return err
	}
	since := quadletKeySince()
	var errs []error
	for _, l := range unitLines(unit) {
		t := strings.TrimSpace(l.text)
		if l.section != "[Container]" || t == "" || t[0] == '#' || t[0] == ';' {
			continue
		}
		key, _, ok := strings.Cut(t, "=")
		if !ok {
			errs = append(errs, fmt.Errorf("line %d in [Container] is not KEY=VALUE: %q", l.no, l.text))
			continue
		}
		key = strings.TrimSpace(key)
		rel, known := since[key]
		switch {
		case !known:
			errs = append(errs, fmt.Errorf("line %d: [Container] key %s is not in the release table; add it with the podman release that introduced it, and raise the floor if that is newer than %s", l.no, key, floor))
		case rel[0] > limit[0] || rel[0] == limit[0] && rel[1] > limit[1]:
			errs = append(errs, fmt.Errorf("line %d: [Container] key %s needs podman %d.%d, above the floor %s; a host on an older podman refuses the unit", l.no, key, rel[0], rel[1], floor))
		}
	}
	return errors.Join(errs...)
}

// quadletKeySince maps each [Container] key either tool writes to the podman
// release that first accepts it. Memory is listed although neither tool writes it:
// it set solace-util's floor at 5.5 before the memory cap moved to
// PodmanArgs=--memory=, and it stays so that writing it again fails the check.
func quadletKeySince() map[string][2]int {
	since := map[string][2]int{}
	for _, r := range []struct {
		release [2]int
		keys    []string
	}{
		{[2]int{4, 4}, []string{"Image", "ContainerName", "User", "Group", "NoNewPrivileges", "DropCapability",
			"PodmanArgs", "Network", "PublishPort", "Volume", "Environment", "Label"}},
		{[2]int{4, 5}, []string{"HealthCmd", "HealthInterval", "HealthTimeout", "HealthRetries",
			"HealthStartPeriod", "Secret"}},
		{[2]int{4, 6}, []string{"HostName"}},
		{[2]int{4, 7}, []string{"ShmSize", "Ulimit"}},
		{[2]int{5, 5}, []string{"Memory"}},
	} {
		for _, k := range r.keys {
			since[k] = r.release
		}
	}
	return since
}

// parseRelease reads a MAJOR.MINOR podman release.
func parseRelease(v string) ([2]int, error) {
	major, minor, ok := strings.Cut(v, ".")
	x, err1 := strconv.Atoi(major)
	y, err2 := strconv.Atoi(minor)
	if !ok || err1 != nil || err2 != nil || x < 0 || y < 0 {
		return [2]int{}, fmt.Errorf("podman floor %q is not MAJOR.MINOR; pass the release the tool documents, such as 4.7", v)
	}
	return [2]int{x, y}, nil
}

// CheckSecretTargets reports a secret that is not mounted at an absolute path. A
// bare name or a relative target lands under the engine's own /run/secrets, where
// the application never looks, and nothing reports it: the service just fails to
// authenticate later.
//
// In a compose file it reads only the service-level secrets: block at indent 4.
// Each item must be the long syntax, a source: with a target: that starts with '/'
// once quotes are stripped. A service-level secrets: written in any other shape is
// reported rather than skipped, since a check that cannot read a block would pass
// it: a value on the key's own line (a flow list), items at the key's own indent,
// or a block with no item at indent 6. The top-level secrets: map and inline
// content: blocks are never read. In a quadlet unit it reads every Secret= line in
// [Container], which must carry a target= option that starts with '/'. CRLF line
// endings are read as LF.
func CheckSecretTargets(artifact string) error {
	artifact = strings.ReplaceAll(artifact, "\r\n", "\n")
	items, errs := composeSecretItems(strings.Split(artifact, "\n"))
	for _, item := range items {
		if err := checkComposeSecret(item); err != nil {
			errs = append(errs, err)
		}
	}
	for _, l := range unitLines(artifact) {
		value, ok := strings.CutPrefix(strings.TrimSpace(l.text), "Secret=")
		if !ok || l.section != "[Container]" {
			continue
		}
		opts := strings.Split(value, ",")
		target, found := "", false
		for _, o := range opts[1:] {
			if v, ok := strings.CutPrefix(o, "target="); ok {
				target, found = v, true
			}
		}
		switch {
		case !found:
			errs = append(errs, fmt.Errorf("line %d: Secret=%s has no target=, so podman mounts it under its own /run/secrets; give it an absolute target=", l.no, opts[0]))
		case !strings.HasPrefix(target, "/"):
			errs = append(errs, fmt.Errorf("line %d: Secret=%s mounts at %q, which is not an absolute path; give it an absolute target=", l.no, opts[0], target))
		}
	}
	return errors.Join(errs...)
}

// composeSecretItems returns the items of every service-level secrets: block,
// each as its trimmed lines, and an error for each service-level secrets: it
// cannot read. A block runs from a line at indent 4 that is "secrets:" and nothing
// else to the next non-blank line at indent 4 or less, and an item starts at a
// "- " line at indent 6.
func composeSecretItems(lines []string) ([][]string, []error) {
	var items [][]string
	var errs []error
	in, cur, open := false, -1, 0
	// end closes the open block, reporting one that held no item it could read.
	end := func() {
		if in && cur < 0 {
			errs = append(errs, fmt.Errorf("line %d: the service's secrets: has no item at indent 6, so its targets cannot be checked; "+
				"write each secret as a \"- source:\" item indented under secrets:", open))
		}
		in, cur = false, -1
	}
	for i, l := range lines {
		t := strings.TrimSpace(l)
		indent := len(l) - len(strings.TrimLeft(l, " "))
		switch {
		case indent == 4 && t == "secrets:":
			end()
			in, cur, open = true, -1, i+1
		case indent == 4 && strings.HasPrefix(t, "secrets:"):
			end()
			errs = append(errs, fmt.Errorf("line %d: the service's %s is not a block this check can read, so its targets cannot be checked; "+
				"write secrets: on its own line with one \"- source:\" item per secret", i+1, t))
		case !in || t == "":
		case indent == 4 && strings.HasPrefix(t, "- "):
			errs = append(errs, fmt.Errorf("line %d: the service's secrets: lists %s at the key's own indent, which this check cannot read; "+
				"indent the items under secrets:", i+1, t))
			in, cur = false, -1
		case indent <= 4:
			end()
		case indent == 6 && strings.HasPrefix(t, "- "):
			items = append(items, []string{t})
			cur = len(items) - 1
		case cur >= 0:
			items[cur] = append(items[cur], t)
		}
	}
	end()
	return items, errs
}

// checkComposeSecret checks one service-level secrets: item.
func checkComposeSecret(item []string) error {
	first := strings.TrimPrefix(item[0], "- ")
	name, long := strings.CutPrefix(first, "source:")
	if !long {
		return fmt.Errorf("service secret %s uses the short syntax, which mounts it under the engine's own /run/secrets; write a source: and an absolute target:", first)
	}
	name = strings.TrimSpace(name)
	for _, l := range item[1:] {
		v, ok := strings.CutPrefix(l, "target:")
		if !ok {
			continue
		}
		if v = strings.Trim(strings.TrimSpace(v), `"'`); !strings.HasPrefix(v, "/") {
			return fmt.Errorf("service secret %s mounts at %q, which is not an absolute path; give it an absolute target:", name, v)
		}
		return nil
	}
	return fmt.Errorf("service secret %s has no target:, so it mounts under the engine's own /run/secrets; give it an absolute target:", name)
}

// unitLine is one line of a unit with the section it sits in, such as
// "[Container]", or "" before the first section.
type unitLine struct {
	section string
	text    string
	no      int
}

// unitLines splits a unit into its lines, leaving out the section headers. A
// section runs from its header to the next line that starts with '['.
func unitLines(unit string) []unitLine {
	var out []unitLine
	section := ""
	for i, text := range strings.Split(unit, "\n") {
		if strings.HasPrefix(text, "[") {
			section = strings.TrimSpace(text)
			continue
		}
		out = append(out, unitLine{section: section, text: text, no: i + 1})
	}
	return out
}

// AssertAbsent fails the test when the artifact carries any of the tokens, such
// as Widening() or HostNetwork().
func AssertAbsent(t testing.TB, name, artifact string, tokens []string) {
	t.Helper()
	report(t, name, artifact, CheckAbsent(artifact, tokens))
}

// AssertComposePosture fails the test when the compose file does not state the
// posture block exactly once.
func AssertComposePosture(t testing.TB, name, artifact string) {
	t.Helper()
	report(t, name, artifact, CheckComposePosture(artifact))
}

// AssertQuadletPosture fails the test when the unit does not state each posture
// line exactly once in [Container].
func AssertQuadletPosture(t testing.TB, name, unit string) {
	t.Helper()
	report(t, name, unit, CheckQuadletPosture(unit))
}

// AssertQuadletKeys fails the test when a [Container] key is unknown or newer than
// floor.
func AssertQuadletKeys(t testing.TB, name, unit, floor string) {
	t.Helper()
	report(t, name, unit, CheckQuadletKeys(unit, floor))
}

// AssertSecretTargets fails the test when a secret is not mounted at an absolute
// path.
func AssertSecretTargets(t testing.TB, name, artifact string) {
	t.Helper()
	report(t, name, artifact, CheckSecretTargets(artifact))
}

// report fails the test with the artifact's name, the problems and the artifact.
func report(t testing.TB, name, artifact string, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: %v\n%s", name, err, artifact)
	}
}
