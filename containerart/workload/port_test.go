package workload

import "testing"

// TestParsePort: E7. A bare port or range means N:N, a bind address needs both
// sides, IPv6 is bracketed, and String always writes both sides.
func TestParsePort(t *testing.T) {
	tests := []struct {
		in   string
		want Port
		str  string
	}{
		{"8080", Port{"", "8080", "8080", ""}, "8080:8080"},
		{"1", Port{"", "1", "1", ""}, "1:1"},
		{"65535", Port{"", "65535", "65535", ""}, "65535:65535"},
		{"8000-8010", Port{"", "8000-8010", "8000-8010", ""}, "8000-8010:8000-8010"},
		{"8080/udp", Port{"", "8080", "8080", "udp"}, "8080:8080/udp"},
		{"8080/tcp", Port{"", "8080", "8080", "tcp"}, "8080:8080/tcp"},
		{"80:8080", Port{"", "80", "8080", ""}, "80:8080"},
		{"127.0.0.1:80:8080/tcp", Port{"127.0.0.1", "80", "8080", "tcp"}, "127.0.0.1:80:8080/tcp"},
		{"0.0.0.0:80:80", Port{"0.0.0.0", "80", "80", ""}, "0.0.0.0:80:80"},
		{"[::1]:80:8080", Port{"::1", "80", "8080", ""}, "[::1]:80:8080"},
		{"[2001:db8::1]:8000-8010:9000-9010/udp", Port{"2001:db8::1", "8000-8010", "9000-9010", "udp"}, "[2001:db8::1]:8000-8010:9000-9010/udp"},
		{"8000-8010:9000-9010", Port{"", "8000-8010", "9000-9010", ""}, "8000-8010:9000-9010"},
		{"8000-8010:9000", Port{"", "8000-8010", "9000", ""}, "8000-8010:9000"},
		{"8000-8000:9000-9000", Port{"", "8000-8000", "9000-9000", ""}, "8000-8000:9000-9000"},
	}
	for _, tc := range tests {
		got, f := ParsePort(tc.in)
		if f != nil {
			t.Errorf("ParsePort(%q) refused: %v", tc.in, f)
			continue
		}
		if got != tc.want {
			t.Errorf("ParsePort(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
		if got.String() != tc.str {
			t.Errorf("Port.String() for %q = %q, want %q", tc.in, got.String(), tc.str)
		}
	}
}

// TestParsePortRefuses: the boundary rows, ranges of unequal size, bad addresses
// and protocols. The marker is never echoed.
func TestParsePortRefuses(t *testing.T) {
	for _, in := range []string{
		"", "0", "65536", "99999999999", "abc", "80:", ":80", "8080:80:70:60", "::1:80:80", "[::1]:80", "[::1:80:80",
		"[1.2.3.4]:80:80", "1.2.3.4.5:80:80", marker + ":80:80", "127.0.0.1:80:80/sctp", "8080/", "8080/" + marker,
		"8010-8000", "8000-8010:9000-9005", "8000-8005:9000-9010", "8000:9000-9010", "08080", "+80", "80-", "-80",
		"1-2-3", "80:" + marker, "[fe80::1%eth0]:80:80", ":80:80", "1.2.3.4::80",
	} {
		_, f := ParsePort(in)
		expectRefusal(t, "ParsePort", in, f, false)
	}
}

// TestParsePortEchoesOnlyNumbers: an out-of-range number is the one value the
// error prints.
func TestParsePortEchoesOnlyNumbers(t *testing.T) {
	_, f := ParsePort("65536")
	if f == nil || f.Value != "65536" {
		t.Errorf("ParsePort(65536) = %v, want the number echoed", f)
	}
	_, f = ParsePort("0")
	if f == nil || f.Value != "0" {
		t.Errorf("ParsePort(0) = %v, want the number echoed", f)
	}
	_, f = ParsePort("80:" + marker)
	if f == nil || f.Value != "" {
		t.Errorf("ParsePort with text = %v, want no value", f)
	}
}
