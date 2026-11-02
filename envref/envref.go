// Package envref resolves a secret that config may give either as a literal value
// or as the name of an environment variable that holds it.
//
// Each secret field has a sibling reference key: semp.adminPass and
// semp.adminPassEnv in one tool, image.pass and image.pass-env in the other. The
// reference keeps the value out of a file that travels. The value of a reference
// key is a NAME this process looks up, never a value and never anything that
// reaches a shell, so the checks here are about actionable errors rather than
// injection: they catch `passEnv: ${FOO}` and `passEnv: my secret`, the two ways
// an operator mistakes the reference key for the value key, before the lookup
// silently fails.
//
// The rules, the same in both tools:
//   - Both keys set is an error rather than a precedence rule. Guessing which one
//     the operator meant is exactly the kind of silently wrong password that only
//     surfaces much later as a login failure.
//   - A variable that is unset, or set but empty, is an error. Deploying with a
//     blank password is worse than refusing to deploy.
//   - No message carries the literal or a resolved value. A name that fails the
//     name rules is not printed either: the likeliest such "name" is the secret
//     itself, pasted into the reference key by mistake. A valid name is printed,
//     because the operator needs it to export the variable.
//
// A tool keeps its own list of Ref values, one per secret field of its schema,
// and its own reserved-name rules on top of these.
package envref

import (
	"fmt"
	"strings"
)

// Ref pairs one secret field with the reference key that may supply it from the
// environment instead.
type Ref struct {
	// Field is the literal key, for messages: "semp.adminPass".
	Field string

	// EnvKey is the reference key, for messages: "semp.adminPassEnv".
	EnvKey string

	// Literal is the literal key's value, "" when unset. It is never printed.
	Literal string

	// Name is the variable EnvKey names, "" when unset.
	Name string
}

// UnsetError reports a reference whose variable is unset or set but empty. A
// validator that only advises -- warning that a deploy will fail until the
// variable is exported -- finds it with errors.As.
type UnsetError struct {
	EnvKey string
	Name   string

	// Empty is true when the variable is set but empty, false when it is unset.
	// The fix differs: a missing export against an export that produced nothing.
	Empty bool
}

func (e *UnsetError) Error() string {
	state := "not set"
	if e.Empty {
		state = "set but empty"
	}
	return fmt.Sprintf("%s names environment variable %s, which is %s; export it before running", e.EnvKey, e.Name, state)
}

// ValidName reports whether name is a portable environment-variable name: a
// letter or '_', then letters, digits or '_', ASCII only.
func ValidName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c == '_', 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z':
		case '0' <= c && c <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

// Check runs the rules that need no environment: the Ref is wired, a named
// variable is a bare valid name, and the literal and the reference are not both
// set. A Ref with neither set passes; whether the field is mandatory is the
// tool's question.
func (r Ref) Check() error {
	if r.Field == "" || r.EnvKey == "" {
		return fmt.Errorf("envref: Ref.Field and Ref.EnvKey must both be set (got %q and %q); name the literal "+
			"key and its reference key", r.Field, r.EnvKey)
	}
	if r.Name == "" {
		return nil
	}
	// Neither message below prints r.Name (see the package doc).
	if strings.Contains(r.Name, "$") {
		return fmt.Errorf("%s holds a variable reference, not a name (the value is not shown, in case it is the "+
			"secret itself): write the bare name, FOO rather than $FOO or ${FOO} -- %s takes the variable NAME, "+
			"%s takes the value itself", r.EnvKey, r.EnvKey, r.Field)
	}
	if !ValidName(r.Name) {
		return fmt.Errorf("%s is not a valid environment variable name (the value is not shown, in case it is the "+
			"secret itself): expected a letter or '_' followed by letters, digits or '_' -- %s takes the variable "+
			"NAME, %s takes the value itself", r.EnvKey, r.EnvKey, r.Field)
	}
	if r.Literal != "" {
		return fmt.Errorf("%s and %s are both set: use one or the other -- %s carries the value, "+
			"%s names the environment variable holding it", r.Field, r.EnvKey, r.Field, r.EnvKey)
	}
	return nil
}

// Resolve returns the secret: the literal when no variable is named, otherwise the
// variable's value through lookup. It runs Check first. lookup is the
// environment seam, normally os.LookupEnv; nil means the calling command has no
// environment access, which is an error only when a variable is named. An unset
// or empty variable is an *UnsetError.
func (r Ref) Resolve(lookup func(string) (string, bool)) (string, error) {
	if err := r.Check(); err != nil {
		return "", err
	}
	if r.Name == "" {
		return r.Literal, nil
	}
	if lookup == nil {
		return "", fmt.Errorf("%s names environment variable %s, but this command has no environment access "+
			"to read it; run a command that deploys", r.EnvKey, r.Name)
	}
	v, set := lookup(r.Name)
	if v == "" {
		return "", &UnsetError{EnvKey: r.EnvKey, Name: r.Name, Empty: set}
	}
	return v, nil
}
