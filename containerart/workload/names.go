package workload

import (
	"strings"
	"unicode"

	"github.com/hafio/go-deploy-utility/envref"
)

// ValidDNSLabel reports whether s is a DNS-1123 label: 1 to 63 lower-case letters,
// digits and '-', starting and ending with a letter or digit. It is the grammar
// kube.ValidLabel will delegate to.
func ValidDNSLabel(s string) bool {
	if s == "" || len(s) > 63 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		alnum := c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
		if !alnum && (c != '-' || i == 0 || i == len(s)-1) {
			return false
		}
	}
	return true
}

// CheckName checks the name of a container a tool creates: a DNS label of at most
// 63 characters. The error does not echo the name, since a refused one may not be
// one.
func CheckName(s string) *FieldError {
	if ValidDNSLabel(s) {
		return nil
	}
	return fail("name", "must be a DNS label: 1 to 63 lower-case letters, digits and '-', starting and ending "+
		"with a letter or digit (the value is not shown); for example solace-primary")
}

// ValidHostName checks a host name against RFC 1123: dot-separated labels of
// letters, digits and '-', each 1 to 63 characters and starting and ending with a
// letter or digit, at most 253 characters in all. An underscore is refused. The
// error does not echo the name.
func ValidHostName(s string) *FieldError {
	ok := s != "" && len(s) <= 253
	if ok {
		for _, label := range strings.Split(s, ".") {
			if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
				ok = false
				break
			}
			for i := 0; i < len(label); i++ {
				if !isAlnum(label[i]) && label[i] != '-' {
					ok = false
				}
			}
		}
	}
	if ok {
		return nil
	}
	return fail("host name", "must be an RFC 1123 host name: dot-separated labels of letters, digits and '-', each "+
		"up to 63 characters and starting and ending with a letter or digit, 253 in all (the value is not shown); "+
		"an underscore is not allowed, write solace-primary")
}

// ValidTimezone checks an IANA time zone database name: parts separated by '/',
// each of letters, digits, '_', '+' and '-', 1 to 14 characters, not starting with
// '-'. A POSIX rule string such as EST5EDT,M3.2.0,M11.1.0 is refused. The error
// does not echo the value.
func ValidTimezone(s string) *FieldError {
	ok := s != ""
	for _, part := range strings.Split(s, "/") {
		if part == "" || len(part) > 14 || part[0] == '-' {
			ok = false
			break
		}
		for i := 0; i < len(part); i++ {
			if c := part[i]; !isAlnum(c) && c != '_' && c != '+' && c != '-' {
				ok = false
			}
		}
	}
	if ok {
		return nil
	}
	return fail("timezone", "must be an IANA time zone name such as Europe/Paris or UTC: parts separated by '/', "+
		"each 1 to 14 letters, digits, '_', '+' or '-' and not starting with '-' (the value is not shown); "+
		"a POSIX rule string is not accepted")
}

// CheckComposeVar checks the name of a variable the tool puts into the compose
// process's environment (C6). It must be a valid environment variable name
// (envref.ValidName), and PATH, HOME, IFS, SHELL, BASH_ENV and any name starting
// LD_, DOCKER_ or COMPOSE_ are refused, in any letter case, because the compose
// process reads them for itself. A refused valid name is echoed; an invalid one,
// which may be a pasted secret, is not.
func CheckComposeVar(name string) *FieldError {
	const field = "compose variable"
	if !envref.ValidName(name) {
		return fail(field, "is not a valid environment variable name (the value is not shown, in case it is a "+
			"secret): use a letter or '_' followed by letters, digits or '_'")
	}
	upper := strings.ToUpper(name)
	switch upper {
	case "PATH", "HOME", "IFS", "SHELL", "BASH_ENV":
		return failValue(field, name, "is read by the compose process itself: choose another name")
	}
	for _, prefix := range [...]string{"LD_", "DOCKER_", "COMPOSE_"} {
		if strings.HasPrefix(upper, prefix) {
			return failValue(field, name, "starts with "+prefix+", a prefix the loader or the compose tools read "+
				"for themselves: choose another name")
		}
	}
	return nil
}

// quadletUnsafe are the characters refused in a quadlet mount source: the quote
// and shell characters, '$' and backslash. '%' is allowed, since the renderer
// doubles it.
const quadletUnsafe = "$\"'`\\"

// CheckMountSource checks the host path of a bind mount (C8, Q7). The path must be
// absolute and hold no control or invisible character (CheckText). Absolute is
// decided here, never by filepath.IsAbs, whose answer depends on the machine
// running the check: a leading '/', or on compose one drive-letter root, C:\ or
// C:/. Compose allows no other ':'. Quadlet allows no ':' at all, and no
// whitespace, '$', quote, backtick or backslash. Kubernetes has no host bind
// mount and is refused. The error does not echo the path.
func CheckMountSource(sink Sink, s string) *FieldError {
	const field = "mount source"
	if f := checkSink(field, sink); f != nil {
		return f
	}
	if sink == SinkKubernetes {
		return fail(field, "has no Kubernetes form: a Deployment cannot bind-mount a host path; remove the mount")
	}
	if f := CheckText(s); f != nil {
		return f.At(field)
	}
	rest := s
	switch {
	case strings.HasPrefix(s, "/"):
	case sink == SinkCompose && hasDriveRoot(s):
		rest = s[3:]
	default:
		return fail(field, "must be an absolute path: start with '/' (compose also takes one drive root such as "+
			"C:\\ or C:/); the value is not shown")
	}
	if strings.Contains(rest, ":") {
		return fail(field, "holds a ':' that would add a field to the mount line (the value is not shown): "+
			"remove it; only a leading drive letter such as C:/ is allowed, and only on compose")
	}
	if sink == SinkQuadlet && (strings.ContainsFunc(s, unicode.IsSpace) || strings.ContainsAny(s, quadletUnsafe)) {
		return fail(field, "holds whitespace, '$', a quote, a backtick or a backslash, which the unit line cannot "+
			"carry (the value is not shown): rename the directory or link to a plain path")
	}
	return nil
}

// hasDriveRoot reports whether s starts with a drive letter and a root, C:\ or C:/.
func hasDriveRoot(s string) bool {
	return len(s) >= 3 && isAlnum(s[0]) && !(s[0] >= '0' && s[0] <= '9') && s[1] == ':' && (s[2] == '\\' || s[2] == '/')
}
