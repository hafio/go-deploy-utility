package workload

import (
	"net/netip"
	"strings"
)

// Port is one published port: `[ip:]host:container[/tcp|/udp]` (E7). Host and
// Container are a port or a lo-hi range, kept as written. IP is empty or an IP
// literal, an IPv6 one without its brackets.
type Port struct {
	IP        string
	Host      string
	Container string
	Protocol  string
}

// ParsePort reads the E7 grammar. A bare port or range means the same on both
// sides, so `8080` is `8080:8080`. With a bind address both sides are written, and
// an IPv6 address is in brackets. The protocol is /tcp or /udp, or none. Ports run
// 1 to 65535 with no sign or leading zero. A container range needs a host range of
// the same size. Only an out-of-range number is echoed.
func ParsePort(s string) (Port, *FieldError) {
	const field = "port"
	var p Port
	body := s
	if i := strings.LastIndexByte(body, '/'); i >= 0 {
		proto := body[i+1:]
		if proto != "tcp" && proto != "udp" {
			return Port{}, fail(field, "the protocol must be /tcp or /udp, or left out (the value is not shown)")
		}
		p.Protocol = proto
		body = body[:i]
	}
	var parts []string
	if rest, ok := strings.CutPrefix(body, "["); ok {
		ip, tail, found := strings.Cut(rest, "]:")
		addr, err := netip.ParseAddr(ip)
		if !found || err != nil || !addr.Is6() || addr.Zone() != "" {
			return Port{}, fail(field, "a bracketed bind address must be an IPv6 literal followed by "+
				":host:container, for example [::1]:8080:8080 (the value is not shown)")
		}
		p.IP = ip
		parts = strings.Split(tail, ":")
		if len(parts) != 2 {
			return Port{}, fail(field, "a bind address needs both sides, [ip]:host:container (the value is not shown)")
		}
	} else {
		parts = strings.Split(body, ":")
		switch len(parts) {
		case 1:
			parts = []string{parts[0], parts[0]}
		case 2:
		case 3:
			addr, err := netip.ParseAddr(parts[0])
			if err != nil || !addr.Is4() {
				return Port{}, fail(field, "a bind address before host:container must be an IPv4 literal; write an "+
					"IPv6 address in brackets, [::1]:8080:8080 (the value is not shown)")
			}
			p.IP = parts[0]
			parts = parts[1:]
		default:
			return Port{}, fail(field, "has too many ':' parts: write [ip:]host:container, with IPv6 in brackets "+
				"(the value is not shown)")
		}
	}
	p.Host, p.Container = parts[0], parts[1]
	hlo, hhi, f := portRange(field, p.Host)
	if f != nil {
		return Port{}, f
	}
	clo, chi, f := portRange(field, p.Container)
	if f != nil {
		return Port{}, f
	}
	if chi > clo && hhi-hlo != chi-clo {
		return Port{}, fail(field, "a container range needs a host range of the same size, for example "+
			"8000-8010:9000-9010 (the value is not shown)")
	}
	return p, nil
}

// portRange parses a port or a lo-hi range and returns its ends.
func portRange(field, s string) (int, int, *FieldError) {
	loS, hiS, isRange := strings.Cut(s, "-")
	lo, f := portNumber(field, loS)
	if f != nil {
		return 0, 0, f
	}
	if !isRange {
		return lo, lo, nil
	}
	hi, f := portNumber(field, hiS)
	if f != nil {
		return 0, 0, f
	}
	if lo > hi {
		return 0, 0, fail(field, "a range must run from the lower port to the higher, lo-hi (the value is not shown)")
	}
	return lo, hi, nil
}

func portNumber(field, s string) (int, *FieldError) {
	n, ok := decimal(s)
	if !ok {
		return 0, fail(field, "a port is a decimal number from 1 to 65535 with no sign or leading zero, and a range "+
			"is lo-hi (the value is not shown)")
	}
	if n < 1 || n > 65535 {
		return 0, failValue(field, s, "is outside 1 to 65535")
	}
	return int(n), nil
}

// String writes both sides, `[ip:]host:container[/protocol]`, with an IPv6
// address in brackets. A bare port never comes out bare.
func (p Port) String() string {
	s := p.Host + ":" + p.Container
	if p.IP != "" {
		ip := p.IP
		if strings.Contains(ip, ":") {
			ip = "[" + ip + "]"
		}
		s = ip + ":" + s
	}
	if p.Protocol != "" {
		s += "/" + p.Protocol
	}
	return s
}
