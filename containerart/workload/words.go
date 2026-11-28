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
		return []string{"Always", "OnFailure", "Never"}
	}
	return nil
}

// kubeKinds returns the seven workload kinds ValidKubeRestart knows, as a fresh
// slice.
func kubeKinds() []string {
	return []string{"Deployment", "StatefulSet", "ReplicaSet", "DaemonSet", "Job", "CronJob", "Pod"}
}

// kubeRestartWords returns the restart words a Kubernetes kind takes, as a fresh
// slice, or nil for a kind that is not one of the seven. A pod template of a
// Deployment, StatefulSet, ReplicaSet or DaemonSet allows Always only; a Job or
// CronJob allows OnFailure and Never; a bare Pod allows all three.
func kubeRestartWords(kind string) []string {
	switch kind {
	case "Deployment", "StatefulSet", "ReplicaSet", "DaemonSet":
		return []string{"Always"}
	case "Job", "CronJob":
		return []string{"OnFailure", "Never"}
	case "Pod":
		return []string{"Always", "OnFailure", "Never"}
	}
	return nil
}

// pullWords returns the fixed pull-policy words a sink takes, as a fresh slice.
// Compose also takes every_<duration>, which is checked apart.
func pullWords(s Sink) []string {
	switch s {
	case SinkCompose:
		return []string{"always", "missing", "never", "daily", "weekly", "if_not_present"}
	case SinkQuadlet:
		return []string{"always", "missing", "never", "newer"}
	case SinkKubernetes:
		return []string{"Always", "IfNotPresent", "Never"}
	}
	return nil
}

// serviceTypes returns the Service types M writes, as a fresh slice.
func serviceTypes() []string {
	return []string{"ClusterIP", "NodePort", "LoadBalancer"}
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
// Kubernetes takes the Kubernetes spellings Always, OnFailure and Never, as a
// union: the word is a member of some kind's list. Which kind takes which word is
// ValidKubeRestart's question, so a caller that knows the workload kind asks that
// one as well. The empty word is allowed on every sink: compose and quadlet render
// always, Kubernetes writes nothing. The word is not echoed, since it may be
// anything.
func ValidRestart(sink Sink, w string) *FieldError {
	const field = "restart"
	if f := checkSink(field, sink); f != nil {
		return f
	}
	if w == "" || slices.Contains(restartWords(sink), w) {
		return nil
	}
	list := strings.Join(restartWords(sink), ", ")
	tail := "or leave it empty for always"
	switch sink {
	case SinkCompose:
		if n, ok := strings.CutPrefix(w, "on-failure:"); ok {
			if v, ok := decimal(n); ok && v >= 1 && v <= maxID {
				return nil
			}
		}
		list += ", on-failure:N with N from 1"
	case SinkKubernetes:
		tail = "or leave it empty to write nothing"
	}
	return failf(field, "is not a %s restart word (the value is not shown): use one of %s, %s",
		sink, list, tail)
}

// ValidKubeRestart checks a Kubernetes restart word against the workload kind
// (P9). Deployment, StatefulSet, ReplicaSet and DaemonSet take Always only; Job
// and CronJob take OnFailure and Never; Pod takes all three. The empty word is
// allowed and writes nothing. The kind is a plain word the renderer passes, and an
// unknown kind is refused, naming the seven kinds, before the word is read. The
// word is not echoed, since it may be anything; the kind is a fixed-list word and
// is named once it is known.
func ValidKubeRestart(kind, w string) *FieldError {
	const field = "restart"
	words := kubeRestartWords(kind)
	if words == nil {
		return failf(field, "the workload kind is not one of %s (the value is not shown): pass one of them",
			strings.Join(kubeKinds(), ", "))
	}
	if w == "" || slices.Contains(words, w) {
		return nil
	}
	return failValue(field, kind, "takes no such restart word (the value is not shown): use one of "+
		strings.Join(words, ", ")+", or leave it empty to write nothing")
}

// ValidPullPolicy checks an image pull policy for a sink, in that sink's words and
// exact case. Compose takes always, missing, never, daily, weekly, if_not_present
// and every_<duration>, where the duration is one or more digit-and-unit pairs with
// the units w, d, h, m and s, such as every_1d12h. build is refused with its own
// sentence, since M never builds an image. Quadlet takes always, missing, never and
// newer. Kubernetes takes Always, IfNotPresent and Never. The empty word is allowed
// and writes nothing, which leaves the engine's or cluster's own default. Only
// build is echoed, as a fixed word; any other refused word is not.
func ValidPullPolicy(sink Sink, w string) *FieldError {
	const field = "pull-policy"
	if f := checkSink(field, sink); f != nil {
		return f
	}
	if w == "" || slices.Contains(pullWords(sink), w) {
		return nil
	}
	list := strings.Join(pullWords(sink), ", ")
	if sink == SinkCompose {
		if d, ok := strings.CutPrefix(w, "every_"); ok && everyDuration(d) {
			return nil
		}
		if w == "build" {
			return failValue(field, w, "is refused: M never builds an image, so a pull policy cannot ask for a build; "+
				"use one of "+list+", or every_<duration>")
		}
		list += ", or every_<duration> built from digits and the units w, d, h, m, s"
	}
	return failf(field, "is not a %s pull policy (the value is not shown): use one of %s, or leave it empty for the default",
		sink, list)
}

// everyDuration reports whether s is one or more digit-and-unit pairs, such as 1d12h,
// with the units w, d, h, m and s and at most nine digits in a pair.
func everyDuration(s string) bool {
	if s == "" {
		return false
	}
	for s != "" {
		n := 0
		for n < len(s) && s[n] >= '0' && s[n] <= '9' {
			n++
		}
		if n == 0 || n > 9 || n == len(s) || strings.IndexByte("wdhms", s[n]) < 0 {
			return false
		}
		s = s[n+1:]
	}
	return true
}

// ValidServiceType checks a Kubernetes Service type: ClusterIP, NodePort or
// LoadBalancer, in exact case. ExternalName is refused with its own sentence, since
// M writes no external name. The empty word is an error: M writes no default, so
// the tool states one. Any other word is refused and not echoed.
func ValidServiceType(w string) *FieldError {
	const field = "service-type"
	list := strings.Join(serviceTypes(), ", ")
	switch {
	case w == "":
		return failf(field, "is empty: M writes no default type, so state one of %s", list)
	case slices.Contains(serviceTypes(), w):
		return nil
	case w == "ExternalName":
		return failValue(field, w, "is refused: M writes no external name; use one of "+list)
	}
	return failf(field, "is not a Service type M writes (the value is not shown): use one of %s", list)
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
