package workload

import (
	"strconv"
	"strings"
)

// maxID is the largest uid or gid a Linux user namespace map accepts.
const maxID = 2147483647

// RunAs is the numeric user and group a container runs as. It tells a stated
// root from a forgotten one (S3): the zero RunAs is not stated and means the
// image's own user, so a renderer writes nothing for it, while a RunAs built from
// 0:0 is stated and writes the root lines. Build one with ParseRunAs or RunAsIDs.
type RunAs struct {
	uid    uint32
	gid    uint32
	stated bool
}

// RunAsIDs states a uid and gid, each 0 to 2147483647.
func RunAsIDs(uid, gid uint32) (RunAs, *FieldError) {
	const field = "run-as"
	for _, id := range [2]uint32{uid, gid} {
		if id > maxID {
			return RunAs{}, failValue(field, strconv.FormatUint(uint64(id), 10),
				"is above 2147483647: a uid and a gid run from 0 to 2147483647")
		}
	}
	return RunAs{uid: uid, gid: gid, stated: true}, nil
}

// ParseRunAs reads `uid:gid`: two decimal numbers, 0 to 2147483647, with no sign
// and no leading zero. An empty string is the zero RunAs, not stated, with no error.
// A name, a bare uid and the word `default` are refused: a tool that offers
// `default` passes the empty string. Only an out-of-range number is echoed.
func ParseRunAs(s string) (RunAs, *FieldError) {
	const field = "run-as"
	if s == "" {
		return RunAs{}, nil
	}
	us, gs, ok := strings.Cut(s, ":")
	uid, uok := decimal(us)
	gid, gok := decimal(gs)
	if !ok || !uok || !gok {
		return RunAs{}, fail(field, "must be numeric uid:gid, for example 1000:0 (the value is not shown): "+
			"a name, a bare uid, a sign or a leading zero is refused; leave the field empty for the image's own user")
	}
	for _, n := range [2]string{us, gs} {
		if v, _ := decimal(n); v > maxID {
			return RunAs{}, failValue(field, n, "is above 2147483647: a uid and a gid run from 0 to 2147483647")
		}
	}
	return RunAs{uid: uint32(uid), gid: uint32(gid), stated: true}, nil
}

// Stated reports whether a user was stated. It is false for the zero RunAs.
func (r RunAs) Stated() bool { return r.stated }

// UID returns the stated uid, and 0 when nothing was stated: test Stated first.
func (r RunAs) UID() uint32 { return r.uid }

// GID returns the stated gid, and 0 when nothing was stated: test Stated first.
func (r RunAs) GID() uint32 { return r.gid }

// String writes uid:gid, or the empty string when nothing was stated.
func (r RunAs) String() string {
	if !r.stated {
		return ""
	}
	return strconv.FormatUint(uint64(r.uid), 10) + ":" + strconv.FormatUint(uint64(r.gid), 10)
}
