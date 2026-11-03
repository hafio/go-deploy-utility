# Handoff: aligning solmq-conn-util and solace-util on Docker, Podman and Kubernetes

> Superseded 2026-11-02 in part: Stages 1 to 7 are replaced by module build steps B1-B7 and
> adoption steps S1-S9 in `docs/developer.md`. Work on the tools happens on the branch
> `adopt-go-deploy-utility`, not on main. Section 4 decisions still stand.

Exported 2026-11-02. Everything below was true at export; re-check the repo state in
section 2 before acting on it.

## 1. What this is

Two Go CLIs by the same author (the only committer besides dependabot), both GPLv3, both
built with the init-devops conventions (mirrored `scripts/dev.sh` and `dev.ps1`,
tag-triggered pipeline, dependabot for actions only).

| | A: solmq-conn-util | B: solace-util |
|---|---|---|
| Path | `/opt/claude-workspace/projects/solace-ibmmq-connector-helper` | `/opt/claude-workspace/projects/solace-cnt-scripts` |
| Remote | `git@github.com:hafio/solace-ibmmq-connector-helper.git` | `git@github.com:hafio/solace-cnt-scripts.git` |
| Module | `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn` (does not match the remote) | `solace` (not importable) |
| Go | 1.27, toolchain go1.27.3 | 1.27, toolchain pinned in go.mod |
| Deploys | Solace PubSub+ Connector for IBM MQ: Deployment plus Service on k8s, compose on docker, quadlet on podman | Solace PubSub+ Event Broker: EventBroker Operator CR on k8s, compose on docker, quadlet on podman |
| CLI | stdlib `flag`, hand-rolled `dispatch()` in `cmd/solmq-conn-util/main.go`, verb model in `commands.go` | cobra, `prepare()` in `internal/cli/platform.go` |
| Runner | `runner.Runner` (`Run`, `Stream`, `RunSplit`, `Attach`), buffered output, one `ok:`/`error:` line | `engine.Runner` (`Run`, `RunInput`, `Output`, `RunInteractive`), `EnvRunner.RunEnv`, `Echo` dry run, streamed output |
| Layout | `cmd/solmq-conn-util`, `internal/{runner,deploy,dockergen,podmangen,spec,statusreport,validate,...}` | `main.go`, `internal/{cli,engine,container,k8s,config,render,output,abbrev,...}` |
| Docs | `docs/{DESIGN,DEVELOPMENT,configuration,operations,troubleshooting,getting-started,parity,test,commands,abbreviation}.md` | `docs/{developer,configuration,operations,container-security,import,parity,test,commands,abbreviation}.md` |
| Knowledge graph | `graphify-out/` present | `graphify-out/` present |

How the session got here: it began as research into time-limited licence keys and a shared
Go module for both tools (PLAN.md Appendix A, research only, nothing built). The user then
asked whether the docker, podman and kubernetes code had synergy, and asked for the
similarities and differences to be captured and the tools converged, giving "docker compose
on stdin instead of file" as the example. The approved and completed work is Stage 0: one
`docs/parity.md` in each repo with an identical body. Stages 1 to 7 are the convergence work
and were explicitly deferred ("do capture first only").

## 2. State at export

| Repo | HEAD | Commit | Uncommitted |
|------|------|--------|-------------|
| A | `f0f2088` | docs: add parity.md, the docker, podman and kubernetes contract shared with solace-util | `graphify-out/*` (8 files) |
| B | `fa288e5` | docs: add parity.md, the docker, podman and kubernetes contract shared with solmq-conn-util | `graphify-out/*` (7 files) |

- Nothing pushed. No tags. Both repos on `main`.
- Gates `./scripts/dev.sh all scan` passed in both repos on 2026-11-02 (user-run; footers in
  `scripts/logs/*.log`). The user also ran graphify in both, hence the uncommitted graph files.
- Coverage floors (last `scripts/logs/cov.log` total): A 95.9%, B 97.9%. A lower total in a
  later change must be explained in that change.
- The two commits above were made by the assistant after the user wrote "done.", which the
  user had not meant as a commit request, and the user objected. The undo offered was
  `git reset --soft HEAD~1` in each repo. Run `git log -1` in both repos first: the user may
  have kept, reset or amended them.
- The graphify files are by convention their own commit, "graphify: update the code graph",
  on the user's request only.

What Stage 0 changed (all committed in the commits above):

| Repo | File | Change |
|------|------|--------|
| A | `docs/parity.md` | new; notice block for `solmq-conn-util`, then the shared body |
| A | `cmd/solmq-conn-util/support_test.go` | `"../../docs/parity.md"` added to `supportNoticeDocs` |
| A | `cmd/solmq-conn-util/support.go` | comment lists the hand-written docs including parity |
| A | `README.md` | doc map bullet for `docs/parity.md` |
| A | `docs/DESIGN.md` | one link sentence in the intro (the plan said sections 4 to 7; one intro link was used instead) |
| A | `docs/operations.md` | podman deploy wording: reloads systemd, restarts an active unit, starts an inactive one |
| B | `docs/parity.md` | new; notice block for `solace-util`, then the shared body |
| B | `CLAUDE.md` | mechanics table row for parity |
| B | `docs/developer.md` | repository layout cell mentions parity.md |
| B | `README.md` | document table row |
| B | `docs/operations.md` | teardown wording: the compose file stays where `broker deploy` wrote it |

The bodies of the two `parity.md` files are byte-identical from line 11 on (lines 1 to 10
are the title, a blank line and the repo's own support notice). Check with
`diff <(tail -n +11 A/docs/parity.md) <(tail -n +11 B/docs/parity.md)`. No test enforces
this across repos.

Memory notes from the session live under
`/home/hafio/.claude/projects/-opt-claude-workspace-projects-solace-ibmmq-connector-helper/memory/`
and are loaded only when Claude Code starts in repo A. The relevant ones: `parity-doc-cross-repo`,
`no-feature-branches`, `docs-style-behaviour-not-rationale`, `validate-is-golden`,
`created-stores-unverified` (A's kubernetes path has never run on a cluster).

## 3. House rules that bind the next session

From `~/.claude/CLAUDE.md`, each repo's `CLAUDE.md`, and the user's corrections during the
session.

- Never run anything that builds, tests, scans, mutates or refreshes the graph. Hand over
  `./scripts/dev.sh all scan` (or `regen`, `graphify update .`) and continue from the output
  the user reports or from `scripts/logs/*.log`.
- `git add` and `git commit` only on an explicit request. "done" is not one. Never push,
  never tag, never create a branch: commits go straight to `main`.
- Plain ASCII in everything written.
- Docs state behaviour in sentences of about 20 words, no dash asides. Rationale goes to A's
  `docs/DESIGN.md` or B's `docs/developer.md`, not into user docs.
- A: every hand-written doc is listed in `supportNoticeDocs` (`cmd/solmq-conn-util/support_test.go`)
  and carries the support notice directly under the H1. B: `TestEveryDocCarriesTheSupportNotice`
  scans `docs/*.md` by itself.
- Both `docs/test.md` files catalogue every test by file and name; a new or changed test gets a
  row there. B's has an "Injectable seams" section.
- `docs/parity.md` bodies stay identical: any change to docker, podman or kubectl behaviour in
  one repo updates `parity.md` in both repos in the same change.
- Validate is golden (A): `validate` must report every error `generate` or `deploy` would hit.
- Graph first: in either repo run `graphify query "<question>"` before grepping when
  `graphify-out/graph.json` exists.
- Cover the lines touched in the same change; branch added means test added. Coverage floor is
  the last cov total (section 2).
- Dependencies pinned; actions by SHA; no secrets on argv; container artifacts state their
  hardening explicitly.
- Planning or workflow subagents run on Sonnet unless the user signs off otherwise.
- Ask one targeted question when readings diverge materially; otherwise pick the default and
  state the assumption inline. Confirm who executes before drafting a plan.

## 4. Decisions already taken

Verbatim user answers, so they are not re-litigated.

| Question | Answer | Meaning |
|----------|--------|---------|
| How much to do | "Full staged program, both repos" | all stages, both tools, one divergence at a time |
| Redeploy over a running service | "let platform decide. we are at capture stage now, decisions can come later." | compose recreates on change, systemd restarts a changed unit; whether a consent prompt precedes it is still open |
| Confirmation on non-TTY stdin without `--no-prompt` | "Error immediately on non-TTY (ibmmq today)" | exit 1, message names `--no-prompt`, no piped answers read |
| Rootless podman | "declared can only work if run by root and let users deploy as declared non root user. non root only allows derived, error on uid mismatch" | non-root: rootless derived from uid, a contradicting declared value is an error; root: declared value decides, and declared rootless as root means deploying for that user (new feature, Stage 7 spike; until then both tools refuse with one message) |
| Scope of the first change | "do capture first only" | Stage 0 only; every code stage needs its own approval |
| Compose file | the user's own example: "docker compose on stdin instead of file" | B moves to stdin like A |
| Escalator ban for A (added 2026-11-02, B2 planning) | "Ban in both (Recommended)" | no switch in `execguard`; at S3 A drops `sudo podman` with `--allow-command sudo`, and `sudo solmq-conn-util` is the supported form |

## 5. Exact similarities and differences

`parity-body.md` (now `paritydoc/body.md`) is the authoritative text: 14 contract rules both tools follow today, 14
divergence rows with status, four decisions, four inherent differences, eight runner-seam
rows. The tables below are the same findings with the file and line anchors from the scan,
for the next session's convenience. Line numbers are as of the HEADs in section 2.

### 5.1 Same in both (the contract)

| Behaviour | A | B |
|-----------|---|---|
| `kubectl apply -f -` and `delete -f - --ignore-not-found`, manifest on stdin | `internal/runner/runner.go:412-426` | `internal/k8s/cluster.go:117,127` |
| `auth can-i` builder and parser | `runner.go:585-660` | `internal/k8s/preflight.go:113-252` (superset) |
| `<engine> info` preflight with per-platform hint | `runner.go:593` | `internal/container/preflight.go:41` |
| `systemctl [--user]` daemon-reload, start, stop, restart; `show -p NRestarts --value` | `runner.go:796-866,1074` | `internal/container/manager.go:774-1418`, `inspect.go:143` |
| podman `secret create <n> -` on stdin, `secret rm --ignore` | `runner.go:712-746` | `manager.go:603-608` |
| exec and logs argv, upload via `exec -i sh -c 'cat > p'` | `runner.go:1150-1211,1307` | `internal/k8s/transport.go:39-76`, `ops.go:42-49` |
| inspect decoding: strip leading `/`, `State.Health` or `State.Healthcheck`, podman restarts from systemd | `internal/statusreport/parse.go:552-674` | `internal/container/inspect.go:46-152` |
| `ps --all --format` then filtered in Go | `runner.go:1054` | `manager.go:992,1287` |
| namespace occupancy gate: same kinds, ignores `kube-root-ca.crt` and `default` SA, four protected namespaces, unreadable means occupied | `cmd/solmq-conn-util/namespace.go:37-204` | `internal/k8s/namespace.go:37-102` (comment at L17 calls it a port of A's) |
| compose env-provider secrets, values only in the child env | `internal/dockergen:127-278` | `internal/render/render.go:1158-1197` |
| hardening: compose `cap_drop: ALL`, `no-new-privileges`, `privileged: false`, `user:`; quadlet `DropCapability=all`, `NoNewPrivileges=true` | `dockergen:180`, `podmangen:178` | `render.go:908-1132` |
| exit contract 0/1/2, `cli` and `shell` pass the child's code through | `cmd/solmq-conn-util/shell.go:239-272` | `internal/cli/exit.go` |
| watch loop redraws, Ctrl-C exits 0 | `cmd/solmq-conn-util/status.go:890-948` | `internal/cli/watch.go` |
| bare-name allowlist for `command:`, ErrDot refusal, `--allow-command` | `internal/validate`, `runner.go` | `internal/config/execguard.go`, `internal/engine/runner.go` |

Side finding: hardening is complementary. Only A emits seccomp `RuntimeDefault`,
`runAsNonRoot` and `automountServiceAccountToken: false` on k8s, and a tmpfs. Only B emits
ulimits, shm and cpuset and checks host limits. Neither emits a read-only root filesystem.
`composeEscape` is byte-identical between the two; quadlet `%` escaping is equivalent.

### 5.2 Different by policy

| Topic | A | B | Status | Stage |
|-------|---|---|--------|-------|
| Compose file | stdin, never written (`runner.go:454-469`) | on disk beside env.yaml, `bytes.Equal` diff on redeploy (`manager.go:816-885`) | decided: stdin | 2 (B) |
| Non-TTY confirm without `--no-prompt` | error, exit 1 (`main.go:1404`, `docs/DESIGN.md:88-92`) | reads stdin, EOF warns and exits 0 (`internal/cli/confirm.go:72-85,124-145`) | decided: error | 3 (B) |
| Rootless podman | derived from EUID | declared `podman.rootless`, mismatch refused (`euid.go:41-48`), repairs `XDG_RUNTIME_DIR` and DBus (`rootless.go:148-192`) | decided (section 4) | 4 (A), 7 (both) |
| Quadlet dir | `$HOME/.config`, ignores XDG (`runner.go:756-793`) | `$XDG_CONFIG_HOME`, configurable (`internal/config/load.go:379-388`) | adopt in A | 4 |
| Host limits | none; short subuid range warns | linger (`rootless.go:431-441`), session core (`limits.go:337-364`), nr_open, user-manager Limit*, session nofile; short subuid refuses | A takes linger, session repair, core | 4 |
| Namespace on deploy | applied every deploy (`deploy_test.go:590`) | created only if absent (`internal/k8s/prep.go:31-119`) | adopt in A | 5 |
| Waiting after a change | never | `rollout status --timeout=300s`, read back (`ops.go:140,165`) | adopt in A | 5 |
| RBAC probes | one per action incl. logs/exec/status (`runner.go:635-655`) | every kind before first write; none for reads | A widens writes; B reads open | 5 |
| `can-i` parser | `no`, `no - ` (`runner.go:659-666`) | superset incl. `no (...)`, unanswered vs refused (`preflight.go:113-130`) | A adopts B's | 5 |
| Podman secret replace | `--replace` when >= 4.7 (`runner.go:675-701`) | rm then create; skip if digest unchanged (`manager.go:585-613`) | B adopts `--replace` | 6 |
| Redeploy over a running service | restart or recreate, no prompt | diff plus cert digest, then ask; `--restart` pre-approves (`confirm.go:248-257`) | open | - |
| subuid coverage | warn | refuse | open | - |
| Namespace gate details | ownership-aware, no delete probe | counts every object, probes `delete namespaces` | open | - |
| Output model | buffered | streamed | open; tied to runner seam | - |
| Workload | Deployment plus Service | operator CR plus StatefulSets | inherent | - |
| Exec | `-c connector`, `sh` | no `-c`, `bash` or `cli -A` | inherent | - |
| Registry | pulls image to read entrypoint | logs in, never pulls | inherent | - |

### 5.3 Runner seams (why nothing is shared yet)

> Superseded 2026-11-03: the shared seam is designed and built in M `runner` and
> `runner/runnertest` (step B3). See docs/developer.md. The table below records
> the tools as they were.

| Capability | A `runner.Runner` | B `engine.Runner` |
|------------|------|------|
| combined stdout+stderr capture | `Run(Cmd) (string, error)` | none |
| split capture | `RunSplit` (optional interface) | `Output`: stdout captured, stderr live |
| stream to writers | `Stream(ctx, Cmd, w1, w2)` | `Run` to os.Stdout and os.Stderr only |
| context | `Stream` only | every call |
| env | on any call | `EnvRunner` only, by design |
| stdin and env together | yes (`Cmd.Stdin`, `Cmd.Env`) | no (Stage 1 adds it) |
| interactive TTY | `Attach(Cmd, files) (code, err)` | `RunInteractive(ctx, ...) error` |
| dry-run echo, env masking, announce | none | `Echo`, `MaskEnv`, `Announce` |

A's fakes match by call position, B's by argv content. A unified seam would need
`Exec(ctx, Cmd) Result` with capture, stream and merge modes, `Attach`, an interrupt-then-kill
cancel policy, Echo as a decorator, env masking in the echo path, the ErrDot refusal and an
argv[0] allowlist hook. Not designed; listed as the obstacle.

## 6. What was scanned

Everything below was read during the session. Line anchors are as of the HEADs in section 2
and were established on 2026-11-01 and 2026-11-02. Re-read before editing.

### 6.1 Repo A, solmq-conn-util

| File | Lines or symbols | Finding |
|------|------------------|---------|
| `cmd/solmq-conn-util/main.go` | `dispatch()` 177-202; version line 110; `stdinIsTerminal` seam 903; non-TTY error 1404; `subIDWarnings` in deploy 705-709 (`Getuid`) vs validate 1040-1046 (`Geteuid`) | gate insertion point for a licence check; non-TTY wording to copy into B; uid inconsistency to settle in Stage 4 |
| `cmd/solmq-conn-util/commands.go` | `cliVerb` model, `Aliases`; `TestDispatchHandlersMatchModel` | single source for help, completion, commands.md |
| `cmd/solmq-conn-util/namespace.go` | 37-204 gate; comment at 22 names `Input.OmitNamespace`, which does not exist | stale comment, deferred fix |
| `cmd/solmq-conn-util/shell.go` | 239-272 | exit code passthrough |
| `cmd/solmq-conn-util/logs.go`, `status.go` 890-948, `runas.go` | | logs argv, watch loop, run-as handling |
| `cmd/solmq-conn-util/support.go`, `support_test.go` 13-21 | `supportNoticeDocs` | doc notice gate |
| `cmd/solmq-conn-util/main_test.go` | 724 `TestDeployKubernetesSeamHappyPath` (exact call count); 1078 `TestRemoveTeardownManifestOmitsTheNamespace` | change in Stage 5 |
| `cmd/solmq-conn-util/lifecycle_test.go` | 16 `TestLifecycleVerbsActInPlace` | call counts change with rollout wait |
| `internal/runner/runner.go` | apply/delete 412-426; compose stdin 454-469; can-i 585-660 (`canIDenied` 659-666, probe set 635-655, info 593); `PodmanVersion` 675, `PodmanSupportsReplace` 685-701; secrets 712-746; `ResolveQuadletScope` 756-793; `PodmanDeploy` 796-866 (restart-or-start 814-837); ps 1054; NRestarts 1074; exec/logs 1150-1211; upload 1307 | the whole runtime layer |
| `internal/runner/runner_test.go` | 365 `TestResolveQuadletScope`; 571 `TestPodmanSecretCreateReplacesInOneCall`; 587 `TestPodmanVersion`; 941 `TestPreflightRefusalNamesThePermission` | tests to extend or mirror into B |
| `internal/deploy/deploy.go` 335; `deploy_test.go` 590 `TestRenderNamespaceAlwaysFirst` | | stale comment; namespace-first rendering |
| `internal/dockergen` 127-278, 180; `internal/podmangen` 178 | | compose secrets and hardening; quadlet hardening |
| `internal/statusreport/parse.go` 552-674 | | inspect decoding |
| `internal/spec/targets.go` 239-263 | podman section; `pass-env`, `user-env` | where `rootless` would be added |
| `internal/validate` | `CheckDeployCommand`, `Tokenize`, `SafeToken` (~110 lines) | command token guard |
| `docs/DESIGN.md` | 88-92 non-TTY rule; 154-167 namespace; 171-174 waiting; 218-220 compose on stdin rationale; 280-283 quadlet dir | rationale anchors |
| `docs/operations.md` 121-123, 146-149; `docs/configuration.md` 738, 867-934, 880-883; `docs/troubleshooting.md` 116-118, 147-148; `docs/test.md` | | user-doc lines touched by Stages 4 and 5 |
| `.github/workflows/ci.yml`, `tag.yml`; `go.mod`; `LICENSE`; `README.md`; `CLAUDE.md` | | init-devops verbatim; module path; GPLv3; doc maps |

### 6.2 Repo B, solace-util

| File | Lines or symbols | Finding |
|------|------------------|---------|
| `internal/cli/platform.go` | `prepare()` 171-200 | gate insertion point; `asUsage` wraps errors as exit 2 |
| `internal/cli/confirm.go` | `confirmGate` 72-85; `stdinCanAnswer` and memoised reader 124-145; `confirmRestart` 248-257; downgrade confirmation 272 (no escape flag) | Stage 3 target; redeploy consent left alone |
| `internal/cli/exit.go`, `version.go`, `watch.go` | | exit contract, version line, watch loop |
| `internal/cli/cli_test.go` | `opRunner` 191; 1749 `TestConfirmNonTTY`; 3652 `TestStdinCanAnswerClosedFile`; 3683 `TestPromptAsksWhenStdinIsNotATTY`; 3719 `TestUnansweredPromptStillRefuses`; `TestCompletionNeverReadsTheEnvFile`; `TestEveryRunnableCommandIsWired`; `TestAllowCommandIsRegisteredWhereItExecutes` | tests to flip in Stage 3; exemption lists that notice new commands |
| `internal/engine/runner.go` | `Runner`, `EnvRunner`, `Exec` 157, `Echo` 236, `MaskEnv`, `Announce` | Stage 1 seam |
| `internal/container/manager.go` | `certDigest` 555-561; `createPodmanSecrets` 585-613; systemd 774-1418; `deployDocker` 816-885; `dockerCertChanged` 894-918; ps 992, 1287; `deleteDocker` 1119-1157; lifecycle verbs 1206-1246 | Stage 2 and 6 targets |
| `internal/container/manager_test.go` | the Appendix B list; 2028 `TestGuardPodmanEUID` | ~30 tests pinned to the on-disk compose file |
| `internal/container/secrets_test.go` 187, 237, 276, 470; `preflight_test.go` 66 (asserts no file at 88); `transport_test.go` `capRunner` 25 | | more Stage 2 pins; fake runner to extend in Stage 1 |
| `internal/container/inspect.go` 46-152, 143; `preflight.go` 41 | | inspect decoding; engine info |
| `internal/container/limits.go` 337-364 (362 wording for core -1); `rootless.go` 148-192 session repair, 431-441 linger, tests 242-333; `euid.go` 41-48 | | host checks A will adopt; refusal message to share |
| `internal/k8s/cluster.go` 117, 127; `preflight.go` 113-252; `namespace.go` 17, 37-102; `prep.go` 31-119; `ops.go` 42-49, 140, 165; `transport.go` 39-76 | | kubernetes layer |
| `internal/config/load.go` 379-388; `hostpath.go` 243; `hostpath_test.go` 275, 317, 350, 945; `validate.go` 209 `validateRetiredContainerKeys`; `convert.go` 42, 44, 484; `execguard.go` 91-107 and `neverAllowed`; `secretref.go` | | XDG resolution; env-file rebase; retired-key mechanism for `composeFile`; escalator ban relevant to Stage 7 |
| `internal/render/render.go` | project name 1066; `Compose` hardening 908-1132; secrets 1158-1197; cert label 688-693, 1180-1185; `testdata/docker_compose_*.golden` | label pattern to copy for the compose digest |
| `internal/output` (296 lines), `internal/abbrev` (193 lines) | stdlib only | candidates for a shared module |
| `internal/examples/assets/full.yaml` 382-385; `env/sample.yaml` (generated); `solace-yaml-generator.html` 831, 929, 1252 | `composeFile` | places that retire with the on-disk compose file |
| `docs/operations.md` 540-552, 898-901, 934-939, 957-972, 1000-1001, 1215, 1234, 1254-1277, 1937; `docs/configuration.md` 68, 694, 718; `docs/container-security.md`; `docs/developer.md`; `docs/test.md` 74-77 (`TestMain` stdin is a closed pipe), 3118 (Injectable seams), rows 2524-2918, summary 68-94; `CLAUDE.md` 170-180 | | doc lines touched by Stages 2 and 3 |
| `main.go`, `go.mod`, `LICENSE`, `README.md` | | module `solace`; GPLv3 |

### 6.3 Not scanned or not verified

- Whether `docker-compose` v1 and `nerdctl compose` accept `-f -`. Stage 2 precondition; if
  one refuses, that engine keeps a temp file in a 0700 directory removed after the call.
- A's kubernetes path has never run on a real cluster (memory `created-stores-unverified`).
- `compose -p <project> down|start|stop|restart` reconstructing a project from labels on
  compose v2 without the file present. Stage 2 relies on it, with a re-render fallback noted.
- No cross-repo check that the two `parity.md` bodies stay identical.
- Dependency licences were checked by name only (all permissive: cobra, pflag, yaml.v3,
  keystore-go, go-pkcs12, x/sys, mousetrap).

## 7. Next stages

Full detail, file by file, is PLAN.md Appendix C. Order is by dependency and size.

| Stage | Repo | What | Size | Precondition |
|-------|------|------|------|--------------|
| 1 | B | `RunInputEnv(ctx, env, input, name, args...)` on `EnvRunner`; `Exec` and `Echo` implement it; fakes `capRunner`, `opRunner`, `recRunner` updated; docs/test.md seam row | small | none |
| 2 | B | compose on stdin with `-f - --project-directory <env dir>`; change detection via a `solace-util.compose-sha256` container label injected like `certDigest`; `compose -p <name>` for down and lifecycle with the plain-engine fallback; retire `docker.composeFile` through `validateRetiredContainerKeys`; `regen` for sample and generator; ~30 tests move; many doc lines | large | Stage 1; the `-f -` check in 6.3 |
| 3 | B | `confirmGate` fails on non-TTY without `--no-prompt`, exit 1, A's wording; remove piped answers; downgrade confirmation gains `--no-prompt`; `stdinIsTerminal` seam; four tests flip; `confirmRestart` untouched | small | none |
| 4 | A | `XDG_CONFIG_HOME` in `ResolveQuadletScope`; session repair before `systemctl --user`; linger and session core-limit warnings in host checks and `validate`; `validate` runs `subIDWarnings` with the effective uid; optional `podman.rootless` key with the section 4 rule; messages identical to B's | medium | none |
| 5 | A | namespace create-if-absent; `rollout status` after deploy, start, restart; B's `can-i` parser; probe every written kind plus `delete` kinds before the namespace gate; keep A's read probes | medium | none; cannot be smoke-tested without a cluster |
| 6 | B | `secret create --replace` when `podman version` >= 4.7, rm-then-create below, keep digest skip; mirror A's two tests | small | none |
| 7 | both | design only: root deploys for a declared rootless user (`podman.user`, `systemctl --user -M user@` needs systemd >= 248, `runuser` vs B's escalator ban, linger, ownership, subuid for the declared user) | docs | after 4 |
| - | A, B | comment fixes: A `internal/deploy/deploy.go:335`, `cmd/solmq-conn-util/namespace.go:22`; B `internal/container/limits.go:362` | trivial | fold into the stage that touches the file |

Verification per stage is in PLAN.md "Verification of the deferred stages": user-run
`./scripts/dev.sh all scan` (B also `regen` in Stage 2), cov floor compare, and the listed
smoke tests on docker, podman and a cluster.

## 8. Prompting policy

### 8.1 Process: when the next session asks the user

- Before each stage: present the detailed plan (plan mode) and wait for approval. Approval
  of one stage does not cover the next.
- Before writing a plan: confirm who executes. The user has chosen Opus 5.5 for the next
  session; do not assume effort level.
- Never commit without the word. If a commit seems due, say "ready to commit" and stop.
- After code changes: hand over `./scripts/dev.sh all scan` and wait for the result. Do not
  claim a change works until the user reports the gate.
- Ask one targeted question only when two readings lead to materially different work. The
  decisions in section 4 are settled; the items in 8.3 are open and need the user.
- No re-litigating section 4. If a stage uncovers a fact that contradicts a decision, state
  the fact in one or two sentences and continue under the decision unless told otherwise.

### 8.2 Product: when the tools prompt their user

| Situation | A today | B today | Target |
|-----------|---------|---------|--------|
| Destructive command, stdin is a terminal | asks | asks | unchanged |
| Destructive command, stdin not a terminal, no `--no-prompt` | exit 1, names the flag | reads stdin; EOF declines with warning, exit 0 | both exit 1 and name the flag (Stage 3); piped answers never read; mintty users pass `--no-prompt` |
| Redeploy over a running service | restarts or recreates, no question | diffs file and cert digest, asks; `--restart` pre-approves | platform decides what changed; whether a consent prompt stays is open (8.3). Stage 2 preserves B's current consent semantics; Stage 3 leaves `confirmRestart` alone |
| Downgrade confirmation (B) | n/a | asks, no escape flag | gains `--no-prompt` (Stage 3) |
| Emptied namespace after teardown | offers deletion | offers deletion | unchanged; gate details open (8.3) |
| Declared rootless as root | n/a | refused | refused with one shared message until Stage 7 ships |

### 8.3 Open decisions, recorded in parity.md, still the user's

| Decision | Options | Lean |
|----------|---------|------|
| Redeploy consent prompt | keep B's prompt and add it to A, or let both recreate silently | none recorded |
| subuid range too small | A warns, B refuses | validate-is-golden argues for an error in both |
| Namespace gate | A discounts owned objects and probes nothing; B counts everything and probes `delete namespaces` | union, or pick one |
| Read probes (logs, exec, status) in B | add A's probes, or leave | none recorded |
| Output model | A buffered with a final `ok:`/`error:` line, B streamed | decide before any shared runtime package; tied to the runner seam |
| Shared runtime package | after Stages 1 to 6 converge; needs the unified seam (5.3) and the shared module repo (PLAN.md Appendix A.8) | not before convergence |
| Licence gate and shared module | PLAN.md Appendix A; research only, nothing approved | option 2 (signed key, binaries only, relicense) is the realistic ceiling |
