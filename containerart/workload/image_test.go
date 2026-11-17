package workload

import (
	"strings"
	"testing"
)

// TestParseImage: every accepted form splits as E3 says and writes back as it came.
func TestParseImage(t *testing.T) {
	digest := "sha256:" + strings.Repeat("ab12", 16)
	tests := []struct {
		in   string
		want Image
	}{
		{"ubuntu:22.04", Image{"", "ubuntu", "22.04", ""}},
		{"solace/pubsubplus-standard:10.26.0", Image{"", "solace/pubsubplus-standard", "10.26.0", ""}},
		{"registry.example.com/solace-pubsub-standard:latest", Image{"registry.example.com", "solace-pubsub-standard", "latest", ""}},
		{"ghcr.io/solace/pubsubplus-standard:10.26.0", Image{"ghcr.io", "solace/pubsubplus-standard", "10.26.0", ""}},
		{"localhost:5000/team/app:1", Image{"localhost:5000", "team/app", "1", ""}},
		{"reg.example.com:443/app:v1", Image{"reg.example.com:443", "app", "v1", ""}},
		{"ghcr.io/solace/app@" + digest, Image{"ghcr.io", "solace/app", "", digest}},
		{"x/y--z:1", Image{"", "x/y--z", "1", ""}},
		{"x/y__z:1", Image{"", "x/y__z", "1", ""}},
		{"ghcr.io/solace/app:1@" + digest, Image{"ghcr.io", "solace/app", "1", digest}},
		{"my_org/a.b-c/d:_x", Image{"", "my_org/a.b-c/d", "_x", ""}},
		{"a:" + strings.Repeat("a", 128), Image{"", "a", strings.Repeat("a", 128), ""}},
	}
	for _, tc := range tests {
		got, f := ParseImage(tc.in)
		if f != nil {
			t.Errorf("ParseImage(%q) refused: %v", tc.in, f)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseImage(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
		if got.String() != tc.in {
			t.Errorf("Image.String() = %q, want %q", got.String(), tc.in)
		}
	}
}

// TestParseImageRefuses: no tag or digest, a host in the repository, bad grammar
// in each part, and S8: no error repeats the reference.
func TestParseImageRefuses(t *testing.T) {
	reject := []string{
		"", "ubuntu", "ghcr.io/solace/app", "ubuntu:", "ghcr.io/solace/app@", "Ubuntu:1", "ghcr.io/Solace/app:1",
		marker + ".example.com:1", marker + "/App:1", "ghcr.io/" + marker + " x:1", marker + " x:1",
		"a//b:1", "ghcr.io/:1", "a/b:1:2", "ubuntu:-" + marker, "ubuntu:.x", "ubuntu:" + strings.Repeat("a", 129),
		"ubuntu@sha256:abc", "ubuntu@md5:" + strings.Repeat("a", 32), "ubuntu@sha256:" + strings.Repeat("A", 64),
		"ghcr.io:0/x:1", "ghcr.io:70000/x:1", "ghcr.io:/x:1", "-a.io/x:1", "a..b/x:1", "a.b-/x:1",
		":5000/x:1", strings.Repeat("a.", 127) + "io/x:1", "reg_x.io/x:1",
		"x/y_:1", "x/-y:1", "x/y..z:1", "x/y___z:1", "x/y-_z:1", "ghcr.io/solace/" + marker + "!:1",
	}
	for _, in := range reject {
		_, f := ParseImage(in)
		expectRefusal(t, "ParseImage", in, f, true)
	}
}
