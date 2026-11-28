package workload

import (
	"slices"
	"strings"
	"testing"
)

// TestValidRestart: E6, ruling 9 and P9, the words per sink and the empty word.
func TestValidRestart(t *testing.T) {
	tests := []struct {
		sink Sink
		w    string
		ok   bool
	}{
		{SinkCompose, "", true},
		{SinkCompose, "no", true},
		{SinkCompose, "always", true},
		{SinkCompose, "unless-stopped", true},
		{SinkCompose, "on-failure", true},
		{SinkCompose, "on-failure:3", true},
		{SinkCompose, "on-failure:2147483647", true},
		{SinkCompose, "on-success", false},
		{SinkCompose, "on-failure: 3", false},
		{SinkCompose, "on-failure:0", false},
		{SinkCompose, "on-failure:-1", false},
		{SinkCompose, "on-failure:x", false},
		{SinkCompose, "Always", false},
		{SinkCompose, "always " + marker, false},
		{SinkQuadlet, "", true},
		{SinkQuadlet, "no", true},
		{SinkQuadlet, "on-success", true},
		{SinkQuadlet, "on-failure", true},
		{SinkQuadlet, "on-abnormal", true},
		{SinkQuadlet, "on-watchdog", true},
		{SinkQuadlet, "on-abort", true},
		{SinkQuadlet, "always", true},
		{SinkQuadlet, "unless-stopped", false},
		{SinkQuadlet, "on-failure:3", false},
		{SinkQuadlet, marker, false},
		{SinkKubernetes, "", true},
		{SinkKubernetes, "Always", true},
		{SinkKubernetes, "OnFailure", true},
		{SinkKubernetes, "Never", true},
		{SinkKubernetes, "always", false},
		{SinkKubernetes, "no", false},
		{SinkKubernetes, "on-failure", false},
		{SinkKubernetes, "unless-stopped", false},
		{SinkKubernetes, "Never " + marker, false},
		{SinkKubernetes, marker, false},
	}
	for _, tc := range tests {
		f := ValidRestart(tc.sink, tc.w)
		if tc.ok {
			if f != nil {
				t.Errorf("ValidRestart(%v, %q) refused: %v", tc.sink, tc.w, f)
			}
			continue
		}
		expectRefusal(t, "ValidRestart", tc.w, f, true)
	}
}

// TestValidKubeRestart: P9. Every kind against every word, the empty word, and an
// unknown kind refused before the word is read.
func TestValidKubeRestart(t *testing.T) {
	allowed := [][]string{
		{"Deployment", "Always"},
		{"StatefulSet", "Always"},
		{"ReplicaSet", "Always"},
		{"DaemonSet", "Always"},
		{"Job", "OnFailure", "Never"},
		{"CronJob", "OnFailure", "Never"},
		{"Pod", "Always", "OnFailure", "Never"},
	}
	if len(allowed) != len(kubeKinds()) {
		t.Fatalf("kubeKinds() has %d kinds, the table has %d", len(kubeKinds()), len(allowed))
	}
	for _, row := range allowed {
		kind, words := row[0], row[1:]
		if !slices.Contains(kubeKinds(), kind) {
			t.Errorf("kubeKinds() lacks %q", kind)
		}
		for _, w := range []string{"Always", "OnFailure", "Never", "always", "no", marker} {
			f := ValidKubeRestart(kind, w)
			if slices.Contains(words, w) {
				if f != nil {
					t.Errorf("ValidKubeRestart(%q, %q) refused: %v", kind, w, f)
				}
				continue
			}
			expectRefusal(t, "ValidKubeRestart", w, f, true)
			if f != nil && f.Value != kind {
				t.Errorf("ValidKubeRestart(%q, %q) = %v, want the kind named", kind, w, f)
			}
		}
		if f := ValidKubeRestart(kind, ""); f != nil {
			t.Errorf("ValidKubeRestart(%q, empty) refused: %v", kind, f)
		}
	}
	for _, kind := range []string{"", "deployment", "Service", marker + "Kind"} {
		for _, w := range []string{"", "Always"} {
			f := ValidKubeRestart(kind, w)
			expectRefusal(t, "ValidKubeRestart kind", kind, f, true)
			if f != nil && !strings.Contains(f.Rule, "CronJob") {
				t.Errorf("ValidKubeRestart(%q, %q) = %v, want the seven kinds named", kind, w, f)
			}
		}
	}
}

// TestValidPullPolicy: the words of each sink in exact case, the every_ grammar
// with its boundaries, build, and an invalid sink.
func TestValidPullPolicy(t *testing.T) {
	tests := []struct {
		sink Sink
		w    string
		ok   bool
	}{
		{SinkCompose, "", true},
		{SinkCompose, "always", true},
		{SinkCompose, "missing", true},
		{SinkCompose, "never", true},
		{SinkCompose, "daily", true},
		{SinkCompose, "weekly", true},
		{SinkCompose, "if_not_present", true},
		{SinkCompose, "every_30s", true},
		{SinkCompose, "every_1d12h", true},
		{SinkCompose, "every_2w3d4h5m6s", true},
		{SinkCompose, "every_999999999s", true},
		{SinkCompose, "every_1", false},
		{SinkCompose, "every_d", false},
		{SinkCompose, "every_1x", false},
		{SinkCompose, "every_1d12", false},
		{SinkCompose, "every_1000000000s", false},
		{SinkCompose, "every_1.5h", false},
		{SinkCompose, "every_-1h", false},
		{SinkCompose, "every_1h " + marker, false},
		{SinkCompose, "Always", false},
		{SinkCompose, "IfNotPresent", false},
		{SinkCompose, "newer", false},
		{SinkCompose, "build", false},
		{SinkCompose, marker, false},
		{SinkQuadlet, "", true},
		{SinkQuadlet, "always", true},
		{SinkQuadlet, "missing", true},
		{SinkQuadlet, "never", true},
		{SinkQuadlet, "newer", true},
		{SinkQuadlet, "daily", false},
		{SinkQuadlet, "every_1d", false},
		{SinkQuadlet, "Never", false},
		{SinkQuadlet, "build", false},
		{SinkQuadlet, marker, false},
		{SinkKubernetes, "", true},
		{SinkKubernetes, "Always", true},
		{SinkKubernetes, "IfNotPresent", true},
		{SinkKubernetes, "Never", true},
		{SinkKubernetes, "always", false},
		{SinkKubernetes, "ifnotpresent", false},
		{SinkKubernetes, "missing", false},
		{SinkKubernetes, "build", false},
		{SinkKubernetes, marker, false},
		{Sink(0), "always", false},
		{Sink(4), "", false},
	}
	for _, tc := range tests {
		f := ValidPullPolicy(tc.sink, tc.w)
		if tc.ok {
			if f != nil {
				t.Errorf("ValidPullPolicy(%v, %q) refused: %v", tc.sink, tc.w, f)
			}
			continue
		}
		expectRefusal(t, "ValidPullPolicy", tc.w, f, true)
	}
}

// TestValidPullPolicyEchoes: build is a fixed word and is echoed, with its own
// sentence, on compose; any other refused word is not.
func TestValidPullPolicyEchoes(t *testing.T) {
	f := ValidPullPolicy(SinkCompose, "build")
	if f == nil || f.Value != "build" || !strings.Contains(f.Rule, "never builds") {
		t.Errorf("ValidPullPolicy(compose, build) = %v, want the word echoed with the never-builds sentence", f)
	}
	for _, tc := range []struct {
		sink Sink
		w    string
	}{
		{SinkCompose, "pull " + marker},
		{SinkCompose, "every_" + marker},
		{SinkCompose, "every_"},
		{SinkQuadlet, "build"},
		{SinkKubernetes, "build"},
		{SinkKubernetes, marker},
	} {
		f := ValidPullPolicy(tc.sink, tc.w)
		if f == nil || f.Value != "" {
			t.Errorf("ValidPullPolicy(%v, %q) = %v, want a refusal that echoes nothing", tc.sink, tc.w, f)
		}
	}
}

// TestValidServiceType: the three types in exact case, ExternalName with its own
// sentence, the empty word as an error, and anything else refused without echo.
func TestValidServiceType(t *testing.T) {
	rows(t, "ValidServiceType", ValidServiceType,
		[]string{"ClusterIP", "NodePort", "LoadBalancer"},
		[]string{"", "clusterip", "CLUSTERIP", "nodeport", "Headless", "ClusterIP " + marker, marker})
	f := ValidServiceType("ExternalName")
	if f == nil || f.Value != "ExternalName" || !strings.Contains(f.Rule, "external name") {
		t.Errorf("ValidServiceType(ExternalName) = %v, want the word echoed with its own sentence", f)
	}
	f = ValidServiceType("")
	if f == nil || !strings.Contains(f.Rule, "no default") {
		t.Errorf("ValidServiceType(empty) = %v, want the no-default sentence", f)
	}
	if f := ValidServiceType(marker); f == nil || f.Value != "" {
		t.Errorf("ValidServiceType(marker) = %v, want a refusal that echoes nothing", f)
	}
}

// TestValidNetwork: the words per sink, the empty word, and Kubernetes refusing
// every word.
func TestValidNetwork(t *testing.T) {
	tests := []struct {
		sink Sink
		w    string
		ok   bool
	}{
		{SinkCompose, "", true},
		{SinkCompose, "bridge", true},
		{SinkCompose, "host", true},
		{SinkCompose, "none", true},
		{SinkCompose, "private", false},
		{SinkCompose, "pasta", false},
		{SinkCompose, "container:" + marker, false},
		{SinkQuadlet, "", true},
		{SinkQuadlet, "private", true},
		{SinkQuadlet, "bridge", true},
		{SinkQuadlet, "pasta", true},
		{SinkQuadlet, "slirp4netns", true},
		{SinkQuadlet, "host", true},
		{SinkQuadlet, "none", true},
		{SinkQuadlet, "container", false},
		{SinkQuadlet, "ns:" + marker, false},
		{SinkKubernetes, "", true},
		{SinkKubernetes, "bridge", false},
		{SinkKubernetes, "host", false},
		{SinkKubernetes, marker, false},
	}
	for _, tc := range tests {
		f := ValidNetwork(tc.sink, tc.w)
		if tc.ok {
			if f != nil {
				t.Errorf("ValidNetwork(%v, %q) refused: %v", tc.sink, tc.w, f)
			}
			continue
		}
		expectRefusal(t, "ValidNetwork", tc.w, f, true)
	}
}

// TestWordListsAreFresh: the lists behind the checks are built per call, so a
// caller or a test changing one cannot change a later answer.
func TestWordListsAreFresh(t *testing.T) {
	a := restartWords(SinkQuadlet)
	a[0] = marker
	if restartWords(SinkQuadlet)[0] != "no" {
		t.Error("restartWords shares its slice between calls")
	}
	b := networkWords(SinkCompose)
	b[0] = marker
	if networkWords(SinkCompose)[0] != "bridge" {
		t.Error("networkWords shares its slice between calls")
	}
	for i, list := range []func() []string{
		kubeKinds,
		serviceTypes,
		kubeSuffixes,
		func() []string { return kubeRestartWords("Pod") },
		func() []string { return pullWords(SinkCompose) },
	} {
		first := list()[0]
		l := list()
		l[0] = marker
		if list()[0] != first {
			t.Errorf("word list %d shares its slice between calls", i)
		}
	}
	// The exported checks refuse an invalid sink first (checkSink); the lists
	// still answer nil for one rather than another sink's words.
	if w := restartWords(Sink(0)); w != nil {
		t.Errorf("restartWords(Sink(0)) = %v, want nil", w)
	}
	if w := pullWords(Sink(0)); w != nil {
		t.Errorf("pullWords(Sink(0)) = %v, want nil", w)
	}
}
