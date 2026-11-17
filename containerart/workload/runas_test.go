package workload

import "testing"

// TestParseRunAs: uid:gid in decimal, and the empty string is not stated.
func TestParseRunAs(t *testing.T) {
	tests := []struct {
		in       string
		uid, gid uint32
		stated   bool
		str      string
	}{
		{"1000:0", 1000, 0, true, "1000:0"},
		{"0:0", 0, 0, true, "0:0"},
		{"1000001:0", 1000001, 0, true, "1000001:0"},
		{"2147483647:2147483647", 2147483647, 2147483647, true, "2147483647:2147483647"},
		{"0:1", 0, 1, true, "0:1"},
		{"", 0, 0, false, ""},
	}
	for _, tc := range tests {
		r, f := ParseRunAs(tc.in)
		if f != nil {
			t.Errorf("ParseRunAs(%q) refused: %v", tc.in, f)
			continue
		}
		if r.Stated() != tc.stated || r.UID() != tc.uid || r.GID() != tc.gid || r.String() != tc.str {
			t.Errorf("ParseRunAs(%q) = stated %v %d:%d %q, want stated %v %d:%d %q",
				tc.in, r.Stated(), r.UID(), r.GID(), r.String(), tc.stated, tc.uid, tc.gid, tc.str)
		}
	}
}

// TestParseRunAsRefuses: names, a bare uid, signs, leading zeros and range. Only an
// out-of-range number may be echoed.
func TestParseRunAsRefuses(t *testing.T) {
	for _, in := range []string{
		"1000", "solace:solace", "solace:" + marker, marker + ":0", "1000:", ":0", ":", "+1:0", "-1:0", "01:0", "1:00",
		"1:2:3", "0x10:0", "1000:0 " + marker, "image", "default", "99999999999:0",
	} {
		_, f := ParseRunAs(in)
		expectRefusal(t, "ParseRunAs", in, f, true)
	}
	// The rule text's own example is 1000:0, so these two are held only to the
	// refusal, not to the echo check.
	for _, in := range []string{" 1000:0", "1000:0 "} {
		_, f := ParseRunAs(in)
		expectRefusal(t, "ParseRunAs", in, f, false)
	}
	for _, in := range []string{"2147483648:0", "0:2147483648"} {
		_, f := ParseRunAs(in)
		expectRefusal(t, "ParseRunAs", in, f, false)
		if f != nil && f.Value != "2147483648" {
			t.Errorf("ParseRunAs(%q).Value = %q, want the out-of-range number", in, f.Value)
		}
	}
}

// TestRunAsIDs: numbers in range are stated, including 0 and 0.
func TestRunAsIDs(t *testing.T) {
	r, f := RunAsIDs(1000, 0)
	if f != nil || !r.Stated() || r.String() != "1000:0" {
		t.Errorf("RunAsIDs(1000, 0) = %v, %v", r, f)
	}
	if _, f := RunAsIDs(2147483648, 0); f == nil {
		t.Error("RunAsIDs accepted a uid above 2147483647")
	}
	if _, f := RunAsIDs(0, 2147483648); f == nil {
		t.Error("RunAsIDs accepted a gid above 2147483647")
	}
}

// TestRunAsRootIsStatedNotEmpty: S3. A stated 0:0 differs from the zero value, so a
// renderer can write the root lines for one and nothing for the other.
func TestRunAsRootIsStatedNotEmpty(t *testing.T) {
	root, f := RunAsIDs(0, 0)
	if f != nil {
		t.Fatalf("RunAsIDs(0, 0) refused: %v", f)
	}
	parsed, _ := ParseRunAs("0:0")
	if !root.Stated() || !parsed.Stated() || root != parsed {
		t.Errorf("stated root differs by construction: %+v and %+v", root, parsed)
	}
	if (RunAs{}).Stated() || root == (RunAs{}) {
		t.Error("the zero RunAs is stated, or equals a stated root")
	}
}
