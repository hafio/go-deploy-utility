package posture_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/containerart/compose"
	"github.com/hafio/go-deploy-utility/containerart/quadlet"
	"github.com/hafio/go-deploy-utility/internal/posture"
)

// marker stands for a secret pasted into a key. No reason may echo it.
const marker = "hunter2"

// The three group sentences compose reasons are drawn from, and the quadlet ones
// besides them. The tests name them as text, since the package keeps them private.
const (
	own   = "is a key M writes itself"
	host  = "reaches the host or another container"
	files = "pulls in files or services"
	flags = "carries any podman flag, including privilege"
	auto  = "makes podman replace the image outside the tool and its redeploy question"
)

// TestComposeLinesMatchRenderer: the posture text is the bytes both tools write
// today, and compose.Posture is that one text.
func TestComposeLinesMatchRenderer(t *testing.T) {
	want := "    privileged: false\n    cap_drop:\n      - ALL\n    security_opt:\n      - no-new-privileges=true\n"
	if posture.ComposeLines != want {
		t.Errorf("ComposeLines = %q, want %q", posture.ComposeLines, want)
	}
	if compose.Posture != posture.ComposeLines {
		t.Errorf("compose.Posture = %q, want posture.ComposeLines", compose.Posture)
	}
}

// TestQuadletLinesMatchRenderer: the posture text is the lines both tools write
// today, and quadlet.Posture is that one text.
func TestQuadletLinesMatchRenderer(t *testing.T) {
	if want := "NoNewPrivileges=true\nDropCapability=all\n"; posture.QuadletLines != want {
		t.Errorf("QuadletLines = %q, want %q", posture.QuadletLines, want)
	}
	if quadlet.Posture != posture.QuadletLines {
		t.Errorf("quadlet.Posture = %q, want posture.QuadletLines", quadlet.Posture)
	}
}

// group is the keys of one group of a deny list with the sentence they share.
type group struct {
	reason string
	keys   []string
}

// composeGroups lists the keys of each group by name, so a key that moves between
// groups is a deliberate change.
func composeGroups() []group {
	return []group{
		{own, []string{
			"image", "pull_policy", "container_name", "hostname", "user", "restart", "privileged", "cap_drop",
			"security_opt", "read_only", "mem_limit", "cpuset", "networks", "network_mode", "ports", "environment",
			"secrets", "configs", "volumes", "healthcheck", "labels",
		}},
		{host, []string{
			"cap_add", "devices", "device_cgroup_rules", "sysctls", "pid", "ipc", "uts", "userns_mode", "cgroup",
			"cgroup_parent", "group_add", "runtime", "gpus", "volumes_from", "use_api_socket", "storage_opt",
			"oom_kill_disable", "links", "external_links", "annotations", "logging", "oom_score_adj",
			"cpu_rt_runtime", "cpu_rt_period", "credential_spec",
		}},
		{files, []string{
			"build", "extends", "env_file", "label_file", "depends_on", "post_start", "pre_stop", "deploy", "provider",
			"develop", "models", "profiles", "scale",
		}},
	}
}

// TestComposeExtraDenied: every listed key is denied with the sentence of its
// group, and the exported list holds exactly those keys.
func TestComposeExtraDenied(t *testing.T) {
	var all []string
	for _, g := range composeGroups() {
		for _, k := range g.keys {
			got, denied := posture.ComposeExtra(k)
			if !denied || got != g.reason {
				t.Errorf("ComposeExtra(%q) = %q, %v; want %q, true", k, got, denied, g.reason)
			}
			all = append(all, k)
		}
	}
	listed := posture.ComposeExtraKeys()
	slices.Sort(all)
	slices.Sort(listed)
	if !slices.Equal(all, listed) {
		t.Errorf("ComposeExtraKeys() = %v, want %v", listed, all)
	}
}

// TestComposeExtraAllowed: keys that are not on the list pass, among them B's
// shm_size and ulimits, and A's entrypoint, command and tmpfs (P4). The
// match is exact and lower case, so a case variant of a denied key passes.
func TestComposeExtraAllowed(t *testing.T) {
	for _, k := range []string{
		"shm_size", "ulimits", "entrypoint", "command", "tmpfs", "working_dir", "stop_grace_period",
		"Privileged", "CAP_ADD", "", marker,
	} {
		if reason, denied := posture.ComposeExtra(k); denied {
			t.Errorf("ComposeExtra(%q) denied (%q), want it to pass", k, reason)
		}
	}
}

// quadletGroups lists the keys of each group by name, in lower case.
func quadletGroups() []group {
	return []group{
		{flags, []string{"podmanargs", "globalargs", "containersconfmodule"}},
		{host, []string{
			"addcapability", "adddevice", "sysctl", "securitylabeldisable", "securitylabeltype", "securitylabellevel",
			"securitylabelfiletype", "securitylabelnested", "seccompprofile", "unmask", "userns", "uidmap", "gidmap",
			"subuidmap", "subgidmap", "groupadd", "mount", "environmenthost", "rootfs", "annotation", "cgroupsmode",
			"logdriver", "logopt", "healthlogdestination",
		}},
		{files, []string{"environmentfile", "pod"}},
		{own, []string{
			"image", "pull", "containername", "hostname", "label", "user", "group", "nonewprivileges", "dropcapability",
			"readonly", "network", "publishport", "environment", "secret", "volume", "timezone", "memory", "healthcmd",
			"healthinterval", "healthtimeout", "healthretries", "healthstartperiod",
		}},
		{auto, []string{"autoupdate"}},
	}
}

// TestQuadletExtraDenied: every listed key is denied with the sentence of its
// group, and the exported list holds exactly those keys.
func TestQuadletExtraDenied(t *testing.T) {
	var all []string
	for _, g := range quadletGroups() {
		for _, k := range g.keys {
			got, denied := posture.QuadletExtra(k)
			if !denied || got != g.reason {
				t.Errorf("QuadletExtra(%q) = %q, %v; want %q, true", k, got, denied, g.reason)
			}
			all = append(all, k)
		}
	}
	listed := posture.QuadletExtraKeys()
	slices.Sort(all)
	slices.Sort(listed)
	if !slices.Equal(all, listed) {
		t.Errorf("QuadletExtraKeys() = %v, want %v", listed, all)
	}
	for _, k := range []string{"ShmSize", "Ulimit", "Tmpfs", "Exec", "Entrypoint", "Notify", marker} {
		if reason, denied := posture.QuadletExtra(k); denied {
			t.Errorf("QuadletExtra(%q) denied (%q), want it to pass", k, reason)
		}
	}
}

// TestQuadletExtraCase: quadlet keys are matched without regard to case.
func TestQuadletExtraCase(t *testing.T) {
	for _, k := range []string{"PodmanArgs", "PODMANARGS", "podmanargs", "AddCapability", "autoupdate", "AutoUpdate", "HealthCmd", "Pull"} {
		if _, denied := posture.QuadletExtra(k); !denied {
			t.Errorf("QuadletExtra(%q) passed, want it denied in any case", k)
		}
	}
}

// TestQuadletExtraKeyGrammar: a key with '=', a space, a line break or another
// control character, or an empty one, is denied with its own reason, and the reason
// does not echo the key.
func TestQuadletExtraKeyGrammar(t *testing.T) {
	for _, k := range []string{
		"", "Key=Value", marker + "=x", "Two Words " + marker, "A\nB " + marker, "A\rB " + marker, "A\tB " + marker,
		"A\x00" + marker, "A\x7f" + marker,
	} {
		reason, denied := posture.QuadletExtra(k)
		if !denied || !strings.Contains(reason, "not a plain key") {
			t.Errorf("QuadletExtra(%q) = %q, %v; want the key-grammar reason", k, reason, denied)
		}
		if strings.Contains(reason, marker) {
			t.Errorf("QuadletExtra(%q) echoed the key: %q", k, reason)
		}
	}
}

// TestExtraListsAreFresh: a caller's edit to a returned list does not change the
// next call or a lookup.
func TestExtraListsAreFresh(t *testing.T) {
	c := posture.ComposeExtraKeys()
	first := c[0]
	c[0] = marker
	if posture.ComposeExtraKeys()[0] != first {
		t.Error("ComposeExtraKeys shares its slice between calls")
	}
	if _, denied := posture.ComposeExtra(first); !denied {
		t.Errorf("ComposeExtra(%q) changed after a caller edited the list", first)
	}
	q := posture.QuadletExtraKeys()
	first = q[0]
	q[0] = marker
	if posture.QuadletExtraKeys()[0] != first {
		t.Error("QuadletExtraKeys shares its slice between calls")
	}
	if _, denied := posture.QuadletExtra(first); !denied {
		t.Errorf("QuadletExtra(%q) changed after a caller edited the list", first)
	}
}

// TestCheckedAgainstNamesVersions: S1, each list names the release it was checked
// against, and the name carries a version number.
func TestCheckedAgainstNamesVersions(t *testing.T) {
	for _, v := range []string{posture.ComposeCheckedAgainst, posture.QuadletCheckedAgainst} {
		if !strings.ContainsAny(v, "0123456789") || !strings.Contains(v, ".") {
			t.Errorf("checked-against name %q lacks an engine name and a version such as 2.40", v)
		}
	}
	if !strings.Contains(posture.ComposeCheckedAgainst, "Compose") || !strings.Contains(posture.QuadletCheckedAgainst, "Podman") {
		t.Errorf("the checked-against names are %q and %q, want Compose and Podman named",
			posture.ComposeCheckedAgainst, posture.QuadletCheckedAgainst)
	}
}
