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
- Least privilege. A command runs with the lowest permission that does the job:
  no escalation, rootless where it can be, and a permission probe only for what
  an operation is about to write. A step that needs more permission only for
  convenience (waiting for a rollout, offering an emptied namespace for deletion,
  reading state for a better message) degrades to a warning when it is refused
  and never blocks the operation; the operation's own result decides the exit
  code.
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
| B5 | `podman`, `systemd`, `rootless` | A `internal/runner` podman version, secret and systemctl functions, `ResolveQuadletScope`, and the subid check in `cmd/solmq-conn-util/runas.go`; B `internal/container` `euid.go`, the session, linger and id-mapping code in `rootless.go`, the core-limit and limit parsing in `limits.go`, and the secret and systemctl calls in `manager.go` and `inspect.go`, plus `xdgConfigHome` in `internal/config/load.go`. A's mount-readability checks and B's nr_open, nofile, user-manager and data-dir checks stay in the tools |
| B6 | `containercli`, `containercli/inspect`, `watch` | A `internal/runner` `Docker`, `DockerLifecycle`, `Preflight`'s engine branch, `EngineList`, `EngineInspectJSON`, `ExecArgv`, `InstallScript` and `LogsArgv`, the decoding and name filter in `internal/statusreport/parse.go`, and the watch loop in `cmd/solmq-conn-util/status.go`; B `internal/container` `preflight.go`, the docker deploy, delete, listing, lifecycle and logs paths in `manager.go`, `transport.go`'s exec and upload, `inspect.go`, and `internal/cli/watch.go`. A's status model, image, stats and status-script code and B's `cp` transport, registry login and cert-digest comparison stay in the tools |
| B7 | `kube`, `kube/rbac`, `kube/nsgate` | A `internal/runner/runner.go` `Kubernetes`, `KubernetesLifecycle`, `Preflight`'s can-i branch and `canIDenied`, the get and list builders and the kubernetes branches of `ExecArgv` and `LogsArgv`, plus `cmd/solmq-conn-util/namespace.go` and `NamespaceManifest`; B `internal/k8s` `cluster.go` apply and delete, `prep.go` namespace read, create and delete, `ops.go` scale, rollout and logs, `transport.go` exec and upload, `preflight.go` and `namespace.go`. A's status parsers, `top` and secret probes and B's operator, CR, `cp` and delete-pod restart code stay in the tools |

The `output` package waits until the output-model decision is made, because only
B uses it until then. Each tool keeps its own version printing.

Notes on the available packages:

- `exitcode` never prints or exits. A tool marks usage errors with `Usagef` or
  `AsUsage`, and its main calls `os.Exit(exitcode.Code(err))`; a tool whose
  handlers return an int uses the constants. 2 means the tool refused before
  acting because of what it was given: argv, the env file, and the files and
  environment variables the env file names (decided 2026-11-05). 1 means
  something ran and failed, or the live system said no. A session's own status
  passes through only when the caller marks it with `ChildExit`. `Exited(name,
  code)` turns the code `runner.Attach` returns into an error that carries it and
  reads `name: exit status N`, as os/exec words it.
- `notice.New` takes the name, the kind word and the covers clause without its
  final period. One greedy wrap at 80 columns reproduces every committed notice.
- `noticetest` checks `README.md` plus `docs/*.md` by default. The glob does not
  descend, so a doc in a subfolder of `docs/` is not checked unless listed in
  `Options.Files`.
- `paritydoc/body.md` is the only copy of the shared parity text that is edited.
  A change there is tagged and then picked up by both tools in one step.
- `confirm` holds one buffered reader per `Prompter`, so a tool keeps one
  Prompter per run. A destructive question goes through `Gate`; a question
  `--no-prompt` must never answer (solace-util's widening of the operator watch
  to every namespace) goes through `Ask` and is refused without a terminal.
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
    a clean end. Its helpers take a context, and its commands pass `bg()`. It
    imports the shared package as `gdurunner`, because its own helper package
    is also named `runner`.
  - B keeps its `NewExec` verbose announcer, which is built on B's `output`
    package. Its `engine` package keeps the process-stream helpers `Run` and
    `Output`, and `Interactive`, which attaches the process's three files and
    returns `exitcode.Exited`.
  - B's `EnvRunner` narrowing becomes `engine.NoEnv`, a wrapper that refuses
    `Cmd.Env` for every holder except compose. `engine.IsEcho` sees through it.
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
- `podman`, `systemd` and `rootless` build commands, read answers and check
  facts; they never execute, read a file or read the environment. Every host
  probe is a `runner.Cmd` or a seam parameter (`getenv`, `home`, `isDir`), so a
  tool runs it through its own runner, a fake sees it and a preview skips it.
  Each tool keeps its own reads (the subid files, the account name), its
  memoising and its reports.
- A podman builder takes `base`, the engine argv the tool built and checked with
  execguard, and copies it.
- Secrets: `CreateSecret` writes
  `secret create [--replace] [--label l]... <name> -` with the value on stdin
  only, as a non-nil copy. When `SupportsReplace` holds for `Version`'s answer,
  one create with replace set does it; otherwise `RemoveSecrets` for the name
  runs first. Each step's error is the tool's to word.
- `systemd.IsActive` exits non-zero for every state but an active one, so a tool
  reads its stdout alone (`runner.Output` or `runner.Split`), ignores the exit
  error and passes it to `State`. `Active` (active or reloading) picks restart
  over start on deploy; `Stopped` (inactive, failed, unknown) is what a teardown
  needs before it may assume the unit is down, and no answer is neither.
- `podman.QuadletDir` is the directory a unit is written to, resolved the way
  `os.UserConfigDir`, and so podman's generator, resolves it. The generator reads
  the user manager's environment, so an `XDG_CONFIG_HOME` set only in a shell
  profile is not seen there; a relative one is refused.
- `rootless.Decide` is the rootless rule: rootless is the default, and rootful
  podman is deployed only when asked for, by root. A tool passes its rootful
  setting (false when unset) and the setting's key, which the refusals name; no
  shared package hard-codes a tool's config key.
- `rootless.UserSession` derives `XDG_RUNTIME_DIR` and the D-Bus address only
  when unset, after the runtime directory is proven to exist; `Env` is the
  derived entries alone, for `os.Setenv` or `Cmd.Env`.
- Subordinate ids are checked against podman's own mapping (`podman.IDMap`,
  `rootless.CheckIDMap`), for rootless deploys only, and a run-as id outside it
  is refused (decided 2026-11-04). The refusal reads: the diagnosis, a period, a
  newline, `rootless.SubIDFix` with a block from `NextFreeRange`, then the
  tool's own alternatives. `CheckCore` and `CheckLinger` are generic too; a tool
  appends its drop-in file or config key to the error.
- `containercli` builds engine commands from `base`, the guarded engine argv,
  and compose commands from `compose`, the guarded compose argv; how a tool
  derives either stays tool config. Compose takes its file on stdin only
  (`-f - --project-directory <dir>`) with secrets in `Cmd.Env`, so nothing is
  written to disk; both tools' files declare `name:`, so `<dir>` only anchors
  relative paths. Lifecycle is plain engine verbs on the container (decided
  2026-11-04).
- A redeploy is left to the platform and asks before anything is written
  (decided 2026-11-06). Compose: the file carries a label interpolated from an
  environment variable set to `containercli.ComposeDigest` of the file and the
  compose child's environment, so compose recreates exactly when either
  changed; the tool compares it with the running container's label to know
  whether to ask. Podman: each secret carries a digest label read back with
  `podman.SecretLabel`, and the unit is compared with the one on disk; only a
  change is written, and only a change restarts an active unit.
- `containercli.InfoError` is the one preflight message: engine, command,
  wrapped cause, the engine's own output when the tool captured it, then
  `EngineHint`.
- `containercli.List` lists with real tab separators and `ParseList` reads it;
  matching is done in Go (`HasName` exact, `ByImage` substring), never with
  `--filter name=`, which is an unanchored regex. A listed name that fails
  `ValidName` is reported, not passed back to the engine.
- `Exec` and `Upload` return an error for an empty name, command or
  destination, and `Exec` for a terminal without stdin. `Upload` single-quotes
  the destination, a POSIX path inside the container.
- `inspect.Containers` asks `container inspect`, which only containers answer.
  `inspect.Decode` strips docker's leading slash, takes health from the first of
  `State.Health` and `State.Healthcheck` with a status, and never decodes the
  environment. A tool prints engine text through `inspect.Printable`.
- `watch.Loop` redraws only when `watch.CanRedraw` says the output is a terminal
  that takes escape sequences, and appends 60-column section lines otherwise. A
  refresh error goes to the tool's `Warn` and the loop carries on; a refresh that
  reports its own failure returns nil. The loop ends with nil when its context
  does, so Ctrl-C exits 0.
- `kube` builders take `base`, the guarded kubectl argv with any `--context` or
  `--kubeconfig`, and put a namespace after the subcommand's own arguments as
  `-n <ns>`. They trust their names: a tool checks every namespace and object
  name with `kube.ValidLabel` at load, since a name starting with `-` would be
  read as a flag. Manifests travel on stdin (`Apply`, `Delete`).
- Namespace create-if-absent, in each tool: `GetNamespace` and `ParseNamespace`;
  an existing namespace needs nothing more and no cluster permission. Absent:
  probe `create namespaces`, then `Apply` the `NamespaceManifest` on its own; a
  refusal is `MissingNamespace`, a real failure. Unknown (the read was refused):
  probe `create namespaces`; allowed, apply it; refused, warn and carry on, since
  the next write's own probe stops a deploy into a missing namespace.
- After deploy, start and restart a tool waits with `RolloutStatus` (300s). The
  wait is a convenience: when `Forbidden` says it was refused, the tool warns
  and succeeds. A rollout that times out or fails is an error.
- `rbac.CanI` asks cluster-wide (`--all-namespaces`) for a cluster-scoped
  resource or an `AllNamespaces` probe, otherwise in its namespace, or in the
  context's namespace when none is given. A tool reads the probe's stdout alone
  and passes it to `rbac.Check`, whose refusal is a `*rbac.RefusedError` (find
  it with `errors.As`). Tools probe only what they are about to write.
- The namespace gate (`kube/nsgate`) runs after a successful remove and never
  fails it: protected, then `List` and `Occupants`, then `rbac.Check` of
  `DeleteProbe`, then the prompt, then `kube.DeleteNamespace`. Any refusal,
  error, occupant, decline, or missing terminal without `--no-prompt` keeps the
  namespace with a warning. `Occupants` takes the tool's own objects by kind and
  name, follows ownerReferences from them, and leaves out terminating objects and
  what Kubernetes puts in every namespace.

## Adoption order

Each step lands in M first. Then each tool gets one commit on its
`adopt-go-deploy-utility` branch. Tests that leave a tool must exist in M in the
same step. When a tool's coverage total falls because code moved out, the step
states the old and new totals and why.

| Step | Component | Behaviour change | Decision needed first |
| --- | --- | --- | --- |
| S1 | exitcode, notice, paritydoc, abbrev; both tools re-copy `tag.yml`, add `go.work` and `go.work.sum` to `.gitignore`, and require v0.1.0 through an untracked `go.work` | A: an env file that is missing, unreadable, unparsable or invalid, the files and environment variables it names, an unknown `--platform` or missing platform section, an unsafe configured command or name, and bad `status --user` or `--management-port` values exit 2 instead of 1; `validate` exits 2 when the env file has errors and 1 when only host checks fail; the alias tables are `abbrev` sets, so a declaration that breaks the shared rules fails at start-up. B: none; `internal/abbrev` and the lifted `exit.go` code move to the module, so its coverage total drops by what they carried. Both: the notice is built by `notice.New` and `docs/parity.md` is checked by `paritydoc.Check`, with no change to any committed doc but B's abbreviation.md naming the module's `abbrev` | none: env-file problems exit 2 in both (decided 2026-11-05) |
| S2 | confirm; the parity body's confirmation row becomes `aligned` | A: one prompter per run, so one piped write answers both `remove` questions, and end of input at a terminal prints a newline; otherwise none. B: a destructive confirmation whose stdin is not a terminal, without `--no-prompt`, fails with exit 1 and names the flag, and a pipe is never read (it used to read stdin, and end of input declined with a warning and exit 0); end of input at a terminal is the same error; `operator deploy` gains `--no-prompt`, which approves a downgrade but never the widening of the operator watch to every namespace, which still needs a terminal; the "refusing to ... without confirmation" warnings become the shared `confirming ... interactively` errors | none: the non-TTY rule was decided; `--no-prompt` approves the downgrade only (decided 2026-11-05) |
| S3 | execguard, envref; the parity body's rule 14 gains the escalator ban | A refuses invisible and C1 characters, invalid UTF-8, any escalator in a command (`sudo podman` goes; run `sudo solmq-conn-util` instead) and `--allow-command sudo`, accepts an embedded `~` and `.Exe`, and takes the shared guard messages; a guard refusal exits 2 wherever it surfaces; its other token and host-path checks are rebuilt from execguard's charset, so they refuse C1, invisible characters and invalid UTF-8 too; a set-but-empty `-env` variable fails deploy and warns at validate, and an invalid name is no longer echoed. B refuses C1 characters, invalid UTF-8 and 8 more escalators (`runuser`, `sg`, `chroot`, `gosu`, `su-exec`, `setuidgid`, `s6-setuidgid`, `chpst`), its escalator and leading-`~` messages change, an invalid `*Env` name is no longer echoed back, and its host paths and probe command refuse C1 | none, already decided (ban in both, 2026-11-02) |
| S4 | runner, and `exitcode.Exited` for B's session status; the parity body's "Runner seams" section is rewritten in the same step | A: every call takes a context and is interrupted, then killed, on cancel (was `Stream` only); a cancel is an error, which `logs --follow` maps to its clean end; errors read `kubectl: exit status 1`; B's resolve messages, and a relative resolution is refused; the optional `Streamer`, `Splitter` and `Attacher` interfaces go; a pipe held open after exit 0 no longer hangs. B: a cancel interrupts, then kills, instead of killing; stdin and env pass together; `RunInteractive` becomes `Attach`, which returns the exit code and is never killed; Echo annotates stdin and env on one line; `Quote` quotes more shell syntax and escapes unprintable runes; a held pipe no longer hangs; `EnvRunner` becomes the `engine.NoEnv` wrapper | none |
| S5 | containerart: compose, quadlet, artifacttest; the parity body's rule 11 states the quoting policy and names the shared packages | B: its artifact bytes stay identical. The timezone, which had no check, is held to the token charset without the tilde rule (as A's is); health-probe tokens and host paths also refuse invisible characters and invalid UTF-8. Its widening test gains A's spellings (`Privileged=true`, `no-new-privileges=false`, `NoNewPrivileges=false`, `SYS_ADMIN`, `label=disable`, `label:disable`, `Unconfined`), and the audit greps in `docs/container-security.md` follow. A: compose quotes the project `name:`, the service key, `container_name`, the `TZ` and `LOGGING_SYSLOG_*` values, volume entries, and secret and config `source:` names and their top-level keys. Every quadlet `Environment=` is quoted whole. The image, numbers and tool constants stay bare. Values routed to `Quote` or `Environment` are checked against execguard's charset, so `java-options` also refuses C1 and invisible characters and invalid UTF-8. The posture tests use `artifacttest`, keeping A's Kubernetes extras, and A gains B's quadlet key floor check (at its own 4.5 floor) and the rendered secret-target check | none: A adopts B's always-quote policy (decided 2026-11-04) |
| S6 | podman, systemd, rootless; the parity body's rootless, quadlet directory, host check, secret replace and subuid rows are updated in the same step | Both: rootless is the default and rootful podman is deployed only when `podman.rootful: true` asks for it, by root; a user asking for it, or root without it, is refused with exit 2 and nothing run. A: `podman.rootful` is new, so root now needs it where the mode used to follow the effective uid; the effective uid everywhere; the quadlet directory honours `XDG_CONFIG_HOME` and refuses a relative one or an unresolvable home; session repair before `systemctl --user`; a linger check, and deploy enables linger for its own account; the subid check asks `podman info` and refuses, in deploy and in validate, replacing the /etc/subuid warning, and its fix names both ranges; validate runs these probes; `is-active` is read as text from stdout and loses `--quiet`; a version with surrounding space still parses; a negative `NRestarts` is refused; no core check, since its unit asks for no core limit. B: `podman.rootless` gives way to `podman.rootful`, so rootless becomes the default, and the old key is refused at load naming the new one; `convert` writes `rootful: true` where the source meant rootful; `secret create --replace` on podman 4.7 and later, with a `podman version` call before the first secret; the `NRestarts` argv becomes `show <unit> -p NRestarts --value`; the core probe becomes `cat /proc/self/limits`, parsed in Go; the linger and core messages no longer name the broker; a reloading unit is restarted; id 0 passes the subid check; a relative `XDG_CONFIG_HOME` or an unresolvable home is a load error, retiring the `~` fallback | rootful only when asked for, by root (decided 2026-11-06); subid ranges are checked through `podman info` and a miss is refused, saying what to correct and how (decided 2026-11-04) |
| S7 | containercli, inspect, and `containercli.ComposeDigest` and `podman.SecretLabel` for the redeploy check; the parity body's preflight, exec, inspect and listing rules and its compose file and redeploy rows are updated in the same step | Both: a redeploy is left to the platform and asks before anything is written. Compose: the file carries a deploy-digest label interpolated from the environment (`ComposeDigest` of the file and every value the compose child receives), so compose itself recreates exactly when the file or a secret changed, and `up -d` always runs; a running container whose label differs asks first. Podman: the unit is compared with the one on disk and every secret with its digest label (`SecretLabel`) before any write; an unchanged secret is not replaced, nothing changed means nothing is written or restarted, and a change to an active unit asks before the restart. A "no" changes nothing and exits 0; `--no-prompt` approves; without a terminal and without `--no-prompt` the deploy fails with exit 1 naming the flag. Lifecycle verbs are `<engine> start <name>`, `stop` and `restart`; inspect is `container inspect`; one preflight message with merged hints; engine text is printed through `Printable`. A: deploy gains `--no-prompt`, so a CI deploy that changes a running service passes it; an unchanged podman deploy no longer restarts the unit; the listing gains a status column and `ValidName` replaces the `SafeToken` re-check of listed names; upload paths are quoted; the `--previous` refusal takes the shared wording. B: compose gets its file on stdin with `--project-directory` set to the env file's directory, and no compose file is written; the deploy digest replaces the on-disk diff and the docker certificate label, so a rotated password is detected; `docker.composeFile` is retired and refused by name; `compose -f <file> start`, `stop` and `restart` give way to engine verbs; `--restart` keeps meaning "recreate even when unchanged", and a stopped container is no longer force-recreated; every podman secret carries a digest label, not only the certificate; the restart question no longer declines silently without a terminal, and the stale-artifact warning goes; `exec -it` becomes `-i -t`; the logs flags take A's order; the listing adds `--no-trunc`; `Printable` also drops C1 controls | none: engine lifecycle verbs, and the redeploy prompt in both with `--no-prompt` (decided 2026-11-04); the redeploy is left to the platform and asks before any write (decided 2026-11-06) |
| S8 | watch | A: no screen clears into a pipe (section lines instead); a refresh reports its own error and returns nil; a fixed pause after each refresh rather than a fixed rate; the header reads `Every <d>: <title>   <time>   Ctrl-C to stop`. B: the same header text, and `--interval` gains a 1h ceiling | none |
| S9 | kube, rbac, nsgate; the parity body's namespace, waiting, permission-probe, can-i and namespace-gate rows are updated in the same step | Both: the namespace gate is the union of both tools and never fails a remove; a rollout wait refused with Forbidden is a warning; a refused namespace read plus a refused create warns and carries on; permission probes cover writes only. A: creates the namespace only when absent (read, probe, then the Namespace document applied alone); waits with `rollout status --timeout=300s` after deploy, start and restart; probes every kind it writes and drops its status, logs and cli probes; reads can-i by its last line, takes `no (...)` as a refusal and uses the shared messages; `-n` replaces `--namespace`; supplies its owned objects with their kinds (`CreatedNames` gains kinds); the gate probes `delete namespaces` and deletes the namespace by name; only the default service account's token Secret is discounted; a `{}` listing keeps the namespace; logs `--tail` is passed as a string. B: the gate lists JSON without `--ignore-not-found` and discounts its own objects (the custom resource as the root), terminating objects and the default token Secret; it probes before the prompt; without a terminal the namespace prompt keeps the namespace instead of failing; a refused rollout wait no longer fails a start or restart; `exec -it -n ns pod --` becomes `exec -i -t pod -n ns --`; `scale statefulset <n>` becomes `scale statefulset/<n>`; `errNotAllowed` becomes `*rbac.RefusedError`; can-i messages name no command, and an empty-namespace refusal says the current namespace; the matechannel exec uses `kube.Exec` | none: the union gate and write-only probes (decided 2026-11-04); least privilege (decided 2026-11-05) |

Compose reads its file from stdin in every implementation the tools allow, so no
interim file is written: docker compose v2 and `docker-compose` v1 document
`-f -`, and `nerdctl compose` loads through compose-go, which reads stdin for
`-` (checked from documentation and source on 2026-11-06; the S7 smoke run
confirms on a host). A's Kubernetes path has never run on a real cluster. Root deploying
on behalf of a rootless user is refused after S6 and is its own step: it needs a
key naming the user and a way to run podman and `systemctl --user` as that user.

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
