package workload

import (
	"slices"
	"strings"
	"testing"
)

// TestUlimitNames: the 16 Linux names, as a fresh slice.
func TestUlimitNames(t *testing.T) {
	names := UlimitNames()
	if len(names) != 16 {
		t.Errorf("UlimitNames() has %d names, want 16: %v", len(names), names)
	}
	for _, want := range []string{"nofile", "nproc", "memlock", "core", "rtprio", "rttime", "nice"} {
		if !slices.Contains(names, want) {
			t.Errorf("UlimitNames() lacks %q", want)
		}
	}
	names[0] = marker
	if UlimitNames()[0] == marker {
		t.Error("UlimitNames shares its slice between calls")
	}
}

// TestParseLimit: S2. A number or -1, soft not above hard, a single value for both.
func TestParseLimit(t *testing.T) {
	tests := []struct {
		in   string
		want Ulimit
	}{
		{"nofile=65536:65536", Ulimit{"nofile", 65536, 65536}},
		{"nofile=1024", Ulimit{"nofile", 1024, 1024}},
		{"nofile=0:0", Ulimit{"nofile", 0, 0}},
		{"memlock=-1", Ulimit{"memlock", -1, -1}},
		{"memlock=-1:-1", Ulimit{"memlock", -1, -1}},
		{"core=0", Ulimit{"core", 0, 0}},
		{"nproc=4096:-1", Ulimit{"nproc", 4096, -1}},
		{"as=1:2", Ulimit{"as", 1, 2}},
	}
	for _, tc := range tests {
		got, f := ParseLimit(tc.in)
		if f != nil || got != tc.want {
			t.Errorf("ParseLimit(%q) = %+v, %v; want %+v", tc.in, got, f, tc.want)
		}
	}
	for _, name := range UlimitNames() {
		_, f := ParseLimit(name + "=1")
		refused := name == "nice" || name == "rtprio" || name == "rttime"
		if refused != (f != nil) {
			t.Errorf("ParseLimit(%q) = %v, refused should be %v", name+"=1", f, refused)
		}
	}
}

// TestParseLimitRefuses: bad shape, names, values and order. A name from the fixed
// list is echoed when refused; nothing else is.
func TestParseLimitRefuses(t *testing.T) {
	for _, in := range []string{
		"", "nofile 1024", "nofile=", "nofile=abc " + marker, "bogus=1", marker + "=1", "NOFILE=1", "nofile=2:1",
		"nofile=-1:5", "nofile=-2", "nofile=1:2:3", "nofile=+1", "nofile=1.5", "nofile=99999999999999999999",
		"nofile=1:", "nofile=:1", "nofile=1k", "nofile=infinity", "nofile=" + marker, "nofile=0x10", "=1",
	} {
		_, f := ParseLimit(in)
		expectRefusal(t, "ParseLimit", in, f, true)
	}
	for _, name := range []string{"nice", "rtprio", "rttime"} {
		_, f := ParseLimit(name + "=1")
		if f == nil || f.Value != name {
			t.Errorf("ParseLimit(%q) = %v, want a refusal naming the limit", name+"=1", f)
		}
	}
	_, f := ParseLimit("nofile=2:1")
	if f == nil || f.Value != "nofile" || !strings.Contains(f.Rule, "soft") {
		t.Errorf("ParseLimit(nofile=2:1) = %v, want the name and the soft-above-hard rule", f)
	}
}

// TestUlimitCheck: the sink rules, and a struct built by hand is held to the same
// rules as a parsed one.
func TestUlimitCheck(t *testing.T) {
	ok := Ulimit{Name: "nofile", Soft: 1024, Hard: 4096}
	for _, sink := range []Sink{SinkCompose, SinkQuadlet} {
		if f := ok.Check(sink); f != nil {
			t.Errorf("Check(%v) refused: %v", sink, f)
		}
	}
	if f := ok.Check(SinkKubernetes); f == nil || f.Field != "ulimit" || !strings.Contains(f.Rule, "Kubernetes") {
		t.Errorf("Check(Kubernetes) = %v, want a refusal naming Kubernetes", f)
	}
	for _, bad := range []Ulimit{
		{"nofile", 4096, 1024},
		{"nofile", -1, 1024},
		{"nofile", -2, -1},
		{"nofile", 1, -2},
		{"rtprio", 1, 1},
		{"", 1, 1},
		{marker, 1, 1},
	} {
		expectRefusal(t, "Ulimit.Check", bad.Name, bad.Check(SinkCompose), false)
	}
}

// sinkSize wraps ValidMemory and ValidShmSize at one sink for the shared rows.
func sinkSize(check func(Sink, string) *FieldError, sink Sink) func(string) *FieldError {
	return func(s string) *FieldError { return check(sink, s) }
}

// TestValidMemoryAndShmSize: S2. A number with an optional b, k, m or g in either
// case on compose and quadlet; Kubernetes waits for E12.
func TestValidMemoryAndShmSize(t *testing.T) {
	accept := []string{"512", "6898m", "8g", "8G", "1024k", "1024K", "64b", "64B", "1M", "0", "007m"}
	reject := []string{
		"", "g", "8gb", "-1", "+1", "1.5g", "8 g", " 8g", "0x10", marker, "8g " + marker, "8t", "1" + strings.Repeat("0", 18),
	}
	for name, check := range map[string]func(Sink, string) *FieldError{"ValidMemory": ValidMemory, "ValidShmSize": ValidShmSize} {
		for _, sink := range []Sink{SinkCompose, SinkQuadlet} {
			rows(t, name, sinkSize(check, sink), accept, reject)
		}
		f := check(SinkKubernetes, "8g")
		if f == nil || !strings.Contains(f.Rule, "Kubernetes") || !strings.Contains(f.Rule, "E12") {
			t.Errorf("%s on Kubernetes = %v, want a refusal that says it waits for E12", name, f)
		}
	}
}

// TestValidCPUSet: lists of CPUs and ranges.
func TestValidCPUSet(t *testing.T) {
	rows(t, "ValidCPUSet", ValidCPUSet,
		[]string{"0", "0-3", "0-3,5", "0,2,4", "1-1", "0-3,8-11", "12"},
		[]string{
			"", "3-0", "0-", "-1", ",1", "1,", "1,,2", "0 - 3", "a", "0-3," + marker, "1-2-3", "1234567", "0-1234567",
			"+1", "0.5", marker,
		})
}
