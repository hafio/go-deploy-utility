package workload

import (
	"testing"
	"time"
)

// TestParseSeconds: S6. A Go duration of whole seconds becomes its seconds;
// a fraction, a bare number and text are refused.
func TestParseSeconds(t *testing.T) {
	tests := []struct {
		in   string
		want int64
	}{
		{"30s", 30},
		{"60s", 60},
		{"1m30s", 90},
		{"2m", 120},
		{"1h", 3600},
		{"0s", 0},
		{"0", 0},
		{"1.0s", 1},
		{"-5s", -5},
	}
	for _, tc := range tests {
		got, f := ParseSeconds(tc.in)
		if f != nil || got != tc.want {
			t.Errorf("ParseSeconds(%q) = %d, %v; want %d", tc.in, got, f, tc.want)
		}
	}
	for _, in := range []string{"1.5s", "500ms", "1m0.5s", "1ns", "", "90", "abc " + marker, marker, "1s " + marker, "s", "--1s"} {
		_, f := ParseSeconds(in)
		expectRefusal(t, "ParseSeconds", in, f, true)
	}
}

// healthy is a valid probe: A's constant readiness check.
func healthy() Health {
	return Health{
		Test:        []string{"curl", "-fs", "http://localhost:5550/health-check/readiness"},
		Interval:    30 * time.Second,
		Timeout:     5 * time.Second,
		StartPeriod: 60 * time.Second,
		Retries:     3,
	}
}

// TestHealthCheck: S6 and Q6 per sink. field is the Field of the expected refusal,
// and empty means the probe is accepted.
func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name   string
		sink   Sink
		mutate func(*Health)
		field  string
	}{
		{"compose accepts", SinkCompose, func(h *Health) {}, ""},
		{"quadlet accepts", SinkQuadlet, func(h *Health) {}, ""},
		{"kubernetes accepts", SinkKubernetes, func(h *Health) {}, ""},
		{"quadlet status script", SinkQuadlet, func(h *Health) { h.Test = []string{"sh", "/app/external/.status-script", "--health"} }, ""},
		{"start period zero", SinkCompose, func(h *Health) { h.StartPeriod = 0 }, ""},
		{"retries one", SinkCompose, func(h *Health) { h.Retries = 1 }, ""},
		{"no test", SinkCompose, func(h *Health) { h.Test = nil }, "health.test"},
		{"empty token", SinkCompose, func(h *Health) { h.Test = []string{"curl", ""} }, "health.test[1]"},
		{"control character", SinkCompose, func(h *Health) { h.Test = []string{"curl", "a\x00" + marker} }, "health.test[1]"},
		{"invisible character", SinkKubernetes, func(h *Health) { h.Test = []string{"curl", "a\u200b" + marker} }, "health.test[1]"},
		{"compose keeps a shell probe", SinkCompose, func(h *Health) { h.Test = []string{"sh", "-c", "curl x || exit 1"} }, ""},
		{"quadlet refuses a shell probe", SinkQuadlet, func(h *Health) { h.Test = []string{"sh", "-c", "curl x || exit 1"} }, "health.test[2]"},
		{"interval zero", SinkCompose, func(h *Health) { h.Interval = 0 }, "health.interval"},
		{"interval negative", SinkCompose, func(h *Health) { h.Interval = -30 * time.Second }, "health.interval"},
		{"interval fraction", SinkKubernetes, func(h *Health) { h.Interval = 1500 * time.Millisecond }, "health.interval"},
		{"timeout zero", SinkQuadlet, func(h *Health) { h.Timeout = 0 }, "health.timeout"},
		{"timeout fraction", SinkQuadlet, func(h *Health) { h.Timeout = 500 * time.Millisecond }, "health.timeout"},
		{"start period negative", SinkCompose, func(h *Health) { h.StartPeriod = -time.Second }, "health.start-period"},
		{"start period fraction", SinkCompose, func(h *Health) { h.StartPeriod = 2500 * time.Millisecond }, "health.start-period"},
		{"retries zero", SinkCompose, func(h *Health) { h.Retries = 0 }, "health.retries"},
		{"retries negative", SinkQuadlet, func(h *Health) { h.Retries = -4 }, "health.retries"},
	}
	for _, tc := range tests {
		h := healthy()
		tc.mutate(&h)
		f := h.Check(tc.sink)
		if tc.field == "" {
			if f != nil {
				t.Errorf("%s: refused: %v", tc.name, f)
			}
			continue
		}
		if f == nil {
			t.Errorf("%s: accepted, want a refusal of %s", tc.name, tc.field)
			continue
		}
		if f.Field != tc.field {
			t.Errorf("%s: Field = %q, want %q", tc.name, f.Field, tc.field)
		}
		expectRefusal(t, tc.name, "", f, false)
	}
}

// TestHealthRetriesEchoesTheNumber: retries is a number, so S8 lets it show.
func TestHealthRetriesEchoesTheNumber(t *testing.T) {
	h := healthy()
	h.Retries = -4
	f := h.Check(SinkQuadlet)
	if f == nil || f.Value != "-4" {
		t.Errorf("Check with retries -4 = %v, want the number echoed", f)
	}
}

// TestHealthQuadletRefusesShellCharacters: Q6. Each character of the list, and
// whitespace, refuses on quadlet and passes on compose and Kubernetes.
func TestHealthQuadletRefusesShellCharacters(t *testing.T) {
	for _, c := range ";|&$<>()`'\"\\*?[]{}~!# " {
		h := healthy()
		h.Test = []string{"curl", "a" + string(c) + "b"}
		if f := h.Check(SinkQuadlet); f == nil || f.Field != "health.test[1]" {
			t.Errorf("quadlet accepted a token holding %q: %v", c, f)
		}
		for _, sink := range []Sink{SinkCompose, SinkKubernetes} {
			if f := h.Check(sink); f != nil {
				t.Errorf("%v refused a token holding %q: %v", sink, c, f)
			}
		}
	}
}
