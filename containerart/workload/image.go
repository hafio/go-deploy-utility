package workload

import "strings"

// Image is a container image reference split into the parts the env files name
// separately (E3). Registry is a host with an optional port, and empty means
// Docker Hub. Repository is the path only and never holds a host. A tag or a
// digest is required (S7), and the word latest is allowed as an explicit tag.
type Image struct {
	// Registry is host[:port], or empty for Docker Hub.
	Registry string

	// Repository is a lower-case path whose first part has no '.' or ':'.
	Repository string

	// Tag is [A-Za-z0-9_][A-Za-z0-9_.-]{0,127}.
	Tag string

	// Digest is sha256: and 64 lower-case hex digits.
	Digest string
}

// ParseImage splits s into an Image and checks it. The first path part is the
// registry when it holds a '.' or a ':', the way the engines read it. The error
// never echoes s: an image reference is free text (S8).
func ParseImage(s string) (Image, *FieldError) {
	const field = "image"
	if s == "" {
		return Image{}, fail(field, "is empty: write [registry/]repository:tag or repository@sha256:<64 hex digits>")
	}
	var img Image
	rest := s
	if i := strings.IndexByte(rest, '@'); i >= 0 {
		img.Digest = rest[i+1:]
		rest = rest[:i]
	}
	last := strings.LastIndexByte(rest, '/') + 1
	if i := strings.IndexByte(rest[last:], ':'); i >= 0 {
		img.Tag = rest[last+i+1:]
		rest = rest[:last+i]
	}
	if i := strings.IndexByte(rest, '/'); i >= 0 && strings.ContainsAny(rest[:i], ".:") {
		img.Registry = rest[:i]
		rest = rest[i+1:]
	}
	img.Repository = rest
	if f := img.check(field); f != nil {
		return Image{}, f
	}
	return img, nil
}

// String writes registry/repository:tag@digest, leaving out the parts that are
// empty.
func (i Image) String() string {
	s := i.Repository
	if i.Registry != "" {
		s = i.Registry + "/" + s
	}
	if i.Tag != "" {
		s += ":" + i.Tag
	}
	if i.Digest != "" {
		s += "@" + i.Digest
	}
	return s
}

// check applies the grammar to each part. No message carries a part.
func (i Image) check(field string) *FieldError {
	if i.Registry != "" && !validRegistry(i.Registry) {
		return fail(field, "the registry must be a host name with an optional :port, such as ghcr.io or "+
			"localhost:5000 (the value is not shown); put the path in the repository")
	}
	if !validRepository(i.Repository) {
		return fail(field, "the repository must be a lower-case path of letters, digits and single '.', '_' or '-' "+
			"separators, and its first part must hold no '.' or ':' (the value is not shown); "+
			"put a registry host in the registry, for example ghcr.io/team/app")
	}
	if i.Tag == "" && i.Digest == "" {
		return fail(field, "needs a tag or a sha256 digest so the image is pinned: write :latest to ask for the "+
			"newest explicitly, or :10.26.0 for a release")
	}
	if i.Tag != "" && !validTag(i.Tag) {
		return fail(field, "the tag must be 1 to 128 letters, digits, '_', '.' or '-' and must not start with '.' "+
			"or '-' (the value is not shown)")
	}
	if i.Digest != "" && !validDigest(i.Digest) {
		return fail(field, "the digest must be sha256: followed by 64 lower-case hex digits (the value is not shown)")
	}
	return nil
}

func isAlnum(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// validRegistry accepts host[:port]: dot-separated labels of letters, digits and
// '-' that start and end with a letter or digit, at most 253 characters, and a
// port from 1 to 65535.
func validRegistry(s string) bool {
	host := s
	if i := strings.LastIndexByte(s, ':'); i >= 0 {
		host = s[:i]
		if p, ok := decimal(s[i+1:]); !ok || p < 1 || p > 65535 {
			return false
		}
	}
	if host == "" || len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			if !isAlnum(label[i]) && label[i] != '-' {
				return false
			}
		}
	}
	return true
}

// validRepository accepts '/'-separated parts of lower-case letters and digits
// with single '.', '_' or '-' between them, the first part holding no '.'.
func validRepository(s string) bool {
	if s == "" {
		return false
	}
	for n, part := range strings.Split(s, "/") {
		if part == "" || (n == 0 && strings.Contains(part, ".")) {
			return false
		}
		// The distribution grammar: lower-case alphanumeric runs joined by one
		// separator, which is ".", "_", "__" or a run of "-".
		lower := func(c byte) bool { return c >= 'a' && c <= 'z' || c >= '0' && c <= '9' }
		for i := 0; i < len(part); {
			if lower(part[i]) {
				i++
				continue
			}
			if i == 0 {
				return false
			}
			j := i + 1
			switch part[i] {
			case '.':
			case '_':
				if j < len(part) && part[j] == '_' {
					j++
				}
			case '-':
				for j < len(part) && part[j] == '-' {
					j++
				}
			default:
				return false
			}
			if j == len(part) || !lower(part[j]) {
				return false
			}
			i = j
		}
	}
	return true
}

func validTag(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !isAlnum(c) && c != '_' && (i == 0 || (c != '.' && c != '-')) {
			return false
		}
	}
	return true
}

func validDigest(s string) bool {
	hex, ok := strings.CutPrefix(s, "sha256:")
	if !ok || len(hex) != 64 {
		return false
	}
	for i := 0; i < len(hex); i++ {
		c := hex[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
