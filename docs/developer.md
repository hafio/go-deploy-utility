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
| B2 | `execguard`, `envref` | B `internal/config/execguard.go` and `secretref.go`; A `internal/validate` |
| B3 | `runner`, `runner/runnertest` | A `internal/runner` seam; B `internal/engine` |
| B4 | `containerart`, `containerart/compose`, `containerart/quadlet`, `containerart/artifacttest` | A `internal/dockergen`, `internal/podmangen`, `internal/yamlwriter`; B `internal/render` |
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

## Adoption order

Each step lands in M first. Then each tool gets one commit on its
`adopt-go-deploy-utility` branch. Tests that leave a tool must exist in M in the
same step. When a tool's coverage total falls because code moved out, the step
states the old and new totals and why.

| Step | Component | Behaviour change | Decision needed first |
| --- | --- | --- | --- |
| S1 | exitcode, notice, paritydoc, abbrev; both tools re-copy `tag.yml` and add `go.work` to `.gitignore` | none when the notice goldens match | env-file exit code, A 1 vs B 2 |
| S2 | confirm | B fails a non-TTY destructive prompt without `--no-prompt` with exit 1 and names the flag; B's downgrade prompt gains `--no-prompt` | none, already decided |
| S3 | execguard, envref | A rejects invisible characters and accepts embedded `~` and `.Exe` | escalator ban for A |
| S4 | runner | B cancels with interrupt, then kill, and can pass stdin and env together; A's ErrDot message names the fix | none |
| S5 | containerart | B's output bytes stay identical; A's change only if A adopts B's quoting | A adopts B's always-quote policy |
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
