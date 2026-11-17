// Package workload holds the value types of the shared container description and
// the checks that decide whether a value may be written into a compose file, a
// quadlet unit or a Kubernetes Deployment.
//
// S10 of the 2026-11-15 surface review splits the checking in two. A tool calls a
// check at load, so an operator hears about a bad value before any question is
// asked. A renderer calls the same check before it writes, so a value that skips
// the tool's validator is still refused. Both call the one definition here, which
// is why each rule is one exported function and no tool keeps a copy.
//
// Every check returns a *FieldError, or nil when the value is fine. The type is a
// pointer so that nil needs no second return value, with one trap: assigning a
// nil *FieldError to an error interface makes a non-nil error. A caller that
// returns the result as an error tests it against nil first.
//
// A FieldError never carries free text (S8). The checked value is echoed only
// when it is a name, a word from a fixed list or a number. Everything else is
// described by the field, the rule and, for text, a byte offset: operator text may
// hold a password, and an error reaches logs. An image reference counts as free
// text, because a registry login can sit in its host part, so ParseImage never
// echoes it.
//
// Checks that depend on the renderer take a Sink. The zero Sink is invalid and is
// refused, so a caller that forgot to choose fails loudly rather than getting the
// widest grammar. A field a sink cannot express is refused for that sink with an
// error naming the field, never dropped silently.
//
// The package imports only the standard library, execguard and envref. It does
// not import kube or containercli: the DNS-label rule is written here, and
// kube.ValidLabel delegates to ValidDNSLabel in a later step. No package variable
// holds state; the word lists are built by functions that return fresh slices.
package workload
