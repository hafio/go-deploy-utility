package workload

import (
	"slices"
	"strings"
)

// restartWords returns the restart words a sink takes, as a fresh slice. Compose
// also takes on-failure:N, which is checked apart.
func restartWords(s Sink) []string {
	switch s {
	case SinkCompose:
		return []string{"no", "always", "unless-stopped", "on-failure"}
	case SinkQuadlet:
		return []string{"no", "on-success", "on-failure", "on-abnormal", "on-watchdog", "on-abort", "always"}
	case SinkKubernetes:
		return []string{"always"}
	}
	return nil
}

// networkWords returns the network words a sink takes, as a fresh slice. A
// Deployment has no network setting, so Kubernetes has none.
func networkWords(s Sink) []string {
	switch s {
	case SinkCompose:
		return []string{"bridge", "host", "none"}
	case SinkQuadlet:
		return []string{"private", "bridge", "pasta", "slirp4netns", "host", "none"}
	}
	return nil
}

// ValidRestart checks a restart word for a sink. Compose takes no, always,
// unless-stopped, on-failure and on-failure:N (N from 1). Quadlet takes systemd's
// words: no, on-success, on-failure, on-abnormal, on-watchdog, on-abort and always.
// Kubernetes takes only always, the one policy a Deployment allows (an assumption
// recorded in the container description draft, P9). The empty word is allowed on
// every sink and renders always. The word is not echoed, since it may be anything.
func ValidRestart(sink Sink, w string) *FieldError {
	const field = "restart"
	if f := checkSink(field, sink); f != nil {
		return f
	}
	if w == "" || slices.Contains(restartWords(sink), w) {
		return nil
	}
	list := strings.Join(restartWords(sink), ", ")
	if sink == SinkCompose {
		if n, ok := strings.CutPrefix(w, "on-failure:"); ok {
			if v, ok := decimal(n); ok && v >= 1 && v <= maxID {
				return nil
			}
		}
		list += ", on-failure:N with N from 1"
	}
	return failf(field, "is not a %s restart word (the value is not shown): use one of %s, or leave it empty for always",
		sink, list)
}

// ValidNetwork checks a network word for a sink. Compose takes bridge, host and
// none. Quadlet takes private, bridge, pasta, slirp4netns, host and none. A
// Deployment has no network setting, so Kubernetes refuses any word. The empty
// word is allowed on every sink and means the sink's own default. The word is not
// echoed, since it may be anything.
func ValidNetwork(sink Sink, w string) *FieldError {
	const field = "network"
	if f := checkSink(field, sink); f != nil {
		return f
	}
	if w == "" {
		return nil
	}
	words := networkWords(sink)
	if words == nil {
		return failf(field, "has no %s form: a Deployment has no network setting; remove the field", sink)
	}
	if slices.Contains(words, w) {
		return nil
	}
	return failf(field, "is not a %s network word (the value is not shown): use one of %s, or leave it empty for the default",
		sink, strings.Join(words, ", "))
}
