# Plan R1: the workload package and its core checks

Approved by the owner, 2026-11-17. R1 is the first build step of the rendering work in the
module (design note 2026-11-15, sections 8 to 11). It adds the value types of the shared
container description and the checks the owner settled, and nothing that waits on an open
question. No renderer, tool or golden changes in this step, and no tag.

Executor: Sonnet agents launched by this session (owner, 2026-11-17). This session reviews the
diff. The owner runs every gate.

## Why this step first

S10: M exports one check per rule it renders by; the renderers call the same ones, and the
tools call them at load. The checks need no renderer, so they can land first and be adopted
by the tools at load before any artifact moves. The consolidation draft
[2026-11-17-container-description.md](2026-11-17-container-description.md) places them in a
new package `containerart/workload` (a working name, its P10).

## Package

`containerart/workload`, standard library plus `execguard` and `envref` only. It does not import `kube`
or `containercli`: the DNS-label rule is written here, and `kube.ValidLabel` delegates to it
in the kube step (N59). That avoids the import cycle the draft's P1 describes. No package
variable holds state: word lists are returned by functions as fresh slices.

The project CLAUDE.md still says "A package needs two consumers". The owner discarded that
rule on 2026-11-15 (design note section 5 lists the rule change); this step updates the line.

## Contents

Each check returns `*FieldError` (rule and, where S8 allows, the value) or nil. S8: a value is
echoed only for names, words from a fixed list and numbers; free text never.

| Name | Rule | Decided by |
| --- | --- | --- |
| `Sink`, `SinkCompose`, `SinkQuadlet`, `SinkKubernetes` | which renderer a check runs for; zero is invalid | S10 |
| `FieldError` | field, rule, optional value; `Error()` never prints free text | S8 |
| `CheckText(s)` | valid UTF-8, no `execguard.IsControl`, no `execguard.IsInvisible` | S8, N13 later |
| `Image`, `ParseImage(s)`, `Image.String()` | registry host[:port] or empty (Docker Hub); repository a lower-case path whose first part has no `.` or `:`; tag or `sha256:` digest required, `latest` allowed | E3, S7 |
| `RunAs`, `ParseRunAs(s)`, `RunAsIDs(uid, gid)`, `Stated()` | `uid:gid`, decimal, 0 to 2147483647, no sign or leading zero; `""` is not stated; a bare uid and a name are refused | S3, E5 |
| `Port`, `ParsePort(s)`, `Port.String()` | bare port or range means N:N; `[ip:]host:container`, then `/tcp` or `/udp` optional; IPv6 in brackets; a container range needs a host range of the same size; `String` writes both sides | E7 |
| `ValidRestart(sink, w)` | compose: `no`, `always`, `unless-stopped`, `on-failure[:N]`; quadlet: `no`, `on-success`, `on-failure`, `on-abnormal`, `on-watchdog`, `on-abort`, `always`; Kubernetes: only `always` (draft P9, an assumption); `""` allowed (renders `always`) | E6, ruling 9 |
| `ValidNetwork(sink, w)` | compose: `bridge`, `host`, `none`; quadlet: `private`, `bridge`, `pasta`, `slirp4netns`, `host`, `none`; Kubernetes: refused; `""` allowed | todo 6, ruling 13 |
| `CheckName(s)` | DNS label, at most 63 characters (container name a tool creates) | rulings 11, 13 |
| `ValidDNSLabel(s)` | the label grammar `kube` will delegate to | N59 later |
| `ValidHostName(s)` | RFC 1123: labels of letters, digits, `-`, at most 63 each, 253 in all, starting with a letter or digit | S12 |
| `ValidTimezone(s)` | IANA name: parts split by `/`, each of letters, digits, `_`, `+`, `-`, at most 14, not starting with `-` | S13 |
| `ParseSeconds(s)`, `Health`, `Health.Check(sink)` | Go duration in whole seconds, written `<n>s`; interval and timeout above 0, start period 0 or more, retries at least 1; on quadlet a command token with whitespace or a shell character is refused | S6, Q6 |
| `Ulimit`, `ParseLimit(s)`, `UlimitNames()` | the 16 Linux names; soft not above hard; a number or -1; `rtprio`, `rttime`, `nice` refused; refused on Kubernetes | S2 |
| `ValidMemory(sink, s)`, `ValidShmSize(sink, s)` | compose and quadlet: a number with an optional `b`, `k`, `m`, `g` in either case; Kubernetes memory waits for E12 (refused in R1, with a message saying so) | S2 |
| `ValidCPUSet(s)` | a list of CPUs and ranges, `0-3,5`; refused on Kubernetes | S2 |
| `CheckComposeVar(name)` | a variable M puts into the compose process environment: `envref` name grammar; refuses `PATH`, `HOME`, `IFS`, `SHELL`, `BASH_ENV` and any `LD_`, `DOCKER_`, `COMPOSE_` prefix | C6 |
| `CheckInlineText(s)` | refuses CR, U+0085, U+2028, U+2029, every other control character but newline and tab, DEL, the C1 range, U+FEFF and invalid UTF-8; the error gives an offset, never the text | C7 |
| `CheckMountSource(sink, s)` | absolute path; compose allows one drive-letter root and no other `:`; quadlet per Q7's answer; `filepath.IsAbs` is not used | C8, Q7 |

Not in R1: the `Spec` struct, the deny lists of S1, label keys and reserved prefixes (label
design note, review item 4), Kubernetes memory and ports' `Name` (E12, K entries), the writer,
the atomic write (S11, Q5), every renderer.

## Tests

`containerart/workload/workload_test.go`, table tests per function: every accepted form and
every refusal named in the table above, including the boundary rows (63 and 64 characters,
port 65535 and 65536, a range of unequal size, `1m30s` -> 90 and `1.5s` refused, `-1` and
`soft > hard`, U+2028 inside inline text, `C:\x` and `C:\x:y`). For every free-text field a
row checks that the error text does not contain the value (S8). No test touches the host.

## Docs in the same step

- `docs/developer.md`: a row for the step in the package table; `workload` under the layout.
- `docs/test.md`: the new test file and its tests.
- Project `CLAUDE.md`: the "two consumers" line becomes the 2026-11-15 rule (common artifacts
  and operations in the module, customisation and orchestration in the tools).
- Design note section 11: a pointer to this plan.

## Gates (the owner runs them)

1. `./scripts/dev.sh all scan`
2. `./scripts/dev.sh cov`; the last recorded total is the floor.

Commit headline when the gates pass: `add the workload package with the core checks of the
shared container description`
