# Align docker, podman and kubernetes behaviour across solmq-conn-util and solace-util

A = /opt/claude-workspace/projects/solace-ibmmq-connector-helper (solmq-conn-util)
B = /opt/claude-workspace/projects/solace-cnt-scripts (solace-util)

## Context

Both tools drive the same three platforms with the same external commands, and B already carries
a hand-ported copy of A's namespace gate (`internal/k8s/namespace.go:17`). Where they differ it is
policy, not platform knowledge: compose file on stdin vs on disk, prompt rules, how rootless is
decided, what is probed before a write. The user wants the similarities and differences captured
in each repo, then the two tools converged one divergence at a time, so a shared runtime package
becomes possible later. Licensing and the shared-module research that led here is kept as
Appendix A.

Decided by the user:
- Full staged program across both repos.
- Redeploy over a running service: the platform decides (compose recreates on change, systemd
  restarts a changed unit). Whether a consent prompt stays is an open decision; capture, do not
  settle.
- Non-TTY confirmation without `--no-prompt`: fail immediately with exit 1 naming the flag (A's
  rule, docs/DESIGN.md L88-92). No piped answers.
- Rootless podman: a non-root invoker gets rootless derived from its uid, and a declared value
  that contradicts it is an error. Only root may use a declared value, and a declared rootless
  deploy as root means deploying on behalf of the declared non-root user. That last part is a
  new feature in both tools and is staged as a design spike.

**This plan covers Stage 0 only: documentation, no code.** Stages 1 to 7 (code changes) are
kept in Appendix C for a later, separate approval.

House rules: one commit per repo on main (no branches, no push); the user runs
`./scripts/dev.sh all scan` and reports; plain ASCII; docs state behaviour in sentences of
about 20 words, no dash asides; rationale goes to A's docs/DESIGN.md or B's docs/developer.md.

## Stage 0: capture (both repos, docs only)

### Files

| Repo | File | Change |
|------|------|--------|
| A | `docs/parity.md` | new |
| A | `cmd/solmq-conn-util/support_test.go:13-21` | add `docs/parity.md` to `supportNoticeDocs` (test list, not behaviour) |
| A | `README.md` doc map, `docs/DESIGN.md` sections 4-7 | one link line each |
| A | `docs/operations.md:146-147` | reword: deploy starts a stopped unit and restarts only an active one (`runner.go:814-837`) |
| B | `docs/parity.md` | new; `TestEveryDocCarriesTheSupportNotice` scans `docs/*.md` by itself |
| B | `CLAUDE.md` doc map, `docs/developer.md` | one link line each |
| B | `docs/operations.md:1215` | reword: teardown does not remove the compose file (`deleteDocker` never does) |

Not in this stage, listed for later with the code stages: stale Go comments at A
`internal/deploy/deploy.go:335` and `cmd/solmq-conn-util/namespace.go:22` (`Input.OmitNamespace`
does not exist), and B `internal/container/limits.go:362` remedy wording for a core value of -1.

### parity.md content

Same body in both repos. Only the support notice block differs (each repo's own wording, directly
under the H1, blank line either side, before the first `## `). Title: `# Platform parity:
solmq-conn-util and solace-util`. Sections:

1. **Scope.** What this doc is: the contract both tools follow for docker, podman and kubectl,
   the divergences with their status, and the decisions taken. Where each tool documents its
   own platform behaviour (A docs/DESIGN.md sections 4-7 and docs/operations.md; B
   docs/operations.md and docs/container-security.md).
2. **Contract.** Each row of the "same in both" table in A.7.1 becomes one rule in prose plus a
   two-column line naming the implementing file in A and B. Fifteen rules: apply and delete on
   stdin; `auth can-i` probe shape; `<engine> info` preflight; systemctl argv and NRestarts;
   podman secrets on stdin; exec, logs and upload argv; inspect decoding (leading slash,
   Health or Healthcheck, systemd restarts on podman); `ps --all` filtered in Go; namespace
   occupancy gate; env-provider compose secrets; hardening lines; exit contract and child code
   passthrough; watch loop; ErrDot refusal and bare-name allowlist.
3. **Divergences.** The "different by policy" table in A.7.1 with columns Topic, solmq-conn-util,
   solace-util, Status. Status values: `decided, not yet applied` (compose on stdin, non-TTY,
   rootless rule), `open` (redeploy consent, subuid, namespace gate, logs/exec probes, output
   model), `adopt` (XDG dir, host checks, namespace create-if-absent, rollout wait, can-i parser,
   secret `--replace`). No stage numbers in the doc; those live in this plan.
4. **Decisions.** Three statements of behaviour both tools will follow:
   - Compose file: handed to compose on stdin with `--project-directory` set to the env file's
     directory; never written to disk.
   - Confirmation without a terminal: a destructive command whose stdin is not a terminal fails
     at once with exit 1 and names `--no-prompt`; piped answers are not read. Known cost: under
     mintty-style shells stdin is a pipe, so those users pass `--no-prompt`.
   - Rootless podman: a non-root invoker is rootless; a declared `rootless: false` is an error
     naming the uid. Root follows the declared value. A declared rootless deploy as root means
     deploying for the declared non-root user; until that ships, both tools refuse it with the
     same message.
   - Redeploy over a running service: the platform decides what changed (compose recreates on
     change, systemd restarts a changed unit). Whether a consent prompt precedes it is open.
5. **Inherent differences.** Deployment vs operator CR and StatefulSets; `-c connector` and
   `sh` vs `bash` and `cli -A`; image pull for the entrypoint check vs registry login;
   broker-only host limits (nofile, memlock, shm).
6. **Runner seams.** The capability table from A.7.1 (eight rows), as the record of why a shared
   driver does not exist yet. No design here.

### Verification

User-run, per repo: `./scripts/dev.sh all scan`. In A, `TestSupportNoticeInHandWrittenDocs`
proves the notice placement and the list entry; in B, `TestEveryDocCarriesTheSupportNotice`
does the same. Open both `parity.md` files side by side and confirm the bodies are identical
below the notice (`diff <(tail -n +N a) <(tail -n +N b)` where N skips the title and notice).
Then one commit per repo on main.

---

# Appendix C: deferred code stages (not in this plan, separate approval)

Also deferred: comment fixes at A `internal/deploy/deploy.go:335`, `cmd/solmq-conn-util/namespace.go:22`
and B `internal/container/limits.go:362` wording.

## Stage 1: B runner seam, minimum for stdin plus env

`internal/engine/runner.go`: `RunEnv` cannot feed stdin and `RunInput` cannot add env. Add one
method that does both (`RunInputEnv(ctx, env, input, name, args...)`) on `EnvRunner`, implement
in `Exec` (L157) and `Echo` (L236: show stdin as a byte count, env as `KEY=***` via `MaskEnv`).
Update the fakes that implement `EnvRunner`: `capRunner` (`internal/container/transport_test.go:25`),
`opRunner` (`internal/cli/cli_test.go:191`), and any `recRunner` embedding. Catalogue in
docs/test.md "Injectable seams" (L3118). No behaviour change yet.

## Stage 2: B compose on stdin, like A

Target shape (A: `internal/runner/runner.go:454-469`): `<compose> -f - --project-directory
<env-file dir> up -d`, file on stdin, secrets in env, nothing written to disk.

Changes in `internal/container/manager.go`:
- `deployDocker` (L816-885): drop `writeArtifact` for compose. Change detection moves from the
  on-disk `bytes.Equal` to a label on the container: render carries
  `solace-util.compose-sha256: "${<EnvVar>_COMPOSE_SHA256}"` beside the existing cert label
  (`internal/render/render.go:1180-1185`), the manager injects the sha256 of the rendered bytes
  the way `certDigest` (L555-561) is injected, and reads it back with the same
  `inspect --type container --format {{index .Config.Labels ...}}` as `dockerCertChanged`
  (L894-918). Flow stays: not running -> `up -d`; running and either label differs -> consent
  path as today; unchanged -> nothing unless `--restart`. `--force-recreate` stays only for the
  `--restart` case. Consent semantics themselves are the open redeploy decision; this stage
  preserves them.
- `deleteDocker` (L1119-1157): `compose -p <project> down` (project name is `name:` in the
  rendered file, `render.go:1066`) with the existing `stopAndRemove` fallback. Lifecycle verbs
  (L1206-1246) go to `compose -p <project> start|stop|restart` with the plain engine fallback.
  Compose v2 reconstructs a project from labels; the v1 and nerdctl paths fall back as today.
  Alternative if `-p` proves unreliable: re-render and pipe like A does for `down`; this needs
  the role, which `Delete`/`Start`/`Stop`/`Restart` do not take today.
- `composeOrDerived` and `ComposeCommand` unchanged; `--project-directory` is the env file's
  directory (B knows it through the rebase in `internal/config/hostpath.go:243`).
- `CheckEnv` (L298) drops the `composeFile=` row.

Config: retire `docker.composeFile` by name through `validateRetiredContainerKeys`
(`internal/config/validate.go:209`) with a message that says the file is no longer written;
remove the `convert.go:42,44,484` legacy mapping (or map it to the retirement error); remove it
from `internal/examples/assets/full.yaml:382-385`, then `regen` for `env/sample.yaml` and the
generator HTML `GEN` block; hand-edit the generator page's non-generated part
(`solace-yaml-generator.html:831,929,1252`) and the cross-check in docs/developer.md.

Tests: the ~30 tests listed in Appendix B (manager_test.go, secrets_test.go, preflight_test.go,
hostpath_test.go) move from "file on disk" to "stdin carries the rendered bytes and env carries
the digests"; `TestManagerLifecycleDockerDryRunUsesCompose` (L1415) asserts the new `-p` argv.
Add: label-driven change detection, `--project-directory` value, nothing under the env dir after
a deploy. Update docs/test.md rows (Appendix B lists them) and the Summary counts.

Docs: docs/operations.md L540-552, 898-901, 934-939, 957-972, 1000-1001, 1215, 1234, 1937;
docs/configuration.md L68 and the retired-keys table (L694, 718); README.md L223. Mirror A's
docs/DESIGN.md L218-220 wording in docs/developer.md for the rationale.

Verify before coding: `docker-compose` v1 and `nerdctl compose` acceptance of `-f -`. If one
refuses, that engine keeps a temp file written to a 0700 directory and removed after the call,
documented as the exception.

> Superseded 2026-11-04: the compose, listing, inspect and lifecycle commands are built in M
> `containercli` and `containercli/inspect` (step B6). Lifecycle verbs are plain engine verbs on the
> container in both tools, not `compose -p`, and inspect is `container inspect`. A compose that
> refuses `-f -` is a decision for S7, not a silent temp file. See docs/developer.md.

## Stage 3: B non-TTY confirmations fail fast

`internal/cli/confirm.go` `confirmGate` (L72-85): when `noPrompt` is false and stdin is not a
terminal, return an error (exit 1, not usage 2) worded like A's
`cmd/solmq-conn-util/main.go:1404`: "confirming <action> interactively: stdin is not a terminal,
so this prompt cannot be answered interactively; pass --no-prompt instead". Remove the piped
answer path (`stdinCanAnswer`, the memoised reader L124-145 stays for the terminal case). Apply
to the destructive confirmations and the downgrade confirmation (L272, which today has no
escape; give it `--no-prompt`). Leave the deploy restart consent (`confirmRestart` L248-257,
`staleWarning`) as is; it belongs to the open redeploy decision.

Tests to flip or replace (`internal/cli/cli_test.go`): `TestPromptAsksWhenStdinIsNotATTY` (3683)
becomes a fail-fast test, `TestUnansweredPromptStillRefuses` (3719) and `TestConfirmNonTTY`
(1749) assert exit 1 and the flag in the message, `TestStdinCanAnswerClosedFile` (3652) goes.
`TestMain` points stdin at a closed pipe for the package (docs/test.md L74-77), so every test
that reaches a gate must pass `--no-prompt` or set the terminal seam; add a `stdinIsTerminal`
seam like A's (`main.go:903`). Docs: docs/operations.md L1254-1277 and CLAUDE.md L170-180
describe the old rule; rewrite both. docs/parity.md row goes to `aligned`.

Known cost, recorded in parity.md: under mintty-style shells stdin is a pipe even at a terminal,
so those users need `--no-prompt`. A has the same limitation today.

## Stage 4: A podman session and host parity

In `internal/runner/runner.go` and `cmd/solmq-conn-util/runas.go`, adopting B's generic checks:
- `ResolveQuadletScope` (L756-793): honour `XDG_CONFIG_HOME` before `$HOME/.config`
  (B: `internal/config/load.go:379-388`). Test `TestResolveQuadletScope` (runner_test.go:365)
  gains the XDG case. docs/configuration.md L880-883 and docs/DESIGN.md L280-283.
- Session repair before any `systemctl --user`: derive `XDG_RUNTIME_DIR=/run/user/<euid>` only
  if unset and the directory exists, `DBUS_SESSION_BUS_ADDRESS=unix:path=<xdg>/bus` only if
  empty, never overwrite inherited values (B: `rootless.go:148-192`, tests
  `TestEnsureUserSession*` 242-333). Memoise per run.
- Linger check for user mode (B: `rootless.go:431-441`): a quadlet with `WantedBy=default.target`
  dies at logout without it. Warn in `podmanHostChecks` and in `validate`; the remedy names
  `loginctl enable-linger`.
- Session core limit (B: `limits.go:337-364`, reading `/proc/self/limits`): a hard core limit of
  0 makes every `podman exec` fail with `setrlimit RLIMIT_CORE`. Warn with B's limits.d remedy.
  Skip nr_open, nofile and memlock: broker-specific.
- `validate` runs `subIDWarnings` too (today only deploy does, main.go L705-709 vs L1040-1046),
  and both use the same uid (`Getuid` at L706 vs `Geteuid` in scope; pick effective).
- Rootless rule, part one: add optional `rootless` (bool) to the podman section in
  `internal/spec/targets.go:239-263`. Unset: derived from euid as today. Set and invoker is
  non-root: `false` is an error "podman.rootless=false requires root, but running as euid N",
  `true` is accepted. Set and invoker is root: the declared value decides; `true` as root is
  refused until Stage 7 ships, with B's message (`internal/container/euid.go:41-48`). Messages
  identical in both tools. Tests mirror B's `TestGuardPodmanEUID` table (manager_test.go:2028).
  B needs no code change for part one; its `podman.rootless` already behaves this way.

Docs: docs/configuration.md podman section (L867-934), docs/operations.md, docs/troubleshooting.md
L116-118, docs/DESIGN.md section 6; docs/test.md rows and snapshot line.

> Superseded 2026-11-04: the session repair, linger, core limit, quadlet directory and rootless
> rule are built in M `podman`, `systemd` and `rootless` (step B5) and adopted at S6. A drops its
> /etc/subuid count for `podman info` and refuses a miss, and a relative `XDG_CONFIG_HOME` is
> refused. See docs/developer.md. The stage above records the plan as it was.

## Stage 5: A kubernetes parity

- Namespace create-if-absent (B: `internal/k8s/prep.go:31-119`): `get namespace <ns> -o name
  --ignore-not-found` first; apply the Namespace document only when absent, in its own stream;
  the rest of `deploy.Render` output stays one `apply -f -`. Preflight probes `create namespaces`
  only when it will be created. Tests `TestRenderNamespaceAlwaysFirst` (deploy_test.go:590),
  `TestRemoveTeardownManifestOmitsTheNamespace` (main_test.go:1078),
  `TestDeployKubernetesSeamHappyPath` (main_test.go:724, exact call count) change; docs
  docs/configuration.md L738, docs/operations.md L121-123, docs/DESIGN.md L154-167.
- Wait after deploy, start and restart: `rollout status deployment/<n> --timeout=300s`
  (B: `internal/k8s/ops.go:140,165`). Call count tests in `TestLifecycleVerbsActInPlace`
  (lifecycle_test.go:16) and the seam test change. Docs docs/operations.md L146-149,
  docs/troubleshooting.md L147-148, docs/DESIGN.md L171-174.
- `auth can-i` parser: adopt B's superset (`internal/k8s/preflight.go:113-130`): last non-empty
  line, `no`, `no - ...`, `no (...)`, anything not `yes` refused, "unanswered" distinct from
  "refused". Replace `canIDenied` (runner.go:659-666); tests `TestPreflightRefusalNamesThePermission`
  (runner_test.go:941) gain the new forms.
- Probe set: before the first write, probe every kind deploy will create (`create` on
  deployments, services, configmaps, secrets, persistentvolumeclaims, and namespaces when
  absent), not only `create deployment` (runner.go:635-655). Remove probes `delete` on the same
  kinds and `delete namespaces` before the namespace gate offers deletion (today never probed,
  namespace.go). Keep A's logs/exec/status probes; B adding them is an open item.

> Superseded 2026-11-05: built in M `kube`, `kube/rbac` and `kube/nsgate` (step B7). Probes cover
> writes only in both tools, so A drops its logs, exec and status probes (decided 2026-11-04). Under
> the least-privilege rule (2026-11-05) a refused rollout wait is a warning and the namespace gate
> never fails a remove. See docs/developer.md.

## Stage 6: podman secrets, both

B `createPodmanSecrets` (manager.go:585-613) adopts A's `secret create --replace <n> -` when
`podman version` reports >= 4.7 (A: `PodmanVersion` L675, `PodmanSupportsReplace` L685-701),
keeping rm-then-create below that and the digest skip for the cert. A keeps `--replace` and
gains nothing here; its always-recreate is correct, the digest skip is an optimisation left open.
Tests: B mirrors `TestPodmanSecretCreateReplacesInOneCall` (A runner_test.go:571) and
`TestPodmanVersion` (587). Docs: B docs/operations.md podman section; docs/parity.md row.

> Superseded 2026-11-04: the version check and the secret builders are built in M `podman` (step
> B5), with their cases in M's tests, and B adopts `--replace` at S6. See docs/developer.md.

## Stage 7: design spike, root deploys for a declared rootless user (both)

Not code. A short design in A's docs/DESIGN.md and B's docs/developer.md answering: schema
(`podman.rootless: true` plus `podman.user: <name>`), how podman and systemctl run as that user
(`systemctl --user -M <user>@` needs systemd >= 248; podman secrets need `runuser -u <user> --
podman ...` or equivalent), linger for that user, `XDG_RUNTIME_DIR=/run/user/<uid>`, ownership
of the unit and data dirs, subuid checks against the declared user, and how B's permanent ban on
`runuser`/`su`/`setpriv` in configured commands (`internal/config/execguard.go:91-107`) coexists
with tool-originated use. Output is a decision, then a stage per tool.

## Open decisions (recorded in parity.md, not settled here)

- Redeploy over a running service: keep B's consent prompt and add it to A, or let the platform
  recreate silently in both.

> Superseded 2026-11-04: decided. Both tools ask before restarting or recreating a running service
> whose artifact or secrets changed, and `--no-prompt` skips the prompt (applied at S7).

- subuid coverage: A warns, B refuses. Validate-is-golden argues for an error in both.

> Superseded 2026-11-04: decided. Both tools check the run-as ids against `podman info`'s own
> mapping and refuse a miss, saying what to correct and how (M `rootless`, step B5; applied at S6).

- Namespace gate: A discounts owned objects, B probes delete permission. Union or pick one.

> Superseded 2026-11-04: decided, the union (M `kube/nsgate`, step B7; applied at S9).

- Logs/exec/status probes in B.

> Superseded 2026-11-04: decided, writes only in both tools; A drops these probes at S9.

- Output model: A buffers and prints `ok:`/`error:` at the end, B streams. Tied to the runner
  seam; decide before any shared runtime package.
- Shared runtime package (argv builders and parsers from 7.1) after stages 1-6 converge; needs
  the unified `Exec(ctx, Cmd) Result` seam described in 7.1 and the shared module repo from
  Appendix A section 8.

## Verification of the deferred stages

Per stage, in the repo touched, user-run:
- `./scripts/dev.sh all scan`; B also `./scripts/dev.sh regen` when examples, goldens or
  generated docs move (Stage 2); A `./scripts/dev.sh regen` only if a verb or flag changes
  (none planned).
- cov floor: compare the new total with the last cov log; explain any drop.
- Stage 2 smoke on a docker host: `solace-util broker deploy --platform docker` twice (second
  run reports unchanged), change a value and deploy (consent path), `broker remove`; confirm no
  compose file appears beside env.yaml; repeat once with `docker-compose` v1 if available.
- Stage 3 smoke: `solace-util broker remove ... < /dev/null` exits 1 and names `--no-prompt`.
- Stage 4 smoke on a rootless podman host: `XDG_CONFIG_HOME` set to a temp dir places the unit
  there; run under `sudo -u <user> -i` with no session variables and confirm `systemctl --user`
  works; `validate` reports linger and core limit state.
- Stage 5 smoke on a cluster: deploy into a fresh namespace (created once), redeploy (no
  namespace apply), `start`/`restart` return only after rollout, `remove` with an occupied
  namespace keeps it.

---

# Appendix A: research notes (licensing and shared module)

## A.1 Facts that shape the answer

- Both repos are GPLv3. Authors: the user (as H and Hamlyn) plus dependabot. No external
  contributor, so relicensing future versions needs nobody's consent.
- Every dependency is permissive (cobra and mousetrap Apache-2.0, pflag BSD-3, yaml.v3
  MIT+Apache, keystore-go MIT, go-pkcs12 BSD-3, x/sys BSD-3). Nothing blocks a closed or
  source-available relicense.
- Neither binary has any licensing, expiry or trial logic today. The only "licence" strings
  are Solace broker product keys that solace-util passes through (`broker.productKeys`).
- The ibmmq helper's `download` verb fetches public Maven jars (IBM MQ client, syslog
  encoder). There is no Solace-entitled download to piggyback a check on.
- Both tools run inside customer infrastructure, often air-gapped or egress-restricted.
- Module paths: ibmmq helper is `github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn`
  (does not match its remote `github.com/hafio/solace-ibmmq-connector-helper`); solace-util is
  the bare `solace` (not importable).

## A.2 Enforcement options, weakest to strongest

| # | Option | What it stops | What it does not stop | Fit |
|---|--------|---------------|-----------------------|-----|
| 1 | Signed expiring key checked offline, source stays GPLv3 | Nothing technically. GPLv3 s.6/s.7 let any recipient delete the check and redistribute | Everything | Advisory: tracks who is licensed, gives a clean expiry UX, supports a PS-engagement term |
| 2 | Same key, but binaries only, new versions relicensed | Casual use past expiry, sharing between customers (licensee in the key) | A motivated person patching the Go binary (flip one branch) | The realistic ceiling for a CLI helper. Old GPL tags stay GPL forever |
| 3 | Capability key (key decrypts embedded templates or assets) | Slightly raises patch effort | One honest decrypt leaks the plaintext for everyone | Not worth the complexity |
| 4 | Online activation or lease (phone home, cached offline lease) | Key sharing, expired keys, revocation | Air-gapped hosts, egress policies; still binary-patchable; needs a server PS must run | Poor fit for where these tools run; viable only as optional activation producing an offline lease |
| 5 | Build-date time bomb, no keys | Use of stale releases | Anyone who rebuilds from source; no idea who is licensed | Simplest, but forces every customer onto an upgrade cadence |

Verdict on "enforced properly": for a deployment helper there is no technical enforcement
that survives a determined user, with or without keys. What is achievable is option 2:
contract plus signed key plus binary distribution. That makes expiry effective against the
normal case (a customer who keeps using it after the engagement) and makes the licensee
visible. If the business goal is community adoption with paid support, option 1 or no key
at all is more honest; the support contract is then the product.

Relicensing mechanics if option 2 is chosen: freeze the public repo at the last GPL tag (or
make it private), state the new license in LICENSE and every doc's support notice, keep
dependabot and the tag pipeline, and add GOPRIVATE plus a read token to CI if the shared
module below is private too.

## A.3 Key format and verification (options 1 and 2)

- Signature: Ed25519 from the Go stdlib, no new dependency. Public key(s) compiled into the
  binary with a key id so a signing key can be rotated without breaking issued keys.
- Token shape: `SOLACE-LIC-1.<base64url payload>.<base64url signature>`. Payload fields:
  `kid`, `id`, `product` (`solmq-conn-util`, `solace-util` or a list), `licensee`, `issued`,
  `notBefore`, `expires`, optional `features` and `maxVersion` (perpetual-fallback style:
  "any release cut before expires keeps working").
- Checks at run time: signature, product match, time window with a few minutes of skew, and
  warn on stderr inside the last 30 days (mirrors the certificate expiry convention in the
  ibmmq helper's storebuild). Expired or missing key: refuse with an actionable message that
  names the licensee, the expiry and where to get a new key.
- Clock rollback: optional high-water-mark file (last seen wall clock) under the tool's
  state dir. Cheap, catches naive rollback, not a real defence. Document it as such or skip.
- Revocation: offline tools cannot revoke. A short embedded deny-list refreshed per release
  is the only option and is weak. Short expiries are the practical substitute.
- Issuer: a `license-issuer` command in the shared module. Private key from a file or env
  var, never in a repo. Subcommands: `issue`, `inspect`, `verify`. An issued-keys ledger
  (CSV in a private place) gives the audit trail PS needs.
- Tests mint keys at run time with a throwaway Ed25519 pair, the way `certfixture` mints
  certificates in the ibmmq helper. No key material is committed.

## A.4 Where the key comes from

Precedence, first hit wins:

1. `SOLACE_LICENSE` environment variable (one name for both tools; the `product` field in
   the payload tells them apart). Works in CI and containers.
2. `/etc/solace/license` on Linux and macOS, `%ProgramData%\Solace\license` on Windows.
   Root-installed, host-wide.
3. `$XDG_CONFIG_HOME/solace/license` (solace-util already has `xdgConfigHome()` for the
   quadlet dir; the ibmmq helper derives the same dir from `os.UserHomeDir`).
4. `license:` (token) or `licenseFile:` (path) in env.yaml. Precedent: solace-util already
   accepts licence strings in `broker.productKeys`, and its `*Env` secret-reference
   convention could give `licenseEnv:` for free.

Not a secret in the password sense, but a bearer token bound to a licensee, so: never echo
the token in logs or status output, print only `licensee` and `expires`. A `license`
subcommand (`show`, `install`) covers the UX without a user hunting for paths.

`validate` must report license state like any other error generate/deploy would hit; it is
the golden check in both tools, so a missing or expired key surfaces there first.

## A.5 Where the gate plugs in

ibmmq helper (hand-rolled dispatch):
- `cmd/solmq-conn-util/main.go` `dispatch()` lines 177-202: after the `verbHandlers`
  lookup and the `<verb> -h` short-circuit, before `h(args[1:], r)`.
- Which verbs need a key becomes a field on `cliVerb` in `commands.go`, so help, completion
  and `docs/commands.md` stay single-source and `TestDispatchHandlersMatchModel` keeps
  gating drift. `version`, `help`, `auto-complete`, `examples` stay free whatever the final
  mix; `validate` reports instead of refusing.
- Exit code: 1 (runtime/config). A new code 3 would be clearer but changes the documented
  contract in docs/DESIGN.md s.2 and `TestExitCodeContract`.
- Seams: `var licenseNow = time.Now` and an injected lookup, same pattern as `promptLine`,
  `downloadFn`. Every existing `dispatch(...)` test then needs either a minted key or the
  bypass seam, which is the main cost of the change.

solace-util (cobra):
- Not on the root `PersistentPreRunE`: `TestCompletionNeverReadsTheEnvFile` forbids a hook
  that fires on every TAB. The gate goes into `prepare()` in `internal/cli/platform.go`
  (lines 171-200) after `app.load(cmd)`, so env.yaml is available as a key source.
- Commands that skip `prepare` (`convert`, `examples`, `version`, `auto-complete`) stay free
  unless given their own check. Hand-kept exemption lists in `TestEveryRunnableCommandIsWired`
  and `TestAllowCommandIsRegisteredWhereItExecutes` are the places that notice a new command.
- `prepare` wraps everything with `asUsage` (exit 2); a license failure should return
  unwrapped so it exits 1.
- A stdlib-only verifier sits at the leaf layer beside `abbrev` and `output`, or comes from
  the shared module; `cli` is the only caller.

Both: a new hand-written doc needs the support notice and an entry in the notice test's doc
list; `docs/test.md` grows with the tests; `regen` updates commands.md and abbreviation.md.

## A.6 Shared module

- Shape: one Go module, stdlib-only `license` package (token, verify, sources) plus
  `cmd/license-issuer`. Nothing in it may import either tool's types.
- Both consumers pin the exact version; dependabot manages actions only, by init-devops
  design, so module bumps are deliberate commits.
- If the module is private: GOPRIVATE and a read token in both CI workflows; a CI
  workflow still calls dev-script task names only.

## A.7 Code the two repos could share (synergy)

Inventory of parallel code. Ranked by shared idea versus coupling to each repo's own types. The
first five are stdlib-only today or trivially made so, and would sit in the same shared module
as `license`.

| Rank | Area | A today | B today | Note |
|------|------|---------|---------|------|
| 1 | Output sink (`==> ` steps, `[ OK ]`/`[WARN]`/`[FAIL]` tags, tables) | none; ad hoc `warning:`/`error:` prints | `internal/output` (296 lines, stdlib only) | Lift B's package as-is; A adopts it |
| 2 | Abbreviation set with collision and charset rules | inside `cmd/solmq-conn-util/commands.go` model | `internal/abbrev` (193 lines, stdlib only) | Lift B's; A's model keeps its `Aliases` fields but validates through it |
| 3 | Support notice text plus the doc-gate test helper | `support.go` constant, `support_test.go` | test-only constant in `commanddoc_test.go`, copy in `importdoc_test.go` | Same text bar tool and product name; one `AssertNoticeAtTop(t, doc, tool, product)` replaces three copies |
| 4 | Version line and exit-code classification | `main.go:110`, `TestExitCodeContract` | `internal/cli/version.go`, `exit.go` (`usageError`, 0/1/2) | Tiny. Forces settling one divergence: missing env file exits 1 in A, 2 in B |
| 5 | Command-token guard (allowlist, flag shape, metachar rejection) | `internal/validate` `CheckDeployCommand`, `Tokenize`, `SafeToken` (~110 lines) | `internal/config/execguard.go` (659 lines, layered) | Same idea. Policy split: B has a `neverAllowed` escalator list (sudo, doas, pkexec); A allows `sudo podman` via `--allow-command` |
| 6 | Compose and quadlet escaping plus the hardening lines | `dockergen.go`, `podmangen.go` | `render.go` `Compose`, `Quadlet` | `composeEscape` is byte-identical; `%` escaping equivalent; both emit user, `cap_drop: ALL`, `no-new-privileges`. A alone adds tmpfs; neither emits read-only root |
| 7 | Quadlet directory and XDG resolution | `runner.go:756-794`, hardcodes `$HOME/.config` | `load.go:379-388`, honours `XDG_CONFIG_HOME` | ~20 lines; sharing B's fixes A's missing XDG support (Stage 4) |
| 8 | Env-name secret reference check (`*Env` sibling keys) | `pass-env`/`user-env` in `internal/spec` | `secretref.go` (name regex, both-set, unset vs empty) | Small helper, no config type needed |

> Superseded 2026-11-04: rank 6 is built in M `containerart/compose`,
> `containerart/quadlet` and `containerart/artifacttest` (step B4), with A
> adopting B's always-quote policy. See docs/developer.md. The table above
> records the tools as they were.

Not worth sharing as Go: shell completion (hand-written scripts vs cobra generators), the
commands.md and abbreviation.md renderers (model-driven vs cobra-tree walk), and certificates
(B has no x509 code at all). The CI workflows differ by 2 to 10 lines and could be one reusable
workflow, a template decision in init-devops rather than a per-repo edit.

### A.7.1 Runtime layer: docker, podman, kubectl

Behaviour agrees far more than code shape suggests. B's `internal/k8s/namespace.go:17` calls its
namespace gate "a port of the same gate in the sibling project solmq-conn-util".

Same in both, no config-type dependency (the contract for docs/parity.md, Stage 0):

| Behaviour | A | B |
|-----------|---|---|
| `kubectl apply -f -` / `delete -f - --ignore-not-found`, manifest on stdin | runner.go:412-426 | k8s/cluster.go:117,127 |
| `auth can-i` builder and parser | runner.go:585-660 | k8s/preflight.go:113-252 (superset) |
| `<engine> info` preflight with per-platform hint | runner.go:593 | container/preflight.go:41 |
| `systemctl [--user]` daemon-reload/start/stop/restart, `show -p NRestarts --value` | runner.go:796-866,1074 | container/manager.go:774-1418, inspect.go:143 |
| podman `secret create <n> -` stdin, `secret rm --ignore` | runner.go:712-746 | container/manager.go:603-608 |
| exec and logs argv shapes, upload via `exec -i sh -c 'cat > p'` | runner.go:1150-1211,1307 | k8s/transport.go:39-76, ops.go:42-49 |
| inspect decoding: strip leading `/`, `State.Health` or `State.Healthcheck`, podman restarts from systemd | statusreport/parse.go:552-674 | container/inspect.go:46-152 |
| `ps --all --format` filtered in Go | runner.go:1054 | container/manager.go:992,1287 |
| namespace occupancy gate: same kinds, ignores `kube-root-ca.crt` and `default` SA, same 4 protected namespaces, unreadable means occupied | cmd/.../namespace.go:37-204 | k8s/namespace.go:37-102 |
| compose env-provider secrets, values only in child env | dockergen.go:127-278 | render.go:1158-1197 |
| hardening snippets: compose cap_drop ALL, no-new-privileges, privileged false; quadlet `DropCapability=all`, `NoNewPrivileges=true` | dockergen.go:180, podmangen.go:178 | render.go:908-1132 |
| exit contract 0/1/2, `cli`/`shell` pass the child's code through | shell.go:239-272 | cli/exit.go |
| watch loop: redraw, Ctrl-C exits 0 | status.go:890-948 | cli/watch.go |

Different by policy (the divergence table for parity.md; the Stage column is this plan's):

| Topic | A | B | Stage |
|-------|---|---|-------|
| Compose file | stdin, never written | on disk, diffed | 2: B -> stdin |
| Non-TTY confirm without `--no-prompt` | error, exit 1 | reads stdin; EOF warns, exit 0 | 3: B -> error |
| Rootless podman | derived from EUID | declared, mismatch refused; repairs XDG_RUNTIME_DIR and DBus | 4 and 7 |
| Quadlet dir | `$HOME/.config`, ignores XDG | `$XDG_CONFIG_HOME`, configurable | 4: A honours XDG |
| Host limits | none | nr_open, user-manager Limit*, session nofile/core, linger | 4: A takes the generic ones |
| Namespace on deploy | re-applied every time | created only if absent | 5: A |
| Waiting | never | `rollout status --timeout=300s`, CR read back | 5: A |
| RBAC probes | one per action incl. logs/exec/status | every kind before first write; none for logs/exec/status | 5: A widens; B open |
| can-i parser | `no`, `no - ` | superset incl. `no (...)`, unanswered vs refused | 5: A adopts B's |
| Podman secret replace | `--replace` if >= 4.7 | rm then create; skip if digest unchanged | 6: B adopts `--replace` |
| Redeploy over a running service | restart or recreate, no prompt | diff and ask; non-TTY declines | open |
| subuid coverage | warn | refuse | open |
| Namespace gate | ownership-aware, no delete probe | ignores ownership, probes delete | open |
| Output model | buffered, `ok:`/`error:` line | streamed live | open; runner seam |
| Workload | Deployment | operator CR + StatefulSets | inherent |
| Exec | `-c connector`, `sh` | no `-c`, `bash` or `cli -A` | inherent |
| Registry | pulls image to check entrypoint | logs in, never pulls | inherent |

Runner seam comparison (the obstacle to a shared driver):

| Capability | A `runner.Runner` + Streamer/Splitter/Attacher | B `engine.Runner` |
|------------|------|------|
| combined stdout+stderr capture | `Run(Cmd) (string, error)` | none |
| split capture | `RunSplit` (optional) | `Output`: stdout captured, stderr live |
| stream to writers | `Stream(ctx, Cmd, w1, w2)` | `Run` to os.Stdout/err only |
| context | Stream only | every call |
| env | on any call | `EnvRunner` only, by design |
| stdin + env together | yes (`Cmd.Stdin`, `Cmd.Env`) | no (Stage 1 adds it) |
| interactive TTY | `Attach(Cmd, files) (code, err)` | `RunInteractive(ctx, ...) error` |
| dry-run echo, env masking, announce | none | `Echo`, `MaskEnv`, `Announce` |

Neither is a superset. A unified seam needs `Exec(ctx, Cmd) Result` with capture/stream/merge
modes, `Attach`, an interrupt-then-kill cancel policy, Echo as a decorator with an `IsDryRun`
capability, env masking in the echo path, the ErrDot refusal, and an argv[0] allowlist hook.
The shared fake must match by call position (A's fakes) and by argv content (B's).

Side finding: hardening is complementary. Only A emits seccomp `RuntimeDefault`,
`runAsNonRoot` and `automountServiceAccountToken: false` on k8s; only B emits ulimits, shm
and cpuset and checks host limits.

## A.8 Shared module: layout and GitHub mechanics

> Superseded 2026-11-02: the module is `github.com/hafio/go-deploy-utility`, licensed GPLv3
> (not Apache-2.0 or MIT), released through the init-devops library shape (not binaries-only),
> and built before convergence rather than after. Layout and order: `docs/developer.md`.

Both repos live under the personal account `hafio`. Dependabot manages GitHub Actions only.
The tag pipeline is a verbatim init-devops template and refuses a repo that ships nothing.

One repo, not a nested module: `github.com/hafio/solace-shared` (name open). A nested module
needs `lib/x/v0.1.0` tags, which the template's `tags: ["v*"]` trigger ignores.

```
solace-shared/
  go.mod                 module github.com/hafio/solace-shared; go 1.27; toolchain pinned
  LICENSE                Apache-2.0 or MIT, not GPL (a relicensed consumer could not import GPL)
  README.md  docs/       support notice as in both tools
  license/  output/  abbrev/  cliconv/  cliconv/clitest/     stdlib only
  cmd/license-issuer/    keygen, issue, inspect, verify
  scripts/ .github/      init-devops verbatim; shape binaries-only (the issuer ships)
```

Versioning: `v0.x.y` until both tools are on it, then `v1.0.0`; breaking after v1 means `/v2`.
Consumers pin exact versions; an untracked `go.work` covers local work across checkouts.
Public repo recommended: proxy and sumdb serve it, no CI changes; nothing in it is secret.
Private needs GOPRIVATE plus a fine-grained PAT applied via a git `insteadOf` rewrite on both
runner legs, as a growth of the init-devops template. Keyring by `kid` in the module; rotation
is append, tag, bump both tools. Private key stays on a PS workstation. Optional reusable
`ci.yml` in the shared repo is a template decision, separate from the module.

Checklist, user-run: create repo and LICENSE; init-devops with BUILD_TARGETS for the issuer;
tag v0.1.0; `go get` in B and repoint `output`/`abbrev`; `go get` in A and adopt them; only
then the license gate in each tool.

---

# Appendix B: B tests and doc lines pinned to the on-disk compose file (Stage 2 input)

internal/container/manager_test.go: TestManagerDeployDockerComposeWritesFile 319,
TestManagerDockerComposeCommandOverride 340, TestManagerRedeployUnchangedIsNoOp 521,
TestManagerRedeployChangedNeedsConsent 575, TestManagerDeployDockerPassesSecretsAsEnv 686,
TestManagerDeployRejectsEmptySecret 812, TestManagerDeployDockerDryRunMasksSecretEnv 829,
TestManagerDeleteDockerComposeDownWhenFileExists 1089, TestManagerDeleteDockerPurgeClearsDataDir
1105, TestManagerDeleteDockerComposeNoFileFallsBackToStopRm 1132,
TestManagerStopAndRemove{ContainerAbsentNoOp 1151, StopFailsContainerRunningBlocks 1172,
StopFailsProbeUnansweredBlocks 1197}, TestStatusAllDetailInspectsEachAndKeepsSecretsOut 1326,
TestManagerLifecycleDockerComposeFile 1356, TestManagerLifecycleDockerNoComposeFile 1387,
TestManagerLifecycleDockerDryRunUsesCompose 1415, TestManagerDeployDockerComposeWriteError 1606,
TestManagerDeployDockerComposeUpError 1620, TestManagerRedeployUnchangedRestartsForRotation 1669,
TestManagerDeployDockerProbeUnansweredErrors 1777, TestManagerRedeployStoppedContainerRecreates
1806, TestManagerRedeployUnchangedHintsRotation 1827, TestManagerDeleteDockerComposeDownError
1883, TestManagerDeleteDockerStopTolerated 1905, TestManagerDeletePurgeError 1928,
TestManagerStatusDockerNoComposeFile 1947, TestManagerStatusDockerRunsNoComposePs 1984.

internal/container/secrets_test.go: TestDeployDockerPassesTheBundleAsEnvNotArgv 187,
TestTeardownVerbsDoNotReadTheCertificate 237, TestDeployDockerFailsBeforeWritingTheComposeFile
276, TestDockerCertLabelDrivesTheRecreate 470.
internal/container/preflight_test.go: TestPreflightFailureStopsTheDeploy 66 (asserts no file, L88).
internal/config/hostpath_test.go: TestValidateHostPathsChecksEveryPlatformsFields 275,
TestValidateHostPathsMatchesTheRebaseList 317, TestRebaseResolvesAgainstTheEnvFileDirectory 350,
TestExpandHomePathsCoversEveryFieldItClaims 945.
Goldens internal/render/testdata/docker_compose_*.golden: body only, unaffected except the new
label line.
docs/test.md rows: 2524, 2543, 2547, 2552, 2556, 2566, 2568, 2577-2580, 2599, 2606, 2608, 2783,
2788, 2918; Summary counts L68-94.
Cert digest label pattern to copy: render.go:688-693, 1180-1185; manager.go:555-561, 894-918.
