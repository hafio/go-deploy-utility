package build

import (
	"reflect"
	"testing"
)

// TestCmdCopiesBase: the argv is base then args, in a slice of its own, so a later
// change to the caller's base does not reach it, and nothing but the argv is set.
func TestCmdCopiesBase(t *testing.T) {
	base := []string{"kubectl", "--context", "prod"}
	c := Cmd(base, "get", "pods")
	if want := []string{"kubectl", "--context", "prod", "get", "pods"}; !reflect.DeepEqual(c.Argv, want) {
		t.Fatalf("Argv = %q, want %q", c.Argv, want)
	}
	base[0] = "oc"
	if c.Argv[0] != "kubectl" {
		t.Errorf("Argv shares the caller's base: %q", c.Argv)
	}
	if c.Stdin != nil || c.Env != nil {
		t.Errorf("Cmd = %+v, want the argv only", c)
	}
	if got := Cmd(nil, "info"); !reflect.DeepEqual(got.Argv, []string{"info"}) {
		t.Errorf("Cmd(nil, info).Argv = %q", got.Argv)
	}
}

// TestStdinIsNeverNil: even an empty body is a non-nil slice, so the child reads
// EOF at once instead of inheriting no input, and the copy shares nothing.
func TestStdinIsNeverNil(t *testing.T) {
	if got := Stdin(nil); got == nil || len(got) != 0 {
		t.Errorf("Stdin(nil) = %#v, want an empty, non-nil slice", got)
	}
	in := []byte("body")
	got := Stdin(in)
	in[0] = 'X'
	if string(got) != "body" {
		t.Errorf("Stdin shares the caller's bytes: %q", got)
	}
}

// TestShQuote: one sh word with nothing expanded, a single quote included.
func TestShQuote(t *testing.T) {
	for in, want := range map[string]string{
		"/opt/app/conf.yml": `'/opt/app/conf.yml'`,
		"it's":              `'it'\''s'`,
		"$HOME *":           `'$HOME *'`,
		"":                  `''`,
	} {
		if got := ShQuote(in); got != want {
			t.Errorf("ShQuote(%q) = %s, want %s", in, got, want)
		}
	}
}

// TestUploadScript: the cat, and the mkdir of the destination's directory first
// when asked, both quoted.
func TestUploadScript(t *testing.T) {
	for _, c := range []struct {
		dest  string
		mkdir bool
		want  string
	}{
		{"/opt/app/conf.yml", false, `cat > '/opt/app/conf.yml'`},
		{"/opt/app/conf.yml", true, `mkdir -p '/opt/app' && cat > '/opt/app/conf.yml'`},
		{"/tmp/it's here", true, `mkdir -p '/tmp' && cat > '/tmp/it'\''s here'`},
	} {
		if got := UploadScript(c.dest, c.mkdir); got != c.want {
			t.Errorf("UploadScript(%q, %v) = %s, want %s", c.dest, c.mkdir, got, c.want)
		}
	}
}
