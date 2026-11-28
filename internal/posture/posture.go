// Package posture holds the one copy of the container privilege posture and the
// deny lists of the renderers. It is internal, at the module root, so that
// containerart's renderers and the top-level artifacttest can read it and no tool
// can import it: a tool never hand-writes a service with these lines.
//
// The package holds data and two lookups. The posture lines are constants. The
// deny lists come from functions as fresh slices, and the lookups search them, so
// no package variable holds state. A lookup returns a fixed reason sentence for the
// group a key belongs to and never builds an error: the renderer names the key, a
// name that S8 lets it echo, and wraps the result in its own FieldError.
//
// The lists answer S1 of the 2026-11-15 surface review with the additions of its
// answer. A deny list is complete only for the versions it was checked against, so
// ComposeCheckedAgainst and QuadletCheckedAgainst name them; a key a later version
// adds passes until it is listed. The Kubernetes lists and posture lines wait for
// the Kubernetes step.
package posture

import "strings"

// ComposeLines is the compose service's privilege posture, stated rather than
// inherited: privileged off, every capability dropped, and no-new-privileges on, as
// five lines at service indent. compose.Posture is this text.
//
// privileged is a bare boolean because the schema of compose v1 rejects a quoted
// one. no-new-privileges takes the '=' separator because dockerd logs a
// deprecation warning for ':' on every container create.
const ComposeLines = "    privileged: false\n" +
	"    cap_drop:\n" +
	"      - ALL\n" +
	"    security_opt:\n" +
	"      - no-new-privileges=true\n"

// QuadletLines is the quadlet container's privilege posture: no-new-privileges on
// and every capability dropped, as two lines that belong in [Container].
// quadlet.Posture is this text. Quadlet has no Privileged= key, since a container
// is unprivileged unless asked, so there is nothing to state for it.
const QuadletLines = "NoNewPrivileges=true\n" +
	"DropCapability=all\n"

// ComposeCheckedAgainst names the compose release the deny list was written
// against. S1 recalls the lists from the compose specification; this is the
// release a hand check on a real engine confirms (docs/smoke.md).
const ComposeCheckedAgainst = "Docker Compose 2.40"

// QuadletCheckedAgainst names the podman release the deny list was written
// against. S1 recalls the lists from the podman 5 quadlet keys; this is the release
// a hand check on a real engine confirms (docs/smoke.md).
const QuadletCheckedAgainst = "Podman 5.6"

// The reason sentences, one per group. A renderer prints the key and then the
// reason: `compose: extra key "cap_add" reaches the host or another container`.
const (
	reasonOwn   = "is a key M writes itself"
	reasonHost  = "reaches the host or another container"
	reasonFiles = "pulls in files or services"
	reasonFlags = "carries any podman flag, including privilege"
	reasonAuto  = "makes podman replace the image outside the tool and its redeploy question"
	reasonKey   = "is not a plain key: it is empty or holds '=', whitespace or a control character"
)

// entry is one denied key with the reason of its group.
type entry struct{ key, reason string }

// composeEntries is the S1 deny list for compose service extras. Group one is the
// keys M writes, cpuset among them (S1, S2); entrypoint, command and tmpfs are not in
// it since P4 (2026-11-27): M no longer writes them, and a tool passes them as extras.
// shm_size and ulimits are not in it either: B passes them as extras today (S1).
func composeEntries() []entry {
	var list []entry
	for _, k := range []string{
		"image", "pull_policy", "container_name", "hostname", "user", "restart", "privileged", "cap_drop",
		"security_opt", "read_only", "mem_limit", "cpuset", "networks", "network_mode", "ports", "environment",
		"secrets", "configs", "volumes", "healthcheck", "labels",
	} {
		list = append(list, entry{k, reasonOwn})
	}
	for _, k := range []string{
		"cap_add", "devices", "device_cgroup_rules", "sysctls", "pid", "ipc", "uts", "userns_mode", "cgroup",
		"cgroup_parent", "group_add", "runtime", "gpus", "volumes_from", "use_api_socket", "storage_opt",
		"oom_kill_disable", "links", "external_links", "annotations", "logging", "oom_score_adj", "cpu_rt_runtime",
		"cpu_rt_period", "credential_spec",
	} {
		list = append(list, entry{k, reasonHost})
	}
	for _, k := range []string{
		"build", "extends", "env_file", "label_file", "depends_on", "post_start", "pre_stop", "deploy", "provider",
		"develop", "models", "profiles", "scale",
	} {
		list = append(list, entry{k, reasonFiles})
	}
	return list
}

// quadletEntries is the S1 deny list for [Container] extras, in lower case since
// quadlet keys are matched without regard to case.
func quadletEntries() []entry {
	var list []entry
	for _, k := range []string{"podmanargs", "globalargs", "containersconfmodule"} {
		list = append(list, entry{k, reasonFlags})
	}
	for _, k := range []string{
		"addcapability", "adddevice", "sysctl", "securitylabeldisable", "securitylabeltype", "securitylabellevel",
		"securitylabelfiletype", "securitylabelnested", "seccompprofile", "unmask", "userns", "uidmap", "gidmap",
		"subuidmap", "subgidmap", "groupadd", "mount", "environmenthost", "rootfs", "annotation", "cgroupsmode",
		"logdriver", "logopt", "healthlogdestination",
	} {
		list = append(list, entry{k, reasonHost})
	}
	for _, k := range []string{"environmentfile", "pod"} {
		list = append(list, entry{k, reasonFiles})
	}
	for _, k := range []string{
		"image", "pull", "containername", "hostname", "label", "user", "group", "nonewprivileges", "dropcapability",
		"readonly", "network", "publishport", "environment", "secret", "volume", "timezone", "memory", "healthcmd",
		"healthinterval", "healthtimeout", "healthretries", "healthstartperiod",
	} {
		list = append(list, entry{k, reasonOwn})
	}
	return append(list, entry{"autoupdate", reasonAuto})
}

// ComposeExtra reports whether key may not be a compose service extra, and why. The
// match is exact and lower case, as compose reads its keys. The reason is a fixed
// sentence for the key's group; the caller names the key.
func ComposeExtra(key string) (reason string, denied bool) {
	for _, e := range composeEntries() {
		if e.key == key {
			return e.reason, true
		}
	}
	return "", false
}

// QuadletExtra reports whether key may not be a [Container] extra, and why. The
// match ignores case. A key that is empty or holds '=', whitespace or a control
// character is denied with its own reason, since it could start another assignment.
func QuadletExtra(key string) (reason string, denied bool) {
	if key == "" || strings.ContainsFunc(key, func(r rune) bool {
		return r == '=' || r == ' ' || r < 0x20 || r == 0x7f
	}) {
		return reasonKey, true
	}
	lower := strings.ToLower(key)
	for _, e := range quadletEntries() {
		if e.key == lower {
			return e.reason, true
		}
	}
	return "", false
}

// ComposeExtraKeys returns every denied compose key, as a fresh slice, for docs and
// tests.
func ComposeExtraKeys() []string {
	var keys []string
	for _, e := range composeEntries() {
		keys = append(keys, e.key)
	}
	return keys
}

// QuadletExtraKeys returns every denied quadlet key in lower case, as a fresh
// slice, for docs and tests.
func QuadletExtraKeys() []string {
	var keys []string
	for _, e := range quadletEntries() {
		keys = append(keys, e.key)
	}
	return keys
}
