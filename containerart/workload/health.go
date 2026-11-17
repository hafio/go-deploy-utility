package workload

import (
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/hafio/go-deploy-utility/execguard"
)

// ParseSeconds reads a Go duration that is a whole number of seconds and returns
// the seconds, so 1m30s is 90 (S6). A fraction of a second, such as 1.5s or 500ms,
// is refused, and so is text that is not a duration, a bare number other than 0
// included. The sign is the caller's rule: -5s parses as -5. The error does not echo
// the text, and does not wrap the parser's own error, which would.
func ParseSeconds(s string) (int64, *FieldError) {
	const field = "duration"
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fail(field, "is not a duration (the value is not shown): write a number and a unit, "+
			"such as 30s or 1m30s")
	}
	if d%time.Second != 0 {
		return 0, fail(field, "must be a whole number of seconds (the value is not shown): "+
			"write 2s rather than 1.5s or 500ms")
	}
	return int64(d / time.Second), nil
}

// Health is the exec-form health probe. Test is an argument list, run with no shell.
// The times are whole seconds (S6), which each renderer writes as <n>s.
type Health struct {
	// Test is the probe's argument list: at least one token, none empty.
	Test []string

	// Interval is the time between probes, above 0.
	Interval time.Duration

	// Timeout is how long one probe may take, above 0.
	Timeout time.Duration

	// StartPeriod is the grace period before failures count, 0 or more.
	StartPeriod time.Duration

	// Retries is the number of consecutive failures that make the container
	// unhealthy, 1 or more.
	Retries int
}

// shellChars are the characters a quadlet probe token may not hold (Q6): the
// execguard charset plus '~'. Whitespace is refused beside them.
const shellChars = execguard.UnsafeChars + "~"

// Check applies the S6 rules for a sink. Every time is a whole number of seconds,
// Interval and Timeout are above 0, StartPeriod is 0 or more and Retries is 1 or
// more. Each Test token passes CheckText. On quadlet the unit line is one string
// that podman may run through a shell, so a token holding whitespace or a shell
// character is refused rather than meaning something else than on compose (Q6).
// The first problem found is returned. A token is never echoed.
func (h Health) Check(sink Sink) *FieldError {
	if f := checkSink("health", sink); f != nil {
		return f
	}
	if len(h.Test) == 0 {
		return fail("health.test", "needs at least one token: the probe command, for example curl -fs http://localhost:8080/")
	}
	for i, tok := range h.Test {
		field := "health.test[" + strconv.Itoa(i) + "]"
		if tok == "" {
			return fail(field, "is empty: every token of the probe command must be written")
		}
		if f := CheckText(tok); f != nil {
			return f.At(field)
		}
		if sink == SinkQuadlet && (strings.ContainsFunc(tok, unicode.IsSpace) || strings.ContainsAny(tok, shellChars)) {
			return fail(field, "holds whitespace or a shell character, which podman would give to a shell (the value is "+
				"not shown): split it into separate tokens, or run the probe through a script file inside the image")
		}
	}
	if f := seconds("health.interval", h.Interval, true); f != nil {
		return f
	}
	if f := seconds("health.timeout", h.Timeout, true); f != nil {
		return f
	}
	if f := seconds("health.start-period", h.StartPeriod, false); f != nil {
		return f
	}
	if h.Retries < 1 {
		return failValue("health.retries", strconv.Itoa(h.Retries), "must be 1 or more: write the number of failures "+
			"that make the container unhealthy, for example 3")
	}
	return nil
}

// seconds checks one health time: whole seconds, and above 0 when positive is set,
// otherwise 0 or more.
func seconds(field string, d time.Duration, positive bool) *FieldError {
	if d%time.Second != 0 {
		return fail(field, "must be a whole number of seconds: write 2s rather than 1.5s or 500ms")
	}
	if d < 0 || (positive && d == 0) {
		if positive {
			return fail(field, "must be above 0 seconds")
		}
		return fail(field, "must be 0 seconds or more")
	}
	return nil
}
