package envref

import (
	"errors"
	"strings"
	"testing"
)

// ref is a well-wired Ref with the given literal and name, using solace-util's
// key spelling.
func ref(literal, name string) Ref {
	return Ref{Field: "semp.adminPass", EnvKey: "semp.adminPassEnv", Literal: literal, Name: name}
}

// env is a lookup over a fixed map, so no test reads the real environment.
func env(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

// TestValidName: the portable shape, with rows from both tools' rejects.
func TestValidName(t *testing.T) {
	for name, want := range map[string]bool{
		"FOO": true, "_X": true, "_": true, "a1": true, "MQ_PASSWORD": true, "SOLACE_TEST_ADMIN": true,
		"": false, "1-bad-name": false, "1A": false, "my secret": false, "${FOO}": false, "A-B": false,
		"A.B": false, "A B": false, "\u00c9T\u00c9": false, "FOO\n": false,
	} {
		if got := ValidName(name); got != want {
			t.Errorf("ValidName(%q) = %v, want %v", name, got, want)
		}
	}
}

// TestCheckBothSet: both keys set is an error naming both, not a precedence rule.
func TestCheckBothSet(t *testing.T) {
	err := ref("literal", "ADMIN_PASS").Check()
	if err == nil {
		t.Fatal("Check accepted both keys set")
	}
	for _, want := range []string{"semp.adminPass and semp.adminPassEnv are both set", "use one or the other"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v, want it to contain %q", err, want)
		}
	}
}

// TestCheckReferenceForm: a shell reference written as the name gets its own
// message, which says what to write instead.
func TestCheckReferenceForm(t *testing.T) {
	for _, name := range []string{"${MQ_PASSWORD}", "$MQ_PASSWORD", "PRE${X}"} {
		err := ref("", name).Check()
		if err == nil || !strings.Contains(err.Error(), "semp.adminPassEnv holds a variable reference, not a name") ||
			!strings.Contains(err.Error(), "FOO rather than $FOO") {
			t.Errorf("Check(%q) = %v, want the reference-form message", name, err)
		}
	}
}

// TestCheckInvalidName: any other bad name names both keys, so the operator sees
// which one takes the name and which the value.
func TestCheckInvalidName(t *testing.T) {
	for _, name := range []string{"my secret", "1-bad-name", "A.B"} {
		err := ref("", name).Check()
		if err == nil {
			t.Fatalf("Check(%q) accepted an invalid name", name)
		}
		for _, want := range []string{"is not a valid environment variable name", "semp.adminPassEnv takes the variable NAME",
			"semp.adminPass takes the value itself"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("Check(%q) = %v, want it to contain %q", name, err, want)
			}
		}
	}
	if err := ref("", "ADMIN_PASS").Check(); err != nil {
		t.Errorf("Check(ADMIN_PASS) = %v, want accepted", err)
	}
	if err := ref("", "").Check(); err != nil {
		t.Errorf("Check with neither key set = %v, want accepted", err)
	}
}

// TestRefWiringErrors: a Ref without its two key names is a wiring mistake.
func TestRefWiringErrors(t *testing.T) {
	for _, r := range []Ref{{}, {Field: "a"}, {EnvKey: "aEnv"}} {
		if err := r.Check(); err == nil || !strings.Contains(err.Error(), "envref: Ref.Field and Ref.EnvKey") {
			t.Errorf("Check(%+v) = %v, want the wiring error", r, err)
		}
		if _, err := r.Resolve(env(nil)); err == nil {
			t.Errorf("Resolve(%+v) passed a wiring mistake", r)
		}
	}
}

// TestResolveLiteralSkipsLookup: with no variable named, the literal is the answer
// and the environment is never consulted, even with no access to it.
func TestResolveLiteralSkipsLookup(t *testing.T) {
	never := func(string) (string, bool) {
		t.Fatal("lookup called for a literal")
		return "", false
	}
	for _, literal := range []string{"s3cret", ""} {
		got, err := ref(literal, "").Resolve(never)
		if err != nil || got != literal {
			t.Errorf("Resolve(literal %q) = %q, %v", literal, got, err)
		}
	}
	if got, err := ref("s3cret", "").Resolve(nil); err != nil || got != "s3cret" {
		t.Errorf("Resolve(literal, nil lookup) = %q, %v", got, err)
	}
}

// TestResolveFromEnv: a named, set, non-empty variable is the answer, and a literal
// that looks like a reference is never expanded.
func TestResolveFromEnv(t *testing.T) {
	got, err := ref("", "ADMIN_PASS").Resolve(env(map[string]string{"ADMIN_PASS": "from-env"}))
	if err != nil || got != "from-env" {
		t.Errorf("Resolve = %q, %v; want from-env", got, err)
	}
	got, err = ref("${ADMIN_PASS}", "").Resolve(env(map[string]string{"ADMIN_PASS": "from-env"}))
	if err != nil || got != "${ADMIN_PASS}" {
		t.Errorf("Resolve(literal ${ADMIN_PASS}) = %q, %v; want it verbatim", got, err)
	}
}

// TestResolveUnsetAndEmpty: both are errors, told apart, and both are an
// *UnsetError a validator can downgrade to advice.
func TestResolveUnsetAndEmpty(t *testing.T) {
	for _, tc := range []struct {
		name  string
		env   map[string]string
		empty bool
		want  string
	}{
		{"unset", nil, false, "semp.adminPassEnv names environment variable ADMIN_PASS, which is not set; export it"},
		{"empty", map[string]string{"ADMIN_PASS": ""}, true, "which is set but empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ref("", "ADMIN_PASS").Resolve(env(tc.env))
			var u *UnsetError
			if !errors.As(err, &u) {
				t.Fatalf("Resolve = %v, want an *UnsetError", err)
			}
			if u.Empty != tc.empty || u.Name != "ADMIN_PASS" || u.EnvKey != "semp.adminPassEnv" {
				t.Errorf("UnsetError = %+v", u)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

// TestResolveNoEnvironmentAccess: a named variable on a command with no
// environment access is an error that says so.
func TestResolveNoEnvironmentAccess(t *testing.T) {
	_, err := ref("", "ADMIN_PASS").Resolve(nil)
	if err == nil || !strings.Contains(err.Error(), "has no environment access") {
		t.Errorf("Resolve(nil lookup) = %v, want the no-access error", err)
	}
	var u *UnsetError
	if errors.As(err, &u) {
		t.Error("no access is not an unset variable")
	}
}

// TestMessagesNeverCarryTheValue: no error path prints the literal or a resolved
// value, nor a "name" that fails the name rules -- the likeliest such name is the
// secret itself, pasted into the reference key.
func TestMessagesNeverCarryTheValue(t *testing.T) {
	const literal, value = "lit-s3cret-1", "env-s3cret-2"
	lookup := env(map[string]string{"ADMIN_PASS": value, "EMPTY": ""})
	for _, pasted := range []string{"pa$$w0rd!", "my s3cret phrase", "9lives"} {
		err := ref("", pasted).Check()
		if err == nil || strings.Contains(err.Error(), pasted) {
			t.Errorf("Check(name %q) = %v; want an error that does not echo the name", pasted, err)
		}
	}
	for _, r := range []Ref{
		ref(literal, "ADMIN_PASS"),
		ref(literal, "${ADMIN_PASS}"),
		ref(literal, "bad name"),
		ref("", "UNSET"),
		ref("", "EMPTY"),
		{Field: "", EnvKey: "", Literal: literal},
	} {
		_, err := r.Resolve(lookup)
		if err == nil {
			t.Fatalf("Resolve(name %q) passed", r.Name)
		}
		if strings.Contains(err.Error(), literal) || strings.Contains(err.Error(), value) {
			t.Errorf("error carries a secret: %v", err)
		}
	}
}
