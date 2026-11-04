// Package compose holds the pieces of a docker compose file that solmq-conn-util
// and solace-util write the same way: the escaping and quoting of scalars, the
// service's privilege posture, and the secret wiring.
//
// Each tool keeps its own renderer, its field order and its product settings. A
// block here ends in a newline and is written as it is; a scalar has none. Both
// tools put the service at indent 4, so the blocks carry that indent.
//
// Quoting follows one policy in both tools: a scalar derived from operator input
// or from a name goes through Quote, while the image reference, numbers and the
// tool's own constants are written bare.
package compose

import (
	"strconv"
	"strings"
)

// Escape doubles every '$'. Docker Compose interpolates the whole document,
// inline config content included, and turns "$$" back into one "$", so an
// escaped value reaches the container exactly as it was written. Unescaped, a
// shell script's variables are blanked, compose refuses a "$(" outright, and a
// "${...}" placeholder meant for the application is filled from the environment
// compose itself runs with, where the secret values are.
//
// A value that is meant to be interpolated, such as a digest label written as
// "${VAR}", must not go through Escape.
func Escape(s string) string { return strings.ReplaceAll(s, "$", "$$") }

// Quote renders s as a double-quoted YAML scalar with every '$' doubled. It is
// strconv.Quote, which is what fmt's %q produces, so the bytes match what
// solace-util's renderer has always written.
//
// The quotes matter for YAML 1.1 as much as for the charset: bare, a name such as
// yes, off, 0123 or 1.5 is read as a boolean or a number, and a service key or
// container_name stops being the string the operator wrote.
//
// For valid UTF-8 every escape strconv.Quote writes is a YAML escape with the same
// meaning: the single-letter ones, \xNN for a control character, and \u or \U for
// any other rune that does not print. Invalid UTF-8 is the exception. A raw byte
// comes out as \xNN, which YAML reads as the code point U+00NN, a different value.
// A tool checks every value it routes here against execguard's charset when the
// config is loaded.
func Quote(s string) string { return strconv.Quote(Escape(s)) }

// FlowList renders items as a one-line YAML flow sequence, each item through
// Quote, the way a healthcheck test: or an entrypoint: is written. No items give
// an empty sequence.
func FlowList(items []string) string {
	q := make([]string, len(items))
	for i, it := range items {
		q[i] = Quote(it)
	}
	return "[" + strings.Join(q, ", ") + "]"
}

// Posture is the service's privilege posture, stated rather than inherited:
// privileged off, every capability dropped, and no-new-privileges on, as five
// lines at service indent.
//
// privileged is a bare boolean because the schema of compose v1 rejects a quoted
// one. no-new-privileges takes the '=' separator because dockerd logs a
// deprecation warning for ':' on every container create. The user: line is not
// part of the block: a tool may leave the image's own user in place, so it writes
// that line itself.
const Posture = "    privileged: false\n" +
	"    cap_drop:\n" +
	"      - ALL\n" +
	"    security_opt:\n" +
	"      - no-new-privileges=true\n"

// Secret is one secret the service mounts. Its value never appears in the file:
// compose reads it from the environment it runs with (the environment provider,
// which needs Docker Compose 2.23.1 or later).
type Secret struct {
	// Name is the compose secret's name. It is written through Quote.
	Name string

	// Target is the absolute path the secret is mounted at. A bare name would
	// mount it under the engine's own /run/secrets instead, where nothing
	// looks. artifacttest.CheckSecretTargets holds a rendered file to this.
	Target string

	// EnvVar is the environment variable compose reads the value from. It is
	// written bare, so it must be a plain variable name.
	EnvVar string
}

// ServiceSecrets renders the service's secrets: block in the long syntax, one
// source and target per secret, in the order given. It is empty when there are no
// secrets, so no key is left holding a null.
func ServiceSecrets(secrets []Secret) string {
	if len(secrets) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("    secrets:\n")
	for _, s := range secrets {
		b.WriteString("      - source: " + Quote(s.Name) + "\n")
		b.WriteString("        target: " + s.Target + "\n")
	}
	return b.String()
}

// TopSecrets renders the top-level secrets: map, each entry read through the
// environment provider from its EnvVar. It is empty when there are no secrets.
func TopSecrets(secrets []Secret) string {
	if len(secrets) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("secrets:\n")
	for _, s := range secrets {
		b.WriteString("  " + Quote(s.Name) + ":\n")
		b.WriteString("    environment: " + s.EnvVar + "\n")
	}
	return b.String()
}
