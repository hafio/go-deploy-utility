package workload

import "testing"

// TestValidRestart: E6 and ruling 9, the words per sink and the empty word.
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
		{SinkKubernetes, "always", true},
		{SinkKubernetes, "no", false},
		{SinkKubernetes, "on-failure", false},
		{SinkKubernetes, "unless-stopped", false},
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
	// The exported checks refuse an invalid sink first (checkSink); the lists
	// still answer nil for one rather than another sink's words.
	if w := restartWords(Sink(0)); w != nil {
		t.Errorf("restartWords(Sink(0)) = %v, want nil", w)
	}
}
