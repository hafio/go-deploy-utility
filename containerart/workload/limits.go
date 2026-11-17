package workload

import (
	"slices"
	"strconv"
	"strings"
)

// UlimitNames returns the 16 Linux resource limit names, as a fresh slice. Three of
// them, nice, rtprio and rttime, are listed but refused by the checks (S2): they
// let a container out-prioritise the host.
func UlimitNames() []string {
	return []string{
		"as", "core", "cpu", "data", "fsize", "locks", "memlock", "msgqueue",
		"nice", "nofile", "nproc", "rss", "rtprio", "rttime", "sigpending", "stack",
	}
}

func refusedLimits() []string { return []string{"nice", "rtprio", "rttime"} }

// Ulimit is one typed resource limit (S2). Soft and Hard are a number or -1 for
// unlimited, and Soft is not above Hard.
type Ulimit struct {
	// Name is one of UlimitNames, less the three refused ones.
	Name string

	// Soft is the limit the kernel enforces, a number from 0 or -1.
	Soft int64

	// Hard is the ceiling Soft may be raised to, a number from 0 or -1.
	Hard int64
}

// ParseLimit reads `name=soft` or `name=soft:hard`, the docker spelling, where a
// single value sets both. Each value is digits or -1. It applies the rules of
// Ulimit.Check that do not depend on a sink. Only a name from the fixed list is
// echoed.
func ParseLimit(s string) (Ulimit, *FieldError) {
	const field = "ulimit"
	name, vals, ok := strings.Cut(s, "=")
	if !ok {
		return Ulimit{}, fail(field, "must be name=soft or name=soft:hard, for example core=-1 "+
			"(the value is not shown)")
	}
	softS, hardS, two := strings.Cut(vals, ":")
	soft, sok := limitValue(softS)
	hard, hok := soft, sok
	if two {
		hard, hok = limitValue(hardS)
	}
	if !sok || !hok {
		return Ulimit{}, fail(field, "a limit is digits or -1 for unlimited, with no sign, fraction or unit "+
			"(the value is not shown): write core=-1 or a soft:hard pair of numbers")
	}
	u := Ulimit{Name: name, Soft: soft, Hard: hard}
	if f := u.check(); f != nil {
		return Ulimit{}, f
	}
	return u, nil
}

// limitValue reads digits or -1.
func limitValue(s string) (int64, bool) {
	if s == "-1" {
		return -1, true
	}
	if !isDigits(s) || len(s) > 18 {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err == nil
}

// Check applies the S2 rules for a sink: the name is one of the 16 and not nice,
// rtprio or rttime, each value is 0 or more or -1, and soft is not above hard,
// where -1 is unlimited and so is above every number. Kubernetes has no ulimit
// field and is refused.
func (u Ulimit) Check(sink Sink) *FieldError {
	if f := checkSink("ulimit", sink); f != nil {
		return f
	}
	if sink == SinkKubernetes {
		return fail("ulimit", "has no Kubernetes form: a Deployment cannot set resource limits; remove the ulimit")
	}
	return u.check()
}

func (u Ulimit) check() *FieldError {
	const field = "ulimit"
	if !slices.Contains(UlimitNames(), u.Name) {
		return fail(field, "the name is not a Linux limit (the value is not shown): use one of "+
			strings.Join(UlimitNames(), ", "))
	}
	if slices.Contains(refusedLimits(), u.Name) {
		return failValue(field, u.Name, "is refused because it lets a container out-prioritise the host: remove it")
	}
	if u.Soft < -1 || u.Hard < -1 {
		return failValue(field, u.Name, "has a value below -1: write a number from 0, or -1 for unlimited")
	}
	if u.Hard != -1 && (u.Soft == -1 || u.Soft > u.Hard) {
		return failValue(field, u.Name, "has a soft limit above its hard limit: lower the soft limit or raise the hard limit")
	}
	return nil
}

// sizeCheck is the grammar of a memory or shared-memory size: digits with an
// optional unit of b, k, m or g in either case. Kubernetes memory waits for the
// Deployment's resources (E12), so it is refused with a message that says so.
func sizeCheck(field string, sink Sink, s string) *FieldError {
	if f := checkSink(field, sink); f != nil {
		return f
	}
	if sink == SinkKubernetes {
		return fail(field, "is not supported on Kubernetes yet: the Deployment's resources are still to be decided "+
			"(E12); remove the field")
	}
	n := s
	if n != "" && strings.IndexByte("bBkKmMgG", n[len(n)-1]) >= 0 {
		n = n[:len(n)-1]
	}
	if !isDigits(n) || len(n) > 18 {
		return fail(field, "must be a number with an optional unit b, k, m or g, such as 512m or 8g (the value is not shown)")
	}
	return nil
}

// ValidMemory checks a memory limit: a number with an optional b, k, m or g in
// either case. Kubernetes is refused until the Deployment's resources are decided
// (E12). The value is not echoed.
func ValidMemory(sink Sink, s string) *FieldError { return sizeCheck("memory", sink, s) }

// ValidShmSize checks a shared-memory size by the same grammar as ValidMemory.
// Kubernetes is refused. The value is not echoed.
func ValidShmSize(sink Sink, s string) *FieldError { return sizeCheck("shm-size", sink, s) }

// ValidCPUSet checks a CPU list of numbers and ranges, such as 0-3,5. It checks the
// grammar only: a renderer refuses the field for Kubernetes, which has no CPU
// list, by naming it. The value is not echoed.
func ValidCPUSet(s string) *FieldError {
	ok := s != ""
	for _, item := range strings.Split(s, ",") {
		loS, hiS, isRange := strings.Cut(item, "-")
		if !isDigits(loS) || len(loS) > 6 {
			ok = false
			break
		}
		if !isRange {
			continue
		}
		lo, _ := strconv.Atoi(loS)
		hi, err := strconv.Atoi(hiS)
		if !isDigits(hiS) || len(hiS) > 6 || err != nil || lo > hi {
			ok = false
			break
		}
	}
	if ok {
		return nil
	}
	return fail("cpu-set", "must be a list of CPU numbers and ranges such as 0-3,5, with no spaces and each range "+
		"running low to high (the value is not shown)")
}
