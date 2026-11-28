package writer_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/containerart/writer"
)

// marker stands for a secret pasted into a value or a key. No error may echo it.
const marker = "hunter2"

// expectRefusal fails unless err is a refusal that repeats neither the marker nor
// the whole input in. It also requires that every joined error is a *writer.Error
// with a path and a rule.
func expectRefusal(t *testing.T, what, in string, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("%s(%q) accepted, want a refusal", what, in)
		return
	}
	msg := err.Error()
	var we *writer.Error
	if !errors.As(err, &we) || we.Path == "" || we.Rule == "" {
		t.Errorf("%s(%q): %v is not a *writer.Error with a path and a rule", what, in, err)
	}
	if strings.Contains(msg, marker) {
		t.Errorf("%s(%q) echoed the value: %q", what, in, msg)
	}
	if len(in) >= 6 && strings.Contains(msg, in) {
		t.Errorf("%s(%q) echoed the value: %q", what, in, msg)
	}
}

// TestErrorText: the path, then the rule.
func TestErrorText(t *testing.T) {
	e := &writer.Error{Path: `env["TZ"]`, Rule: "is refused"}
	if got, want := e.Error(), `env["TZ"]: is refused`; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
