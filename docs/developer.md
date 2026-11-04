# Developer guide

> [!WARNING]
> **Not a supported Solace product.** `go-deploy-utility` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this module, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this module only, not the Solace products that the
> tools built on it deploy.

This module holds the code that `solmq-conn-util` and `solace-util` share. The
packages are built here first. Each tool then adopts them one component at a
time, both tools in the same step.

| Short name | Repo | Path |
| --- | --- | --- |
| M | this module | `/opt/claude-workspace/projects/go-deploy-utility` |
| A | `solmq-conn-util` | `/opt/claude-workspace/projects/solace-ibmmq-connector-helper` |
| B | `solace-util` | `/opt/claude-workspace/projects/solace-cnt-scripts` |

## Rules for shared code

- No package imports a type or config from either tool. Each tool adapts to the
  shared API.
- A package exists only when both tools use it, or will in the step that lifts it.
- The standard library comes first. A new dependency is pinned exactly, and the
  change that adds it says why.
- Packages are ordered leaf first, and the import graph has no cycles.
- Builders and parsers are pure. They return `runner.Cmd` values or parsed data
  and never execute. Only code that takes a `runner.Runner` executes.
- Shared code has no globals, registers no flags and never calls `os.Exit`.
  Seams are function fields or parameters. A golden test's update switch is a
  parameter, never a package flag. The one exception is the parity body in
  `paritydoc`: `go:embed` needs a package-level variable, which stays
  unexported and is never assigned.
- Every text file is LF in the repo and in every checkout (`.gitattributes`),
  because embedded text and notices are compared byte for byte.
- An error names what failed and the offending value, with secrets masked, and
  says what to do next. A wrapped cause uses `%w`.
- A lifted test keeps its original name, and [test.md](test.md) records its source.

## Package layout and build order

Packages are built in this order. Each step is planned and approved on its own.

| Step | Packages | Lifted from |
| --- | --- | --- |
| B1 | `exitcode` | B `internal/cli/exit.go`; the cobra parts stay in B |
| B1 | `notice`, `notice/noticetest` | A `cmd/solmq-conn-util/support.go` and `support_test.go`; B `internal/cli/commanddoc_test.go` |
| B1 | `paritydoc` | `paritydoc/body.md` (moved from `docs/handoff/parity-body.md`), embedded, with a check each tool runs on its own `docs/parity.md` |
| B1 | `abbrev` | B `internal/abbrev` |
| B1 | `confirm` | A `promptLine` and `stdinIsTerminal`; B `internal/cli/confirm.go` |
| B2 | `execguard`, `envref` | B `internal/config/execguard.go` and `secretref.go`; A `internal/validate` and `internal/gen` `ResolveCredentials` |
| B3 | `runner`, `runner/runnertest` | A `internal/runner` seam; B `internal/engine` |
| B4 | `containerart/compose`, `containerart/quadlet`, `containerart/artifacttest` | A `internal/dockergen` and `internal/podmangen` escaping, hardening and secret lines; B `internal/render` `Compose` and `Quadlet` and their posture tests. A's `internal/yamlwriter` stays in A, its only user |
| B5 | `podman`, `systemd`, `rootless` | A `internal/runner` podman and systemd parts, `runas.go`; B `internal/container` rootless, limits and euid code |
| B6 | `containercli`, `containercli/inspect`, `watch` | A compose, ps and info code, `internal/statusreport`; B `internal/container` docker paths, `inspect.go`, `internal/cli/watch.go` |
| B7 | `kube`, `kube/rbac`, `kube/nsgate` | A kubectl code and `namespace.go`; B `internal/k8s` |

The `output` package waits until the output-model decision is made, because only
B uses it until then. Each tool keeps its own version printing.

Notes on the available packages:

- `exitcode` never prints or exits. A tool marks usage errors with `Usagef` or
  `AsUsage`, and its main calls `os.Exit(exitcode.Code(err))`.
- `notice.New` takes the name, the kind word and the covers clause without its
  final period. One greedy wrap at 80 columns reproduces every committed notice.
- `noticetest` checks `README.md` plus `docs/*.md` by default. The glob does not
  descend, so a doc in a subfolder of `docs/` is not checked unless listed in
  `Options.Files`.
- `paritydoc/body.md` is the only copy of the shared parity text that is edited.
  A change there is tagged and then picked up by both tools in one step.
- `confirm` holds one buffered reader per `Prompter`, so a tool keeps one
  Prompter per run. B's redeploy consent prompt stays in B until the
  redeploy-consent decision is made.
- `execguard.Check` is the one command guard. A tool calls it from its
  validator and again in every executor right before argv is built, with one
  `Rules` value per command field. `Approve` turns `--allow-command` values into
  `Approvals`, and is the only way to widen an allowlist. A privilege-escalation
  wrapper (22 names, case-folded, `.exe` stripped) is never permitted, by either
  route. Leading-`~` expansion stays in each tool; the guard refuses a leading
  `~` it is handed.
- `execguard.UnsafeChars`, `PathSeparators`, `IsControl` and `IsInvisible` are
  the one charset definition. A host-path gate derives its own set from them
  rather than copying it.
- `envref.Ref` pairs a secret field with its reference key. `Check` needs no
  environment; `Resolve` takes the lookup seam (`os.LookupEnv`, or nil for a
  command with no environment access). An unset or empty variable is an
  `*UnsetError`, which a validator may downgrade to a warning. No message
  prints a value, or a name that fails the name rules, since that is most
  likely the secret pasted into the wrong key. Reserved-name rules, such as A's
  `_GEN_` prefix, stay in the tool.
- `runner.Runner` has two methods, `Run` and `Attach`. Every capture mode is a
  helper over `Run`: `Combined` (A's `Run`), `Split` (A's `RunSplit`) and
  `Output` (B's `Output`). Each returns what was captured even when the run
  fails. The package doc maps every old method of both tools.
- `Exec` interrupts a cancelled command and kills it when `Grace` ends (2s by
  default). A cancelled run is an error wrapping `ctx.Err()`, never success. A
  command that exits 0 while a process it started holds its output open
  succeeds when the grace period ends. `Attach` takes no context, and the runner
  never interrupts or kills a session.
- Errors read `argv0: cause` and never carry an argument, the stdin bytes or an
  env value. `Resolve` refuses ErrDot and any other relative result.
  `Cmd.Check`, `CheckRun` and `CheckAttach` are the shared preconditions, and
  `Exec`, `Echo` and the fake all apply them.
- `Echo` is a value type, so a tool may type-assert against it. It prints
  `+ argv`, with the stdin size and the env keys with masked values, and writes
  nothing to the command's own streams. `Quote` is for display only. It
  single-quotes a token with shell syntax from `execguard.UnsafeChars` or a
  leading `~`, and Go-escapes a token with an unprintable rune.
- `runnertest.Fake` records calls and answers each from its `Script`: `Seq`
  matches by position, while `FailOn` or any other function matches by argv
  content. It refuses what `Exec` refuses, and records nothing for a refused
  call.
- Some runner policy stays in the tools:
  - A keeps its `report` ok/error line, and maps a cancelled `logs --follow` to
    a clean end.
  - B keeps its `NewExec` verbose announcer, which is built on B's `output`
    package.
  - B's `EnvRunner` narrowing becomes a B-side wrapper that refuses `Cmd.Env`
    for every holder except compose.
- `containerart/compose` and `containerart/quadlet` hold pieces, not
  renderers. Each tool keeps its own renderer, field order and product
  settings. A block ends in a newline and is written as it is; a line or scalar
  has none. Both tools put the service at indent 4.
- One quoting policy, solace-util's, decided 2026-11-04: a compose scalar
  derived from operator input or from a name goes through `compose.Quote`
  (`strconv.Quote` after `Escape` doubles every `$`); the image, numbers and
  tool constants stay bare; every quadlet assignment is `quadlet.Environment`,
  quoted whole. A value meant for interpolation, such as a digest label
  `"${VAR}"`, is written by the tool and never escaped.
- `compose.Quote` cannot carry invalid UTF-8, since YAML has no escape for a raw
  byte. `quadlet.Escape` writes a control character as a C escape that podman
  decodes, so a newline cannot add keys to a unit. Each tool still checks every
  value it routes to either against execguard's charset at load.
- `compose.Posture` and `quadlet.Posture` are the privilege lines. The user line
  stays with the tool, which may keep the image's own user.
  `compose.ServiceSecrets`, `compose.TopSecrets` and `quadlet.Secret` are the
  secret wiring: values come from the environment compose runs with or from
  podman's secret store, never from the file, and targets are absolute.
- `artifacttest` checks rendered artifacts, and a tool passes every shape it
  renders. It covers:
  - the widening spellings of both tools (`Widening()`);
  - host networking (`HostNetwork()`), kept separate because solace-util
    writes it when asked;
  - the posture blocks;
  - quadlet `[Container]` keys against the tool's own podman floor
    (`CheckQuadletKeys(unit, "4.7")`);
  - absolute secret targets.

  A new quadlet key goes into the release table in `quadletKeySince`, with the
  podman release that introduced it. The Kubernetes spellings
  (`allowPrivilegeEscalation: true`, `hostNetwork`, `hostPID`, `hostIPC`) stay
  in solmq-conn-util until B7.

## Adoption order

Each step lands in M first. Then each tool gets one commit on its
`adopt-go-deploy-utility` branch. Tests that leave a tool must exist in M in the
same step. When a tool's coverage total falls because code moved out, the step
states the old and new totals and why.

| Step | Component | Behaviour change | Decision needed first |
| --- | --- | --- | --- |
| S1 | exitcode, notice, paritydoc, abbrev; both tools re-copy `tag.yml` and add `go.work` to `.gitignore` | none when the notice goldens match | env-file exit code, A 1 vs B 2 |
| S2 | confirm | B fails a non-TTY destructive prompt without `--no-prompt` with exit 1 and names the flag; B's downgrade prompt gains `--no-prompt` | none, already decided |
| S3 | execguard, envref | A refuses invisible and C1 characters, invalid UTF-8, any escalator in a command (`sudo podman` goes; run `sudo solmq-conn-util` instead) and a set-but-empty `-env` variable at deploy, accepts embedded `~` and `.Exe`, and takes B's guard messages; B refuses C1 characters, invalid UTF-8 and 8 more escalators, its escalator messages change, and an invalid `*Env` name is no longer echoed back | none, already decided (ban in both, 2026-11-02) |
| S4 | runner; the parity body's "Runner seams" section is rewritten in the same step | A: every call takes a context and is interrupted, then killed, on cancel (was `Stream` only); a cancel is an error, which `logs --follow` maps to its clean end; errors read `kubectl: exit status 1`; B's resolve messages, and a relative resolution is refused; the optional `Streamer`, `Splitter` and `Attacher` interfaces go; a pipe held open after exit 0 no longer hangs. B: a cancel interrupts, then kills, instead of killing; stdin and env pass together; `RunInteractive` becomes `Attach`, which returns the exit code and is never killed; Echo annotates stdin and env on one line; `Quote` quotes more shell syntax and escapes unprintable runes; a held pipe no longer hangs; `EnvRunner` becomes a B-side wrapper | none |
| S5 | containerart: compose, quadlet, artifacttest | B: its artifact bytes stay identical. Timezone and health-probe tokens are checked against execguard's charset at load. Its widening test gains A's spellings (`Privileged=true`, `no-new-privileges=false`, `NoNewPrivileges=false`, `SYS_ADMIN`, `label=disable`, `label:disable`, `Unconfined`), and the audit greps in `docs/container-security.md` follow. A: compose quotes the project `name:`, the service key, `container_name`, the `TZ` and `LOGGING_SYSLOG_*` values, volume entries, and secret and config `source:` names and their top-level keys. Every quadlet `Environment=` is quoted whole. The image, numbers and tool constants stay bare. Values routed to `Quote` or `Environment` are checked against execguard's charset. The posture tests use `artifacttest`, keeping A's Kubernetes extras | none: A adopts B's always-quote policy (decided 2026-11-04) |
| S6 | podman, systemd, rootless | the decided rootless rule; A gains the XDG quadlet dir, session repair, linger and core-limit checks, and uses the effective uid; B uses `secret create --replace` | subuid warn vs refuse |
| S7 | containercli, inspect | B hands compose its file on stdin and writes no compose file; a digest label drives change detection; `docker.composeFile` is retired | redeploy consent prompt; docker lifecycle verbs |
| S8 | watch | A stops sending screen clears into a pipe | none |
| S9 | kube, rbac, nsgate | A creates the namespace only when absent, waits for rollout, probes every written kind and uses B's can-i parser | namespace gate policy; read probes in B |

S7 needs `docker-compose` v1 and `nerdctl compose` checked for `-f -` first. A's
Kubernetes path has never run on a real cluster. The root deploy for a declared
rootless user stays a design task after S6.

## Adding a package

1. Read both source implementations and their tests in A and B.
2. Pick the superset or better-tested source, and note the choice in the change.
3. Cut every dependency on tool types: take plain values, small option structs
   or function seams instead.
4. Lift the tests under their original names and add rows to [test.md](test.md).
5. Add the package to the table in the README and in this guide.
6. Hand over `./scripts/dev.sh all scan` and record the new coverage total.

## Dev tasks

| Task | What it does |
| --- | --- |
| `build` | `go build ./...` as a compile check; writes nothing |
| `vet` | `go vet ./...` plus a `gofmt -l` check that reports and never rewrites |
| `test` | `go test -count=1 ./...`; `DEV_RACE=1` adds `-race` |
| `cov` | coverage profile and HTML, total printed into `scripts/logs/cov.log` |
| `scan` | `go tool govulncheck ./...`; any reachable finding fails |
| `graphify` | refreshes the knowledge graph; local only |

`all` runs build, vet and test, and CI runs `all scan`. `full` adds cov, scan
and graphify, and is the sweep before a tag. The tag pipeline uses the
init-devops library shape, which needs the repo variable `SHIP_LIBRARY=true`.
Releases and the order across the three repos are in [consuming.md](consuming.md).
